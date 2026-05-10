package handlers

import (
	"net/http"

	"docflow.local/pkg/auth"
	"docflow.local/pkg/httputil"
)

func MeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := auth.UserIDIntFromContext(r.Context())
		if !ok {
			httputil.Error(w, http.StatusUnauthorized, "not authenticated")
			return
		}

		login, _ := auth.UserLoginFromContext(r.Context())
		role, _ := auth.SystemRoleFromContext(r.Context())
		isHead, _ := auth.IsHeadFromContext(r.Context())
		deptID, _ := auth.DepartmentIDFromContext(r.Context())

		httputil.OK(w, map[string]any{
			"user_id":       userID,
			"login":         login,
			"system_role":   role,
			"is_head":       isHead,
			"department_id": deptID,
		})
	}
}
