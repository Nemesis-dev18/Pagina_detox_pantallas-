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

	if err := recargarFiltro(context.Background()); err != nil {
		log.Println("aviso: no se pudo cargar la lista de palabras excluidas (¿ya creaste la tabla?): ", err)
	}

	http.Handle("/static/", sinCache(http.StripPrefix("/static/", http.FileServer(http.Dir("static")))))
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

	// Administración de palabras excluidas (solo admins)
	http.HandleFunc("GET /admin/palabras", listarPalabras)
	http.HandleFunc("POST /admin/palabras", exigirCSRF(agregarPalabra))
	http.HandleFunc("DELETE /admin/palabras/{id}", exigirCSRF(borrarPalabra))
	http.HandleFunc("GET /admin/coincidencias", listarCoincidencias)

	fmt.Println("Servidor en http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", conCSRF(http.DefaultServeMux)))
}

// sinCache le dice al navegador que revalide los archivos estáticos en cada
// carga. FileServer ya manda Last-Modified, así que si el archivo no cambió
// el servidor responde 304 (sin reenviarlo) y si cambió, llega la versión nueva.
func sinCache(siguiente http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")
		siguiente.ServeHTTP(w, r)
	})
}
