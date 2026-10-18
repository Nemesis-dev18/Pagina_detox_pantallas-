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

	// Migraciones seguras (solo agregan; ver migraciones/). Si fallan, no se arranca a medias.
	if err := aplicarMigraciones(context.Background()); err != nil {
		log.Fatal("no se pudieron aplicar las migraciones: ", err)
	}
	if _, err := claveDatos(); err != nil {
		log.Println("ATENCION: ", err, "— el registro de especialistas queda desactivado")
	}

	if err := recargarFiltro(context.Background()); err != nil {
		log.Println("no se pudo cargar el filtro de palabras: ", err)
	}

	// En producción (Render) el puerto lo asigna la plataforma con la variable PORT.
	// En tu computador, si no existe, se usa el 8080.
	puerto := os.Getenv("PORT")
	if puerto == "" {
		puerto = "8080"
	}

	fmt.Println("Servidor escuchando en el puerto " + puerto)
	log.Fatal(http.ListenAndServe(":"+puerto, rutas()))
}

// rutas arma todas las rutas del sitio (antes estaban dentro de main; así las
// pruebas pueden usar exactamente las mismas).
func rutas() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.Handle("/", http.RedirectHandler("/static/index.html", http.StatusFound))

	mux.HandleFunc("GET /robots.txt", robotsTxt)
	mux.HandleFunc("GET /sitemap.xml", sitemapXML)

	mux.HandleFunc("GET /publicaciones", listarPublicaciones)
	mux.HandleFunc("POST /publicaciones", exigirCSRF(crearPublicacion))
	mux.HandleFunc("GET /publicaciones/{id}/comentarios", listarComentarios)
	mux.HandleFunc("POST /publicaciones/{id}/comentarios", exigirCSRF(crearComentario))
	mux.HandleFunc("DELETE /publicaciones/{id}", exigirCSRF(borrarPublicacion))
	mux.HandleFunc("PUT /publicaciones/{id}", exigirCSRF(editarPublicacion))
	mux.HandleFunc("PUT /publicaciones/{id}/comentarios/{comentarioId}", exigirCSRF(editarComentario))
	mux.HandleFunc("DELETE /publicaciones/{id}/comentarios/{comentarioId}", exigirCSRF(borrarComentario))
	mux.HandleFunc("/recuperar", exigirCSRF(recuperar))
	mux.HandleFunc("/restablecer", exigirCSRF(restablecer))
	mux.HandleFunc("/registro", exigirCSRF(registrar))
	mux.HandleFunc("/login", exigirCSRF(iniciarSesion))
	mux.HandleFunc("/logout", exigirCSRF(cerrarSesion))
	mux.HandleFunc("/yo", yo)

	mux.HandleFunc("GET /admin/palabras", listarPalabras)
	mux.HandleFunc("POST /admin/palabras", exigirCSRF(agregarPalabra))
	mux.HandleFunc("DELETE /admin/palabras/{id}", exigirCSRF(borrarPalabra))
	mux.HandleFunc("GET /admin/coincidencias", listarCoincidencias)

	// Especialistas verificados
	mux.HandleFunc("/registro-especialista", exigirCSRF(registrarEspecialista))
	mux.HandleFunc("GET /yo/especialista", miSolicitudEspecialista)
	mux.HandleFunc("GET /especialistas", listarEspecialistas)
	mux.HandleFunc("GET /especialistas/{id}", verEspecialista)

	// Revisión de solicitudes (solo administradores)
	mux.HandleFunc("GET /admin/especialistas", listarSolicitudesEspecialistas)
	mux.HandleFunc("GET /admin/especialistas/{id}", verSolicitudEspecialista)
	mux.HandleFunc("POST /admin/especialistas/{id}/consultar-institucion", exigirCSRF(consultarInstitucionEspecialista))
	mux.HandleFunc("POST /admin/especialistas/{id}/revisar", exigirCSRF(revisarEspecialista))

	// Acompañamiento voluntario
	mux.HandleFunc("POST /acompanamientos", exigirCSRF(solicitarAcompanamiento))
	mux.HandleFunc("GET /acompanamientos", listarAcompanamientos)
	mux.HandleFunc("POST /acompanamientos/{id}/responder", exigirCSRF(responderAcompanamiento))
	mux.HandleFunc("POST /acompanamientos/{id}/cerrar", exigirCSRF(cerrarAcompanamiento))
	mux.HandleFunc("GET /acompanamientos/{id}/mensajes", listarMensajesPrivados)
	mux.HandleFunc("POST /acompanamientos/{id}/mensajes", exigirCSRF(enviarMensajePrivado))

	return conCSRF(mux)
}
