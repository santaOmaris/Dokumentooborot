package handlers

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDecodeJSONCatalog(t *testing.T) {
	req := httptest.NewRequest("POST", "/", strings.NewReader(`{"name":"folder"}`))
	var v struct {
		Name string `json:"name"`
	}
	if err := decodeJSON(req, &v); err != nil {
		t.Fatalf("decodeJSON error: %v", err)
	}
	if v.Name != "folder" {
		t.Fatalf("unexpected decoded value: %+v", v)
	}
}

func TestDecodeJSONCatalogInvalid(t *testing.T) {
	req := httptest.NewRequest("POST", "/", strings.NewReader(`{"name"`))
	var v map[string]string
	if err := decodeJSON(req, &v); err == nil {
		t.Fatal("expected decode error")
	}
}

func TestPathInt32Catalog(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.SetPathValue("id", "42")
	n, err := pathInt32(req, "id")
	if err != nil {
		t.Fatalf("pathInt32 error: %v", err)
	}
	if n != 42 {
		t.Fatalf("unexpected parsed value: %d", n)
	}
}

func TestPathInt32CatalogInvalid(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.SetPathValue("id", "abc")
	if _, err := pathInt32(req, "id"); err == nil {
		t.Fatal("expected parse error")
	}
}
