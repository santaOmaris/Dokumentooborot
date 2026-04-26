package handlers

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"docflow.local/pkg/auth"
	"docflow.local/pkg/httputil"
	filepb "docflow.local/pkg/pb/file"
	db "catalog-service/db/generated"
	"catalog-service/dto"
	"catalog-service/service"

	"google.golang.org/grpc"
)

const maxUploadSize = 32 << 20 // 32 MB

type DocumentHandlers struct {
	Q          *db.Queries
	FileClient filepb.FileServiceClient
}

func NewDocumentHandlers(q *db.Queries, fileConn *grpc.ClientConn) *DocumentHandlers {
	return &DocumentHandlers{
		Q:          q,
		FileClient: filepb.NewFileServiceClient(fileConn),
	}
}

func (h *DocumentHandlers) UploadDocumentHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(maxUploadSize); err != nil {
			httputil.Error(w, http.StatusBadRequest, "file too large or invalid form")
			return
		}

		title := r.FormValue("title")
		if title == "" {
			httputil.Error(w, http.StatusBadRequest, "title is required")
			return
		}
		description := r.FormValue("description")
		folderID, err := parseInt32Form(r, "folder_id")
		if err != nil {
			httputil.Error(w, http.StatusBadRequest, "folder_id is required")
			return
		}
		deptID, err := parseInt32Form(r, "department_id")
		if err != nil {
			httputil.Error(w, http.StatusBadRequest, "department_id is required")
			return
		}

		ctx := r.Context()
		isAdmin := auth.HasRole(ctx, "ADMIN")
		isHead, _ := auth.IsHeadFromContext(ctx)
		callerDeptStr, _ := auth.DepartmentIDFromContext(ctx)

		callerDeptID := int32(0)
		if callerDeptStr != "" {
			if parsed, parseErr := strconv.ParseInt(callerDeptStr, 10, 32); parseErr == nil {
				callerDeptID = int32(parsed)
			}
		}

		if !isAdmin {
			if callerDeptID == 0 || deptID != callerDeptID {
				httputil.Error(w, http.StatusForbidden, "forbidden: upload only to your department")
				return
			}
		}

		folder, err := h.Q.GetFolder(ctx, folderID)
		if err != nil {
			httputil.Error(w, http.StatusBadRequest, "folder not found")
			return
		}
		if folder.DepartmentID != deptID {
			httputil.Error(w, http.StatusForbidden, "forbidden: folder is not in target department")
			return
		}
		if folder.IsSystem && folder.Name == "head_only" && !isAdmin && !isHead {
			httputil.Error(w, http.StatusForbidden, "forbidden: head_only folder is restricted")
			return
		}

		var typeID *int32
		if rawType := r.FormValue("type_id"); rawType != "" {
			n, parseErr := strconv.ParseInt(rawType, 10, 32)
			if parseErr != nil {
				httputil.Error(w, http.StatusBadRequest, "invalid type_id")
				return
			}
			t := int32(n)
			typeID = &t
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			httputil.Error(w, http.StatusBadRequest, "file is required")
			return
		}
		defer file.Close()

		content, err := io.ReadAll(io.LimitReader(file, maxUploadSize))
		if err != nil {
			httputil.Error(w, http.StatusInternalServerError, "failed to read file")
			return
		}

		bucket := bucketName(deptID)
		uploadResp, err := h.FileClient.UploadFile(ctx, &filepb.UploadFileRequest{
			Filename: header.Filename,
			Bucket:   bucket,
			Content:  content,
		})
		if err != nil {
			httputil.Error(w, http.StatusInternalServerError, "file upload failed: "+err.Error())
			return
		}

		callerID, _ := auth.UserIDIntFromContext(ctx)

		docID, err := service.CreateDocument(ctx, h.Q, service.UploadDocumentParams{
			Title:        title,
			Description:  description,
			TypeID:       typeID,
			FolderID:     folderID,
			FilePath:     uploadResp.ObjectPath,
			OriginalName: header.Filename,
			DepartmentID: deptID,
			CreatedBy:    callerID,
		})
		if err != nil {
			httputil.Error(w, http.StatusInternalServerError, err.Error())
			return
		}

		httputil.Created(w, map[string]int32{"document_id": docID})
	}
}

func (h *DocumentHandlers) DownloadDocumentHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		docID, err := pathInt32(r, "id")
		if err != nil {
			httputil.Error(w, http.StatusBadRequest, "invalid id")
			return
		}
		doc, err := service.GetDocument(r.Context(), h.Q, docID)
		if err != nil {
			httputil.Error(w, http.StatusNotFound, "document not found")
			return
		}
		if doc.IsHidden {
			if !auth.HasRole(r.Context(), "ADMIN") {
				httputil.Error(w, http.StatusForbidden, "document is hidden")
				return
			}
		}

		bucket := bucketName(doc.DepartmentID)
		resp, err := h.FileClient.DownloadFile(r.Context(), &filepb.DownloadFileRequest{
			Bucket:     bucket,
			ObjectPath: doc.FilePath,
		})
		if err != nil {
			httputil.Error(w, http.StatusInternalServerError, "file download failed: "+err.Error())
			return
		}

		w.Header().Set("Content-Disposition", "attachment; filename="+resp.Filename)
		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
		io.Copy(w, bytes.NewReader(resp.Content))
	}
}

func ListDocumentsByFolderHandler(q *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		folderID, err := pathInt32(r, "id")
		if err != nil {
			httputil.Error(w, http.StatusBadRequest, "invalid id")
			return
		}
		docs, err := service.ListDocumentsByFolder(r.Context(), q, folderID)
		if err != nil {
			httputil.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		httputil.OK(w, docs)
	}
}

func GetDocumentHandler(q *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := pathInt32(r, "id")
		if err != nil {
			httputil.Error(w, http.StatusBadRequest, "invalid id")
			return
		}
		doc, err := service.GetDocument(r.Context(), q, id)
		if err != nil {
			httputil.Error(w, http.StatusNotFound, "not found")
			return
		}
		if doc.IsHidden && !auth.HasRole(r.Context(), "ADMIN") {
			httputil.Error(w, http.StatusForbidden, "document is hidden")
			return
		}
		httputil.OK(w, doc)
	}
}

func MoveDocumentHandler(q *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := pathInt32(r, "id")
		if err != nil {
			httputil.Error(w, http.StatusBadRequest, "invalid id")
			return
		}
		var req dto.MoveDocumentRequest
		if err := decodeJSON(r, &req); err != nil {
			httputil.Error(w, http.StatusBadRequest, "invalid json")
			return
		}
		login, _ := auth.UserLoginFromContext(r.Context())
		if err := service.MoveDocument(r.Context(), q, id, req.FolderID, login); err != nil {
			httputil.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		httputil.OK(w, map[string]string{"status": "ok"})
	}
}

func (h *DocumentHandlers) HideDocumentHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := pathInt32(r, "id")
		if err != nil {
			httputil.Error(w, http.StatusBadRequest, "invalid id")
			return
		}

		ctx := r.Context()
		isHead, _ := auth.IsHeadFromContext(ctx)
		isAdmin := auth.HasRole(ctx, "ADMIN")

		if !isHead && !isAdmin {
			httputil.Error(w, http.StatusForbidden, "only head or admin can hide documents")
			return
		}

		if isAdmin {
			if err := service.HideDocument(ctx, h.Q, id); err != nil {
				httputil.Error(w, http.StatusInternalServerError, err.Error())
				return
			}
		} else {
			doc, err := service.GetDocument(ctx, h.Q, id)
			if err != nil {
				httputil.Error(w, http.StatusNotFound, "document not found")
				return
			}
			login, _ := auth.UserLoginFromContext(ctx)
			if err := service.MoveToHeadOnly(ctx, h.Q, id, doc.DepartmentID, login); err != nil {
				httputil.Error(w, http.StatusInternalServerError, err.Error())
				return
			}
		}

		httputil.OK(w, map[string]string{"status": "ok"})
	}
}

func UnhideDocumentHandler(q *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !auth.HasRole(r.Context(), "ADMIN") {
			httputil.Error(w, http.StatusForbidden, "admin only")
			return
		}
		id, err := pathInt32(r, "id")
		if err != nil {
			httputil.Error(w, http.StatusBadRequest, "invalid id")
			return
		}
		if err := service.UnhideDocument(r.Context(), q, id); err != nil {
			httputil.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		httputil.OK(w, map[string]string{"status": "ok"})
	}
}

func SearchDocumentsHandler(q *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		deptID, err := pathInt32(r, "dept_id")
		if err != nil {
			httputil.Error(w, http.StatusBadRequest, "invalid dept_id")
			return
		}
		query := r.URL.Query().Get("q")
		if query == "" {
			httputil.Error(w, http.StatusBadRequest, "q is required")
			return
		}
		docs, err := service.SearchDocuments(r.Context(), q, deptID, query)
		if err != nil {
			httputil.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		httputil.OK(w, docs)
	}
}

func GetDocumentHistoryHandler(q *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := pathInt32(r, "id")
		if err != nil {
			httputil.Error(w, http.StatusBadRequest, "invalid id")
			return
		}
		history, err := service.GetDocumentHistory(r.Context(), q, id)
		if err != nil {
			httputil.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		httputil.OK(w, history)
	}
}


func bucketName(deptID int32) string {
	return fmt.Sprintf("dept-%d", deptID)
}

func parseInt32Form(r *http.Request, key string) (int32, error) {
	v := r.FormValue(key)
	n, err := strconv.ParseInt(v, 10, 32)
	return int32(n), err
}
