package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Pruebas de las piezas críticas de seguridad. No necesitan base de datos.
// Ejecutar con: go test ./...

func TestCifradoIdaYVuelta(t *testing.T) {
	t.Setenv("DATOS_CLAVE", strings.Repeat("ab", 32))
	cifrado, err := cifrar("CC:12345678")
	if err != nil {
		t.Fatal(err)
	}
	plano, err := descifrar(cifrado)
	if err != nil || plano != "CC:12345678" {
		t.Fatalf("descifrar = %q, %v; se esperaba CC:12345678", plano, err)
	}
}

func TestExigirCSRF(t *testing.T) {
	ok := exigirCSRF(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) })
	casos := []struct {
		nombre     string
		cookie     string
		encabezado string
		esperado   int
	}{
		{"sin cookie", "", "x", http.StatusForbidden},
		{"sin encabezado", "tok", "", http.StatusForbidden},
		{"encabezado distinto", "tok", "otro", http.StatusForbidden},
		{"correcto", "tok", "tok", http.StatusNoContent},
	}
	for _, c := range casos {
		r := httptest.NewRequest(http.MethodPost, "/x", nil)
		if c.cookie != "" {
			r.AddCookie(&http.Cookie{Name: cookieCSRF, Value: c.cookie})
		}
		if c.encabezado != "" {
			r.Header.Set("X-CSRF-Token", c.encabezado)
		}
		w := httptest.NewRecorder()
		ok(w, r)
		if w.Code != c.esperado {
			t.Errorf("%s: código %d, se esperaba %d", c.nombre, w.Code, c.esperado)
		}
	}
}

func TestMismoNIT(t *testing.T) {
	if !mismoNIT("899999063", "899.999.063-3") {
		t.Error("debería coincidir ignorando puntos y dígito de verificación")
	}
	if mismoNIT("800123456", "899.999.063-3") {
		t.Error("no debería coincidir con otro NIT")
	}
	if mismoNIT("", "123") {
		t.Error("un NIT vacío nunca coincide")
	}
}

func TestNormalizarNIT(t *testing.T) {
	if got := normalizarNIT("800.123.456-7"); got != "800123456" {
		t.Errorf("normalizarNIT = %q", got)
	}
	if got := normalizarNIT("abc"); got != "" {
		t.Errorf("letras no deberían ser un NIT, salió %q", got)
	}
}

func TestEscaparLike(t *testing.T) {
	if got := escaparLike("50%_"); got != `50\%\_` {
		t.Errorf("escaparLike = %q", got)
	}
}

func TestBloqueoTrasVariosFallos(t *testing.T) {
	const clave = "prueba-bloqueo@ejemplo.com"
	defer limpiarIntentos(clave)
	for i := 0; i < intentosMaximos; i++ {
		registrarFallo(clave)
	}
	if bloq, _ := bloqueado(clave); !bloq {
		t.Error("después de los intentos máximos la cuenta debería estar bloqueada")
	}
}

func TestCabecerasDeSeguridad(t *testing.T) {
	h := conCabeceras(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
	if w.Header().Get("X-Frame-Options") != "DENY" {
		t.Error("falta X-Frame-Options: DENY")
	}
	if w.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Error("falta X-Content-Type-Options: nosniff")
	}
}
