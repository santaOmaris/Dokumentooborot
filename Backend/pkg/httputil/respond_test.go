package httputil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOK(t *testing.T) {
	rr := httptest.NewRecorder()
	OK(rr, map[string]string{"k": "v"})

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if rr.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("unexpected content-type: %s", rr.Header().Get("Content-Type"))
	}

	var got map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("json unmarshal error: %v", err)
	}
	if got["k"] != "v" {
		t.Fatalf("unexpected payload: %+v", got)
	}
}

func TestCreated(t *testing.T) {
	rr := httptest.NewRecorder()
	Created(rr, map[string]int{"id": 1})

	if rr.Code != http.StatusCreated {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
}

func TestError(t *testing.T) {
	rr := httptest.NewRecorder()
	Error(rr, http.StatusForbidden, "forbidden")

	if rr.Code != http.StatusForbidden {
		t.Fatalf("unexpected status: %d", rr.Code)
	}

	var got map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("json unmarshal error: %v", err)
	}
	if got["error"] != "forbidden" {
		t.Fatalf("unexpected payload: %+v", got)
	}
}

func TestNoContent(t *testing.T) {
	rr := httptest.NewRecorder()
	NoContent(rr)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
}
