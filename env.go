package main

import (
	"bufio"
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
