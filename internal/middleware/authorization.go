package middleware

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

const (
	CookieName = "authorization"
	SecretKey  = "secretkey"
	TokenExp   = time.Hour * 12
)

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie(CookieName)

		var userID string

		if err == nil {
			userID = GetUserID(c.Value)
		}

		if userID == "" {
			tokenString, fail := BuildJWTString()
			if fail != nil {
				log.Println("Failed to build JWT token")
				next.ServeHTTP(w, r)
			}

			http.SetCookie(w, &http.Cookie{
				Name:  CookieName,
				Value: tokenString,
			})
		}

		ctx := context.WithValue(r.Context(), "user_id", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func BuildJWTString() (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExp)),
		},
		UserID: newUserID(),
	})

	tokenString, err := token.SignedString([]byte(SecretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func newUserID() string {
	return uuid.NewString()
}

func GetUserID(tokenString string) string {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			return []byte(SecretKey), nil
		})
	if err != nil {
		return ""
	}

	if !token.Valid {
		log.Println("Token is not valid")
		return ""
	}

	log.Println("Token is valid")

	return claims.UserID
}
