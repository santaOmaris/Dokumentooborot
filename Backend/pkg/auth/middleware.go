package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("super_secret_demo_key")

type ctxKey string

const (
	ctxUserLogin    ctxKey = "user_login"
	ctxSystemRole   ctxKey = "system_role"
	ctxIsHead       ctxKey = "is_head"
	ctxDepartmentID ctxKey = "department_id"
	ctxUserID       ctxKey = "user_id_int"
)

func GenerateToken(userID int32, login string, systemRole string, isHead bool, departmentID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id":       userID,
		"login":         login,
		"system_role":   systemRole,
		"is_head":       isHead,
		"department_id": departmentID,
		"exp":           time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func ValidateToken(tokenString string) (map[string]interface{}, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims structure")
	}
	return claims, nil
}

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("jwt_token")
		if err != nil {
			http.Error(w, "No token", http.StatusUnauthorized)
			return
		}

		claims, err := ValidateToken(cookie.Value)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		login, ok := claims["login"].(string)
		if !ok || login == "" {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}
		systemRole, _ := claims["system_role"].(string)

		var isHead bool
		if v, ok := claims["is_head"].(bool); ok {
			isHead = v
		} else if v, ok := claims["is_head"].(float64); ok {
			isHead = v == 1
		}

		var departmentID string
		if v, ok := claims["department_id"].(string); ok {
			departmentID = v
		} else if v, ok := claims["department_id"].(float64); ok {
			departmentID = fmt.Sprintf("%.0f", v)
		}

		var userID int32
		if v, ok := claims["user_id"].(float64); ok {
			userID = int32(v)
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, ctxUserLogin, login)
		ctx = context.WithValue(ctx, ctxSystemRole, systemRole)
		ctx = context.WithValue(ctx, ctxIsHead, isHead)
		ctx = context.WithValue(ctx, ctxDepartmentID, departmentID)
		ctx = context.WithValue(ctx, ctxUserID, userID)

		next(w, r.WithContext(ctx))
	}
}

func UserLoginFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(ctxUserLogin).(string)
	return v, ok
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	return UserLoginFromContext(ctx)
}

func UserIDIntFromContext(ctx context.Context) (int32, bool) {
	v, ok := ctx.Value(ctxUserID).(int32)
	return v, ok
}

func SystemRoleFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(ctxSystemRole).(string)
	return v, ok
}

func IsHeadFromContext(ctx context.Context) (bool, bool) {
	v, ok := ctx.Value(ctxIsHead).(bool)
	return v, ok
}

func DepartmentIDFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(ctxDepartmentID).(string)
	return v, ok
}

func HasRole(ctx context.Context, allowedRoles ...string) bool {
	role, ok := SystemRoleFromContext(ctx)
	if !ok {
		return false
	}
	for _, allowed := range allowedRoles {
		if role == allowed {
			return true
		}
	}
	return false
}
