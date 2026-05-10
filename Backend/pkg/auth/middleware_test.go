package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGenerateTokenAndValidateToken(t *testing.T) {
	token, err := GenerateToken(11, "alice", "ADMIN", true, "5")
	if err != nil {
		t.Fatalf("GenerateToken error: %v", err)
	}

	claims, err := ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken error: %v", err)
	}

	if claims["login"] != "alice" {
		t.Fatalf("unexpected login claim: %v", claims["login"])
	}
	if claims["system_role"] != "ADMIN" {
		t.Fatalf("unexpected system_role claim: %v", claims["system_role"])
	}
	if claims["is_head"] != true {
		t.Fatalf("unexpected is_head claim: %v", claims["is_head"])
	}
	if claims["department_id"] != "5" {
		t.Fatalf("unexpected department_id claim: %v", claims["department_id"])
	}
}

func TestValidateTokenRejectsGarbage(t *testing.T) {
	if _, err := ValidateToken("not-a-jwt"); err == nil {
		t.Fatal("expected error for invalid token")
	}
}

func TestAuthMiddlewareRejectsMissingToken(t *testing.T) {
	called := false
	h := AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h(rr, req)

	if called {
		t.Fatal("next handler should not be called")
	}
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
}

func TestAuthMiddlewareSetsContext(t *testing.T) {
	token, err := GenerateToken(7, "head_dev", "USER", true, "2")
	if err != nil {
		t.Fatalf("GenerateToken error: %v", err)
	}

	h := AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		login, ok := UserLoginFromContext(ctx)
		if !ok || login != "head_dev" {
			t.Fatalf("unexpected login from context: %q ok=%v", login, ok)
		}

		uid, ok := UserIDIntFromContext(ctx)
		if !ok || uid != 7 {
			t.Fatalf("unexpected user_id from context: %d ok=%v", uid, ok)
		}

		role, ok := SystemRoleFromContext(ctx)
		if !ok || role != "USER" {
			t.Fatalf("unexpected role from context: %q ok=%v", role, ok)
		}

		isHead, ok := IsHeadFromContext(ctx)
		if !ok || !isHead {
			t.Fatalf("unexpected is_head from context: %v ok=%v", isHead, ok)
		}

		dept, ok := DepartmentIDFromContext(ctx)
		if !ok || dept != "2" {
			t.Fatalf("unexpected department_id from context: %q ok=%v", dept, ok)
		}

		if !HasRole(ctx, "USER") {
			t.Fatal("HasRole should be true for USER")
		}
		if HasRole(ctx, "ADMIN") {
			t.Fatal("HasRole should be false for ADMIN")
		}

		aliasID, ok := UserIDFromContext(ctx)
		if !ok || aliasID != "head_dev" {
			t.Fatalf("unexpected UserIDFromContext value: %q ok=%v", aliasID, ok)
		}

		w.WriteHeader(http.StatusNoContent)
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "jwt_token", Value: token})
	h(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
}

func TestHasRoleWithoutContextRole(t *testing.T) {
	if HasRole(context.Background(), "ADMIN") {
		t.Fatal("HasRole should be false when context has no role")
	}
}
