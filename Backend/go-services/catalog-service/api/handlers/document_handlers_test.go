package handlers

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"docflow.local/pkg/auth"
	db "catalog-service/db/generated"

	"github.com/DATA-DOG/go-sqlmock"
)

func buildMultipartRequest(t *testing.T, fields map[string]string) (*bytes.Buffer, string) {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for k, v := range fields {
		if err := writer.WriteField(k, v); err != nil {
			t.Fatalf("WriteField error: %v", err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("writer.Close error: %v", err)
	}
	return body, writer.FormDataContentType()
}

func TestUploadDocumentHandlerForbiddenForForeignDepartment(t *testing.T) {
	h := &DocumentHandlers{}
	handler := auth.AuthMiddleware(h.UploadDocumentHandler())

	token, err := auth.GenerateToken(4, "user2", "USER", false, "3")
	if err != nil {
		t.Fatalf("GenerateToken error: %v", err)
	}

	body, contentType := buildMultipartRequest(t, map[string]string{
		"title":         "forbidden",
		"folder_id":     "17",
		"department_id": "2",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/catalog/documents", body)
	req.Header.Set("Content-Type", contentType)
	req.AddCookie(&http.Cookie{Name: "jwt_token", Value: token})

	rr := httptest.NewRecorder()
	handler(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("unexpected status: %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestUploadDocumentHandlerForbiddenForHeadOnlyFolder(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New error: %v", err)
	}
	defer sqlDB.Close()

	q := db.New(sqlDB)
	h := &DocumentHandlers{Q: q}
	handler := auth.AuthMiddleware(h.UploadDocumentHandler())

	token, err := auth.GenerateToken(4, "user2", "USER", false, "3")
	if err != nil {
		t.Fatalf("GenerateToken error: %v", err)
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, department_id, parent_id, name, is_system FROM folders WHERE id = $1")).
		WithArgs(int32(17)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "department_id", "parent_id", "name", "is_system"}).
			AddRow(int32(17), int32(3), nil, "head_only", true))

	body, contentType := buildMultipartRequest(t, map[string]string{
		"title":         "forbidden",
		"folder_id":     "17",
		"department_id": "3",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/catalog/documents", body)
	req.Header.Set("Content-Type", contentType)
	req.AddCookie(&http.Cookie{Name: "jwt_token", Value: token})

	rr := httptest.NewRecorder()
	handler(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("unexpected status: %d body=%s", rr.Code, rr.Body.String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestUploadDocumentHandlerValidationErrors(t *testing.T) {
	h := &DocumentHandlers{}

	body1, ct1 := buildMultipartRequest(t, map[string]string{"folder_id": "1", "department_id": "1"})
	req1 := httptest.NewRequest(http.MethodPost, "/api/catalog/documents", body1)
	req1.Header.Set("Content-Type", ct1)
	rr1 := httptest.NewRecorder()
	h.UploadDocumentHandler()(rr1, req1)
	if rr1.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status for missing title: %d", rr1.Code)
	}

	body2, ct2 := buildMultipartRequest(t, map[string]string{"title": "x", "department_id": "1"})
	req2 := httptest.NewRequest(http.MethodPost, "/api/catalog/documents", body2)
	req2.Header.Set("Content-Type", ct2)
	rr2 := httptest.NewRecorder()
	h.UploadDocumentHandler()(rr2, req2)
	if rr2.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status for missing folder_id: %d", rr2.Code)
	}

	body3, ct3 := buildMultipartRequest(t, map[string]string{"title": "x", "folder_id": "1"})
	req3 := httptest.NewRequest(http.MethodPost, "/api/catalog/documents", body3)
	req3.Header.Set("Content-Type", ct3)
	rr3 := httptest.NewRecorder()
	h.UploadDocumentHandler()(rr3, req3)
	if rr3.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status for missing department_id: %d", rr3.Code)
	}

	body4, ct4 := buildMultipartRequest(t, map[string]string{"title": "x", "folder_id": "1", "department_id": "a"})
	req4 := httptest.NewRequest(http.MethodPost, "/api/catalog/documents", body4)
	req4.Header.Set("Content-Type", ct4)
	rr4 := httptest.NewRecorder()
	h.UploadDocumentHandler()(rr4, req4)
	if rr4.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status for bad department_id: %d", rr4.Code)
	}
}

func TestDownloadAndGetHandlersInvalidID(t *testing.T) {
	h := &DocumentHandlers{}

	rr1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodGet, "/api/catalog/documents/bad/download", nil)
	req1.SetPathValue("id", "bad")
	h.DownloadDocumentHandler()(rr1, req1)
	if rr1.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status for download bad id: %d", rr1.Code)
	}

	rr2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/api/catalog/documents/bad", nil)
	req2.SetPathValue("id", "bad")
	GetDocumentHandler(nil)(rr2, req2)
	if rr2.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status for get bad id: %d", rr2.Code)
	}
}

func TestFolderHistorySearchAndMoveValidation(t *testing.T) {
	rr1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodGet, "/api/catalog/folders/bad/documents", nil)
	req1.SetPathValue("id", "bad")
	ListDocumentsByFolderHandler(nil)(rr1, req1)
	if rr1.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status for list by folder bad id: %d", rr1.Code)
	}

	rr2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/api/catalog/documents/bad/history", nil)
	req2.SetPathValue("id", "bad")
	GetDocumentHistoryHandler(nil)(rr2, req2)
	if rr2.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status for history bad id: %d", rr2.Code)
	}

	rr3 := httptest.NewRecorder()
	req3 := httptest.NewRequest(http.MethodGet, "/api/catalog/departments/bad/search?q=x", nil)
	req3.SetPathValue("dept_id", "bad")
	SearchDocumentsHandler(nil)(rr3, req3)
	if rr3.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status for search bad dept id: %d", rr3.Code)
	}

	rr4 := httptest.NewRecorder()
	req4 := httptest.NewRequest(http.MethodGet, "/api/catalog/departments/1/search", nil)
	req4.SetPathValue("dept_id", "1")
	SearchDocumentsHandler(nil)(rr4, req4)
	if rr4.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status for missing query: %d", rr4.Code)
	}

	rr5 := httptest.NewRecorder()
	req5 := httptest.NewRequest(http.MethodPatch, "/api/catalog/documents/bad/move", strings.NewReader(`{"folder_id":1}`))
	req5.SetPathValue("id", "bad")
	MoveDocumentHandler(nil)(rr5, req5)
	if rr5.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status for move bad id: %d", rr5.Code)
	}

	rr6 := httptest.NewRecorder()
	req6 := httptest.NewRequest(http.MethodPatch, "/api/catalog/documents/1/move", strings.NewReader(`{"folder_id"`))
	req6.SetPathValue("id", "1")
	MoveDocumentHandler(nil)(rr6, req6)
	if rr6.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status for move bad json: %d", rr6.Code)
	}
}

func TestHideUnhideValidationAndPermissions(t *testing.T) {
	h := &DocumentHandlers{}

	rr1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodPatch, "/api/catalog/documents/bad/hide", nil)
	req1.SetPathValue("id", "bad")
	h.HideDocumentHandler()(rr1, req1)
	if rr1.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status for hide bad id: %d", rr1.Code)
	}

	rr2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPatch, "/api/catalog/documents/1/hide", nil)
	req2.SetPathValue("id", "1")
	h.HideDocumentHandler()(rr2, req2)
	if rr2.Code != http.StatusForbidden {
		t.Fatalf("unexpected status for hide without head/admin: %d", rr2.Code)
	}

	rr3 := httptest.NewRecorder()
	req3 := httptest.NewRequest(http.MethodPatch, "/api/catalog/documents/1/unhide", nil)
	req3.SetPathValue("id", "1")
	UnhideDocumentHandler(nil)(rr3, req3)
	if rr3.Code != http.StatusForbidden {
		t.Fatalf("unexpected status for unhide without admin: %d", rr3.Code)
	}
}
