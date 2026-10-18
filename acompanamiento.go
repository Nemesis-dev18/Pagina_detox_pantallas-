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
	"github.com/jackc/pgx/v5/pgconn"
)

// Acompañamiento voluntario: una persona pide hablar con un especialista
// APROBADO; el especialista acepta o rechaza; si acepta, ambos pueden
// intercambiar mensajes privados.
//
// Privacidad (regla única, en SQL): una conversación solo la ven o escriben sus
// dos participantes. Un especialista, además, debe seguir "aprobado". Ni
// administradores ni otros especialistas tienen acceso. Quien no participa
// recibe 404, igual que si la conversación no existiera.

const maxSolicitudesPendientes = 3

// Condición de acceso reutilizada. $1 = id de la conversación, $2 = usuario que consulta.
const condicionParticipante = `
	a.id = $1 AND (
		a.solicitante_id = $2
		OR (a.especialista_id = $2 AND EXISTS (
			SELECT 1 FROM perfiles_especialistas WHERE usuario_id = $2 AND estado = 'aprobado')))`

func escribirJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "private, no-store")
	json.NewEncoder(w).Encode(v)
}

// POST /acompanamientos   (especialista_id, mensaje, autorizo=1)
func solicitarAcompanamiento(w http.ResponseWriter, r *http.Request) {
	u, err := usuarioActual(r)
	if err != nil {
		http.Error(w, "debes iniciar sesión para solicitar acompañamiento", http.StatusUnauthorized)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<16)

	especialistaID := r.FormValue("especialista_id")
	mensaje := strings.TrimSpace(r.FormValue("mensaje"))
	if !reUUID.MatchString(especialistaID) {
		http.Error(w, "especialista no encontrado", http.StatusNotFound)
		return
	}
	if n := utf8.RuneCountInString(mensaje); n < 1 || n > 500 || !sinControl(mensaje) {
		http.Error(w, "cuéntale al especialista, en máximo 500 caracteres, qué necesitas", http.StatusBadRequest)
		return
	}
	// Lo que se escribe aquí puede ser dato sensible de salud: exige autorización expresa (Ley 1581 de 2012)
	if r.FormValue("autorizo") != "1" {
		http.Error(w, "debes autorizar que el especialista vea tu mensaje y la conversación", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	tx, err := pool.Begin(ctx)
	if err != nil {
		http.Error(w, "error interno", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(ctx)
	// Una solicitud a la vez por persona, para que el tope de pendientes no se salte con peticiones paralelas
	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock(hashtext($1::text))", u.ID); err != nil {
		http.Error(w, "error interno", http.StatusInternalServerError)
		return
	}
	var pendientes int
	if err := tx.QueryRow(ctx,
		"SELECT count(*) FROM acompanamientos WHERE solicitante_id = $1 AND estado = 'pendiente'", u.ID).Scan(&pendientes); err != nil {
		http.Error(w, "error interno", http.StatusInternalServerError)
		return
	}
	if pendientes >= maxSolicitudesPendientes {
		http.Error(w, "ya tienes varias solicitudes esperando respuesta; espera o cancela alguna", http.StatusTooManyRequests)
		return
	}

	// Solo se inserta si el especialista está aprobado AHORA en la base
	var id string
	err = tx.QueryRow(ctx, `
		INSERT INTO acompanamientos (solicitante_id, especialista_id, mensaje, autorizado_en)
		SELECT $1, pe.usuario_id, $3, now() FROM perfiles_especialistas pe
		WHERE pe.usuario_id = $2 AND pe.estado = 'aprobado'
		RETURNING id`, u.ID, especialistaID, mensaje).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		switch {
		case errors.As(err, &pgErr) && pgErr.Code == "23505":
			http.Error(w, "ya tienes una solicitud activa con este especialista", http.StatusConflict)
		case errors.As(err, &pgErr) && pgErr.Code == "23514":
			http.Error(w, "no puedes solicitarte acompañamiento a ti mismo", http.StatusBadRequest)
		case errors.Is(err, pgx.ErrNoRows):
			http.Error(w, "ese especialista no está disponible", http.StatusNotFound)
		default:
			log.Println("solicitar acompañamiento:", err)
			http.Error(w, "error guardando la solicitud", http.StatusInternalServerError)
		}
		return
	}
	if err := tx.Commit(ctx); err != nil {
		http.Error(w, "error interno", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": id})
}

type solicitudVista struct {
	ID           string     `json:"id"`
	Estado       string     `json:"estado"`
	Mensaje      string     `json:"mensaje"`
	CreadoEn     time.Time  `json:"creado_en"`
	RespondidoEn *time.Time `json:"respondido_en,omitempty"`
	// enviadas: datos del especialista; recibidas: alias de quien consulta
	Nombre string `json:"nombre"`
	Alias  string `json:"alias"`
}

// GET /acompanamientos → {enviadas, recibidas}
func listarAcompanamientos(w http.ResponseWriter, r *http.Request) {
	u, err := usuarioActual(r)
	if err != nil {
		http.Error(w, "no has iniciado sesión", http.StatusUnauthorized)
		return
	}
	leer := func(consulta string) ([]solicitudVista, error) {
		rows, err := pool.Query(r.Context(), consulta, u.ID)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		lista := []solicitudVista{}
		for rows.Next() {
			var s solicitudVista
			if err := rows.Scan(&s.ID, &s.Estado, &s.Mensaje, &s.CreadoEn, &s.RespondidoEn, &s.Nombre, &s.Alias); err != nil {
				return nil, err
			}
			lista = append(lista, s)
		}
		return lista, rows.Err()
	}

	// Enviadas: el nombre real solo si el especialista sigue aprobado
	enviadas, err := leer(`
		SELECT a.id, a.estado, a.mensaje, a.creado_en, a.respondido_en,
		       CASE WHEN pe.estado = 'aprobado' THEN concat_ws(' ', pe.nombres, pe.primer_apellido, pe.segundo_apellido) ELSE '' END,
		       ue.alias
		FROM acompanamientos a
		JOIN perfiles_especialistas pe ON pe.usuario_id = a.especialista_id
		JOIN usuarios ue ON ue.id = a.especialista_id
		WHERE a.solicitante_id = $1 ORDER BY a.creado_en DESC LIMIT 100`)
	if err != nil {
		http.Error(w, "error consultando la base", http.StatusInternalServerError)
		return
	}
	// Recibidas: solo si quien consulta es un especialista aprobado; ve el alias, nunca el correo
	recibidas, err := leer(`
		SELECT a.id, a.estado, a.mensaje, a.creado_en, a.respondido_en, '', us.alias
		FROM acompanamientos a
		JOIN usuarios us ON us.id = a.solicitante_id
		WHERE a.especialista_id = $1
		  AND EXISTS (SELECT 1 FROM perfiles_especialistas WHERE usuario_id = $1 AND estado = 'aprobado')
		ORDER BY (a.estado = 'pendiente') DESC, a.creado_en DESC LIMIT 100`)
	if err != nil {
		http.Error(w, "error consultando la base", http.StatusInternalServerError)
		return
	}
	escribirJSON(w, map[string]any{"enviadas": enviadas, "recibidas": recibidas})
}

// POST /acompanamientos/{id}/responder   (decision=aceptar|rechazar)
func responderAcompanamiento(w http.ResponseWriter, r *http.Request) {
	u, err := usuarioActual(r)
	if err != nil {
		http.Error(w, "no has iniciado sesión", http.StatusUnauthorized)
		return
	}
	id := r.PathValue("id")
	nuevo := map[string]string{"aceptar": "aceptada", "rechazar": "rechazada"}[r.FormValue("decision")]
	if !reUUID.MatchString(id) || nuevo == "" {
		http.Error(w, "solicitud o decisión no válida", http.StatusBadRequest)
		return
	}
	tag, err := pool.Exec(r.Context(), `
		UPDATE acompanamientos a SET estado = $3, respondido_en = now()
		WHERE a.id = $1 AND a.especialista_id = $2 AND a.estado = 'pendiente'
		  AND EXISTS (SELECT 1 FROM perfiles_especialistas WHERE usuario_id = $2 AND estado = 'aprobado')`,
		id, u.ID, nuevo)
	if err != nil {
		http.Error(w, "error guardando la respuesta", http.StatusInternalServerError)
		return
	}
	if tag.RowsAffected() == 0 {
		http.Error(w, "solicitud no encontrada o ya respondida", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// POST /acompanamientos/{id}/cerrar
// La persona puede cancelar una solicitud pendiente; cualquiera de los dos puede
// finalizar un acompañamiento aceptado.
func cerrarAcompanamiento(w http.ResponseWriter, r *http.Request) {
	u, err := usuarioActual(r)
	if err != nil {
		http.Error(w, "no has iniciado sesión", http.StatusUnauthorized)
		return
	}
	id := r.PathValue("id")
	if !reUUID.MatchString(id) {
		http.Error(w, "solicitud no encontrada", http.StatusNotFound)
		return
	}
	tag, err := pool.Exec(r.Context(), `
		UPDATE acompanamientos SET
		  estado = CASE WHEN estado = 'pendiente' THEN 'cancelada' ELSE 'finalizada' END,
		  respondido_en = COALESCE(respondido_en, now())
		WHERE id = $1 AND (
		  (estado = 'pendiente' AND solicitante_id = $2)
		  OR (estado = 'aceptada' AND (solicitante_id = $2 OR especialista_id = $2)))`, id, u.ID)
	if err != nil {
		http.Error(w, "error guardando el cambio", http.StatusInternalServerError)
		return
	}
	if tag.RowsAffected() == 0 {
		http.Error(w, "solicitud no encontrada", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type mensajePrivado struct {
	ID           string    `json:"id"`
	Mio          bool      `json:"mio"`
	Especialista bool      `json:"del_especialista"`
	Contenido    string    `json:"contenido"`
	CreadoEn     time.Time `json:"creado_en"`
}

// GET /acompanamientos/{id}/mensajes
func listarMensajesPrivados(w http.ResponseWriter, r *http.Request) {
	u, err := usuarioActual(r)
	if err != nil {
		http.Error(w, "no has iniciado sesión", http.StatusUnauthorized)
		return
	}
	id := r.PathValue("id")
	if !reUUID.MatchString(id) {
		http.Error(w, "conversación no encontrada", http.StatusNotFound)
		return
	}
	var permitido bool
	err = pool.QueryRow(r.Context(), `
		SELECT EXISTS (SELECT 1 FROM acompanamientos a WHERE `+condicionParticipante+` AND a.estado IN ('aceptada', 'finalizada'))`,
		id, u.ID).Scan(&permitido)
	if err != nil {
		http.Error(w, "error consultando la base", http.StatusInternalServerError)
		return
	}
	if !permitido {
		http.Error(w, "conversación no encontrada", http.StatusNotFound)
		return
	}
	rows, err := pool.Query(r.Context(), `
		SELECT m.id, m.autor_id = $2, m.autor_id = a.especialista_id, m.contenido, m.creado_en
		FROM acompanamiento_mensajes m JOIN acompanamientos a ON a.id = m.acompanamiento_id
		WHERE m.acompanamiento_id = $1 ORDER BY m.creado_en LIMIT 500`, id, u.ID)
	if err != nil {
		http.Error(w, "error consultando la base", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	lista := []mensajePrivado{}
	for rows.Next() {
		var m mensajePrivado
		if err := rows.Scan(&m.ID, &m.Mio, &m.Especialista, &m.Contenido, &m.CreadoEn); err != nil {
			http.Error(w, "error leyendo datos", http.StatusInternalServerError)
			return
		}
		lista = append(lista, m)
	}
	escribirJSON(w, lista)
}

// POST /acompanamientos/{id}/mensajes   (contenido)
func enviarMensajePrivado(w http.ResponseWriter, r *http.Request) {
	u, err := usuarioActual(r)
	if err != nil {
		http.Error(w, "no has iniciado sesión", http.StatusUnauthorized)
		return
	}
	id := r.PathValue("id")
	if !reUUID.MatchString(id) {
		http.Error(w, "conversación no encontrada", http.StatusNotFound)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<16)
	contenido := strings.TrimSpace(r.FormValue("contenido"))
	if n := utf8.RuneCountInString(contenido); n < 1 || n > 2000 || !sinControl(contenido) {
		http.Error(w, "escribe un mensaje de máximo 2000 caracteres", http.StatusBadRequest)
		return
	}
	// El acceso y el estado "aceptada" se comprueban en la misma sentencia que inserta
	tag, err := pool.Exec(r.Context(), `
		INSERT INTO acompanamiento_mensajes (acompanamiento_id, autor_id, contenido)
		SELECT a.id, $2, $3 FROM acompanamientos a WHERE `+condicionParticipante+` AND a.estado = 'aceptada'`,
		id, u.ID, contenido)
	if err != nil {
		http.Error(w, "error guardando el mensaje", http.StatusInternalServerError)
		return
	}
	if tag.RowsAffected() == 0 {
		http.Error(w, "conversación no encontrada o ya finalizada", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusCreated)
}
