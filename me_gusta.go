package main

import (
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5/pgconn"
)

// Me gusta en publicaciones y comentarios. Solo con sesión.
// PUT da el me gusta (si ya lo tenía, no pasa nada); DELETE lo quita.

// PUT y DELETE /publicaciones/{id}/me-gusta
func meGustaPublicacion(w http.ResponseWriter, r *http.Request) {
	u, err := usuarioActual(r)
	if err != nil {
		http.Error(w, "debes iniciar sesión para dar me gusta", http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")
	if !reUUID.MatchString(id) {
		http.Error(w, "publicación no encontrada", http.StatusNotFound)
		return
	}

	if r.Method == http.MethodDelete {
		_, err = pool.Exec(r.Context(),
			"DELETE FROM me_gusta WHERE usuario_id = $1 AND publicacion_id = $2", u.ID, id)
	} else {
		var existe bool
		err = pool.QueryRow(r.Context(),
			"SELECT EXISTS (SELECT 1 FROM publicaciones WHERE id = $1)", id).Scan(&existe)
		if err == nil && !existe {
			http.Error(w, "publicación no encontrada", http.StatusNotFound)
			return
		}
		if err == nil {
			_, err = pool.Exec(r.Context(),
				"INSERT INTO me_gusta (usuario_id, publicacion_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", u.ID, id)
		}
	}
	if err != nil {
		http.Error(w, "error guardando el me gusta", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// PUT y DELETE /publicaciones/{id}/comentarios/{comentarioId}/me-gusta
// Se exige que el comentario sea de esa publicación.
func meGustaComentario(w http.ResponseWriter, r *http.Request) {
	u, err := usuarioActual(r)
	if err != nil {
		http.Error(w, "debes iniciar sesión para dar me gusta", http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")
	comentarioID := r.PathValue("comentarioId")
	if !reUUID.MatchString(id) || !reUUID.MatchString(comentarioID) {
		http.Error(w, "comentario no encontrado", http.StatusNotFound)
		return
	}

	if r.Method == http.MethodDelete {
		_, err = pool.Exec(r.Context(),
			"DELETE FROM me_gusta WHERE usuario_id = $1 AND comentario_id = $2", u.ID, comentarioID)
	} else {
		var existe bool
		err = pool.QueryRow(r.Context(),
			"SELECT EXISTS (SELECT 1 FROM comentarios WHERE id = $1 AND publicacion_id = $2)",
			comentarioID, id).Scan(&existe)
		if err == nil && !existe {
			http.Error(w, "comentario no encontrado", http.StatusNotFound)
			return
		}
		if err == nil {
			_, err = pool.Exec(r.Context(),
				"INSERT INTO me_gusta (usuario_id, comentario_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", u.ID, comentarioID)
		}
	}
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			http.Error(w, "comentario no encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, "error guardando el me gusta", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
