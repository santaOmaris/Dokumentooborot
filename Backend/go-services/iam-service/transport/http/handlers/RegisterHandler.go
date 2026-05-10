package handlers

import (
	"encoding/json"
	"net/http"
	"net/mail"
	"strings"

	"docflow.local/pkg/auth"
	"docflow.local/pkg/httputil"
	db "iam-service/db/generated"
	"iam-service/service"
	"iam-service/dto"
)

func RegisterHandler(q *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httputil.Error(w, http.StatusBadRequest, "invalid json payload")
			return
		}

		if len(req.Password) < 8 {
			httputil.Error(w, http.StatusBadRequest, "password must be at least 8 characters")
			return
		}
		if _, err := mail.ParseAddress(req.Email); err != nil {
			httputil.Error(w, http.StatusBadRequest, "invalid email format")
			return
		}
		if strings.ContainsAny(req.Login, " \t") || req.Login == "" {
			httputil.Error(w, http.StatusBadRequest, "login must not be empty or contain spaces")
			return
		}
		if req.SystemRole != "USER" && req.SystemRole != "ADMIN" {
			httputil.Error(w, http.StatusBadRequest, "system_role must be USER or ADMIN")
			return
		}
		parts := strings.Fields(req.FullName)
		if len(parts) < 3 {
			httputil.Error(w, http.StatusBadRequest, "full_name must contain at least 3 words")
			return
		}

		if req.IsHead || req.SystemRole == "ADMIN" {
			if !auth.HasRole(r.Context(), "ADMIN") {
				httputil.Error(w, http.StatusForbidden, "only admins can create privileged accounts")
				return
			}
		}

		id, err := service.RegisterUser(r.Context(), q,
			req.Login, req.Password, req.Email, req.FullName,
			req.IsHead, req.SystemRole,
		)
		if err != nil {
			if strings.Contains(err.Error(), "login already exists") {
				httputil.Error(w, http.StatusConflict, "login already exists")
			} else {
				httputil.Error(w, http.StatusInternalServerError, "internal error")
			}
			return
		}

		httputil.Created(w, dto.RegisterResponse{UserID: id})
	}
}

