package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// writeJSON helper
func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

// registerHandler crea un usuario nuevo
func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "use POST"})
		return
	}
	if err := r.ParseForm(); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid form"})
		return
	}
	email := r.Form.Get("email")
	pass := r.Form.Get("password")
	if email == "" || pass == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing fields"})
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "server error"})
		return
	}
	_, err = db.Exec("INSERT INTO users (email, password_hash) VALUES ($1, $2)", email, string(hash))
	if err != nil {
		// posible clave duplicada (email)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "could not create user"})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"message": "created"})
}

// loginHandler autentica
func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "use POST"})
		return
	}
	if err := r.ParseForm(); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid form"})
		return
	}
	email := r.Form.Get("username")
	pass := r.Form.Get("password")
	if email == "" || pass == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing fields"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	var id int
	var hash string
	err := db.QueryRowContext(ctx, "SELECT id, password_hash FROM users WHERE email = $1", email).Scan(&id, &hash)
	if err != nil {
		if err == sql.ErrNoRows {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "server error"})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(pass)); err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		return
	}
	// En producción: genera JWT o establece cookie segura.
	writeJSON(w, http.StatusOK, map[string]any{"message": "ok", "user_id": id})
}

// requestResetHandler genera token y envía email
func requestResetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "use POST"})
		return
	}
	if err := r.ParseForm(); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid form"})
		return
	}
	email := r.Form.Get("email")
	if email == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing email"})
		return
	}

	var userID int
	err := db.QueryRow("SELECT id FROM users WHERE email = $1", email).Scan(&userID)
	if err == sql.ErrNoRows {
		// Responder genérico para no revelar existencia
		writeJSON(w, http.StatusOK, map[string]string{"message": "if the email exists, you'll receive a link"})
		return
	} else if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "server error"})
		return
	}

	token := generateToken(32)
	expires := time.Now().Add(1 * time.Hour)
	_, err = db.Exec("INSERT INTO password_resets (user_id, token, expires_at) VALUES ($1, $2, $3)", userID, token, expires)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "server error"})
		return
	}

	// envía email de forma no bloqueante
	cfg := getEmailConfigFromEnv()
	resetLink := fmt.Sprintf("%s/reset-password?token=%s", os.Getenv("BASE_URL"), token)
	go sendPasswordReset(cfg, email, resetLink)

	writeJSON(w, http.StatusOK, map[string]string{"message": "if the email exists, you'll receive a link"})
}

// resetPasswordHandler recibe token y nueva contraseña (espera JSON)
func resetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "use POST"})
		return
	}
	var body struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}
	if body.Token == "" || body.NewPassword == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing fields"})
		return
	}

	var prID, userID int
	var expires time.Time
	var used bool
	err := db.QueryRow("SELECT id, user_id, expires_at, used FROM password_resets WHERE token = $1", body.Token).Scan(&prID, &userID, &expires, &used)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid or expired token"})
		return
	}
	if used || time.Now().After(expires) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid or expired token"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(body.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "server error"})
		return
	}

	tx, err := db.Begin()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "server error"})
		return
	}
	_, err = tx.Exec("UPDATE users SET password_hash = $1 WHERE id = $2", string(hash), userID)
	if err != nil {
		tx.Rollback()
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "server error"})
		return
	}
	_, err = tx.Exec("UPDATE password_resets SET used = true WHERE id = $1", prID)
	if err != nil {
		tx.Rollback()
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "server error"})
		return
	}
	if err := tx.Commit(); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "server error"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "password reset successful"})
}

// generateToken genera una cadena segura para tokens (URL-safe)
func generateToken(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}