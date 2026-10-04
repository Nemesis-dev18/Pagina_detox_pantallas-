package main

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Visible para todos los archivos del paquete
var pool *pgxpool.Pool

func conectarDB(ctx context.Context) error {
	var err error
	pool, err = pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		return err
	}
	return pool.Ping(ctx)
}
