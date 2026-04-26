package handlers

import (
	"encoding/json"
	"net/http"

	"docflow.local/pkg/auth"
	"docflow.local/pkg/httputil"
	db "iam-service/db/generated"
	"iam-service/service"
	"iam-service/dto"
)

func ListDepartmentsHandler(q *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		depts, err := service.ListDepartments(r.Context(), q)
		if err != nil {
			httputil.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		httputil.OK(w, depts)
	}
}

func CreateDepartmentHandler(q *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !auth.HasRole(r.Context(), "ADMIN") {
			httputil.Error(w, http.StatusForbidden, "admin only")
			return
		}
		var req dto.CreateDepartmentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httputil.Error(w, http.StatusBadRequest, "invalid json payload")
			return
		}
		if req.Name == "" {
			httputil.Error(w, http.StatusBadRequest, "name is required")
			return
		}
		id, err := service.CreateDepartment(r.Context(), q, req.Name, req.ParentID)
		if err != nil {
			httputil.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		httputil.Created(w, map[string]int32{"id": id})
	}
}

func SetDepartmentParentHandler(q *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !auth.HasRole(r.Context(), "ADMIN") {
			httputil.Error(w, http.StatusForbidden, "admin only")
			return
		}
		deptID, err := parseIDParam(r, "id")
		if err != nil {
			httputil.Error(w, http.StatusBadRequest, "invalid department id")
			return
		}
		var req dto.SetParentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httputil.Error(w, http.StatusBadRequest, "invalid json payload")
			return
		}
		if err := service.SetDepartmentParent(r.Context(), q, deptID, req.ParentID); err != nil {
			httputil.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		httputil.OK(w, map[string]string{"status": "ok"})
	}
}

func DeleteDepartmentHandler(q *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !auth.HasRole(r.Context(), "ADMIN") {
			httputil.Error(w, http.StatusForbidden, "admin only")
			return
		}
		deptID, err := parseIDParam(r, "id")
		if err != nil {
			httputil.Error(w, http.StatusBadRequest, "invalid department id")
			return
		}
		if err := service.DeleteDepartment(r.Context(), q, deptID); err != nil {
			httputil.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		httputil.NoContent(w)
	}
}
