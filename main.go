package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
)

func main() {
	cargarEnv(".env")

	if err := conectarDB(context.Background()); err != nil {
		log.Fatal("no se pudo conectar a la base: ", err)
	}
	defer pool.Close()
	fmt.Println("Conectado a Postgres")

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.Handle("/", http.RedirectHandler("/static/index.html", http.StatusFound))

	http.HandleFunc("GET /publicaciones", listarPublicaciones)
	http.HandleFunc("POST /publicaciones", exigirCSRF(crearPublicacion))
	http.HandleFunc("GET /publicaciones/{id}/comentarios", listarComentarios)
	http.HandleFunc("POST /publicaciones/{id}/comentarios", exigirCSRF(crearComentario))
	http.HandleFunc("DELETE /publicaciones/{id}", exigirCSRF(borrarPublicacion))
	http.HandleFunc("PUT /publicaciones/{id}", exigirCSRF(editarPublicacion))
	http.HandleFunc("PUT /publicaciones/{id}/comentarios/{comentarioId}", exigirCSRF(editarComentario))
	http.HandleFunc("DELETE /publicaciones/{id}/comentarios/{comentarioId}", exigirCSRF(borrarComentario))
	http.HandleFunc("/recuperar", exigirCSRF(recuperar))
	http.HandleFunc("/restablecer", exigirCSRF(restablecer))
	http.HandleFunc("/registro", exigirCSRF(registrar))
	http.HandleFunc("/login", exigirCSRF(iniciarSesion))
	http.HandleFunc("/logout", exigirCSRF(cerrarSesion))
	http.HandleFunc("/yo", yo)

	fmt.Println("Servidor en http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", conCSRF(http.DefaultServeMux)))
}