package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"strings"
)

// DATOS_CLAVE: 64 caracteres hexadecimales (32 bytes). Se genera con: openssl rand -hex 32
// Cifra el número de documento de los especialistas. Si se pierde, esos documentos
// no se pueden recuperar; si se filtra, hay que rotarla.
var errSinClave = errors.New("falta DATOS_CLAVE (64 caracteres hexadecimales)")

func claveDatos() ([]byte, error) {
	k, err := hex.DecodeString(strings.TrimSpace(os.Getenv("DATOS_CLAVE")))
	if err != nil || len(k) != 32 {
		return nil, errSinClave
	}
	return k, nil
}

// cifrar devuelve nonce||texto cifrado (AES-256-GCM).
func cifrar(plano string) ([]byte, error) {
	k, err := claveDatos()
	if err != nil {
		return nil, err
	}
	bloque, err := aes.NewCipher(k)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(bloque)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	return gcm.Seal(nonce, nonce, []byte(plano), nil), nil
}

func descifrar(datos []byte) (string, error) {
	k, err := claveDatos()
	if err != nil {
		return "", err
	}
	bloque, err := aes.NewCipher(k)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(bloque)
	if err != nil {
		return "", err
	}
	if len(datos) < gcm.NonceSize() {
		return "", errors.New("dato cifrado inválido")
	}
	plano, err := gcm.Open(nil, datos[:gcm.NonceSize()], datos[gcm.NonceSize():], nil)
	return string(plano), err
}

// huella permite detectar el mismo documento dos veces sin guardarlo en claro.
// Usa una subclave distinta a la de cifrado.
func huella(texto string) (string, error) {
	k, err := claveDatos()
	if err != nil {
		return "", err
	}
	sub := sha256.Sum256(append([]byte("huella:"), k...))
	h := hmac.New(sha256.New, sub[:])
	h.Write([]byte(texto))
	return hex.EncodeToString(h.Sum(nil)), nil
}
