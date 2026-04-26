package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFolderHandlersValidation(t *testing.T) {
	rr1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodGet, "/api/catalog/departments/bad/folders", nil)
	req1.SetPathValue("dept_id", "bad")
	ListFoldersHandler(nil)(rr1, req1)
	if rr1.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status for list folders bad dept: %d", rr1.Code)
	}

	rr2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPost, "/api/catalog/departments/bad/folders", nil)
	req2.SetPathValue("dept_id", "bad")
	CreateFolderHandler(nil)(rr2, req2)
	if rr2.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status for create folder bad dept: %d", rr2.Code)
	}

	rr3 := httptest.NewRecorder()
	req3 := httptest.NewRequest(http.MethodDelete, "/api/catalog/folders/bad", nil)
	req3.SetPathValue("id", "bad")
	DeleteFolderHandler(nil)(rr3, req3)
	if rr3.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status for delete folder bad id: %d", rr3.Code)
	}

	rr4 := httptest.NewRecorder()
	req4 := httptest.NewRequest(http.MethodPatch, "/api/catalog/folders/bad/rename", strings.NewReader(`{"name":"x"}`))
	req4.SetPathValue("id", "bad")
	RenameFolderHandler(nil)(rr4, req4)
	if rr4.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status for rename folder bad id: %d", rr4.Code)
	}

	rr5 := httptest.NewRecorder()
	req5 := httptest.NewRequest(http.MethodPatch, "/api/catalog/folders/1/rename", strings.NewReader(`{"name"`))
	req5.SetPathValue("id", "1")
	RenameFolderHandler(nil)(rr5, req5)
	if rr5.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status for rename folder bad json: %d", rr5.Code)
	}

	rr6 := httptest.NewRecorder()
	req6 := httptest.NewRequest(http.MethodPatch, "/api/catalog/folders/bad/move", strings.NewReader(`{"parent_id":1}`))
	req6.SetPathValue("id", "bad")
	MoveFolderHandler(nil)(rr6, req6)
	if rr6.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status for move folder bad id: %d", rr6.Code)
	}

	rr7 := httptest.NewRecorder()
	req7 := httptest.NewRequest(http.MethodPatch, "/api/catalog/folders/1/move", strings.NewReader(`{"parent_id"`))
	req7.SetPathValue("id", "1")
	MoveFolderHandler(nil)(rr7, req7)
	if rr7.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status for move folder bad json: %d", rr7.Code)
	}
}

func TestInitDepartmentFoldersHandlerPermissionsAndValidation(t *testing.T) {
	rr1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodPost, "/api/catalog/departments/1/init", nil)
	req1.SetPathValue("dept_id", "1")
	InitDepartmentFoldersHandler(nil)(rr1, req1)
	if rr1.Code != http.StatusForbidden {
		t.Fatalf("unexpected status without admin role: %d", rr1.Code)
	}
}

func TestDocumentTypeHandlersValidation(t *testing.T) {
	rr1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodPost, "/api/catalog/types", strings.NewReader(`{"name"`))
	CreateDocumentTypeHandler(nil)(rr1, req1)
	if rr1.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status for create type bad json: %d", rr1.Code)
	}

	rr2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodDelete, "/api/catalog/types/bad", nil)
	req2.SetPathValue("id", "bad")
	DeleteDocumentTypeHandler(nil)(rr2, req2)
	if rr2.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status for delete type bad id: %d", rr2.Code)
	}
}
