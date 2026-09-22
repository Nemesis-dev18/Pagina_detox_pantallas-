package main

import (
    "database/sql"
    "fmt"
    "net/http"

    "golang.org/x/crypto/bcrypt"
)

func loginHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Use POST", http.StatusMethodNotAllowed)
        return
    }

    email := r.FormValue("username")
    password := r.FormValue("password")

    if email == "" || password == "" {
        http.Error(w, "faltan campos", http.StatusBadRequest)
        return
    }

    var id int
    var hash string
    err := db.QueryRow("SELECT id, password_hash FROM users WHERE email = ?", email).Scan(&id, &hash)
    if err == sql.ErrNoRows {
        http.Error(w, "credenciales inválidas", http.StatusUnauthorized)
        return
    } else if err != nil {
        http.Error(w, "error del servidor", http.StatusInternalServerError)
        return
    }

    if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
        http.Error(w, "credenciales inválidas", http.StatusUnauthorized)
        return
    }

    // Autenticación OK: aquí puedes redirigir o establecer cookie de sesión
    fmt.Fprintf(w, "Bienvenido usuario %d", id)
}