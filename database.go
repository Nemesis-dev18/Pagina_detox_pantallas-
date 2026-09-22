package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

// postgresDSNFromEnv obtiene la DSN desde DATABASE_URL o devuelve vacío
func postgresDSNFromEnv() string {
	if d := os.Getenv("DATABASE_URL"); d != "" {
		return d
	}
	// fallback simple (no recomendado para producción)
	user := os.Getenv("PGUSER")
	pass := os.Getenv("PGPASSWORD")
	host := os.Getenv("PGHOST")
	port := os.Getenv("PGPORT")
	name := os.Getenv("PGDATABASE")
	if port == "" {
		port = "5432"
	}
	if user == "" || pass == "" || host == "" || name == "" {
		return ""
	}
	return "postgresql://" + user + ":" + pass + "@" + host + ":" + port + "/" + name + "?sslmode=disable"
}

// initDB abre la conexión, verifica y crea esquema mínimo
func initDB() *sql.DB {
	dsn := postgresDSNFromEnv()
	if dsn == "" {
		log.Fatal("DATABASE_URL no definido y variables PG* incompletas")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("sql.Open: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("db.Ping: %v", err)
	}
	if err := createSchema(db); err != nil {
		log.Fatalf("createSchema: %v", err)
	}
	return db
}

func createSchema(db *sql.DB) error {
	schema := `
CREATE TABLE IF NOT EXISTS users (
  id SERIAL PRIMARY KEY,
  email TEXT UNIQUE NOT NULL,
  password_hash TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS password_resets (
  id SERIAL PRIMARY KEY,
  user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token TEXT NOT NULL UNIQUE,
  expires_at TIMESTAMPTZ NOT NULL,
  used BOOLEAN NOT NULL DEFAULT false
);
`
	_, err := db.Exec(schema)
	return err
}