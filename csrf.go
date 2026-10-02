package main

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
)

const cookieCSRF = "csrf_token"

func generarTokenCSRF() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// conCSRF envuelve TODO el servidor: se asegura de que cada visitante tenga
// la cookie csrf_token. A diferencia de la cookie "sesion", esta NO es
// HttpOnly, porque el JavaScript necesita leerla para reenviarla en el
// encabezado X-CSRF-Token de cada petición que modifica datos.
func conCSRF(siguiente http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := r.Cookie(cookieCSRF); err != nil {
			if token, genErr := generarTokenCSRF(); genErr == nil {
				http.SetCookie(w, &http.Cookie{
					Name:     cookieCSRF,
					Value:    token,
					Path:     "/",
					SameSite: http.SameSiteLaxMode,
					// Secure: true, // activar cuando el sitio use HTTPS
				})
			}
		}
		siguiente.ServeHTTP(w, r)
	})
}

// exigirCSRF protege un handler que modifica datos (POST/PUT/DELETE).
// Un sitio externo puede lograr que el navegador mande la cookie de sesión,
// pero no puede leer la cookie csrf_token ni agregar este encabezado — así
// que una petición forjada desde otra página nunca pasa esta comprobación.
func exigirCSRF(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie(cookieCSRF)
		if err != nil {
			http.Error(w, "falta el token de seguridad, recarga la página", http.StatusForbidden)
			return
		}
		encabezado := r.Header.Get("X-CSRF-Token")
		if encabezado == "" || subtle.ConstantTimeCompare([]byte(encabezado), []byte(c.Value)) != 1 {
			http.Error(w, "token de seguridad inválido, recarga la página", http.StatusForbidden)
			return
		}
		h(w, r)
	}
}