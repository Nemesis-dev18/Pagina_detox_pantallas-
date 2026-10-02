package main

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

// Regla: si alguien escribe el MISMO comentario más de 5 veces seguidas dentro
// de 4 minutos, se le bloquea para comentar durante 10 minutos.
// (Se permiten 5; el sexto es el que dispara el bloqueo y no se guarda.)
const (
	spamMaxRepetidos = 4
	spamVentana      = 4 * time.Minute
	spamBloqueo      = 10 * time.Minute
)

// errBloqueado se devuelve cuando la persona no puede comentar todavía.
type errBloqueado struct {
	Hasta time.Time
}

func (e errBloqueado) Error() string {
	return "usuario bloqueado para comentar hasta " + e.Hasta.Format(time.RFC3339)
}

func (e errBloqueado) segundos() int {
	s := int(math.Ceil(time.Until(e.Hasta).Seconds()))
	if s < 1 {
		s = 1
	}
	return s
}

// responderBloqueo contesta 429 con un mensaje claro y el encabezado Retry-After.
func responderBloqueo(w http.ResponseWriter, e errBloqueado) {
	s := e.segundos()
	minutos := (s + 59) / 60
	w.Header().Set("Retry-After", strconv.Itoa(s))
	http.Error(w,
		fmt.Sprintf("Bloqueado por spam: repetiste el mismo comentario demasiadas veces. Podrás comentar de nuevo en %d min.", minutos),
		http.StatusTooManyRequests)
}

// claveSpam deja el texto "plano" para que "Hola", "hola!!" o "h o l a" cuenten
// como el mismo comentario. Si no queda nada (solo emojis, por ejemplo), usa el
// texto tal cual en minúsculas.
func claveSpam(texto string) string {
	_, plano, _ := aplanar(texto)
	if len(plano) == 0 {
		return strings.ToLower(strings.TrimSpace(texto))
	}
	return string(plano)
}

// guardarComentario revisa el bloqueo y las repeticiones y, si todo está bien,
// inserta el comentario. Todo ocurre en una transacción con un candado por
// usuario, para que mandar 10 peticiones a la vez no se salte la regla.
func guardarComentario(ctx context.Context, usuarioID, publicacionID, contenido string) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) // si ya se hizo Commit, no hace nada

	// Un comentario a la vez por usuario
	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock(hashtext($1::text))", usuarioID); err != nil {
		return err
	}

	// 1) ¿Sigue bloqueado?
	var hasta time.Time
	err = tx.QueryRow(ctx,
		"SELECT hasta FROM bloqueos_comentarios WHERE usuario_id = $1 AND hasta > now()",
		usuarioID).Scan(&hasta)
	if err == nil {
		return errBloqueado{Hasta: hasta}
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	// 2) ¿Cuántos comentarios idénticos lleva seguidos en la ventana de 4 min?
	rows, err := tx.Query(ctx, `
		SELECT contenido FROM comentarios
		WHERE autor_id = $1 AND creado_en > now() - make_interval(secs => $2)
		ORDER BY creado_en DESC
		LIMIT $3`,
		usuarioID, spamVentana.Seconds(), spamMaxRepetidos)
	if err != nil {
		return err
	}
	clave := claveSpam(contenido)
	seguidos := 0
	for rows.Next() {
		var previo string
		if err := rows.Scan(&previo); err != nil {
			rows.Close()
			return err
		}
		if claveSpam(previo) != clave {
			break // un comentario distinto en medio corta la racha
		}
		seguidos++
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	if seguidos >= spamMaxRepetidos {
		// Este sería el sexto igual: se bloquea y NO se guarda.
		// Aquí sí hay que hacer Commit, o el bloqueo se desharía con el Rollback.
		err := tx.QueryRow(ctx, `
			INSERT INTO bloqueos_comentarios (usuario_id, hasta)
			VALUES ($1, now() + make_interval(secs => $2))
			ON CONFLICT (usuario_id) DO UPDATE SET hasta = EXCLUDED.hasta
			RETURNING hasta`,
			usuarioID, spamBloqueo.Seconds()).Scan(&hasta)
		if err != nil {
			return err
		}
		if err := tx.Commit(ctx); err != nil {
			return err
		}
		return errBloqueado{Hasta: hasta}
	}

	// 3) Todo bien: se guarda
	if _, err := tx.Exec(ctx,
		"INSERT INTO comentarios (publicacion_id, autor_id, contenido) VALUES ($1, $2, $3)",
		publicacionID, usuarioID, contenido); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
