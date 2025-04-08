# Imagen base
FROM golang:1.24

# Establecer directorio de trabajo
WORKDIR /app

# Copiar archivos del backend
COPY . .

# Descargar dependencias
RUN go mod tidy

# Comando por defecto (también puede ir en el docker-compose)
CMD ["go", "run", "main.go"]
