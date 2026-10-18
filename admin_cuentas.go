package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
)

type cuentaAdmin struct {
	ID       string    `json:"id"`
	Alias    string    `json:"alias"`
	Correo   string    `json:"correo"`
	Tipo     string    `json:"tipo"`
	EsAdmin  bool      `json:"es_admin"`
	CreadoEn time.Time `json:"creado_en"`
}

// GET /admin/cuentas?q=texto  → busca por alias o correo (solo administradores)
func listarCuentas(w http.ResponseWriter, r *http.Request) {
	if _, ok := exigirAdmin(w, r); !ok {
		return
	}
	q := limpiar(r.URL.Query().Get("q"))
	if utf8.RuneCountInString(q) > 80 {
		http.Error(w, "búsqueda muy larga", http.StatusBadRequest)
		return
	}
	patron := "%" + escaparLike(q) + "%"

	rows, err := pool.Query(r.Context(), `
		SELECT id, alias, correo, tipo, es_admin, creado_en
		FROM usuarios
		WHERE ($1 = '%%' OR alias ILIKE $1 OR correo ILIKE $1)
		ORDER BY creado_en DESC
		LIMIT 100`, patron)
	if err != nil {
		http.Error(w, "error consultando la base", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	lista := []cuentaAdmin{}
	for rows.Next() {
		var c cuentaAdmin
		if err := rows.Scan(&c.ID, &c.Alias, &c.Correo, &c.Tipo, &c.EsAdmin, &c.CreadoEn); err != nil {
			http.Error(w, "error leyendo datos", http.StatusInternalServerError)
			return
		}
		lista = append(lista, c)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, "error leyendo datos", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "private, no-store")
	json.NewEncoder(w).Encode(lista)
}

// DELETE /admin/cuentas/{id}
// Borra la cuenta y lo que depende de ella. Reglas:
//   - no se puede borrar la propia cuenta ni la de otro administrador;
//   - se borran sus publicaciones y comentarios (no tienen borrado en cascada);
//   - el resto (sesiones, perfil de especialista, centros, acompañamientos) se
//     borra solo por las claves foráneas.
//
// Todo ocurre en una transacción: si algo falla, no queda nada a medias.
func borrarCuenta(w http.ResponseWriter, r *http.Request) {
	admin, ok := exigirAdmin(w, r)
	if !ok {
		return
	}
	id := r.PathValue("id")
	if !reUUID.MatchString(id) {
		http.Error(w, "cuenta no encontrada", http.StatusNotFound)
		return
	}
	if id == admin.ID {
		http.Error(w, "no puedes borrar tu propia cuenta desde aquí", http.StatusForbidden)
		return
	}

	ctx := r.Context()
	tx, err := pool.Begin(ctx)
	if err != nil {
		http.Error(w, "error interno", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(ctx)

	var esAdmin bool
	var alias string
	err = tx.QueryRow(ctx, "SELECT es_admin, alias FROM usuarios WHERE id = $1 FOR UPDATE", id).Scan(&esAdmin, &alias)
	if errors.Is(err, pgx.ErrNoRows) {
		http.Error(w, "cuenta no encontrada", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "error consultando la base", http.StatusInternalServerError)
		return
	}
	if esAdmin {
		http.Error(w, "no se puede borrar a otro administrador desde aquí", http.StatusForbidden)
		return
	}

	// Lo que no tiene borrado en cascada, se borra primero
	if _, err := tx.Exec(ctx, "DELETE FROM comentarios WHERE autor_id = $1", id); err != nil {
		http.Error(w, "error borrando los comentarios", http.StatusInternalServerError)
		return
	}
	if _, err := tx.Exec(ctx, "DELETE FROM publicaciones WHERE autor_id = $1", id); err != nil {
		http.Error(w, "error borrando las publicaciones", http.StatusInternalServerError)
		return
	}
	if _, err := tx.Exec(ctx, "DELETE FROM respuestas_cuestionario WHERE usuario_id = $1", id); err != nil {
		http.Error(w, "error borrando las respuestas", http.StatusInternalServerError)
		return
	}

	tag, err := tx.Exec(ctx, "DELETE FROM usuarios WHERE id = $1 AND es_admin = false", id)
	if err != nil {
		http.Error(w, "error borrando la cuenta", http.StatusInternalServerError)
		return
	}
	if tag.RowsAffected() == 0 {
		http.Error(w, "cuenta no encontrada", http.StatusNotFound)
		return
	}
	if err := tx.Commit(ctx); err != nil {
		http.Error(w, "error interno", http.StatusInternalServerError)
		return
	}

	// Se anota quién borró a quién, nunca el correo ni el contenido
	log.Printf("admin %s borró la cuenta %s (alias %s)", admin.ID, id, strings.TrimSpace(alias))
	w.WriteHeader(http.StatusNoContent)
}
