package main

import (
	"bufio"
	"log"
	"net/url"
	"os"
	"strings"
)

// cargarEnv lee un archivo .env (CLAVE=valor) y define esas variables de entorno.
// Si una variable ya existe en el sistema, NO la pisa (así en producción manda el servidor).
func cargarEnv(ruta string) {
	f, err := os.Open(ruta)
	if err != nil {
		return // si no hay .env, se usan las variables del sistema
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		linea := strings.TrimSpace(sc.Text())
		if linea == "" || strings.HasPrefix(linea, "#") {
			continue
		}
		linea = strings.TrimPrefix(linea, "export ")
		clave, valor, ok := strings.Cut(linea, "=")
		if !ok {
			continue
		}
		clave = strings.TrimSpace(clave)
		valor = strings.TrimSpace(valor)
		if len(valor) >= 2 {
			p, u := valor[0], valor[len(valor)-1]
			if (p == '"' && u == '"') || (p == '\'' && u == '\'') {
				valor = valor[1 : len(valor)-1]
			}
		}
		if _, existe := os.LookupEnv(clave); !existe {
			os.Setenv(clave, valor)
		}
	}
}

// avisarBaseDeDatos escribe en el log a qué host y usuario se va a conectar la app.
// Nunca imprime la contraseña. Sirve para confirmar en Render que usa Supabase.
func avisarBaseDeDatos() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Println("ATENCION: DATABASE_URL está vacía")
		return
	}
	u, err := url.Parse(dsn)
	if err != nil {
		log.Println("ATENCION: DATABASE_URL no se pudo interpretar")
		return
	}
	log.Printf("conectando a host=%s usuario=%s", u.Host, u.User.Username())
}
