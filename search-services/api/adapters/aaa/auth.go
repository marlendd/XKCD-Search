package aaa

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/marlendd/XKCD-Search/api/core"
)

const adminRole = "superuser" // token subject

// Authentication, Authorization, Accounting
type AAA struct {
	users    map[string]string
	secret   []byte
	tokenTTL time.Duration
	log      *slog.Logger
}

func New(tokenTTL time.Duration, log *slog.Logger) (AAA, error) {
	const adminUser = "ADMIN_USER"
	const adminPass = "ADMIN_PASSWORD"
	const jwtSecret = "JWT_SECRET"
	user := os.Getenv(adminUser)
	if user == "" {
		return AAA{}, fmt.Errorf("could not get admin user from environment")
	}
	password := os.Getenv(adminPass)
	if password == "" {
		return AAA{}, fmt.Errorf("could not get admin password from environment")
	}
	secret := os.Getenv(jwtSecret)
	if secret == "" {
		return AAA{}, fmt.Errorf("could not get JWT secret from environment")
	}

	return AAA{
		users:    map[string]string{user: password},
		secret:   []byte(secret),
		tokenTTL: tokenTTL,
		log:      log,
	}, nil
}

func (a AAA) Login(name, password string) (string, error) {
	storedPassword, ok := a.users[name]
	if !ok || storedPassword != password {
		a.log.Info("wrong login or password")
		return "", core.ErrNotAuthorized
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   adminRole,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(a.tokenTTL)),
	})

	tokenString, err := token.SignedString(a.secret)
	if err != nil {
		return "", core.ErrNotAuthorized
	}

	return tokenString, nil
}

func (a AAA) Verify(tokenString string) error {
	claims := jwt.RegisteredClaims{}

	token, err := jwt.ParseWithClaims(
		tokenString, &claims, func(t *jwt.Token) (any, error) {
			return a.secret, nil
		}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil || !token.Valid || claims.Subject != adminRole {
		a.log.Error("invalid token", "error", err)
		return core.ErrNotAuthorized
	}

	return nil
}
