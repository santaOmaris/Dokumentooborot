package handlers

import (
	"errors"
	"net/http"
	"strconv"

	db "catalog-service/db/generated"
	"catalog-service/dto"
	"catalog-service/service"
	"docflow.local/pkg/auth"
	"docflow.local/pkg/httputil"

	"github.com/jackc/pgx/v5/pgconn"
)

func ListFoldersHandler(q *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		deptID, err := pathInt32(r, "dept_id")
		if err != nil {
			httputil.Error(w, http.StatusBadRequest, "invalid dept_id")
			return
		}
		if !auth.HasRole(r.Context(), "ADMIN") {
			callerDept, _ := auth.DepartmentIDFromContext(r.Context())
			if callerDept == "" || callerDept != strconv.FormatInt(int64(deptID), 10) {
				httputil.Error(w, http.StatusForbidden, "forbidden: only your department folders are allowed")
				return
			}
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
		if !auth.HasRole(r.Context(), "ADMIN") {
			callerDept, _ := auth.DepartmentIDFromContext(r.Context())
			if callerDept == "" || callerDept != strconv.FormatInt(int64(deptID), 10) {
				httputil.Error(w, http.StatusForbidden, "forbidden: only your department folders are allowed")
				return
			}
		}
		var req dto.CreateFolderRequest
		if err := decodeJSON(r, &req); err != nil || req.Name == "" {
			httputil.Error(w, http.StatusBadRequest, "name is required")
			return
		}
		if req.ParentID == nil {
			httputil.Error(w, http.StatusBadRequest, "parent_id is required")
			return
		}
		parent, err := q.GetFolder(r.Context(), *req.ParentID)
		if err != nil {
			httputil.Error(w, http.StatusBadRequest, "parent folder not found")
			return
		}
		if parent.DepartmentID != deptID {
			httputil.Error(w, http.StatusForbidden, "forbidden: parent folder must be in same department")
			return
		}
		folder, err := service.CreateFolder(r.Context(), q, deptID, req.ParentID, req.Name)
		if err != nil {
			if isUniqueViolation(err) {
				httputil.Error(w, http.StatusConflict, "folder with this name already exists in parent folder")
				return
			}
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
		folder, err := q.GetFolder(r.Context(), id)
		if err != nil {
			httputil.Error(w, http.StatusNotFound, "folder not found")
			return
		}
		if !auth.HasRole(r.Context(), "ADMIN") {
			callerDept, _ := auth.DepartmentIDFromContext(r.Context())
			isHead, _ := auth.IsHeadFromContext(r.Context())
			if callerDept == "" || callerDept != strconv.FormatInt(int64(folder.DepartmentID), 10) {
				httputil.Error(w, http.StatusForbidden, "forbidden: only your department folders are allowed")
				return
			}
			if folder.IsSystem && folder.Name == "head_only" && !isHead {
				httputil.Error(w, http.StatusForbidden, "forbidden: head_only folder is restricted")
				return
			}
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
		folder, err := q.GetFolder(r.Context(), id)
		if err != nil {
			httputil.Error(w, http.StatusNotFound, "folder not found")
			return
		}
		if !auth.HasRole(r.Context(), "ADMIN") {
			callerDept, _ := auth.DepartmentIDFromContext(r.Context())
			isHead, _ := auth.IsHeadFromContext(r.Context())
			if callerDept == "" || callerDept != strconv.FormatInt(int64(folder.DepartmentID), 10) {
				httputil.Error(w, http.StatusForbidden, "forbidden: only your department folders are allowed")
				return
			}
			if folder.IsSystem && folder.Name == "head_only" && !isHead {
				httputil.Error(w, http.StatusForbidden, "forbidden: head_only folder is restricted")
				return
			}
		}
		if err := service.RenameFolder(r.Context(), q, id, req.Name); err != nil {
			if isUniqueViolation(err) {
				httputil.Error(w, http.StatusConflict, "folder with this name already exists in parent folder")
				return
			}
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
		folder, err := q.GetFolder(r.Context(), id)
		if err != nil {
			httputil.Error(w, http.StatusNotFound, "folder not found")
			return
		}
		if req.ParentID != nil {
			parent, parentErr := q.GetFolder(r.Context(), *req.ParentID)
			if parentErr != nil {
				httputil.Error(w, http.StatusBadRequest, "parent folder not found")
				return
			}
			if parent.DepartmentID != folder.DepartmentID {
				httputil.Error(w, http.StatusForbidden, "forbidden: parent folder must be in same department")
				return
			}
		}
		if !auth.HasRole(r.Context(), "ADMIN") {
			callerDept, _ := auth.DepartmentIDFromContext(r.Context())
			isHead, _ := auth.IsHeadFromContext(r.Context())
			if callerDept == "" || callerDept != strconv.FormatInt(int64(folder.DepartmentID), 10) {
				httputil.Error(w, http.StatusForbidden, "forbidden: only your department folders are allowed")
				return
			}
			if folder.IsSystem && folder.Name == "head_only" && !isHead {
				httputil.Error(w, http.StatusForbidden, "forbidden: head_only folder is restricted")
				return
			}
		}
		callerLogin, _ := auth.UserLoginFromContext(r.Context())
		_ = callerLogin
		if err := service.MoveFolder(r.Context(), q, id, req.ParentID); err != nil {
			if isUniqueViolation(err) {
				httputil.Error(w, http.StatusConflict, "folder with this name already exists in parent folder")
				return
			}
			httputil.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		httputil.OK(w, map[string]string{"status": "ok"})
	}
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
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
