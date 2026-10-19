package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/mail"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

var categoriasValidas = map[string]bool{
	"pregunta":    true,
	"experiencia": true,
	"consejo":     true,
}

type Publicacion struct {
	ID                string    `json:"id"`
	Alias             string    `json:"alias"`
	Titulo            string    `json:"titulo"`
	Contenido         string    `json:"contenido"`
	Categoria         string    `json:"categoria"`
	CreadoEn          time.Time `json:"creado_en"`
	TituloOriginal    string    `json:"titulo_original,omitempty"`
	ContenidoOriginal string    `json:"contenido_original,omitempty"`
	// Solo se llenan si el autor es un especialista APROBADO en la base (ver especialistas.go)
	NombreReal string `json:"nombre_real,omitempty"`
	Verificado bool   `json:"profesional_verificado,omitempty"`
	Likes      int    `json:"likes"`
	MeGusta    bool   `json:"me_gusta"`
}

func listarPublicaciones(w http.ResponseWriter, r *http.Request) {
	u, errSesion := usuarioActual(r)
	haySesion := errSesion == nil
	// Visitantes: uid nil, así "me gusta" sale en falso
	var uid *string
	if haySesion {
		uid = &u.ID
	}

	rows, err := pool.Query(r.Context(), `
		SELECT p.id, u.alias, p.titulo, p.contenido, COALESCE(p.categoria, ''), p.creado_en, `+columnasInsignia+`,
			(SELECT COUNT(*) FROM me_gusta g WHERE g.publicacion_id = p.id),
			EXISTS (SELECT 1 FROM me_gusta g WHERE g.publicacion_id = p.id AND g.usuario_id = $1)
		FROM publicaciones p
		JOIN usuarios u ON u.id = p.autor_id
		`+unirPerfilAprobado+`
		ORDER BY p.creado_en DESC
		LIMIT 50`, uid)
	if err != nil {
		http.Error(w, "error consultando la base", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	publicaciones := []Publicacion{}
	for rows.Next() {
		var p Publicacion
		if err := rows.Scan(&p.ID, &p.Alias, &p.Titulo, &p.Contenido, &p.Categoria, &p.CreadoEn, &p.NombreReal, &p.Verificado, &p.Likes, &p.MeGusta); err != nil {
			http.Error(w, "error leyendo datos", http.StatusInternalServerError)
			return
		}
		// En la base queda el texto original; al público se le muestra tapado.
		// El autor (o un admin) recibe además el original para poder editarlo.
		tituloOrig, contenidoOrig := p.Titulo, p.Contenido
		p.Titulo = censurar(p.Titulo)
		p.Contenido = censurar(p.Contenido)
		if haySesion && (u.EsAdmin || u.Alias == p.Alias) {
			p.TituloOriginal = tituloOrig
			p.ContenidoOriginal = contenidoOrig
		}
		publicaciones = append(publicaciones, p)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, "error leyendo datos", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "private, no-store")
	json.NewEncoder(w).Encode(publicaciones)
}

func crearPublicacion(w http.ResponseWriter, r *http.Request) {
	// Sin sesión no se publica
	u, err := usuarioActual(r)
	if err != nil {
		http.Error(w, "debes iniciar sesión para publicar", http.StatusUnauthorized)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // máximo 1 MB por petición

	titulo := strings.TrimSpace(r.FormValue("titulo"))
	contenido := strings.TrimSpace(r.FormValue("contenido"))
	categoria := strings.TrimSpace(r.FormValue("categoria"))

	if titulo == "" || contenido == "" {
		http.Error(w, "el título y el contenido son obligatorios", http.StatusBadRequest)
		return
	}
	if len([]rune(titulo)) > 120 {
		http.Error(w, "el título es muy largo (máximo 120 caracteres)", http.StatusBadRequest)
		return
	}
	if len([]rune(contenido)) > 5000 {
		http.Error(w, "el contenido es muy largo (máximo 5000 caracteres)", http.StatusBadRequest)
		return
	}
	if !categoriasValidas[categoria] {
		http.Error(w, "categoría no válida", http.StatusBadRequest)
		return
	}

	var id string
	err = pool.QueryRow(r.Context(),
		"INSERT INTO publicaciones (autor_id, titulo, contenido, categoria) VALUES ($1, $2, $3, $4) RETURNING id",
		u.ID, titulo, contenido, categoria).Scan(&id)
	if err != nil {
		http.Error(w, "error guardando la publicación", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": id})
}
func registrar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
		return
	}

	alias := strings.TrimSpace(r.FormValue("alias"))
	correo := strings.ToLower(strings.TrimSpace(r.FormValue("correo")))
	password := r.FormValue("password")
	confirmar := r.FormValue("confirmar_password")

	if alias == "" || correo == "" || password == "" {
		http.Error(w, "faltan datos", http.StatusBadRequest)
		return
	}
	if password != confirmar {
		http.Error(w, "las contraseñas no coinciden", http.StatusBadRequest)
		return
	}
	if len(password) < 8 {
		http.Error(w, "la contraseña debe tener mínimo 8 caracteres", http.StatusBadRequest)
		return
	}
	// Validación del servidor (antes solo la tenía el navegador)
	if !reAlias.MatchString(alias) {
		http.Error(w, "el alias debe tener entre 3 y 30 caracteres (letras, números, espacio, punto, guion)", http.StatusBadRequest)
		return
	}
	if a, err := mail.ParseAddress(correo); err != nil || a.Address != correo || len(correo) > 254 {
		http.Error(w, "el correo electrónico no es válido", http.StatusBadRequest)
		return
	}
	if len(password) > 72 { // límite de bcrypt: más allá, el hash no sirve
		http.Error(w, "la contraseña es muy larga (máximo 72 caracteres)", http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "error interno", http.StatusInternalServerError)
		return
	}

	_, err = pool.Exec(r.Context(),
		"INSERT INTO usuarios (alias, correo, password_hash) VALUES ($1, $2, $3)",
		alias, correo, string(hash))
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			http.Error(w, "ese alias o correo ya está en uso", http.StatusConflict)
			return
		}
		http.Error(w, "error guardando el usuario", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/static/inicio_sesion.html", http.StatusSeeOther)
}

// ---------- Recuperar contraseña ----------

func generarToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func baseURL() string {
	if u := os.Getenv("BASE_URL"); u != "" {
		return u
	}
	return "http://localhost:8080"
}

func enviarEnlaceRecuperacion(ctx context.Context, usuarioID, correo string) error {
	token, err := generarToken()
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx,
		"INSERT INTO recuperaciones (token_hash, usuario_id, expira_en) VALUES ($1, $2, now() + interval '30 minutes')",
		hashToken(token), usuarioID)
	if err != nil {
		return err
	}

	enlace := baseURL() + "/static/restablecer.html?token=" + token
	cuerpo := "Hola,\n\nPara elegir una nueva contraseña entra a este enlace (vale 30 minutos):\n\n" +
		enlace + "\n\nSi no fuiste tú, ignora este mensaje.\n"
	return enviarCorreo(correo, "Recuperar contraseña | R-HOPE", cuerpo)
}

func recuperar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
		return
	}

	correo := strings.ToLower(strings.TrimSpace(r.FormValue("correo")))

	var usuarioID string
	err := pool.QueryRow(r.Context(), "SELECT id FROM usuarios WHERE correo = $1", correo).Scan(&usuarioID)
	// Si el correo no tiene cuenta, la respuesta es la misma: no revelamos qué correos están registrados
	if errors.Is(err, pgx.ErrNoRows) {
		responderRecuperacion(w)
		return
	}
	if err != nil {
		log.Println("recuperación:", err)
		http.Error(w, "error interno", http.StatusInternalServerError)
		return
	}

	if err := enviarEnlaceRecuperacion(r.Context(), usuarioID, correo); err != nil {
		log.Println("recuperación:", err)
		http.Error(w, "no pudimos enviar el correo, intenta de nuevo en unos minutos", http.StatusInternalServerError)
		return
	}

	responderRecuperacion(w)
}

func responderRecuperacion(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(w, "Si el correo está registrado, te enviamos un enlace para restablecer tu contraseña. Revisa tu bandeja de entrada.")
}

func restablecer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
		return
	}

	token := r.FormValue("token")
	password := r.FormValue("password")
	confirmar := r.FormValue("confirmar_password")

	if token == "" || password == "" {
		http.Error(w, "faltan datos", http.StatusBadRequest)
		return
	}
	if len(password) > 72 { // límite de bcrypt
		http.Error(w, "la contraseña es muy larga (máximo 72 caracteres)", http.StatusBadRequest)
		return
	}
	if password != confirmar {
		http.Error(w, "las contraseñas no coinciden", http.StatusBadRequest)
		return
	}
	if len(password) < 8 {
		http.Error(w, "la contraseña debe tener mínimo 8 caracteres", http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
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

	// Marca el token como usado y devuelve de qué usuario es (solo si sirve)
	var usuarioID string
	err = tx.QueryRow(ctx,
		"UPDATE recuperaciones SET usado = true WHERE token_hash = $1 AND usado = false AND expira_en > now() RETURNING usuario_id",
		hashToken(token)).Scan(&usuarioID)
	if errors.Is(err, pgx.ErrNoRows) {
		http.Error(w, "el enlace no es válido o ya venció", http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, "error interno", http.StatusInternalServerError)
		return
	}

	if _, err = tx.Exec(ctx, "UPDATE usuarios SET password_hash = $1 WHERE id = $2", string(hash), usuarioID); err != nil {
		http.Error(w, "error interno", http.StatusInternalServerError)
		return
	}
	if _, err = tx.Exec(ctx, "DELETE FROM sesiones WHERE usuario_id = $1", usuarioID); err != nil {
		http.Error(w, "error interno", http.StatusInternalServerError)
		return
	}
	if err = tx.Commit(ctx); err != nil {
		http.Error(w, "error interno", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/static/inicio_sesion.html", http.StatusSeeOther)
}

// ---------- Login y sesiones ----------

type Usuario struct {
	ID      string
	Alias   string
	EsAdmin bool
	Tipo    string // "usuario" o "especialista"
}

// Devuelve quién es el usuario según la cookie, o un error si no hay sesión válida
func usuarioActual(r *http.Request) (*Usuario, error) {
	c, err := r.Cookie("sesion")
	if err != nil {
		return nil, err
	}

	var u Usuario
	err = pool.QueryRow(r.Context(),
		`SELECT u.id, u.alias, u.es_admin, u.tipo
		 FROM sesiones s
		 JOIN usuarios u ON u.id = s.usuario_id
		 WHERE s.token_hash = $1 AND s.expira_en > now()`,
		hashToken(c.Value)).Scan(&u.ID, &u.Alias, &u.EsAdmin, &u.Tipo)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// hashRelleno se compara cuando el correo no existe, para que la respuesta
// tarde lo mismo que con una cuenta real (no se revela qué correos existen).
var hashRelleno, _ = bcrypt.GenerateFromPassword([]byte("relleno-sin-uso"), bcrypt.DefaultCost)

func iniciarSesion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
		return
	}

	// El campo del formulario se llama "username", pero trae el correo
	correo := strings.ToLower(strings.TrimSpace(r.FormValue("username")))
	password := r.FormValue("password")

	if bloq, restante := bloqueado(correo); bloq {
		minutos := int(restante.Minutes()) + 1
		http.Error(w, fmt.Sprintf("demasiados intentos fallidos, intenta de nuevo en %d minutos", minutos), http.StatusTooManyRequests)
		return
	}

	var id, hash string
	err := pool.QueryRow(r.Context(),
		"SELECT id, password_hash FROM usuarios WHERE correo = $1", correo).Scan(&id, &hash)
	existe := err == nil
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		http.Error(w, "error interno", http.StatusInternalServerError)
		return
	}
	if !existe {
		hash = string(hashRelleno)
	}

	// Correo inexistente o contraseña mal: mismo mensaje, mismo tiempo y mismo contador de intentos
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil || !existe {
		registrarFallo(correo)
		http.Error(w, "correo o contraseña incorrectos", http.StatusUnauthorized)
		return
	}
	limpiarIntentos(correo)

	abrirSesion(w, r, id, "/static/index.html")
}

// abrirSesion crea la sesión en la base, pone la cookie y redirige a destino.
// La comparten el inicio de sesión normal y el de especialistas.
func abrirSesion(w http.ResponseWriter, r *http.Request, usuarioID, destino string) {
	token, err := generarToken()
	if err != nil {
		http.Error(w, "error interno", http.StatusInternalServerError)
		return
	}
	_, err = pool.Exec(r.Context(),
		"INSERT INTO sesiones (token_hash, usuario_id, expira_en) VALUES ($1, $2, now() + interval '30 days')",
		hashToken(token), usuarioID)
	if err != nil {
		http.Error(w, "error interno", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "sesion",
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(30 * 24 * time.Hour), // los mismos 30 días de la base
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   cookieSegura(),
	})
	http.Redirect(w, r, destino, http.StatusSeeOther)
}

func cerrarSesion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
		return
	}
	if c, err := r.Cookie("sesion"); err == nil {
		pool.Exec(r.Context(), "DELETE FROM sesiones WHERE token_hash = $1", hashToken(c.Value))
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "sesion",
		Value:    "",
		Path:     "/",
		MaxAge:   -1, // le dice al navegador que borre la cookie
		HttpOnly: true,
		Secure:   cookieSegura(),
	})
	http.Redirect(w, r, "/static/index.html", http.StatusSeeOther)
}

// Ruta de prueba: dice quién eres según tu sesión
func yo(w http.ResponseWriter, r *http.Request) {
	u, err := usuarioActual(r)
	if err != nil {
		http.Error(w, "no has iniciado sesión", http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	datos := map[string]interface{}{
		"alias":    u.Alias,
		"es_admin": u.EsAdmin,
		"tipo":     u.Tipo,
	}
	if u.Tipo == "especialista" {
		// El estado sale siempre de la base; el navegador no puede decidirlo
		var estado string
		pool.QueryRow(r.Context(), "SELECT estado FROM perfiles_especialistas WHERE usuario_id = $1", u.ID).Scan(&estado)
		datos["especialista_estado"] = estado
		datos["profesional_verificado"] = estado == "aprobado"
	}
	json.NewEncoder(w).Encode(datos)
}

// ---------- Comentarios ----------

var reUUID = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

type Comentario struct {
	ID                string    `json:"id"`
	Alias             string    `json:"alias"`
	Contenido         string    `json:"contenido"`
	ContenidoOriginal string    `json:"contenido_original,omitempty"`
	CreadoEn          time.Time `json:"creado_en"`
	NombreReal        string    `json:"nombre_real,omitempty"`
	Verificado        bool      `json:"profesional_verificado,omitempty"`
	Likes             int       `json:"likes"`
	MeGusta           bool      `json:"me_gusta"`
}

func listarComentarios(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !reUUID.MatchString(id) {
		http.Error(w, "publicación no encontrada", http.StatusNotFound)
		return
	}

	u, errSesion := usuarioActual(r)
	haySesion := errSesion == nil
	var uid *string
	if haySesion {
		uid = &u.ID
	}

	rows, err := pool.Query(r.Context(), `
		SELECT c.id, u.alias, c.contenido, c.creado_en, `+columnasInsignia+`,
			(SELECT COUNT(*) FROM me_gusta g WHERE g.comentario_id = c.id),
			EXISTS (SELECT 1 FROM me_gusta g WHERE g.comentario_id = c.id AND g.usuario_id = $1)
		FROM comentarios c
		JOIN usuarios u ON u.id = c.autor_id
		`+unirPerfilAprobado+`
		WHERE c.publicacion_id = $2
		ORDER BY c.creado_en ASC`, uid, id)
	if err != nil {
		http.Error(w, "error consultando la base", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	comentarios := []Comentario{}
	for rows.Next() {
		var c Comentario
		if err := rows.Scan(&c.ID, &c.Alias, &c.Contenido, &c.CreadoEn, &c.NombreReal, &c.Verificado, &c.Likes, &c.MeGusta); err != nil {
			http.Error(w, "error leyendo datos", http.StatusInternalServerError)
			return
		}
		// En la base queda el texto original; al público se le muestra tapado.
		// El autor (o un admin) recibe además el original para poder editarlo.
		original := c.Contenido
		c.Contenido = censurar(c.Contenido)
		if haySesion && (u.EsAdmin || u.Alias == c.Alias) {
			c.ContenidoOriginal = original
		}
		comentarios = append(comentarios, c)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, "error leyendo datos", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "private, no-store")
	json.NewEncoder(w).Encode(comentarios)
}

func crearComentario(w http.ResponseWriter, r *http.Request) {
	u, err := usuarioActual(r)
	if err != nil {
		http.Error(w, "debes iniciar sesión para responder", http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")
	if !reUUID.MatchString(id) {
		http.Error(w, "publicación no encontrada", http.StatusNotFound)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	contenido := strings.TrimSpace(r.FormValue("contenido"))
	if contenido == "" {
		http.Error(w, "escribe algo para responder", http.StatusBadRequest)
		return
	}
	if len([]rune(contenido)) > 2000 {
		http.Error(w, "la respuesta es muy larga (máximo 2000 caracteres)", http.StatusBadRequest)
		return
	}

	// Revisa bloqueo por spam y guarda el comentario (ver antispam.go)
	err = guardarComentario(r.Context(), u.ID, id, contenido)
	if err != nil {
		var bloq errBloqueado
		if errors.As(err, &bloq) {
			responderBloqueo(w, bloq)
			return
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			http.Error(w, "la publicación no existe", http.StatusNotFound)
			return
		}
		http.Error(w, "error guardando la respuesta", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
func borrarPublicacion(w http.ResponseWriter, r *http.Request) {
	u, err := usuarioActual(r)
	if err != nil {
		http.Error(w, "debes iniciar sesión", http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")
	if !reUUID.MatchString(id) {
		http.Error(w, "publicación no encontrada", http.StatusNotFound)
		return
	}

	var tag pgconn.CommandTag
	if u.EsAdmin {
		tag, err = pool.Exec(r.Context(), "DELETE FROM publicaciones WHERE id = $1", id)
	} else {
		tag, err = pool.Exec(r.Context(),
			"DELETE FROM publicaciones WHERE id = $1 AND autor_id = $2", id, u.ID)
	}
	if err != nil {
		http.Error(w, "error borrando la publicación", http.StatusInternalServerError)
		return
	}
	if tag.RowsAffected() == 0 {
		http.Error(w, "no encontrada, o no es tuya", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DELETE /publicaciones/{id}/comentarios/{comentarioId}
// Borra un comentario. Solo el autor o un admin.
func borrarComentario(w http.ResponseWriter, r *http.Request) {
	u, err := usuarioActual(r)
	if err != nil {
		http.Error(w, "debes iniciar sesión", http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")
	comentarioID := r.PathValue("comentarioId")
	if !reUUID.MatchString(id) || !reUUID.MatchString(comentarioID) {
		http.Error(w, "comentario no encontrado", http.StatusNotFound)
		return
	}

	// Se exige también que el comentario sea de esa publicación
	var tag pgconn.CommandTag
	if u.EsAdmin {
		tag, err = pool.Exec(r.Context(),
			"DELETE FROM comentarios WHERE id = $1 AND publicacion_id = $2",
			comentarioID, id)
	} else {
		tag, err = pool.Exec(r.Context(),
			"DELETE FROM comentarios WHERE id = $1 AND publicacion_id = $2 AND autor_id = $3",
			comentarioID, id, u.ID)
	}
	if err != nil {
		http.Error(w, "error borrando el comentario", http.StatusInternalServerError)
		return
	}
	if tag.RowsAffected() == 0 {
		http.Error(w, "no encontrado, o no es tuyo", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
