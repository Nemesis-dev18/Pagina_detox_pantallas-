package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)
type Publicacion struct {
	ID        string `json:"id"`
	Titulo    string `json:"titulo"`
	Contenido string `json:"contenido"`
}

func listarPublicaciones(w http.ResponseWriter, r *http.Request) {
	rows, err := pool.Query(r.Context(), "SELECT id, titulo, contenido FROM publicaciones")
	if err != nil {
		http.Error(w, "error consultando la base", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	publicaciones := []Publicacion{}
	for rows.Next() {
		var p Publicacion
		if err := rows.Scan(&p.ID, &p.Titulo, &p.Contenido); err != nil {
			http.Error(w, "error leyendo datos", http.StatusInternalServerError)
			return
		}
		publicaciones = append(publicaciones, p)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(publicaciones)
}
func registrar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
		return
	}

	alias := strings.TrimSpace(r.FormValue("alias"))
	correo := strings.ToLower(strings.TrimSpace(r.FormValue("correo")))
	password := r.FormValue("password")
	confirmar := r.FormValue("confirmar_password")

	if alias == "" || correo == "" || password == "" {
		http.Error(w, "faltan datos", http.StatusBadRequest)
		return
	}
	if password != confirmar {
		http.Error(w, "las contraseñas no coinciden", http.StatusBadRequest)
		return
	}
	if len(password) < 8 {
		http.Error(w, "la contraseña debe tener mínimo 8 caracteres", http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "error interno", http.StatusInternalServerError)
		return
	}

	_, err = pool.Exec(r.Context(),
		"INSERT INTO usuarios (alias, correo, password_hash) VALUES ($1, $2, $3)",
		alias, correo, string(hash))
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			http.Error(w, "ese alias o correo ya está en uso", http.StatusConflict)
			return
		}
		http.Error(w, "error guardando el usuario", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/static/inicio_sesion.html", http.StatusSeeOther)
}