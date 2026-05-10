package handlers

import (
	"net/http"

	"docflow.local/pkg/auth"
	"docflow.local/pkg/httputil"
	db "iam-service/db/generated"
	"iam-service/service"
)

func ListUsersHandler(q *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		users, err := service.ListUsers(r.Context(), q)
		if err != nil {
			httputil.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		httputil.OK(w, users)
	}
}

func ListUsersByDeptHandler(q *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		deptID, err := parseIDParam(r, "id")
		if err != nil {
			httputil.Error(w, http.StatusBadRequest, "invalid department id")
			return
		}

		if !auth.HasRole(r.Context(), "ADMIN") {
			callerDept, _ := auth.DepartmentIDFromContext(r.Context())
			if callerDept == "" || callerDept != r.PathValue("id") {
				httputil.Error(w, http.StatusForbidden, "forbidden: only your department is allowed")
				return
			}
		}

		users, err := service.ListUsersByDepartment(r.Context(), q, deptID)
		if err != nil {
			httputil.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		httputil.OK(w, users)
	}
}
