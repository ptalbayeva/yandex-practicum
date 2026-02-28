package g

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type contextKey string

const userIDContextKey contextKey = "userID"

const (
	cookieName = "authorization"
	tokenTTL   = 3 * time.Hour
)

// Claims данные, используемые при генерации токена
type Claims struct {
	jwt.RegisteredClaims
	UserID string `json:"user_id"`
}

// UserIDFromContext получение user_id из контекста
func UserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDContextKey).(string)

	return userID, ok
}

// Auth авторизация пользователя
func Auth(key []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cookie, err := r.Cookie(cookieName); err == nil {
				if claims, err := parseToken(cookie.Value, key); err == nil && claims.UserID != "" {
					ctx := context.WithValue(r.Context(), userIDContextKey, claims.UserID)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}

			userID := uuid.NewString()

			token, err := createToken(userID, key)
			if err != nil {
				http.Error(w, "could not create token", http.StatusInternalServerError)
				return
			}

			setAuthCookie(w, token)

			ctx := context.WithValue(r.Context(), userIDContextKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func createToken(userID string, key []byte) (string, error) {
	now := time.Now()

	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(tokenTTL)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(key)
}

func parseToken(tokenStr string, key []byte) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("failed to validate token")
		}

		return key, nil
	})

	if err != nil || !token.Valid {
		return nil, err
	}

	return claims, nil
}

func setAuthCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    token,
		Expires:  time.Now().Add(tokenTTL),
		MaxAge:   int(tokenTTL.Seconds()),
		HttpOnly: true,
	})
}
