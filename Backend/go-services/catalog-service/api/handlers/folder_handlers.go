package handlers

import (
	"net/http"
	"strconv"

	"docflow.local/pkg/auth"
	"docflow.local/pkg/httputil"
	db "catalog-service/db/generated"
	"catalog-service/dto"
	"catalog-service/service"
)

func ListFoldersHandler(q *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		deptID, err := pathInt32(r, "dept_id")
		if err != nil {
			httputil.Error(w, http.StatusBadRequest, "invalid dept_id")
			return
		}

		if err := service.InitDepartmentFolders(r.Context(), q, deptID); err != nil {
			httputil.Error(w, http.StatusInternalServerError, err.Error())
			return
		}

		folders, err := service.ListFoldersByDepartment(r.Context(), q, deptID)
		if err != nil {
			httputil.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		httputil.OK(w, folders)
	}
}

func CreateFolderHandler(q *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		deptID, err := pathInt32(r, "dept_id")
		if err != nil {
			httputil.Error(w, http.StatusBadRequest, "invalid dept_id")
			return
		}
		var req dto.CreateFolderRequest
		if err := decodeJSON(r, &req); err != nil || req.Name == "" {
			httputil.Error(w, http.StatusBadRequest, "name is required")
			return
		}
		folder, err := service.CreateFolder(r.Context(), q, deptID, req.ParentID, req.Name)
		if err != nil {
			httputil.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		httputil.Created(w, folder)
	}
}

func DeleteFolderHandler(q *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := pathInt32(r, "id")
		if err != nil {
			httputil.Error(w, http.StatusBadRequest, "invalid id")
			return
		}
		if err := service.DeleteFolder(r.Context(), q, id); err != nil {
			httputil.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		httputil.NoContent(w)
	}
}

func RenameFolderHandler(q *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := pathInt32(r, "id")
		if err != nil {
			httputil.Error(w, http.StatusBadRequest, "invalid id")
			return
		}
		var req dto.RenameFolderRequest
		if err := decodeJSON(r, &req); err != nil || req.Name == "" {
			httputil.Error(w, http.StatusBadRequest, "name is required")
			return
		}
		if err := service.RenameFolder(r.Context(), q, id, req.Name); err != nil {
			httputil.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		httputil.OK(w, map[string]string{"status": "ok"})
	}
}

func MoveFolderHandler(q *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := pathInt32(r, "id")
		if err != nil {
			httputil.Error(w, http.StatusBadRequest, "invalid id")
			return
		}
		var req dto.MoveFolderRequest
		if err := decodeJSON(r, &req); err != nil {
			httputil.Error(w, http.StatusBadRequest, "invalid json")
			return
		}
		callerLogin, _ := auth.UserLoginFromContext(r.Context())
		_ = callerLogin
		if err := service.MoveFolder(r.Context(), q, id, req.ParentID); err != nil {
			httputil.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		httputil.OK(w, map[string]string{"status": "ok"})
	}
}

func InitDepartmentFoldersHandler(q *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !auth.HasRole(r.Context(), "ADMIN") {
			httputil.Error(w, http.StatusForbidden, "admin only")
			return
		}
		deptStr := r.PathValue("dept_id")
		n, err := strconv.ParseInt(deptStr, 10, 32)
		if err != nil {
			httputil.Error(w, http.StatusBadRequest, "invalid dept_id")
			return
		}
		if err := service.InitDepartmentFolders(r.Context(), q, int32(n)); err != nil {
			httputil.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		httputil.OK(w, map[string]string{"status": "ok"})
	}
}
