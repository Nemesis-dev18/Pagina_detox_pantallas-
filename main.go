package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	db := initDB()
	defer db.Close()

	mux := http.NewServeMux()
	// servir index y estáticos (asume carpeta static/)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/index.html")
	})
	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// handlers deben estar definidos en handlers.go
	mux.HandleFunc("/login", loginHandler)
	mux.HandleFunc("/register", registerHandler)
	mux.HandleFunc("/request-password-reset", requestResetHandler)
	mux.HandleFunc("/reset-password", resetPasswordHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}