package main

import (
	"context"
	"embed"
	"fmt"
	"sort"
)

//go:embed migraciones/*.sql
var migracionesFS embed.FS

// aplicarMigraciones ejecuta, en orden y una sola vez cada una, los archivos de
// migraciones/. Todo ocurre en una transacción con candado, así que si algo
// falla no queda nada a medias y dos instancias no migran a la vez.
func aplicarMigraciones(ctx context.Context) error {
	archivos, err := migracionesFS.ReadDir("migraciones")
	if err != nil {
		return err
	}
	nombres := []string{}
	for _, a := range archivos {
		nombres = append(nombres, a.Name())
	}
	sort.Strings(nombres)

	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock(727001)"); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `CREATE TABLE IF NOT EXISTS migraciones_aplicadas (
		nombre text PRIMARY KEY, aplicada_en timestamptz NOT NULL DEFAULT now())`); err != nil {
		return err
	}
	for _, nombre := range nombres {
		var existe bool
		if err := tx.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM migraciones_aplicadas WHERE nombre = $1)", nombre).Scan(&existe); err != nil {
			return err
		}
		if existe {
			continue
		}
		sql, err := migracionesFS.ReadFile("migraciones/" + nombre)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, string(sql)); err != nil {
			return fmt.Errorf("migración %s: %w", nombre, err)
		}
		if _, err := tx.Exec(ctx, "INSERT INTO migraciones_aplicadas (nombre) VALUES ($1)", nombre); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}
