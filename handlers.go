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
	ID        string    `json:"id"`
	Alias     string    `json:"alias"`
	Titulo    string    `json:"titulo"`
	Contenido string    `json:"contenido"`
	Categoria string    `json:"categoria"`
	CreadoEn  time.Time `json:"creado_en"`
}

func listarPublicaciones(w http.ResponseWriter, r *http.Request) {
	rows, err := pool.Query(r.Context(), `
		SELECT p.id, u.alias, p.titulo, p.contenido, COALESCE(p.categoria, ''), p.creado_en
		FROM publicaciones p
		JOIN usuarios u ON u.id = p.autor_id
		ORDER BY p.creado_en DESC
		LIMIT 50`)
	if err != nil {
		http.Error(w, "error consultando la base", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	publicaciones := []Publicacion{}
	for rows.Next() {
		var p Publicacion
		if err := rows.Scan(&p.ID, &p.Alias, &p.Titulo, &p.Contenido, &p.Categoria, &p.CreadoEn); err != nil {
			http.Error(w, "error leyendo datos", http.StatusInternalServerError)
			return
		}
		publicaciones = append(publicaciones, p)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, "error leyendo datos", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
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
	return enviarCorreo(correo, "Recuperar contraseña | DETOX_WEB", cuerpo)
}

func recuperar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
		return
	}

	correo := strings.ToLower(strings.TrimSpace(r.FormValue("correo")))

	var usuarioID string
	err := pool.QueryRow(r.Context(), "SELECT id FROM usuarios WHERE correo = $1", correo).Scan(&usuarioID)
	if err == nil {
		if err := enviarEnlaceRecuperacion(r.Context(), usuarioID, correo); err != nil {
			log.Println("recuperación:", err)
		}
	} else if !errors.Is(err, pgx.ErrNoRows) {
		log.Println("recuperación:", err)
	}

	// Siempre la misma respuesta, exista o no el correo
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(w, "Si ese correo está registrado, te enviamos un enlace para restablecer tu contraseña. Revisa tu bandeja de entrada (en desarrollo: http://localhost:8025).")
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

// Se usa cuando el correo no existe, para que el login tarde lo mismo
var hashFalso, _ = bcrypt.GenerateFromPassword([]byte("contraseña-falsa"), bcrypt.DefaultCost)
type Usuario struct {
	ID      string
	Alias   string
	EsAdmin bool
}

// Devuelve quién es el usuario según la cookie, o un error si no hay sesión válida
func usuarioActual(r *http.Request) (*Usuario, error) {
	c, err := r.Cookie("sesion")
	if err != nil {
		return nil, err
	}

	var u Usuario
	err = pool.QueryRow(r.Context(),
		`SELECT u.id, u.alias, u.es_admin
		 FROM sesiones s
		 JOIN usuarios u ON u.id = s.usuario_id
		 WHERE s.token_hash = $1 AND s.expira_en > now()`,
		hashToken(c.Value)).Scan(&u.ID, &u.Alias, &u.EsAdmin)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
func iniciarSesion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
		return
	}

	// El campo del formulario se llama "username", pero trae el correo
	correo := strings.ToLower(strings.TrimSpace(r.FormValue("username")))
	password := r.FormValue("password")

	var id, hash string
	encontrado := true
	err := pool.QueryRow(r.Context(),
		"SELECT id, password_hash FROM usuarios WHERE correo = $1", correo).Scan(&id, &hash)
	if errors.Is(err, pgx.ErrNoRows) {
		encontrado = false
		hash = string(hashFalso)
	} else if err != nil {
		http.Error(w, "error interno", http.StatusInternalServerError)
		return
	}

	coincide := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
	if !encontrado || !coincide {
		http.Error(w, "correo o contraseña incorrectos", http.StatusUnauthorized)
		return
	}

	token, err := generarToken()
	if err != nil {
		http.Error(w, "error interno", http.StatusInternalServerError)
		return
	}
	_, err = pool.Exec(r.Context(),
		"INSERT INTO sesiones (token_hash, usuario_id, expira_en) VALUES ($1, $2, now() + interval '30 days')",
		hashToken(token), id)
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
		// Secure: true, // activar cuando el sitio se publique con HTTPS
	})
	http.Redirect(w, r, "/static/index.html", http.StatusSeeOther)
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
	json.NewEncoder(w).Encode(map[string]interface{}{
		"alias":    u.Alias,
		"es_admin": u.EsAdmin,
	})
}
// ---------- Comentarios ----------

var reUUID = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

type Comentario struct {
	ID        string    `json:"id"`
	Alias     string    `json:"alias"`
	Contenido string    `json:"contenido"`
	CreadoEn  time.Time `json:"creado_en"`
}

func listarComentarios(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !reUUID.MatchString(id) {
		http.Error(w, "publicación no encontrada", http.StatusNotFound)
		return
	}

	rows, err := pool.Query(r.Context(), `
		SELECT c.id, u.alias, c.contenido, c.creado_en
		FROM comentarios c
		JOIN usuarios u ON u.id = c.autor_id
		WHERE c.publicacion_id = $1
		ORDER BY c.creado_en ASC`, id)
	if err != nil {
		http.Error(w, "error consultando la base", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	comentarios := []Comentario{}
	for rows.Next() {
		var c Comentario
		if err := rows.Scan(&c.ID, &c.Alias, &c.Contenido, &c.CreadoEn); err != nil {
			http.Error(w, "error leyendo datos", http.StatusInternalServerError)
			return
		}
		comentarios = append(comentarios, c)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, "error leyendo datos", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
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

	_, err = pool.Exec(r.Context(),
		"INSERT INTO comentarios (publicacion_id, autor_id, contenido) VALUES ($1, $2, $3)",
		id, u.ID, contenido)
	if err != nil {
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