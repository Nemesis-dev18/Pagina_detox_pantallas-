package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/mail"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

// ---------- Insignia "Profesional verificado" ----------
//
// La insignia y el nombre real salen SOLO de este JOIN: una fila de
// perfiles_especialistas con estado = 'aprobado', que únicamente cambia un
// administrador (admin_especialistas.go). Ningún dato enviado por el navegador
// influye. Se usa en las consultas de publicaciones y comentarios.
const unirPerfilAprobado = `LEFT JOIN perfiles_especialistas pe ON pe.usuario_id = u.id AND pe.estado = 'aprobado'`
const columnasInsignia = `COALESCE(concat_ws(' ', pe.nombres, pe.primer_apellido, pe.segundo_apellido), ''), pe.usuario_id IS NOT NULL`

var profesiones = map[string]string{
	"psicologia":          "Psicología",
	"psiquiatria":         "Psiquiatría",
	"medicina":            "Medicina",
	"enfermeria":          "Enfermería",
	"trabajo_social":      "Trabajo social",
	"terapia_ocupacional": "Terapia ocupacional",
}

var (
	reNombrePersona = regexp.MustCompile(`^\p{L}[\p{L} .'-]{1,79}$`)
	reAlias         = regexp.MustCompile(`^[\p{L}\p{N}][\p{L}\p{N} ._-]{2,29}$`)
	reDocumento     = regexp.MustCompile(`^[A-Z0-9]{5,15}$`)
	reRegistroProf  = regexp.MustCompile(`^[A-Za-z0-9./ -]{3,30}$`)
	reNIT           = regexp.MustCompile(`^[0-9]{6,10}$`)
)

// limpiar junta espacios repetidos y quita los de los extremos.
func limpiar(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func sinControl(s string) bool {
	if !utf8.ValidString(s) {
		return false
	}
	for _, c := range s {
		if unicode.IsControl(c) && c != '\n' {
			return false
		}
	}
	return true
}

// normalizarNIT acepta "800.123.456-7" o "800123456" y deja solo los dígitos
// sin el dígito de verificación. Devuelve "" si no es un NIT razonable.
func normalizarNIT(s string) string {
	s = strings.ReplaceAll(strings.ReplaceAll(s, ".", ""), " ", "")
	if i := strings.Index(s, "-"); i >= 0 {
		s = s[:i]
	}
	if !reNIT.MatchString(s) {
		return ""
	}
	return s
}

func normalizarDocumento(s string) string {
	s = strings.ToUpper(s)
	for _, c := range []string{" ", ".", "-"} {
		s = strings.ReplaceAll(s, c, "")
	}
	return s
}

// datosRegistroEspecialista son los campos ya limpios y validados.
type datosRegistroEspecialista struct {
	Alias, Correo, Password                    string
	Nombres, PrimerApellido, SegundoApellido   string
	Profesion, Especialidad, Institucion       string
	NIT, TipoDocumento, Documento, NumRegistro string
	Descripcion                                string
	Centros                                    []string // la principal primero, luego adicionales
}

// validarRegistroEspecialista devuelve el primer error en un mensaje claro.
// Es la validación del servidor: la del navegador es solo una ayuda.
func validarRegistroEspecialista(r *http.Request) (*datosRegistroEspecialista, string) {
	d := &datosRegistroEspecialista{
		Alias:           limpiar(r.FormValue("alias")),
		Correo:          strings.ToLower(strings.TrimSpace(r.FormValue("correo"))),
		Password:        r.FormValue("password"),
		Nombres:         limpiar(r.FormValue("nombres")),
		PrimerApellido:  limpiar(r.FormValue("primer_apellido")),
		SegundoApellido: limpiar(r.FormValue("segundo_apellido")),
		Profesion:       strings.TrimSpace(r.FormValue("profesion")),
		Especialidad:    limpiar(r.FormValue("especialidad")),
		Institucion:     limpiar(r.FormValue("institucion")),
		TipoDocumento:   strings.ToUpper(strings.TrimSpace(r.FormValue("tipo_documento"))),
		Documento:       normalizarDocumento(r.FormValue("numero_documento")),
		NumRegistro:     limpiar(r.FormValue("numero_registro")),
		Descripcion:     strings.TrimSpace(r.FormValue("descripcion")),
	}

	if !reNombrePersona.MatchString(d.Nombres) {
		return nil, "escribe tus nombres (solo letras, entre 2 y 80 caracteres)"
	}
	if !reNombrePersona.MatchString(d.PrimerApellido) {
		return nil, "escribe tu primer apellido (solo letras, entre 2 y 80 caracteres)"
	}
	if d.SegundoApellido != "" && !reNombrePersona.MatchString(d.SegundoApellido) {
		return nil, "el segundo apellido solo puede llevar letras"
	}
	if !reAlias.MatchString(d.Alias) {
		return nil, "el usuario debe tener entre 3 y 30 caracteres (letras, números, espacio, punto, guion)"
	}
	if a, err := mail.ParseAddress(d.Correo); err != nil || a.Address != d.Correo || len(d.Correo) > 254 {
		return nil, "el correo electrónico no es válido"
	}
	if len(d.Password) < 8 {
		return nil, "la contraseña debe tener mínimo 8 caracteres"
	}
	if len(d.Password) > 72 { // límite real de bcrypt
		return nil, "la contraseña es muy larga (máximo 72 caracteres)"
	}
	if d.Password != r.FormValue("confirmar_password") {
		return nil, "las contraseñas no coinciden"
	}
	if _, ok := profesiones[d.Profesion]; !ok {
		return nil, "elige una profesión de la lista"
	}
	if n := utf8.RuneCountInString(d.Especialidad); n < 3 || n > 100 || !sinControl(d.Especialidad) {
		return nil, "la especialidad debe tener entre 3 y 100 caracteres"
	}
	if n := utf8.RuneCountInString(d.Institucion); n < 3 || n > 150 || !sinControl(d.Institucion) {
		return nil, "la institución debe tener entre 3 y 150 caracteres"
	}
	d.Centros = []string{d.Institucion}
	for _, c := range r.Form["otros_centros"] {
		c = limpiar(c)
		if c == "" {
			continue
		}
		if n := utf8.RuneCountInString(c); n < 3 || n > 150 || !sinControl(c) {
			return nil, "cada centro adicional debe tener entre 3 y 150 caracteres"
		}
		if len(d.Centros) >= 3 {
			return nil, "puedes registrar hasta 3 centros en total"
		}
		repetido := false
		for _, existente := range d.Centros {
			if strings.EqualFold(existente, c) {
				repetido = true
			}
		}
		if !repetido {
			d.Centros = append(d.Centros, c)
		}
	}
	if nit := strings.TrimSpace(r.FormValue("institucion_nit")); nit != "" {
		if d.NIT = normalizarNIT(nit); d.NIT == "" {
			return nil, "el NIT de la institución no es válido (solo números, sin dígito de verificación)"
		}
	}
	if d.TipoDocumento != "CC" && d.TipoDocumento != "CE" && d.TipoDocumento != "PPT" {
		return nil, "elige el tipo de documento"
	}
	if !reDocumento.MatchString(d.Documento) {
		return nil, "el número de documento no es válido"
	}
	if !reRegistroProf.MatchString(d.NumRegistro) {
		return nil, "el número de tarjeta o registro profesional no es válido"
	}
	if utf8.RuneCountInString(d.Descripcion) > 600 || !sinControl(d.Descripcion) {
		return nil, "la presentación es muy larga (máximo 600 caracteres)"
	}
	// Dos autorizaciones separadas y explícitas (Ley 1581 de 2012)
	if r.FormValue("acepta_datos") != "1" {
		return nil, "debes autorizar el tratamiento de tus datos para poder verificar tu perfil"
	}
	if r.FormValue("acepta_perfil") != "1" {
		return nil, "debes aceptar que, si te aprueban, tu nombre y datos profesionales sean públicos"
	}
	return d, ""
}

// POST /registro-especialista
// Crea una cuenta común (tabla usuarios, tipo 'especialista') más su perfil en
// estado 'pendiente'. NO da insignia ni muestra el nombre real hasta que un
// administrador la apruebe.
func registrarEspecialista(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<16)

	d, msg := validarRegistroEspecialista(r)
	if msg != "" {
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	docCifrado, err := cifrar(d.TipoDocumento + ":" + d.Documento)
	if err == errSinClave {
		log.Println("registro de especialistas desactivado:", err)
		http.Error(w, "el registro de especialistas no está disponible por ahora", http.StatusServiceUnavailable)
		return
	}
	if err != nil {
		http.Error(w, "error interno", http.StatusInternalServerError)
		return
	}
	huellaDoc, err := huella(d.TipoDocumento + ":" + d.Documento)
	if err != nil {
		http.Error(w, "error interno", http.StatusInternalServerError)
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(d.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "error interno", http.StatusInternalServerError)
		return
	}

	ctx := r.Context()
	tx, err := pool.Begin(ctx)
	if err != nil {
		http.Error(w, "error interno", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(ctx)

	var id string
	err = tx.QueryRow(ctx,
		"INSERT INTO usuarios (alias, correo, password_hash, tipo) VALUES ($1, $2, $3, 'especialista') RETURNING id",
		d.Alias, d.Correo, string(hash)).Scan(&id)
	if err == nil {
		_, err = tx.Exec(ctx, `
			INSERT INTO perfiles_especialistas
			  (usuario_id, nombres, primer_apellido, segundo_apellido, profesion, especialidad, institucion,
			   institucion_nit, tipo_documento, documento_cifrado, documento_huella, numero_registro, descripcion,
			   acepta_datos_en, acepta_perfil_en)
			VALUES ($1, $2, $3, NULLIF($4, ''), $5, $6, $7, NULLIF($8, ''), $9, $10, $11, $12, NULLIF($13, ''), now(), now())`,
			id, d.Nombres, d.PrimerApellido, d.SegundoApellido, d.Profesion, d.Especialidad, d.Institucion,
			d.NIT, d.TipoDocumento, docCifrado, huellaDoc, d.NumRegistro, d.Descripcion)
	}
	for i, c := range d.Centros {
		if err != nil {
			break
		}
		_, err = tx.Exec(ctx,
			"INSERT INTO especialista_centros (usuario_id, institucion, es_principal) VALUES ($1, $2, $3)",
			id, c, i == 0)
	}
	if err == nil {
		err = tx.Commit(ctx)
	}
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			if strings.Contains(pgErr.ConstraintName, "documento") {
				http.Error(w, "ya existe una solicitud con ese documento", http.StatusConflict)
			} else {
				http.Error(w, "ese alias o correo ya está en uso", http.StatusConflict)
			}
			return
		}
		log.Println("registro de especialista:", err)
		http.Error(w, "error guardando la solicitud", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/static/inicio_sesion.html?registro=ok", http.StatusSeeOther)
}

// ---------- Consulta pública ----------

type especialistaPublico struct {
	ID             string `json:"id"`
	Nombre         string `json:"nombre"`
	Alias          string `json:"alias"`
	Profesion      string `json:"profesion"`
	ProfesionClave string `json:"profesion_clave"`
	Especialidad   string `json:"especialidad"`
	Institucion    string `json:"institucion"`
	Descripcion    string `json:"descripcion,omitempty"`
}

const columnasPublicas = `
	pe.usuario_id, concat_ws(' ', pe.nombres, pe.primer_apellido, pe.segundo_apellido), u.alias,
	pe.profesion, pe.especialidad, pe.institucion, COALESCE(pe.descripcion, '')`

func leerEspecialistaPublico(row pgx.Row) (*especialistaPublico, error) {
	var e especialistaPublico
	if err := row.Scan(&e.ID, &e.Nombre, &e.Alias, &e.ProfesionClave, &e.Especialidad, &e.Institucion, &e.Descripcion); err != nil {
		return nil, err
	}
	e.Profesion = profesiones[e.ProfesionClave]
	return &e, nil
}

// escaparLike evita que % o _ escritos por quien busca funcionen como comodines.
func escaparLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}

// GET /especialistas?profesion=psicologia&q=ansiedad
// Lista pública: solo especialistas con estado 'aprobado'.
func listarEspecialistas(w http.ResponseWriter, r *http.Request) {
	profesion := r.URL.Query().Get("profesion")
	if profesion != "" {
		if _, ok := profesiones[profesion]; !ok {
			http.Error(w, "profesión no válida", http.StatusBadRequest)
			return
		}
	}
	q := limpiar(r.URL.Query().Get("q"))
	if utf8.RuneCountInString(q) > 80 {
		http.Error(w, "búsqueda muy larga", http.StatusBadRequest)
		return
	}
	patron := "%" + escaparLike(q) + "%"

	rows, err := pool.Query(r.Context(), `
		SELECT `+columnasPublicas+`
		FROM perfiles_especialistas pe
		JOIN usuarios u ON u.id = pe.usuario_id
		WHERE pe.estado = 'aprobado'
		  AND ($1 = '' OR pe.profesion = $1)
		  AND ($2 = '%%' OR concat_ws(' ', pe.nombres, pe.primer_apellido, pe.segundo_apellido, pe.especialidad, pe.institucion) ILIKE $2)
		ORDER BY pe.nombres, pe.primer_apellido
		LIMIT 100`, profesion, patron)
	if err != nil {
		http.Error(w, "error consultando la base", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	lista := []especialistaPublico{}
	for rows.Next() {
		e, err := leerEspecialistaPublico(rows)
		if err != nil {
			http.Error(w, "error leyendo datos", http.StatusInternalServerError)
			return
		}
		lista = append(lista, *e)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, "error leyendo datos", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "private, no-store")
	json.NewEncoder(w).Encode(lista)
}

// GET /especialistas/{id}  (perfil público; solo si está aprobado)
func verEspecialista(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !reUUID.MatchString(id) {
		http.Error(w, "especialista no encontrado", http.StatusNotFound)
		return
	}
	e, err := leerEspecialistaPublico(pool.QueryRow(r.Context(), `
		SELECT `+columnasPublicas+`
		FROM perfiles_especialistas pe
		JOIN usuarios u ON u.id = pe.usuario_id
		WHERE pe.usuario_id = $1 AND pe.estado = 'aprobado'`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		http.Error(w, "especialista no encontrado", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "error consultando la base", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "private, no-store")
	json.NewEncoder(w).Encode(e)
}

// GET /yo/especialista: la propia solicitud (estado de la verificación).
func miSolicitudEspecialista(w http.ResponseWriter, r *http.Request) {
	u, err := usuarioActual(r)
	if err != nil {
		http.Error(w, "no has iniciado sesión", http.StatusUnauthorized)
		return
	}
	var v struct {
		Estado            string    `json:"estado"`
		CredencialEstado  string    `json:"credencial_estado"`
		InstitucionEstado string    `json:"institucion_estado"`
		Motivo            string    `json:"motivo,omitempty"`
		Nombre            string    `json:"nombre"`
		Profesion         string    `json:"profesion"`
		Institucion       string    `json:"institucion"`
		CreadoEn          time.Time `json:"creado_en"`
	}
	var clave string
	err = pool.QueryRow(r.Context(), `
		SELECT estado, credencial_estado, institucion_estado, COALESCE(motivo, ''),
		       concat_ws(' ', nombres, primer_apellido, segundo_apellido), profesion, institucion, creado_en
		FROM perfiles_especialistas WHERE usuario_id = $1`, u.ID).
		Scan(&v.Estado, &v.CredencialEstado, &v.InstitucionEstado, &v.Motivo, &v.Nombre, &clave, &v.Institucion, &v.CreadoEn)
	if errors.Is(err, pgx.ErrNoRows) {
		http.Error(w, "no eres especialista", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "error consultando la base", http.StatusInternalServerError)
		return
	}
	v.Profesion = profesiones[clave]
	if v.Estado != "rechazado" && v.Estado != "suspendido" {
		v.Motivo = ""
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "private, no-store")
	json.NewEncoder(w).Encode(v)
}
