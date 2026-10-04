package main

import (
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
)

// PUT /publicaciones/{id}
// Edita título, categoría y contenido. Solo el autor o un admin.
func editarPublicacion(w http.ResponseWriter, r *http.Request) {
	u, err := usuarioActual(r)
	if err != nil {
		http.Error(w, "debes iniciar sesión", http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")
	if !reUUID.MatchString(id) {
		http.Error(w, "publicación no encontrada", http.StatusNotFound)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // máximo 1 MB por petición

	titulo := strings.TrimSpace(r.FormValue("titulo"))
	contenido := strings.TrimSpace(r.FormValue("contenido"))
	categoria := strings.TrimSpace(r.FormValue("categoria"))

	// Mismas reglas que al crear
	if titulo == "" || contenido == "" {
		http.Error(w, "el título y el contenido son obligatorios", http.StatusBadRequest)
		return
	}
	// Si llega texto ya tapado, no se guarda: pisaría el original
	if strings.Contains(titulo, "*****") || strings.Contains(contenido, "*****") {
		http.Error(w, "no puedes guardar texto censurado", http.StatusBadRequest)
		return
	}
	if len([]rune(titulo)) > 120 {
		http.Error(w, "el título es muy largo (máximo 120 caracteres)", http.StatusBadRequest)
		return
	}
	if len([]rune(contenido)) > 5000 {
		http.Error(w, "el contenido es muy largo (máximo 5000 caracteres)", http.StatusBadRequest)
		return
	}
	if !categoriasValidas[categoria] {
		http.Error(w, "categoría no válida", http.StatusBadRequest)
		return
	}

	var tag pgconn.CommandTag
	if u.EsAdmin {
		tag, err = pool.Exec(r.Context(),
			"UPDATE publicaciones SET titulo = $1, contenido = $2, categoria = $3, editado_en = now() WHERE id = $4",
			titulo, contenido, categoria, id)
	} else {
		tag, err = pool.Exec(r.Context(),
			"UPDATE publicaciones SET titulo = $1, contenido = $2, categoria = $3, editado_en = now() WHERE id = $4 AND autor_id = $5",
			titulo, contenido, categoria, id, u.ID)
	}
	if err != nil {
		http.Error(w, "error guardando los cambios", http.StatusInternalServerError)
		return
	}
	if tag.RowsAffected() == 0 {
		http.Error(w, "no encontrada, o no es tuya", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// PUT /publicaciones/{id}/comentarios/{comentarioId}
// Edita el texto de un comentario. Solo el autor o un admin.
func editarComentario(w http.ResponseWriter, r *http.Request) {
	u, err := usuarioActual(r)
	if err != nil {
		http.Error(w, "debes iniciar sesión", http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")
	comentarioID := r.PathValue("comentarioId")
	if !reUUID.MatchString(id) || !reUUID.MatchString(comentarioID) {
		http.Error(w, "comentario no encontrado", http.StatusNotFound)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	contenido := strings.TrimSpace(r.FormValue("contenido"))
	if contenido == "" {
		http.Error(w, "escribe algo para responder", http.StatusBadRequest)
		return
	}
	// Si llega texto ya tapado, no se guarda: pisaría el original
	if strings.Contains(contenido, "*****") {
		http.Error(w, "no puedes guardar texto censurado", http.StatusBadRequest)
		return
	}
	if len([]rune(contenido)) > 2000 {
		http.Error(w, "la respuesta es muy larga (máximo 2000 caracteres)", http.StatusBadRequest)
		return
	}

	// Se exige también que el comentario sea de esa publicación
	var tag pgconn.CommandTag
	if u.EsAdmin {
		tag, err = pool.Exec(r.Context(),
			"UPDATE comentarios SET contenido = $1, editado_en = now() WHERE id = $2 AND publicacion_id = $3",
			contenido, comentarioID, id)
	} else {
		tag, err = pool.Exec(r.Context(),
			"UPDATE comentarios SET contenido = $1, editado_en = now() WHERE id = $2 AND publicacion_id = $3 AND autor_id = $4",
			contenido, comentarioID, id, u.ID)
	}
	if err != nil {
		http.Error(w, "error guardando los cambios", http.StatusInternalServerError)
		return
	}
	if tag.RowsAffected() == 0 {
		http.Error(w, "no encontrado, o no es tuyo", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
