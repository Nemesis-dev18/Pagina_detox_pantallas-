package main

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
)

// EmailConfig contiene la configuración SMTP leída desde el entorno
type EmailConfig struct {
	Host string
	Port string
	User string
	Pass string
	From string
}

// getEmailConfigFromEnv lee los valores desde variables de entorno
func getEmailConfigFromEnv() EmailConfig {
	return EmailConfig{
		Host: os.Getenv("SMTP_HOST"),
		Port: os.Getenv("SMTP_PORT"),
		User: os.Getenv("SMTP_USER"),
		Pass: os.Getenv("SMTP_PASS"),
		From: os.Getenv("FROM_EMAIL"),
	}
}

// sendEmail envía un email simple usando net/smtp
func sendEmail(cfg EmailConfig, to, subject, body string) error {
	if cfg.Host == "" || cfg.Port == "" || cfg.From == "" {
		return fmt.Errorf("email configuration incomplete")
	}
	addr := cfg.Host + ":" + cfg.Port
	auth := smtp.PlainAuth("", cfg.User, cfg.Pass, cfg.Host)
	msg := []byte("To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" + body + "\r\n")
	if err := smtp.SendMail(addr, auth, cfg.From, []string{to}, msg); err != nil {
		return err
	}
	return nil
}

// sendPasswordReset envía el correo con el link de reset (no bloqueante)
func sendPasswordReset(cfg EmailConfig, to, link string) {
	subject := "Restablecer contraseña"
	body := fmt.Sprintf("Si solicitaste restablecer tu contraseña, haz clic aquí: %s\n\nSi no fuiste tú, ignora este correo.", link)
	if err := sendEmail(cfg, to, subject, body); err != nil {
		log.Printf("failed sending reset email to %s: %v", to, err)
	}
}