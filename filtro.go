package main

import (
	"context"
	"regexp"
	"sort"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"
)

// Lo que se muestra en lugar de una palabra excluida. Siempre el mismo largo,
// para no regalar pistas sobre cuál era la palabra.
const mascara = "#####"

// Cada letra puede venir "disfrazada" con estos símbolos (leetspeak).
var equivalentes = map[rune]string{
	'a': "a4@",
	'e': "e3",
	'i': "i1!|",
	'l': "l1|",
	'o': "o0",
	's': "s5$",
	't': "t7",
	'g': "g9",
	'b': "b8",
	'z': "z2",
}

// Acentos y letras de otros alfabetos que se ven igual que una letra latina.
var plegar = func() map[rune]rune {
	m := map[rune]rune{}
	for base, variantes := range map[rune]string{
		'a': "áàäâã",
		'e': "éèëê",
		'i': "íìïî",
		'o': "óòöôõ",
		'u': "úùüû",
		'n': "ñ",
		'c': "ç",
		// Cirílico que imita letras latinas
		'p': "р",
		'x': "х",
		'y': "у",
	} {
		for _, v := range variantes {
			m[v] = base
		}
	}
	m['а'] = 'a'
	m['е'] = 'e'
	m['о'] = 'o'
	m['с'] = 'c'
	m['і'] = 'i'
	return m
}()

type palabraFiltro struct {
	Palabra string
	re      *regexp.Regexp
}

var (
	filtroMu       sync.RWMutex
	filtroPalabras []palabraFiltro
)

func esAlnum(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}

// aplanar pasa el texto a minúsculas, quita acentos y descarta todo lo que
// separa letras (espacios, puntos, guiones, caracteres invisibles...).
// Devuelve el texto original en runas, el texto "plano" y, para cada runa
// plana, su posición en el original (para saber qué tapar después).
func aplanar(texto string) (orig []rune, plano []rune, idx []int) {
	orig = []rune(texto)
	for i, r := range orig {
		r = unicode.ToLower(r)
		if p, ok := plegar[r]; ok {
			r = p
		}
		if esAlnum(r) || strings.ContainsRune("@$!|", r) {
			plano = append(plano, r)
			idx = append(idx, i)
		}
	}
	return orig, plano, idx
}

// compilarPalabra convierte una palabra excluida en una expresión regular que
// tolera: mayúsculas, acentos, letras repetidas (puuuta), separadores
// (p.u.t.a / p u t a), números y símbolos parecidos (pu7@) y plural (s / es).
// Devuelve nil si la palabra no sirve (menos de 3 caracteres, por ejemplo).
func compilarPalabra(palabra string) *regexp.Regexp {
	_, plano, _ := aplanar(palabra)
	if len(plano) < 3 {
		return nil
	}
	var b strings.Builder
	b.WriteString("(?:")
	for _, r := range plano {
		if eq, ok := equivalentes[r]; ok {
			b.WriteString("[" + eq + "]+")
		} else {
			b.WriteString(regexp.QuoteMeta(string(r)) + "+")
		}
	}
	b.WriteString(")(?:e?[s5$])?")
	re, err := regexp.Compile(b.String())
	if err != nil {
		return nil
	}
	return re
}

// recargarFiltro lee la lista desde la base y reemplaza la que está en memoria.
// Se llama al arrancar y cada vez que el admin agrega o quita una palabra.
func recargarFiltro(ctx context.Context) error {
	rows, err := pool.Query(ctx, "SELECT palabra FROM palabras_excluidas")
	if err != nil {
		return err
	}
	defer rows.Close()

	nuevas := []palabraFiltro{}
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return err
		}
		if re := compilarPalabra(p); re != nil {
			nuevas = append(nuevas, palabraFiltro{Palabra: p, re: re})
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	filtroMu.Lock()
	filtroPalabras = nuevas
	filtroMu.Unlock()
	return nil
}

// coincidencia: Inicio y Fin son posiciones (en runas) dentro del texto
// original; Fin no está incluido.
type coincidencia struct {
	Palabra string
	Inicio  int
	Fin     int
}

func buscarCoincidencias(texto string) []coincidencia {
	filtroMu.RLock()
	palabras := filtroPalabras
	filtroMu.RUnlock()
	if len(palabras) == 0 {
		return nil
	}

	orig, plano, idx := aplanar(texto)
	if len(plano) == 0 {
		return nil
	}
	s := string(plano)

	var out []coincidencia
	for _, p := range palabras {
		for _, m := range p.re.FindAllStringIndex(s, -1) {
			a := utf8.RuneCountInString(s[:m[0]])
			b := utf8.RuneCountInString(s[:m[1]])
			ini, fin := idx[a], idx[b-1]+1

			// Solo palabras completas: "computadora" no debe disparar "puta".
			if ini > 0 && esAlnum(orig[ini-1]) {
				continue
			}
			if fin < len(orig) && esAlnum(orig[fin]) {
				continue
			}
			out = append(out, coincidencia{Palabra: p.Palabra, Inicio: ini, Fin: fin})
		}
	}
	return out
}

// censurar devuelve el texto con cada palabra excluida cambiada por #####.
func censurar(texto string) string {
	ms := buscarCoincidencias(texto)
	if len(ms) == 0 {
		return texto
	}
	sort.Slice(ms, func(i, j int) bool { return ms[i].Inicio < ms[j].Inicio })

	orig := []rune(texto)
	var b strings.Builder
	pos := 0
	for i := 0; i < len(ms); {
		ini, fin := ms[i].Inicio, ms[i].Fin
		i++
		// Une coincidencias que se pisan para no tapar dos veces lo mismo
		for i < len(ms) && ms[i].Inicio <= fin {
			if ms[i].Fin > fin {
				fin = ms[i].Fin
			}
			i++
		}
		b.WriteString(string(orig[pos:ini]))
		b.WriteString(mascara)
		pos = fin
	}
	b.WriteString(string(orig[pos:]))
	return b.String()
}
