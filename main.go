package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	cargarEnv(".env")
	avisarBaseDeDatos()

	if err := conectarDB(context.Background()); err != nil {
		log.Fatal("no se pudo conectar a la base: ", err)
	}
	defer pool.Close()
	fmt.Println("Conectado a Postgres")

	if err := recargarFiltro(context.Background()); err != nil {
		log.Println("no se pudo cargar el filtro de palabras: ", err)
	}

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.Handle("/", http.RedirectHandler("/static/index.html", http.StatusFound))

	http.HandleFunc("GET /robots.txt", robotsTxt)
	http.HandleFunc("GET /sitemap.xml", sitemapXML)

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

	http.HandleFunc("GET /admin/palabras", listarPalabras)
	http.HandleFunc("POST /admin/palabras", exigirCSRF(agregarPalabra))
	http.HandleFunc("DELETE /admin/palabras/{id}", exigirCSRF(borrarPalabra))
	http.HandleFunc("GET /admin/coincidencias", listarCoincidencias)

	// En producción (Render) el puerto lo asigna la plataforma con la variable PORT.
	// En tu computador, si no existe, se usa el 8080.
	puerto := os.Getenv("PORT")
	if puerto == "" {
		puerto = "8080"
	}

	fmt.Println("Servidor escuchando en el puerto " + puerto)
	log.Fatal(http.ListenAndServe(":"+puerto, conCSRF(http.DefaultServeMux)))
}
