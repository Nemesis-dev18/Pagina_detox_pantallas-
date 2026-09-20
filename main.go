package main

import (
    "fmt"
    "net/http"
)

func registro(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
        return
    }

    nombre := r.FormValue("nombre")
    correo := r.FormValue("correo")
    password := r.FormValue("password")

    if nombre == "" || correo == "" || password == "" {
        http.Error(w, "Todos los campos son obligatorios", http.StatusBadRequest)
        return
    }

    fmt.Printf("Registro: %s - %s\n", nombre, correo)

    http.Redirect(w, r, "/registro.html?registrado=true", http.StatusSeeOther)
}

func main() {
    http.HandleFunc("/registro", registro)

    // Sirve tus HTML, CSS, imágenes y JavaScript.
    archivos := http.FileServer(http.Dir("."))
    http.Handle("/", archivos)

    fmt.Println("Servidor en http://localhost:8080")
    http.ListenAndServe(":8080", nil)
}