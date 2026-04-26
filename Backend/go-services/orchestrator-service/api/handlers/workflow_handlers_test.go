package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"docflow.local/pkg/auth"
)

func TestGetStatusHandlerInvalidID(t *testing.T) {
	h := &WorkflowHandlers{}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/orchestrator/documents/bad/status", nil)
	req.SetPathValue("id", "bad")
	h.GetStatusHandler()(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
}

func TestGetHistoryHandlerInvalidID(t *testing.T) {
	h := &WorkflowHandlers{}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/orchestrator/documents/bad/history", nil)
	req.SetPathValue("id", "bad")
	h.GetHistoryHandler()(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
}

func TestSendForVisaHandlerInvalidID(t *testing.T) {
	h := &WorkflowHandlers{}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/orchestrator/documents/bad/send-for-visa", nil)
	req.SetPathValue("id", "bad")
	h.SendForVisaHandler()(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
}

func TestSendForVisaHandlerInvalidJSON(t *testing.T) {
	h := &WorkflowHandlers{}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/orchestrator/documents/1/send-for-visa", strings.NewReader(`{"note"`))
	req.SetPathValue("id", "1")
	h.SendForVisaHandler()(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
}

func TestApproveHandlerInvalidID(t *testing.T) {
	h := &WorkflowHandlers{}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/orchestrator/documents/bad/approve", nil)
	req.SetPathValue("id", "bad")
	h.ApproveHandler()(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
}

func TestRejectHandlerInvalidInput(t *testing.T) {
	h := &WorkflowHandlers{}

	rr1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodPost, "/api/orchestrator/documents/bad/reject", nil)
	req1.SetPathValue("id", "bad")
	h.RejectHandler()(rr1, req1)
	if rr1.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status for bad id: %d", rr1.Code)
	}

	rr2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPost, "/api/orchestrator/documents/1/reject", strings.NewReader(`{"revision_note"`))
	req2.SetPathValue("id", "1")
	h.RejectHandler()(rr2, req2)
	if rr2.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status for bad json: %d", rr2.Code)
	}
}

func TestRequestApprovalHandlerInvalidInput(t *testing.T) {
	h := &WorkflowHandlers{}

	rr1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodPost, "/api/orchestrator/documents/bad/request-approval", nil)
	req1.SetPathValue("id", "bad")
	h.RequestApprovalHandler()(rr1, req1)
	if rr1.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status for bad id: %d", rr1.Code)
	}

	rr2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPost, "/api/orchestrator/documents/1/request-approval", strings.NewReader(`{"question"`))
	req2.SetPathValue("id", "1")
	h.RequestApprovalHandler()(rr2, req2)
	if rr2.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status for bad json: %d", rr2.Code)
	}
}

func TestDelegateHandlerForbiddenForRegularUser(t *testing.T) {
	h := &WorkflowHandlers{}
	handler := auth.AuthMiddleware(h.DelegateHandler())

	token, err := auth.GenerateToken(4, "user2", "USER", false, "3")
	if err != nil {
		t.Fatalf("GenerateToken error: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/orchestrator/documents/3/delegate", strings.NewReader(`{"assignee_id":2}`))
	req.SetPathValue("id", "3")
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "jwt_token", Value: token})

	rr := httptest.NewRecorder()
	handler(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("unexpected status: %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestDelegateHandlerBadRequestWithoutAssignee(t *testing.T) {
	h := &WorkflowHandlers{}
	handler := auth.AuthMiddleware(h.DelegateHandler())

	token, err := auth.GenerateToken(1, "admin", "ADMIN", false, "1")
	if err != nil {
		t.Fatalf("GenerateToken error: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/orchestrator/documents/3/delegate", strings.NewReader(`{}`))
	req.SetPathValue("id", "3")
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "jwt_token", Value: token})

	rr := httptest.NewRecorder()
	handler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d body=%s", rr.Code, rr.Body.String())
	}
}
