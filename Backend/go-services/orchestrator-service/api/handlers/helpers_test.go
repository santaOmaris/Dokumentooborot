package handlers

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDecodeJSONOrchestrator(t *testing.T) {
	req := httptest.NewRequest("POST", "/", strings.NewReader(`{"question":"q"}`))
	var v struct {
		Question string `json:"question"`
	}
	if err := decodeJSON(req, &v); err != nil {
		t.Fatalf("decodeJSON error: %v", err)
	}
	if v.Question != "q" {
		t.Fatalf("unexpected decoded value: %+v", v)
	}
}

func TestPathInt32Orchestrator(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.SetPathValue("id", "99")
	n, err := pathInt32(req, "id")
	if err != nil {
		t.Fatalf("pathInt32 error: %v", err)
	}
	if n != 99 {
		t.Fatalf("unexpected parsed value: %d", n)
	}
}

func TestPathInt32OrchestratorInvalid(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.SetPathValue("id", "oops")
	if _, err := pathInt32(req, "id"); err == nil {
		t.Fatal("expected parse error")
	}
}
