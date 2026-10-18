package main

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
)

// Un especialista puede editar su perfil como máximo dos veces por mes calendario.
// Solo se editan la especialidad y la presentación. La profesión, el número de
// registro, el documento y la institución pasan por verificación: los cambia un admin.
const maxEdicionesMes = 2

// consultor lo cumplen tanto el pool como una transacción.
type consultor interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func contarEdicionesMes(ctx context.Context, c consultor, usuarioID string) (int, error) {
	var n int
	err := c.QueryRow(ctx, `
		SELECT count(*) FROM perfil_ediciones
		WHERE usuario_id = $1 AND date_trunc('month', creado_en) = date_trunc('month', now())`,
		usuarioID).Scan(&n)
	return n, err
}

// GET /yo/especialista/perfil → datos editables y cuántas ediciones quedan este mes
func miPerfilEspecialista(w http.ResponseWriter, r *http.Request) {
	u, err := usuarioActual(r)
	if err != nil {
		http.Error(w, "no has iniciado sesión", http.StatusUnauthorized)
		return
	}
	var v struct {
		Especialidad       string `json:"especialidad"`
		Descripcion        string `json:"descripcion"`
		Estado             string `json:"estado"`
		EdicionesRestantes int    `json:"ediciones_restantes"`
	}
	err = pool.QueryRow(r.Context(),
		`SELECT especialidad, COALESCE(descripcion, ''), estado FROM perfiles_especialistas WHERE usuario_id = $1`,
		u.ID).Scan(&v.Especialidad, &v.Descripcion, &v.Estado)
	if errors.Is(err, pgx.ErrNoRows) {
		http.Error(w, "no eres especialista", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "error consultando la base", http.StatusInternalServerError)
		return
	}
	usadas, err := contarEdicionesMes(r.Context(), pool, u.ID)
	if err != nil {
		http.Error(w, "error consultando la base", http.StatusInternalServerError)
		return
	}
	v.EdicionesRestantes = max(0, maxEdicionesMes-usadas)
	escribirJSON(w, v)
}

// PUT /yo/especialista/perfil   (especialidad, descripcion)
func editarMiPerfilEspecialista(w http.ResponseWriter, r *http.Request) {
	u, err := usuarioActual(r)
	if err != nil {
		http.Error(w, "no has iniciado sesión", http.StatusUnauthorized)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<14)

	especialidad := limpiar(r.FormValue("especialidad"))
	descripcion := strings.TrimSpace(r.FormValue("descripcion"))
	if n := utf8.RuneCountInString(especialidad); n < 3 || n > 100 || !sinControl(especialidad) {
		http.Error(w, "la especialidad debe tener entre 3 y 100 caracteres", http.StatusBadRequest)
		return
	}
	if utf8.RuneCountInString(descripcion) > 600 || !sinControl(descripcion) {
		http.Error(w, "la presentación es muy larga (máximo 600 caracteres)", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	tx, err := pool.Begin(ctx)
	if err != nil {
		http.Error(w, "error interno", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(ctx)

	// Un cambio a la vez por persona, para que dos peticiones no se salten el límite
	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock(hashtext($1::text))", u.ID); err != nil {
		http.Error(w, "error interno", http.StatusInternalServerError)
		return
	}
	var estado string
	err = tx.QueryRow(ctx, "SELECT estado FROM perfiles_especialistas WHERE usuario_id = $1 FOR UPDATE", u.ID).Scan(&estado)
	if errors.Is(err, pgx.ErrNoRows) {
		http.Error(w, "no eres especialista", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "error consultando la base", http.StatusInternalServerError)
		return
	}
	if estado == "suspendido" {
		http.Error(w, "tu perfil está suspendido; no puedes editarlo", http.StatusForbidden)
		return
	}

	usadas, err := contarEdicionesMes(ctx, tx, u.ID)
	if err != nil {
		http.Error(w, "error consultando la base", http.StatusInternalServerError)
		return
	}
	if usadas >= maxEdicionesMes {
		http.Error(w, "ya usaste tus 2 ediciones de este mes; podrás editar de nuevo el próximo mes", http.StatusTooManyRequests)
		return
	}

	if _, err := tx.Exec(ctx,
		"UPDATE perfiles_especialistas SET especialidad = $2, descripcion = NULLIF($3, '') WHERE usuario_id = $1",
		u.ID, especialidad, descripcion); err != nil {
		http.Error(w, "error guardando los cambios", http.StatusInternalServerError)
		return
	}
	if _, err := tx.Exec(ctx, "INSERT INTO perfil_ediciones (usuario_id) VALUES ($1)", u.ID); err != nil {
		http.Error(w, "error guardando los cambios", http.StatusInternalServerError)
		return
	}
	if err := tx.Commit(ctx); err != nil {
		http.Error(w, "error interno", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
