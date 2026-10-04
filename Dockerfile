# ---------- Etapa 1: compilar el programa en Go ----------
FROM golang:alpine AS build
WORKDIR /src

# Primero las dependencias (así Docker las guarda en caché si no cambian)
COPY go.* ./
RUN go mod download

# Luego todo el código y se compila a un solo ejecutable llamado "server"
COPY . .
RUN CGO_ENABLED=0 go build -o /src/server ./backend

# ---------- Etapa 2: imagen final, liviana ----------
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app

# Copia el proyecto completo (HTML, CSS, JS, fuentes, imágenes) + el ejecutable.
# Se hace así para que las rutas relativas de tu código sigan funcionando.
COPY --from=build /src /app

# El puerto lo define la plataforma con la variable PORT (tu código debe leerla)
CMD ["./server"]

