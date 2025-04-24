package aaa

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	secretKey = "something secret here" // token sign key
	adminRole = "superuser"             // token subject
)

type AAA struct {
	users    map[string]string
	tokenTTL time.Duration
	log      *slog.Logger
}

func New(tokenTTL time.Duration, log *slog.Logger) (AAA, error) {
	const adminUser = "ADMIN_USER"
	const adminPass = "ADMIN_PASSWORD"
	user, ok := os.LookupEnv(adminUser)
	if !ok {
		return AAA{}, fmt.Errorf("could not get admin user from enviroment")
	}
	password, ok := os.LookupEnv(adminPass)
	if !ok {
		return AAA{}, fmt.Errorf("could not get admin password from enviroment")
	}

	return AAA{
		users:    map[string]string{user: password},
		tokenTTL: tokenTTL,
		log:      log,
	}, nil
}

func (a AAA) Login(name, password string) (string, error) {
	if name == "" {
		return "", errors.New("empty user")
	}

	storedPass, exists := a.users[name]
	if !exists || storedPass != password {
		a.log.Warn("invalid login attempt", "user", name)
		return "", errors.New("invalid credentials")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": adminRole,
		"exp": time.Now().Add(a.tokenTTL).Unix(),
	})

	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		a.log.Error("failed to sign token", "error", err)
		return "", fmt.Errorf("token generation failed")
	}

	return tokenString, nil
}

func (a AAA) Verify(tokenString string) error {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(secretKey), nil
	})

	if err != nil || !token.Valid {
		a.log.Warn("invalid token", "error", err)
		return errors.New("invalid token")
	}

	subject, err := token.Claims.GetSubject()
	if err != nil {
		a.log.Error("no subject", "error", err)
		return errors.New("incomplete token")
	}
	if subject != adminRole {
		a.log.Error("not admin", "subject", subject)
		return errors.New("not authorized")
	}
	return nil

}
