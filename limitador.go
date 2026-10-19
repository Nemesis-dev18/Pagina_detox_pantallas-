package main

import (
	"sync"
	"time"
)

const (
	intentosMaximos = 4
	bloqueoDuracion = 15 * time.Minute
)

type intentoLogin struct {
	fallos         int
	bloqueadoHasta time.Time
}

var (
	intentosMu sync.Mutex
	intentos   = map[string]*intentoLogin{}
)

// bloqueado dice si esta clave (el correo) está bloqueada por demasiados
// intentos fallidos recientes, y cuánto tiempo falta si lo está.
func bloqueado(clave string) (bool, time.Duration) {
	intentosMu.Lock()
	defer intentosMu.Unlock()

	i, existe := intentos[clave]
	if !existe {
		return false, 0
	}
	restante := time.Until(i.bloqueadoHasta)
	if restante <= 0 {
		return false, 0
	}
	return true, restante
}

func registrarFallo(clave string) {
	intentosMu.Lock()
	defer intentosMu.Unlock()

	// shortcut: tope de memoria simple (se reinicia el conteo si hay muchos correos distintos);
	// subir a un almacén compartido (Redis o la base) si hay varias instancias.
	if len(intentos) >= 50000 {
		intentos = map[string]*intentoLogin{}
	}

	i, existe := intentos[clave]
	if !existe {
		i = &intentoLogin{}
		intentos[clave] = i
	}
	i.fallos++
	if i.fallos >= intentosMaximos {
		i.bloqueadoHasta = time.Now().Add(bloqueoDuracion)
		i.fallos = 0
	}
}

func limpiarIntentos(clave string) {
	intentosMu.Lock()
	defer intentosMu.Unlock()
	delete(intentos, clave)
}
