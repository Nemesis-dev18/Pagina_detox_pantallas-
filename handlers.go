package main

import (
    "database/sql"
    "fmt"
    "log"
    "net/http"

    "golang.org/x/crypto/bcrypt"
    _ "modernc.org/sqlite"
)

var db *sql.DB

func main() {
    var err error
    db, err = sql.Open("sqlite", "app.db")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    http.HandleFunc("/login", loginHandler)
    log.Println("listening :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Use POST", http.StatusMethodNotAllowed)
        return
    }

    // Leer campos del formulario (OK con application/x-www-form-urlencoded o multipart/form-data)
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
        // No revelar detalles: responder genérico
        http.Error(w, "credenciales inválidas", http.StatusUnauthorized)
        return
    } else if err != nil {
        http.Error(w, "error del servidor", http.StatusInternalServerError)
        return
    }

    // Comparar contraseña con bcrypt
    if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
        http.Error(w, "credenciales inválidas", http.StatusUnauthorized)
        return
    }

    // Autenticación OK -> en producción: crear sesión o JWT
    fmt.Fprintf(w, "Bienvenido usuario %d", id)
}