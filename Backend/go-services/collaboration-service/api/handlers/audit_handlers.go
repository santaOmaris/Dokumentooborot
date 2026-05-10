package handlers

import (
	"net/http"

	"docflow.local/pkg/auth"
	"docflow.local/pkg/httputil"
	db "collaboration-service/db/generated"
	"collaboration-service/service"
)

func ListAuditByDepartmentHandler(q *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		deptID, err := pathInt32(r, "dept_id")
		if err != nil {
			httputil.Error(w, http.StatusBadRequest, "invalid dept_id")
			return
		}

		if !auth.HasRole(r.Context(), "ADMIN") {
			callerDept, _ := auth.DepartmentIDFromContext(r.Context())
			if callerDept == "" || callerDept != r.PathValue("dept_id") {
				httputil.Error(w, http.StatusForbidden, "forbidden: only your department audit is allowed")
				return
			}
		}

		limit := queryInt32(r, "limit", 50)
		offset := queryInt32(r, "offset", 0)

		logs, err := service.ListAuditByDepartment(r.Context(), q, deptID, limit, offset)
		if err != nil {
			httputil.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		httputil.OK(w, logs)
	}
}

func ListAuditByDocumentHandler(q *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		docID, err := pathInt32(r, "id")
		if err != nil {
			httputil.Error(w, http.StatusBadRequest, "invalid document id")
			return
		}
		logs, err := service.ListAuditByDocument(r.Context(), q, docID)
		if err != nil {
			httputil.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		httputil.OK(w, logs)
	}
}
