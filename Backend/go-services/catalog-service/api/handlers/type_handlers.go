package handlers

import (
	"net/http"

	db "catalog-service/db/generated"
	"catalog-service/service"

	"docflow.local/pkg/httputil"
)

func ListDocumentTypesHandler(q *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		types, err := service.ListDocumentTypes(r.Context(), q)
		if err != nil {
			httputil.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		httputil.OK(w, types)
	}
}

func CreateDocumentTypeHandler(q *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Name string `json:"name"`
		}
		if err := decodeJSON(r, &req); err != nil || req.Name == "" {
			httputil.Error(w, http.StatusBadRequest, "name is required")
			return
		}
		t, err := service.CreateDocumentType(r.Context(), q, req.Name)
		if err != nil {
			httputil.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		httputil.Created(w, t)
	}
}

func DeleteDocumentTypeHandler(q *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := pathInt32(r, "id")
		if err != nil {
			httputil.Error(w, http.StatusBadRequest, "invalid id")
			return
		}
		if err := service.DeleteDocumentType(r.Context(), q, id); err != nil {
			httputil.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		httputil.NoContent(w)
	}
}
