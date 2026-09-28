package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/jackc/pgx/v5"
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

// ---------- Recuperar contraseña ----------

func generarToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func baseURL() string {
	if u := os.Getenv("BASE_URL"); u != "" {
		return u
	}
	return "http://localhost:8080"
}

func enviarEnlaceRecuperacion(ctx context.Context, usuarioID, correo string) error {
	token, err := generarToken()
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx,
		"INSERT INTO recuperaciones (token_hash, usuario_id, expira_en) VALUES ($1, $2, now() + interval '30 minutes')",
		hashToken(token), usuarioID)
	if err != nil {
		return err
	}

	enlace := baseURL() + "/static/restablecer.html?token=" + token
	cuerpo := "Hola,\n\nPara elegir una nueva contraseña entra a este enlace (vale 30 minutos):\n\n" +
		enlace + "\n\nSi no fuiste tú, ignora este mensaje.\n"
	return enviarCorreo(correo, "Recuperar contraseña | DETOX_WEB", cuerpo)
}

func recuperar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
		return
	}

	correo := strings.ToLower(strings.TrimSpace(r.FormValue("correo")))

	var usuarioID string
	err := pool.QueryRow(r.Context(), "SELECT id FROM usuarios WHERE correo = $1", correo).Scan(&usuarioID)
	if err == nil {
		if err := enviarEnlaceRecuperacion(r.Context(), usuarioID, correo); err != nil {
			log.Println("recuperación:", err)
		}
	} else if !errors.Is(err, pgx.ErrNoRows) {
		log.Println("recuperación:", err)
	}

	// Siempre la misma respuesta, exista o no el correo
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, `<p>Si ese correo está registrado, te enviamos un enlace para restablecer tu contraseña. Revisa tu bandeja de entrada.</p><p><a href="/static/inicio_sesion.html">Volver a iniciar sesión</a></p>`)
}

func restablecer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
		return
	}

	token := r.FormValue("token")
	password := r.FormValue("password")
	confirmar := r.FormValue("confirmar_password")

	if token == "" || password == "" {
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

	ctx := r.Context()
	tx, err := pool.Begin(ctx)
	if err != nil {
		http.Error(w, "error interno", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(ctx)

	// Marca el token como usado y devuelve de qué usuario es (solo si sirve)
	var usuarioID string
	err = tx.QueryRow(ctx,
		"UPDATE recuperaciones SET usado = true WHERE token_hash = $1 AND usado = false AND expira_en > now() RETURNING usuario_id",
		hashToken(token)).Scan(&usuarioID)
	if errors.Is(err, pgx.ErrNoRows) {
		http.Error(w, "el enlace no es válido o ya venció", http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, "error interno", http.StatusInternalServerError)
		return
	}

	if _, err = tx.Exec(ctx, "UPDATE usuarios SET password_hash = $1 WHERE id = $2", string(hash), usuarioID); err != nil {
		http.Error(w, "error interno", http.StatusInternalServerError)
		return
	}
	if _, err = tx.Exec(ctx, "DELETE FROM sesiones WHERE usuario_id = $1", usuarioID); err != nil {
		http.Error(w, "error interno", http.StatusInternalServerError)
		return
	}
	if err = tx.Commit(ctx); err != nil {
		http.Error(w, "error interno", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/static/inicio_sesion.html", http.StatusSeeOther)
}