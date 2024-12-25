package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

func createJWTToken(password string) (string, error) {
	secret := []byte(password)

	jwtToken := jwt.New(jwt.SigningMethodHS256)

	signedToken, err := jwtToken.SignedString(secret)
	if err != nil {
		return "", fmt.Errorf(`{"error": Authorization error: %w}`, err)
	}

	return signedToken, nil
}

func verifyToken(token string, password string) bool {
	secret := []byte(password)
	jwtToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf(`{"error: "Wrong password"}`)
		}
		return secret, nil
	})
	if err != nil {
		return false
	}
	if !jwtToken.Valid {
		return false
	}

	return true
}

func (app *App) SignInHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	pass := os.Getenv("TODO_PASSWORD")

	m := map[string]string{}
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, `{"error": Server error}`, http.StatusInternalServerError)
		return
	}

	if v, ok := m["password"]; !ok {
		http.Error(w, `{"error": Server error}`, http.StatusInternalServerError)
		return
	} else if v != pass {
		http.Error(w, `{"error": Wrong password}`, http.StatusUnauthorized)
		return
	}

	token, err := createJWTToken(pass)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	response := map[string]any{"token": token}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, `{"error": Server error}`, http.StatusInternalServerError)
		return
	}
}

func auth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// смотрим наличие пароля
		pass := os.Getenv("TODO_PASSWORD")
		if len(pass) > 0 {
			var jwt string // JWT-токен из куки
			// получаем куку
			cookie, err := r.Cookie("token")
			if err == nil {
				jwt = cookie.Value
			}
			var valid bool
			// здесь код для валидации и проверки JWT-токена
			valid = verifyToken(jwt, pass)

			if !valid {
				// возвращаем ошибку авторизации 401
				http.Error(w, "Authentification required", http.StatusUnauthorized)
				return
			}
		}
		next(w, r)
	})
}
