-- ENUM para EstadoAsiento lo explico Bacilio para el proyecto pasado, que es mejor debido a que con los CHECK dentro del dato pueden existir problemas
CREATE TYPE estado_asiento_enum AS ENUM ('disponible', 'reservado');

-- ENUM para EstadoReserva
CREATE TYPE estado_reserva_enum AS ENUM ('pendiente', 'confirmada', 'cancelada');

-- Evento
CREATE TABLE Evento (
    id_evento SERIAL PRIMARY KEY,
    nombre VARCHAR(100) NOT NULL UNIQUE,
    fecha DATE NOT NULL,
    capacidad INT NOT NULL CHECK (capacidad > 0)
);

-- Asiento
CREATE TABLE Asiento (
    id_asiento SERIAL PRIMARY KEY,
    id_evento INT NOT NULL REFERENCES Evento(id_evento) ON DELETE CASCADE,
    numero_asiento INT NOT NULL CHECK (numero_asiento > 0),
    estado estado_asiento_enum NOT NULL DEFAULT 'disponible'
);

--  Usuario
CREATE TABLE Usuario (
    id_usuario SERIAL PRIMARY KEY,
    nombre VARCHAR(100) NOT NULL,
    email VARCHAR(100) NOT NULL UNIQUE CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$') 
);

-- Reserva
CREATE TABLE Reserva (
    id_reserva SERIAL PRIMARY KEY,
    id_usuario INT NOT NULL REFERENCES Usuario(id_usuario) ON DELETE CASCADE,
    id_asiento INT NOT NULL REFERENCES Asiento(id_asiento) ON DELETE CASCADE,
    fecha_reserva TIMESTAMP NOT NULL DEFAULT NOW() CHECK (fecha_reserva >= NOW()),
    estado estado_reserva_enum NOT NULL DEFAULT 'pendiente'
);
