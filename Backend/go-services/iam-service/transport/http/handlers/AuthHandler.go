package handlers

import (
	"encoding/json"
	"net/http"

	"docflow.local/pkg/httputil"
	db "iam-service/db/generated"
	"iam-service/service"
	"iam-service/dto"
)

func AuthHandler(q *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.AuthRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httputil.Error(w, http.StatusBadRequest, "invalid json payload")
			return
		}
		if req.Login == "" || req.Password == "" {
			httputil.Error(w, http.StatusBadRequest, "login and password are required")
			return
		}

		token, err := service.AuthUser(r.Context(), q, req.Login, req.Password)
		if err != nil {
			httputil.Error(w, http.StatusUnauthorized, "invalid credentials")
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "jwt_token",
			Value:    token,
			Path:     "/",
			MaxAge:   86400,
			HttpOnly: true,
			Secure:   false, // true в проде
			SameSite: http.SameSiteStrictMode,
		})
		httputil.OK(w, map[string]string{"status": "ok"})
	}
}

func LogoutHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:     "jwt_token",
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
			Secure:   false,
			SameSite: http.SameSiteStrictMode,
		})
		httputil.OK(w, map[string]string{"status": "ok"})
	}
}

