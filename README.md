# BD1-Proyecto2 - Simulación de Reservas Concurrentes

Este proyecto simula un sistema de reservas concurrentes a una base de datos PostgreSQL, utilizando distintos niveles de aislamiento para observar el comportamiento de la concurrencia y la consistencia.

## 🧑‍💻 Autores y contexto  

Dulce Rebeca Ambrosio Jimenez 231143   
Maria Jose Giron Isidro 23559  
Paula De Leon 23202  
Leonardo Dufrey Mejia Mejia 23648  

Durante el desarrollo del proyecto, trabajamos de manera presencial en clase. Debido a que nuestra compañera **Dulce** tenía experiencia previa con Docker, decidimos enviarle nuestras partes del código por separado para que ella las integrara y preparara el entorno completo en Docker. Por ello, **el primer commit del repositorio refleja ya la integración total del sistema**.

Nuestro docente **Bacilio** fue testigo del progreso constante y del trabajo en equipo que realizamos durante las sesiones.

---

## 📦 Estructura del Proyecto

- `backend/main.go`: Código en Go que ejecuta la simulación de múltiples usuarios reservando un asiento.
- `docker-compose.yml`: Archivo que levanta la base de datos PostgreSQL y el servicio de la app.
- `db/ddl.sql`: Script para crear la estructura de la base de datos.
- `db/data.sql`: Script para insertar los datos iniciales.

---

## ⚙️ Requisitos

- [Docker](https://www.docker.com/)
- [Docker Compose](https://docs.docker.com/compose/)

---

## 🚀 Cómo ejecutar la simulación

Sigue estos pasos para levantar los servicios y ejecutar la simulación automáticamente:

1. **Clona el repositorio:**
   ```bash
   git clone https://github.com/tu_usuario/BD1-Proyecto2.git
   cd BD1-Proyecto2

2. **Ejecutar Docker Compose:**
   docker-compose up --build

   Este comando:

  Levanta un contenedor PostgreSQL con los scripts de creación e inserción de datos.
  Compila y ejecuta la aplicación en Go con 5000 usuarios y el nivel de aislamiento "repeatable    read". 

  **SALIDA**
  === Resultados ===  
Nivel de Aislamiento: repeatable read  
Usuarios Simulados: 5000  
Reservas Exitosas: 1  
Conflictos: 4999  
Errores: 0  
Estado Final del Asiento: reservado  
Reservas Confirmadas: 1  

Si quieren modificar los parametros para la simulacion se debe cambiar el compose de docker, el nombre exacto dedl archivo es docker-compose.yml 

En esta linea:
command: ["go", "run", "main.go", "--users=100", "--isolation=serializable"]

Los niveles de aislamiento válidos son:  
read committed  
repeatable read  
serializable  

Solo tenemos un asiento, ya que lo manejamos segùn el id y queriamos ver su concurrencia
  

