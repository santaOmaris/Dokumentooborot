package handlers

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDecodeJSONCollaboration(t *testing.T) {
	req := httptest.NewRequest("POST", "/", strings.NewReader(`{"content":"hello"}`))
	var v struct {
		Content string `json:"content"`
	}
	if err := decodeJSON(req, &v); err != nil {
		t.Fatalf("decodeJSON error: %v", err)
	}
	if v.Content != "hello" {
		t.Fatalf("unexpected decoded value: %+v", v)
	}
}

func TestPathInt32Collaboration(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.SetPathValue("id", "7")
	n, err := pathInt32(req, "id")
	if err != nil {
		t.Fatalf("pathInt32 error: %v", err)
	}
	if n != 7 {
		t.Fatalf("unexpected parsed value: %d", n)
	}
}

func TestQueryInt32Collaboration(t *testing.T) {
	req := httptest.NewRequest("GET", "/?limit=15&offset=3", nil)
	if got := queryInt32(req, "limit", 30); got != 15 {
		t.Fatalf("unexpected limit: %d", got)
	}
	if got := queryInt32(req, "offset", 0); got != 3 {
		t.Fatalf("unexpected offset: %d", got)
	}
	if got := queryInt32(req, "missing", 99); got != 99 {
		t.Fatalf("unexpected default value: %d", got)
	}
}

func TestQueryInt32CollaborationInvalid(t *testing.T) {
	req := httptest.NewRequest("GET", "/?limit=bad", nil)
	if got := queryInt32(req, "limit", 30); got != 30 {
		t.Fatalf("unexpected fallback value: %d", got)
	}
}
