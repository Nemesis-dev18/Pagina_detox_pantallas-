package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
)

func main() {
	if err := conectarDB(context.Background()); err != nil {
		log.Fatal("no se pudo conectar a la base: ", err)
	}
	defer pool.Close()
	fmt.Println("Conectado a Postgres")

	// Tu página: la carpeta static se entrega en /static/
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.Handle("/", http.RedirectHandler("/static/index.html", http.StatusFound))

	// Tu API
http.HandleFunc("GET /publicaciones", listarPublicaciones)
http.HandleFunc("POST /publicaciones", crearPublicacion)
	http.HandleFunc("/recuperar", recuperar)
    http.HandleFunc("/restablecer", restablecer)
    http.HandleFunc("/registro", registrar)
	http.HandleFunc("/login", iniciarSesion)
    http.HandleFunc("/logout", cerrarSesion)
    http.HandleFunc("/yo", yo)
	
	fmt.Println("Servidor en http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}