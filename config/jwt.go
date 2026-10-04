package config

import (
	"os"
	"time"
)

var (
	JWTSecret     []byte
	JWTExpiration time.Duration
)

func InitJWT() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "uts-rahasia-super-secret-044"
	}
	JWTSecret = []byte(secret)
	JWTExpiration = 15 * time.Minute
}
