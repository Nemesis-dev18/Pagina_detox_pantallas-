package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/smtp"
	"os"
	"time"
)

var clienteHTTP = &http.Client{Timeout: 15 * time.Second}

// enviarCorreo manda un correo de texto a CUALQUIER dirección (gmail, hotmail,
// outlook, yahoo...). El proveedor del destinatario no importa: siempre sale
// desde un mismo remitente.
//
//   - En producción (Render): si existe BREVO_API_KEY, se envía por la API web
//     de Brevo (HTTPS). Render bloquea los puertos SMTP en el plan gratis.
//   - En tu computador: sin BREVO_API_KEY se usa SMTP (Mailpit en el puerto 1025).
func enviarCorreo(destino, asunto, cuerpo string) error {
	if clave := os.Getenv("BREVO_API_KEY"); clave != "" {
		return enviarPorBrevo(clave, destino, asunto, cuerpo)
	}
	return enviarPorSMTP(destino, asunto, cuerpo)
}

func enviarPorBrevo(clave, destino, asunto, cuerpo string) error {
	remitente := os.Getenv("EMAIL_REMITENTE") // debe estar verificado en Brevo
	if remitente == "" {
		return fmt.Errorf("falta la variable EMAIL_REMITENTE")
	}
	nombre := os.Getenv("EMAIL_NOMBRE")
	if nombre == "" {
		nombre = "DETOX_WEB"
	}

	datos := map[string]any{
		"sender":      map[string]string{"name": nombre, "email": remitente},
		"to":          []map[string]string{{"email": destino}},
		"subject":     asunto,
		"textContent": cuerpo,
	}
	cuerpoJSON, err := json.Marshal(datos)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, "https://api.brevo.com/v3/smtp/email", bytes.NewReader(cuerpoJSON))
	if err != nil {
		return err
	}
	req.Header.Set("api-key", clave)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := clienteHTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		detalle, _ := io.ReadAll(io.LimitReader(resp.Body, 500))
		return fmt.Errorf("brevo respondió %d: %s", resp.StatusCode, detalle)
	}
	return nil
}

func enviarPorSMTP(destino, asunto, cuerpo string) error {
	host := os.Getenv("SMTP_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("SMTP_PORT")
	if port == "" {
		port = "1025"
	}
	from := os.Getenv("SMTP_FROM")
	if from == "" {
		from = "no-responder@detox-web.local"
	}

	asunto = mime.QEncoding.Encode("UTF-8", asunto) // para que la ñ y las tildes se vean bien
	mensaje := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		from, destino, asunto, cuerpo)

	return smtp.SendMail(host+":"+port, nil, from, []string{destino}, []byte(mensaje))
}