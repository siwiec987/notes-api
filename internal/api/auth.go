package api

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

type contextKey string
const userKey contextKey = "user_id"

var jwtKey []byte

func init() {
	key := os.Getenv("JWT_SECRET")
	if key == "" {
		if testing.Testing() {
			key = "test-secret-key"
		} else {
			log.Fatal("JWT_SECRET not set!")
		}
	}
	jwtKey = []byte(key)
	fmt.Println(key)
}

func generateToken(userID int) (string, error) {
	expirationTime := time.Now().Add(15 * time.Hour)

	claims := &Claims {
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	
	return token.SignedString(jwtKey)
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			sendError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		claims := &Claims {}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (any, error) {
			return jwtKey, nil
		})

		switch {
		case err == nil && token.Valid:
			ctx := context.WithValue(r.Context(), userKey, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
			return

		case errors.Is(err, jwt.ErrTokenExpired):
			sendError(w, http.StatusUnauthorized, "Token expired")
			return

		default:
			sendError(w, http.StatusUnauthorized, "Invalid token")
			return
		}
	}
}