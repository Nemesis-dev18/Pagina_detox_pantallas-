package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()

	// La URL de conexión: usuario, contraseña, host, puerto, base de datos
	dbURL := os.Getenv("DATABASE_URL")

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatal("no se pudo crear el pool: ", err)
	}
	defer pool.Close()

	// Probamos que la conexión realmente responde
	if err := pool.Ping(ctx); err != nil {
		log.Fatal("no hay conexión con la base de datos: ", err)
	}
	fmt.Println("Conectado a Postgres")

	// Leemos las publicaciones
	rows, err := pool.Query(ctx, "SELECT id, titulo, contenido FROM publicaciones")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id, titulo, contenido string
		if err := rows.Scan(&id, &titulo, &contenido); err != nil {
			log.Fatal(err)
		}
		fmt.Println(id, "|", titulo, "|", contenido)
	}
}