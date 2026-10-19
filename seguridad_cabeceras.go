package main

import "net/http"

// conCabeceras añade cabeceras de seguridad a todas las respuestas.
// Son cabeceras que no rompen el sitio actual (no bloquean scripts ni estilos):
//   - X-Frame-Options / frame-ancestors: nadie puede meter la página en un iframe (clickjacking).
//   - nosniff: el navegador no adivina tipos de archivo.
//   - HSTS: solo cuando el sitio va por HTTPS (BASE_URL con https://).
//
// shortcut: falta una Content-Security-Policy completa; requiere quitar los <script> en línea
// (inicio_sesion.html y restablecer.html) y pasarlos a archivos .js. Luego subir a script-src 'self'.
func conCabeceras(siguiente http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		h.Set("Content-Security-Policy", "frame-ancestors 'none'; base-uri 'self'; object-src 'none'; form-action 'self'")
		if cookieSegura() {
			h.Set("Strict-Transport-Security", "max-age=31536000")
		}
		siguiente.ServeHTTP(w, r)
	})
}
