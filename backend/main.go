package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

func main() {
	// Definir los flags que se pasarán al ejecutar el programa.
	var numUsers int
	var isolationLevel string
	flag.IntVar(&numUsers, "users", 5, "Número de usuarios concurrentes")                                                                  // Número de usuarios que simula la concurrencia.
	flag.StringVar(&isolationLevel, "isolation", "read committed", "Nivel de aislamiento (read committed, repeatable read, serializable)") // Nivel de aislamiento de la base de datos.
	flag.Parse()

	// Parsear el nivel de aislamiento desde la entrada del usuario.
	isoLevel := parseIsolationLevel(isolationLevel)
	if isoLevel == nil {
		log.Fatal("Nivel de aislamiento inválido") // Si el nivel de aislamiento no es válido, finalizar el programa.
	}

	// Conectar a la base de datos.
	db, err := sql.Open("postgres", getDBURL())
	if err != nil {
		log.Fatal("Error conectando a la base de datos:", err) // Si ocurre un error en la conexión, terminar el programa.
	}
	defer db.Close() // Asegurarse de cerrar la conexión al final.

	// ID del asiento que se va a reservar (solo hay un asiento en este caso).
	seatID := 1
	// Resetear el estado del asiento antes de comenzar las reservas.
	if err := resetSeat(db, seatID); err != nil {
		log.Fatal("Error reseteando asiento:", err) // Si hay un error al resetear el asiento, finalizar el programa.
	}

	// Variables para contar los resultados.
	var success, conflicts, errors int64
	var wg sync.WaitGroup // Usar un grupo de espera para sincronizar las goroutines.

	wg.Add(numUsers) // Añadir la cantidad de usuarios al grupo de espera.

	// Lanzar goroutines para simular la concurrencia de usuarios reservando el asiento.
	for i := 0; i < numUsers; i++ {
		userID := (i % 8) + 1 // IDs de usuario cíclicos del 1 al 8.
		go func(uid int) {
			defer wg.Done()                                                        // Decrementar el contador del grupo de espera cuando termine la goroutine.
			reserveSeat(db, uid, seatID, *isoLevel, &success, &conflicts, &errors) // Llamar a la función de reserva de asiento.
		}(userID)
	}

	wg.Wait() // Esperar que todas las goroutines terminen.

	// Imprimir los resultados finales de la simulación.
	printResults(db, seatID, isolationLevel, numUsers, success, conflicts, errors)
}

// parseIsolationLevel convierte el nivel de aislamiento de string a un valor de sql.IsolationLevel.
func parseIsolationLevel(level string) *sql.IsolationLevel {
	switch strings.ToLower(level) {
	case "read committed":
		l := sql.LevelReadCommitted
		return &l
	case "repeatable read":
		l := sql.LevelRepeatableRead
		return &l
	case "serializable":
		l := sql.LevelSerializable
		return &l
	default:
		return nil // Si el nivel no es válido, devolver nil.
	}
}

// getDBURL construye la URL de conexión a la base de datos usando variables de entorno.
func getDBURL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)
}

// resetSeat resetea el estado del asiento en la base de datos a "disponible" y borra las reservas existentes.
func resetSeat(db *sql.DB, seatID int) error {
	tx, err := db.Begin() // Iniciar una transacción.
	if err != nil {
		return err
	}
	defer tx.Rollback() // Asegurarse de hacer rollback si hay un error.

	// Cambiar el estado del asiento a 'disponible'.
	if _, err := tx.Exec("UPDATE Asiento SET estado = 'disponible' WHERE id_asiento = $1", seatID); err != nil {
		return err
	}

	// Eliminar todas las reservas previas para ese asiento.
	if _, err := tx.Exec("DELETE FROM Reserva WHERE id_asiento = $1", seatID); err != nil {
		return err
	}

	// Confirmar los cambios en la base de datos.
	return tx.Commit()
}

// reserveSeat simula el proceso de un usuario reservando un asiento.
func reserveSeat(db *sql.DB, userID, seatID int, isoLevel sql.IsolationLevel, success, conflicts, errors *int64) {
	ctx := context.Background()                                     // Crear un contexto de ejecución para la transacción.
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: isoLevel}) // Iniciar una transacción con el nivel de aislamiento especificado.
	if err != nil {
		atomic.AddInt64(errors, 1) // Si hay un error, contarlo como un error.
		return
	}

	var estado string
	// Obtener el estado del asiento antes de intentar reservarlo.
	if err := tx.QueryRowContext(ctx, "SELECT estado FROM Asiento WHERE id_asiento = $1", seatID).Scan(&estado); err != nil {
		tx.Rollback() // Si hay un error al obtener el estado, hacer rollback.
		atomic.AddInt64(errors, 1)
		return
	}

	// Si el asiento no está disponible, hacer rollback.
	if estado != "disponible" {
		tx.Rollback()
		atomic.AddInt64(conflicts, 1) // Contar como un conflicto si el asiento no está disponible.
		return
	}

	// Intentar reservar el asiento (cambiar el estado a 'reservado').
	res, err := tx.ExecContext(ctx, "UPDATE Asiento SET estado = 'reservado' WHERE id_asiento = $1 AND estado = 'disponible'", seatID)
	if err != nil {
		handleTxError(err, conflicts, errors) // Manejar errores de la transacción.
		tx.Rollback()
		return
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		atomic.AddInt64(conflicts, 1) // Si no se afectaron filas, hay un conflicto (otro usuario reservó el asiento).
		tx.Rollback()
		return
	}

	// Insertar la reserva en la tabla de reservas.
	if _, err := tx.ExecContext(ctx, "INSERT INTO Reserva (id_usuario, id_asiento, estado) VALUES ($1, $2, 'confirmada')", userID, seatID); err != nil {
		handleTxError(err, conflicts, errors) // Manejar errores al insertar la reserva.
		tx.Rollback()
		return
	}

	// Confirmar la transacción si todo salió bien.
	if err := tx.Commit(); err != nil {
		handleTxError(err, conflicts, errors) // Manejar cualquier error al hacer commit.
		return
	}

	// Contar como éxito si la reserva se completó correctamente.
	atomic.AddInt64(success, 1)
}

// handleTxError maneja errores de la transacción, diferenciando entre conflictos y errores generales.
func handleTxError(err error, conflicts, errors *int64) {
	if isSerializationError(err) {
		atomic.AddInt64(conflicts, 1) // Si es un error de serialización (conflicto), contar como conflicto.
	} else {
		atomic.AddInt64(errors, 1) // Si es otro tipo de error, contar como error.
	}
}

// isSerializationError verifica si el error es un error de serialización (conflicto de concurrencia).
func isSerializationError(err error) bool {
	pqErr, ok := err.(*pq.Error)
	return ok && pqErr.Code == "40001" // Código de error de serialización.
}

// printResults imprime los resultados finales de la simulación de reservas.
func printResults(db *sql.DB, seatID int, isolation string, users int, s, c, e int64) {
	var estado string
	// Obtener el estado final del asiento.
	db.QueryRow("SELECT estado FROM Asiento WHERE id_asiento = $1", seatID).Scan(&estado)

	var count int
	// Contar cuántas reservas confirmadas hay para el asiento.
	db.QueryRow("SELECT COUNT(*) FROM Reserva WHERE id_asiento = $1 AND estado = 'confirmada'", seatID).Scan(&count)

	// Imprimir los resultados de la simulación.
	fmt.Println("=== Resultados ===")
	fmt.Printf("Nivel de Aislamiento: %s\n", isolation)
	fmt.Printf("Usuarios Simulados: %d\n", users)
	fmt.Printf("Reservas Exitosas: %d\n", s)
	fmt.Printf("Conflictos: %d\n", c)
	fmt.Printf("Errores: %d\n", e)
	fmt.Printf("Estado Final del Asiento: %s\n", estado)
	fmt.Printf("Reservas Confirmadas: %d\n", count)
}
