package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
	"unicode"
)

// Devuelve el usuario si es admin; si no, responde el error y devuelve false.
func exigirAdmin(w http.ResponseWriter, r *http.Request) (*Usuario, bool) {
	u, err := usuarioActual(r)
	if err != nil {
		http.Error(w, "no has iniciado sesión", http.StatusUnauthorized)
		return nil, false
	}
	if !u.EsAdmin {
		http.Error(w, "solo para administradores", http.StatusForbidden)
		return nil, false
	}
	return u, true
}

type palabraVista struct {
	ID       string    `json:"id"`
	Palabra  string    `json:"palabra"`
	CreadoEn time.Time `json:"creado_en"`
}

func listarPalabras(w http.ResponseWriter, r *http.Request) {
	if _, ok := exigirAdmin(w, r); !ok {
		return
	}
	rows, err := pool.Query(r.Context(),
		"SELECT id, palabra, creado_en FROM palabras_excluidas ORDER BY palabra")
	if err != nil {
		http.Error(w, "error consultando la base", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	palabras := []palabraVista{}
	for rows.Next() {
		var p palabraVista
		if err := rows.Scan(&p.ID, &p.Palabra, &p.CreadoEn); err != nil {
			http.Error(w, "error leyendo datos", http.StatusInternalServerError)
			return
		}
		palabras = append(palabras, p)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, "error leyendo datos", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(palabras)
}

func agregarPalabra(w http.ResponseWriter, r *http.Request) {
	if _, ok := exigirAdmin(w, r); !ok {
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<16)

	palabra := strings.Join(strings.Fields(strings.ToLower(r.FormValue("palabra"))), " ")
	for _, c := range palabra {
		if !unicode.IsLetter(c) && !unicode.IsDigit(c) && c != ' ' {
			http.Error(w, "usa solo letras, números y espacios", http.StatusBadRequest)
			return
		}
	}
	if len([]rune(palabra)) > 40 || compilarPalabra(palabra) == nil {
		http.Error(w, "la palabra debe tener entre 3 y 40 caracteres", http.StatusBadRequest)
		return
	}

	tag, err := pool.Exec(r.Context(),
		"INSERT INTO palabras_excluidas (palabra) VALUES ($1) ON CONFLICT (palabra) DO NOTHING",
		palabra)
	if err != nil {
		http.Error(w, "error guardando la palabra", http.StatusInternalServerError)
		return
	}
	if tag.RowsAffected() == 0 {
		http.Error(w, "esa palabra ya está en la lista", http.StatusConflict)
		return
	}
	if err := recargarFiltro(r.Context()); err != nil {
		http.Error(w, "se guardó, pero falló al recargar el filtro", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func borrarPalabra(w http.ResponseWriter, r *http.Request) {
	if _, ok := exigirAdmin(w, r); !ok {
		return
	}
	id := r.PathValue("id")
	if !reUUID.MatchString(id) {
		http.Error(w, "palabra no encontrada", http.StatusNotFound)
		return
	}
	tag, err := pool.Exec(r.Context(), "DELETE FROM palabras_excluidas WHERE id = $1", id)
	if err != nil {
		http.Error(w, "error borrando la palabra", http.StatusInternalServerError)
		return
	}
	if tag.RowsAffected() == 0 {
		http.Error(w, "palabra no encontrada", http.StatusNotFound)
		return
	}
	if err := recargarFiltro(r.Context()); err != nil {
		http.Error(w, "se borró, pero falló al recargar el filtro", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type coincidenciaVista struct {
	ComentarioID string    `json:"comentario_id"`
	Alias        string    `json:"alias"`
	Palabra      string    `json:"palabra"`
	Fragmento    string    `json:"fragmento"` // tal como lo escribió la persona
	Contenido    string    `json:"contenido"` // comentario original, sin censurar
	CreadoEn     time.Time `json:"creado_en"`
}

// Revisa los últimos comentarios y devuelve los que el filtro tapó, con el
// texto ORIGINAL y la forma exacta en que escribieron la palabra. Sirve para
// ver cómo intentan saltarse la regla y agregar nuevas variantes a la lista.
func listarCoincidencias(w http.ResponseWriter, r *http.Request) {
	if _, ok := exigirAdmin(w, r); !ok {
		return
	}
	rows, err := pool.Query(r.Context(), `
		SELECT c.id, u.alias, c.contenido, c.creado_en
		FROM comentarios c
		JOIN usuarios u ON u.id = c.autor_id
		ORDER BY c.creado_en DESC
		LIMIT 500`)
	if err != nil {
		http.Error(w, "error consultando la base", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	vistas := []coincidenciaVista{}
	for rows.Next() {
		var id, alias, contenido string
		var creado time.Time
		if err := rows.Scan(&id, &alias, &contenido, &creado); err != nil {
			http.Error(w, "error leyendo datos", http.StatusInternalServerError)
			return
		}
		runas := []rune(contenido)
		for _, m := range buscarCoincidencias(contenido) {
			if m.Inicio < 0 || m.Fin > len(runas) || m.Inicio >= m.Fin {
				continue
			}
			vistas = append(vistas, coincidenciaVista{
				ComentarioID: id,
				Alias:        alias,
				Palabra:      m.Palabra,
				Fragmento:    string(runas[m.Inicio:m.Fin]),
				Contenido:    contenido,
				CreadoEn:     creado,
			})
		}
	}
	if err := rows.Err(); err != nil {
		http.Error(w, "error leyendo datos", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(vistas)
}
