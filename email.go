package main

import (
	"fmt"
	"mime"
	"net/smtp"
	"os"
)

func enviarCorreo(destino, asunto, cuerpo string) error {
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
