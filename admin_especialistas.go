package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
)

var estadosVerificacion = map[string]bool{"pendiente": true, "verificada": true, "rechazada": true}

type solicitudAdmin struct {
	ID                string    `json:"id"`
	Alias             string    `json:"alias"`
	Nombre            string    `json:"nombre"`
	Profesion         string    `json:"profesion"`
	Especialidad      string    `json:"especialidad"`
	Institucion       string    `json:"institucion"`
	Estado            string    `json:"estado"`
	CredencialEstado  string    `json:"credencial_estado"`
	InstitucionEstado string    `json:"institucion_estado"`
	CreadoEn          time.Time `json:"creado_en"`
	// Solo en el detalle:
	Correo         string          `json:"correo,omitempty"`
	InstitucionNIT string          `json:"institucion_nit,omitempty"`
	TipoDocumento  string          `json:"tipo_documento,omitempty"`
	Documento      string          `json:"documento,omitempty"`
	NumeroRegistro string          `json:"numero_registro,omitempty"`
	Descripcion    string          `json:"descripcion,omitempty"`
	Notas          string          `json:"notas_revision,omitempty"`
	Motivo         string          `json:"motivo,omitempty"`
	Evidencia      json.RawMessage `json:"evidencia_institucion,omitempty"`
	RevisadoEn     *time.Time      `json:"revisado_en,omitempty"`
	URLReTHUS      string          `json:"url_rethus,omitempty"`
}

// GET /admin/especialistas?estado=pendiente
func listarSolicitudesEspecialistas(w http.ResponseWriter, r *http.Request) {
	if _, ok := exigirAdmin(w, r); !ok {
		return
	}
	estado := r.URL.Query().Get("estado")
	if estado != "" && estado != "pendiente" && estado != "aprobado" && estado != "rechazado" && estado != "suspendido" {
		http.Error(w, "estado no válido", http.StatusBadRequest)
		return
	}
	rows, err := pool.Query(r.Context(), `
		SELECT pe.usuario_id, u.alias, concat_ws(' ', pe.nombres, pe.primer_apellido, pe.segundo_apellido),
		       pe.profesion, pe.especialidad, pe.institucion, pe.estado, pe.credencial_estado, pe.institucion_estado, pe.creado_en
		FROM perfiles_especialistas pe JOIN usuarios u ON u.id = pe.usuario_id
		WHERE ($1 = '' OR pe.estado = $1)
		ORDER BY (pe.estado = 'pendiente') DESC, pe.creado_en DESC
		LIMIT 200`, estado)
	if err != nil {
		http.Error(w, "error consultando la base", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	lista := []solicitudAdmin{}
	for rows.Next() {
		var s solicitudAdmin
		var clave string
		if err := rows.Scan(&s.ID, &s.Alias, &s.Nombre, &clave, &s.Especialidad, &s.Institucion,
			&s.Estado, &s.CredencialEstado, &s.InstitucionEstado, &s.CreadoEn); err != nil {
			http.Error(w, "error leyendo datos", http.StatusInternalServerError)
			return
		}
		s.Profesion = profesiones[clave]
		lista = append(lista, s)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, "error leyendo datos", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "private, no-store")
	json.NewEncoder(w).Encode(lista)
}

// GET /admin/especialistas/{id}: detalle con el documento descifrado.
// Se registra en el log QUIÉN lo consultó (nunca el número).
func verSolicitudEspecialista(w http.ResponseWriter, r *http.Request) {
	admin, ok := exigirAdmin(w, r)
	if !ok {
		return
	}
	id := r.PathValue("id")
	if !reUUID.MatchString(id) {
		http.Error(w, "solicitud no encontrada", http.StatusNotFound)
		return
	}
	var s solicitudAdmin
	var clave string
	var cifrado []byte
	var evidencia []byte
	err := pool.QueryRow(r.Context(), `
		SELECT pe.usuario_id, u.alias, u.correo, concat_ws(' ', pe.nombres, pe.primer_apellido, pe.segundo_apellido),
		       pe.profesion, pe.especialidad, pe.institucion, COALESCE(pe.institucion_nit, ''), pe.tipo_documento,
		       pe.documento_cifrado, pe.numero_registro, COALESCE(pe.descripcion, ''),
		       pe.estado, pe.credencial_estado, pe.institucion_estado, COALESCE(pe.notas_revision, ''),
		       COALESCE(pe.motivo, ''), pe.evidencia_institucion, pe.creado_en, pe.revisado_en
		FROM perfiles_especialistas pe JOIN usuarios u ON u.id = pe.usuario_id
		WHERE pe.usuario_id = $1`, id).
		Scan(&s.ID, &s.Alias, &s.Correo, &s.Nombre, &clave, &s.Especialidad, &s.Institucion, &s.InstitucionNIT,
			&s.TipoDocumento, &cifrado, &s.NumeroRegistro, &s.Descripcion, &s.Estado, &s.CredencialEstado,
			&s.InstitucionEstado, &s.Notas, &s.Motivo, &evidencia, &s.CreadoEn, &s.RevisadoEn)
	if errors.Is(err, pgx.ErrNoRows) {
		http.Error(w, "solicitud no encontrada", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "error consultando la base", http.StatusInternalServerError)
		return
	}
	s.Profesion = profesiones[clave]
	plano, err := descifrar(cifrado)
	if err != nil {
		log.Println("no se pudo descifrar el documento:", err)
		http.Error(w, "no se pudo leer el documento (¿falta DATOS_CLAVE?)", http.StatusInternalServerError)
		return
	}
	// plano = "CC:123456"; el tipo ya está en su columna
	if i := strings.Index(plano, ":"); i >= 0 {
		s.Documento = plano[i+1:]
	}
	if len(evidencia) > 0 {
		s.Evidencia = evidencia
	}
	s.URLReTHUS = urlConsultaReTHUS
	log.Printf("admin %s consultó el detalle del especialista %s", admin.ID, id)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "private, no-store")
	json.NewEncoder(w).Encode(s)
}

// POST /admin/especialistas/{id}/consultar-institucion
// Consulta datos.gov.co y guarda el resultado como EVIDENCIA. No cambia ningún estado.
func consultarInstitucionEspecialista(w http.ResponseWriter, r *http.Request) {
	if _, ok := exigirAdmin(w, r); !ok {
		return
	}
	id := r.PathValue("id")
	if !reUUID.MatchString(id) {
		http.Error(w, "solicitud no encontrada", http.StatusNotFound)
		return
	}
	var nombre, nit string
	err := pool.QueryRow(r.Context(),
		"SELECT institucion, COALESCE(institucion_nit, '') FROM perfiles_especialistas WHERE usuario_id = $1", id).Scan(&nombre, &nit)
	if errors.Is(err, pgx.ErrNoRows) {
		http.Error(w, "solicitud no encontrada", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "error consultando la base", http.StatusInternalServerError)
		return
	}
	ev := consultarFuentesOficiales(r.Context(), nombre, nit)
	bytes, _ := json.Marshal(ev)
	if _, err := pool.Exec(r.Context(),
		"UPDATE perfiles_especialistas SET evidencia_institucion = $1 WHERE usuario_id = $2", bytes, id); err != nil {
		http.Error(w, "error guardando la evidencia", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}

// POST /admin/especialistas/{id}/revisar
// Campos: credencial, institucion (pendiente|verificada|rechazada), notas, motivo,
// decision (aprobar|rechazar|suspender|reabrir). Toda la lógica de permisos y de
// "cuándo se puede aprobar" vive aquí, en el servidor.
func revisarEspecialista(w http.ResponseWriter, r *http.Request) {
	admin, ok := exigirAdmin(w, r)
	if !ok {
		return
	}
	id := r.PathValue("id")
	if !reUUID.MatchString(id) {
		http.Error(w, "solicitud no encontrada", http.StatusNotFound)
		return
	}
	if id == admin.ID {
		http.Error(w, "no puedes revisar tu propia solicitud", http.StatusForbidden)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<16)

	credencial := r.FormValue("credencial")
	institucion := r.FormValue("institucion")
	decision := r.FormValue("decision")
	notas := strings.TrimSpace(r.FormValue("notas"))
	motivo := strings.TrimSpace(r.FormValue("motivo"))
	_, enviaNotas := r.PostForm["notas"]

	if (credencial != "" && !estadosVerificacion[credencial]) || (institucion != "" && !estadosVerificacion[institucion]) {
		http.Error(w, "estado de verificación no válido", http.StatusBadRequest)
		return
	}
	switch decision {
	case "", "aprobar", "rechazar", "suspender", "reabrir":
	default:
		http.Error(w, "decisión no válida", http.StatusBadRequest)
		return
	}
	if utf8.RuneCountInString(notas) > 1000 || utf8.RuneCountInString(motivo) > 500 {
		http.Error(w, "las notas (máx. 1000) o el motivo (máx. 500) son muy largos", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	tx, err := pool.Begin(ctx)
	if err != nil {
		http.Error(w, "error interno", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(ctx)

	var estado, credActual, instActual, correo string
	err = tx.QueryRow(ctx, `
		SELECT pe.estado, pe.credencial_estado, pe.institucion_estado, u.correo
		FROM perfiles_especialistas pe JOIN usuarios u ON u.id = pe.usuario_id
		WHERE pe.usuario_id = $1 FOR UPDATE OF pe`, id).Scan(&estado, &credActual, &instActual, &correo)
	if errors.Is(err, pgx.ErrNoRows) {
		http.Error(w, "solicitud no encontrada", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "error interno", http.StatusInternalServerError)
		return
	}

	nuevaCred, nuevaInst := credActual, instActual
	if credencial != "" {
		nuevaCred = credencial
	}
	if institucion != "" {
		nuevaInst = institucion
	}
	if estado == "aprobado" && (nuevaCred != credActual || nuevaInst != instActual) {
		http.Error(w, "el especialista está aprobado: suspéndelo antes de cambiar las verificaciones", http.StatusConflict)
		return
	}

	nuevoEstado := estado
	fijarMotivo := false
	switch decision {
	case "aprobar":
		if estado == "aprobado" {
			http.Error(w, "ya está aprobado", http.StatusConflict)
			return
		}
		if nuevaCred != "verificada" || nuevaInst != "verificada" {
			http.Error(w, "para aprobar, la credencial y la institución deben estar verificadas", http.StatusConflict)
			return
		}
		nuevoEstado, motivo, fijarMotivo = "aprobado", "", true
	case "rechazar":
		if estado != "pendiente" {
			http.Error(w, "solo se puede rechazar una solicitud pendiente", http.StatusConflict)
			return
		}
		if motivo == "" {
			http.Error(w, "escribe el motivo del rechazo", http.StatusBadRequest)
			return
		}
		nuevoEstado, fijarMotivo = "rechazado", true
	case "suspender":
		if estado != "aprobado" {
			http.Error(w, "solo se puede suspender a un especialista aprobado", http.StatusConflict)
			return
		}
		if motivo == "" {
			http.Error(w, "escribe el motivo de la suspensión", http.StatusBadRequest)
			return
		}
		nuevoEstado, fijarMotivo = "suspendido", true
	case "reabrir":
		if estado != "rechazado" && estado != "suspendido" {
			http.Error(w, "solo se reabre una solicitud rechazada o suspendida", http.StatusConflict)
			return
		}
		nuevoEstado, motivo, fijarMotivo = "pendiente", "", true
	}

	_, err = tx.Exec(ctx, `
		UPDATE perfiles_especialistas SET
		  credencial_estado = $2, institucion_estado = $3, estado = $4,
		  notas_revision = CASE WHEN $5 THEN NULLIF($6, '') ELSE notas_revision END,
		  motivo = CASE WHEN $7 THEN NULLIF($8, '') ELSE motivo END,
		  revisado_por = $9, revisado_en = now()
		WHERE usuario_id = $1`,
		id, nuevaCred, nuevaInst, nuevoEstado, enviaNotas, notas, fijarMotivo, motivo, admin.ID)
	if err != nil {
		log.Println("revisar especialista:", err)
		http.Error(w, "error guardando la revisión", http.StatusInternalServerError)
		return
	}
	if err := tx.Commit(ctx); err != nil {
		http.Error(w, "error interno", http.StatusInternalServerError)
		return
	}
	log.Printf("admin %s revisó al especialista %s: estado=%s credencial=%s institucion=%s", admin.ID, id, nuevoEstado, nuevaCred, nuevaInst)

	if decision == "aprobar" || decision == "rechazar" || decision == "suspender" {
		go avisarDecision(correo, decision, motivo)
	}
	w.WriteHeader(http.StatusNoContent)
}

// avisarDecision manda un correo con el resultado; si falla solo se anota en el log.
func avisarDecision(correo, decision, motivo string) {
	var asunto, cuerpo string
	switch decision {
	case "aprobar":
		asunto = "Tu perfil de especialista fue aprobado | DETOX_WEB"
		cuerpo = "Hola,\n\nTu solicitud fue aprobada. Desde ahora tu nombre y la insignia de «Profesional verificado» aparecen junto a tus publicaciones y comentarios.\n"
	case "rechazar":
		asunto = "Tu solicitud de especialista no fue aprobada | DETOX_WEB"
		cuerpo = "Hola,\n\nPor ahora no pudimos aprobar tu solicitud.\nMotivo: " + motivo + "\n\nSi crees que es un error, responde a este correo.\n"
	case "suspender":
		asunto = "Tu perfil de especialista fue suspendido | DETOX_WEB"
		cuerpo = "Hola,\n\nTu insignia de «Profesional verificado» fue retirada temporalmente.\nMotivo: " + motivo + "\n"
	}
	if err := enviarCorreo(correo, asunto, cuerpo); err != nil {
		log.Println("no se pudo avisar al especialista:", err)
	}
}
