package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"docflow.local/pkg/auth"
	"docflow.local/pkg/httputil"
	db "iam-service/db/generated"
	"iam-service/service"
	"iam-service/dto"
)

func FireHandler(q *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.FireRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httputil.Error(w, http.StatusBadRequest, "invalid json payload")
			return
		}
		if req.UserID == 0 {
			httputil.Error(w, http.StatusBadRequest, "user_id is required")
			return
		}

		ctx := r.Context()
		callerLogin, _ := auth.UserLoginFromContext(ctx)
		callerDeptID, _ := auth.DepartmentIDFromContext(ctx)
		callerIsHead, _ := auth.IsHeadFromContext(ctx)
		callerRole, _ := auth.SystemRoleFromContext(ctx)

		if err := service.FireUser(ctx, q, callerLogin, callerDeptID, callerIsHead, callerRole, req.UserID); err != nil {
			if strings.HasPrefix(err.Error(), "forbidden") {
				httputil.Error(w, http.StatusForbidden, err.Error())
			} else {
				httputil.Error(w, http.StatusBadRequest, err.Error())
			}
			return
		}

		httputil.OK(w, map[string]string{"status": "ok"})
	}
}

func MoveUserHandler(q *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !auth.HasRole(r.Context(), "ADMIN") {
			httputil.Error(w, http.StatusForbidden, "admin only")
			return
		}
		var req dto.MoveUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httputil.Error(w, http.StatusBadRequest, "invalid json payload")
			return
		}
		if err := service.MoveUserToDepartment(r.Context(), q, req.UserID, req.DepartmentID); err != nil {
			httputil.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		httputil.OK(w, map[string]string{"status": "ok"})
	}
}

func PromoteHandler(q *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !auth.HasRole(r.Context(), "ADMIN") {
			httputil.Error(w, http.StatusForbidden, "admin only")
			return
		}
		id, err := parseIDParam(r, "id")
		if err != nil {
			httputil.Error(w, http.StatusBadRequest, "invalid user id")
			return
		}
		if err := service.PromoteUser(r.Context(), q, id); err != nil {
			httputil.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		httputil.OK(w, map[string]string{"status": "ok"})
	}
}

func DemoteHandler(q *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !auth.HasRole(r.Context(), "ADMIN") {
			httputil.Error(w, http.StatusForbidden, "admin only")
			return
		}
		id, err := parseIDParam(r, "id")
		if err != nil {
			httputil.Error(w, http.StatusBadRequest, "invalid user id")
			return
		}
		if err := service.DemoteUser(r.Context(), q, id); err != nil {
			httputil.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		httputil.OK(w, map[string]string{"status": "ok"})
	}
}

func parseIDParam(r *http.Request, name string) (int32, error) {
	v := r.PathValue(name)
	n, err := strconv.ParseInt(v, 10, 32)
	return int32(n), err
}

