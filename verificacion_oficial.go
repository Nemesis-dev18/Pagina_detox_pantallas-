package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Fuentes OFICIALES abiertas (datos.gov.co, API Socrata, sin llave):
//   - REPS, Registro Especial de Prestadores y Sedes de Servicios de Salud (MinSalud): c36g-9fc2
//   - Instituciones de Educación Superior (MinEducación): n5yy-8nav
//
// Solo prueban que la INSTITUCIÓN existe. No prueban que la persona trabaje o
// estudie allí, ni verifican su credencial: eso lo decide una persona
// administradora (el ReTHUS abierto en datos.gov.co es agregado, sin personas).
var (
	urlREPS = "https://www.datos.gov.co/resource/c36g-9fc2.json"
	urlMEN  = "https://www.datos.gov.co/resource/n5yy-8nav.json"
)

// Consulta pública del ReTHUS (MinSalud/SISPRO). No tiene API: la revisa una persona.
const urlConsultaReTHUS = "https://web.sispro.gov.co/THS/Cliente/ConsultasPublicas/ConsultaPublicaDeTHxIdentificacion.aspx"

type candidatoOficial struct {
	Fuente      string `json:"fuente"`
	Nombre      string `json:"nombre"`
	NIT         string `json:"nit"`
	Detalle     string `json:"detalle"`
	CoincideNIT bool   `json:"coincide_nit"`
}

type evidenciaOficial struct {
	ConsultadoEn time.Time          `json:"consultado_en"`
	Aviso        string             `json:"aviso"`
	Candidatos   []candidatoOficial `json:"candidatos"`
	Errores      []string           `json:"errores,omitempty"`
}

func soloDigitos(s string) string {
	var b strings.Builder
	for _, c := range s {
		if c >= '0' && c <= '9' {
			b.WriteRune(c)
		}
	}
	return b.String()
}

// mismoNIT compara el NIT declarado (sin DV) con el del dataset, que puede traer
// puntos y dígito de verificación ("899.999.063-3") o no ("838000096").
func mismoNIT(declarado, oficial string) bool {
	d, o := soloDigitos(declarado), soloDigitos(oficial)
	if d == "" || o == "" {
		return false
	}
	return o == d || (len(o) > 1 && o[:len(o)-1] == d)
}

func consultarDataset(ctx context.Context, base string, params url.Values) ([]map[string]any, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := clienteHTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("datos.gov.co respondió %d", resp.StatusCode)
	}
	var filas []map[string]any
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&filas); err != nil {
		return nil, err
	}
	return filas, nil
}

func texto(fila map[string]any, clave string) string {
	s, _ := fila[clave].(string)
	return s
}

// consultarFuentesOficiales busca la institución declarada por nombre (y por NIT
// en el REPS, cuyo formato es solo dígitos) y devuelve candidatos como EVIDENCIA.
func consultarFuentesOficiales(ctx context.Context, nombre, nit string) evidenciaOficial {
	ev := evidenciaOficial{
		ConsultadoEn: time.Now().UTC(),
		Aviso:        "Estas fuentes solo muestran que la institución existe. No confirman que la persona pertenezca a ella: eso debe verificarse por separado.",
		Candidatos:   []candidatoOficial{},
	}
	vistos := map[string]bool{}
	agregar := func(c candidatoOficial) {
		clave := c.Fuente + "|" + c.NIT + "|" + c.Nombre
		if vistos[clave] {
			return
		}
		vistos[clave] = true
		c.CoincideNIT = nit != "" && mismoNIT(nit, c.NIT)
		ev.Candidatos = append(ev.Candidatos, c)
	}

	// REPS
	colsREPS := "nombreprestador,numeroidentificacion,claseprestador"
	consultasREPS := []url.Values{{"$select": {colsREPS}, "$group": {colsREPS}, "$q": {nombre}, "$limit": {"5"}}}
	if nit != "" { // nit ya viene validado: solo dígitos
		consultasREPS = append(consultasREPS, url.Values{"$select": {colsREPS}, "$group": {colsREPS},
			"$where": {"numeroidentificacion='" + nit + "'"}, "$limit": {"5"}})
	}
	for _, p := range consultasREPS {
		filas, err := consultarDataset(ctx, urlREPS, p)
		if err != nil {
			ev.Errores = append(ev.Errores, "REPS: "+err.Error())
			break
		}
		for _, f := range filas {
			agregar(candidatoOficial{Fuente: "REPS (MinSalud)", Nombre: texto(f, "nombreprestador"),
				NIT: texto(f, "numeroidentificacion"), Detalle: texto(f, "claseprestador")})
		}
	}

	// Educación superior (MinEducación)
	colsMEN := "nombre_instituci_n,n_mero_identificaci_n,estado,sector"
	filas, err := consultarDataset(ctx, urlMEN, url.Values{"$select": {colsMEN}, "$group": {colsMEN}, "$q": {nombre}, "$limit": {"5"}})
	if err != nil {
		ev.Errores = append(ev.Errores, "MinEducación: "+err.Error())
	}
	for _, f := range filas {
		agregar(candidatoOficial{Fuente: "Educación superior (MinEducación)", Nombre: texto(f, "nombre_instituci_n"),
			NIT: texto(f, "n_mero_identificaci_n"), Detalle: strings.TrimSpace(texto(f, "estado") + " · " + texto(f, "sector"))})
	}
	return ev
}
