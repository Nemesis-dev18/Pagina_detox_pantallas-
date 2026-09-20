package main

import (
    "database/sql"
    "fmt"
    "net/http"
    "strings"

    "golang.org/x/crypto/bcrypt"
    _ "modernc.org/sqlite"
)

var db *sql.DB

func iniciarBaseDatos() error {
    var err error
    db, err = sql.Open("sqlite", "usuarios.db")
    if err != nil {
        return err
    }

    _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS usuarios (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            nombre TEXT NOT NULL,
            correo TEXT NOT NULL UNIQUE,
            password_hash TEXT NOT NULL,
            creado_en DATETIME DEFAULT CURRENT_TIMESTAMP
        )
    `)
    return err
}

func registro(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
        return
    }

    nombre := r.FormValue("nombre")
    correo := strings.ToLower(strings.TrimSpace(r.FormValue("correo")))
    password := r.FormValue("password")
    confirmacion := r.FormValue("confirmar_password")

    if nombre == "" || correo == "" || password == "" || confirmacion == "" {
        http.Redirect(w, r, "/registro.html?error=campos", http.StatusSeeOther)
        return
    }
    if password != confirmacion {
        http.Redirect(w, r, "/registro.html?error=confirmacion", http.StatusSeeOther)
        return
    }

    var existe bool
    err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM usuarios WHERE correo = ?)", correo).Scan(&existe)
    if err != nil {
        http.Error(w, "No se pudo consultar la base de datos", http.StatusInternalServerError)
        return
    }
    if existe {
        http.Redirect(w, r, "/registro.html?error=correo", http.StatusSeeOther)
        return
    }

    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        http.Error(w, "No se pudo proteger la contraseña", http.StatusInternalServerError)
        return
    }

    _, err = db.Exec("INSERT INTO usuarios (nombre, correo, password_hash) VALUES (?, ?, ?)", nombre, correo, string(hash))
    if err != nil {
        http.Error(w, "No se pudo guardar el usuario", http.StatusInternalServerError)
        return
    }

    http.Redirect(w, r, "/index.html?registrado=true", http.StatusSeeOther)
}

func inicioSesion(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
        return
    }

    correo := strings.ToLower(strings.TrimSpace(r.FormValue("username")))
    password := r.FormValue("password")

    if correo == "" || password == "" {
        http.Redirect(w, r, "/inicio_sesion.html?error=campos", http.StatusSeeOther)
        return
    }

    var passwordHash string
    err := db.QueryRow("SELECT password_hash FROM usuarios WHERE correo = ?", correo).Scan(&passwordHash)
    if err == sql.ErrNoRows || bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)) != nil {
        http.Redirect(w, r, "/inicio_sesion.html?error=credenciales", http.StatusSeeOther)
        return
    }
    if err != nil {
        http.Error(w, "No se pudo consultar la base de datos", http.StatusInternalServerError)
        return
    }

    http.Redirect(w, r, "/index.html?sesion=iniciada", http.StatusSeeOther)
}

func main() {
    if err := iniciarBaseDatos(); err != nil {
        panic(err)
    }
    defer db.Close()

    http.HandleFunc("/registro", registro)
    http.HandleFunc("/inicio_sesion", inicioSesion)

    // Sirve tus HTML, CSS, imágenes y JavaScript.
    archivos := http.FileServer(http.Dir("."))
    http.Handle("/", archivos)

    fmt.Println("Servidor en http://localhost:8080")
    http.ListenAndServe(":8080", nil)
}