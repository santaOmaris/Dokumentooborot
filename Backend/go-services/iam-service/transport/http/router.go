package iamhttp

import (
	"net/http"

	"docflow.local/pkg/auth"
	"docflow.local/pkg/middleware"
	db "iam-service/db/generated"
	"iam-service/transport/http/handlers"
)

func NewRouter(q *db.Queries, allowedOrigin string) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/iam/auth/login", handlers.AuthHandler(q))
	mux.HandleFunc("POST /api/iam/auth/logout", handlers.LogoutHandler())
	mux.HandleFunc("GET /api/iam/auth/me",
		auth.AuthMiddleware(handlers.MeHandler()),
	)
	mux.HandleFunc("POST /api/iam/auth/register",
		auth.AuthMiddleware(handlers.RegisterHandler(q)),
	)
	mux.HandleFunc("GET /api/iam/users",
		auth.AuthMiddleware(handlers.ListUsersHandler(q)),
	)

	mux.HandleFunc("GET /api/iam/departments/{id}/users",
		auth.AuthMiddleware(handlers.ListUsersByDeptHandler(q)),
	)
	mux.HandleFunc("POST /api/iam/users/fire",
		auth.AuthMiddleware(handlers.FireHandler(q)),
	)
	mux.HandleFunc("POST /api/iam/users/move",
		auth.AuthMiddleware(handlers.MoveUserHandler(q)),
	)

	mux.HandleFunc("POST /api/iam/users/{id}/promote",
		auth.AuthMiddleware(handlers.PromoteHandler(q)),
	)
	mux.HandleFunc("POST /api/iam/users/{id}/demote",
		auth.AuthMiddleware(handlers.DemoteHandler(q)),
	)

	mux.HandleFunc("GET /api/iam/departments",
		auth.AuthMiddleware(handlers.ListDepartmentsHandler(q)),
	)
	mux.HandleFunc("POST /api/iam/departments",
		auth.AuthMiddleware(handlers.CreateDepartmentHandler(q)),
	)
	mux.HandleFunc("PATCH /api/iam/departments/{id}/parent",
		auth.AuthMiddleware(handlers.SetDepartmentParentHandler(q)),
	)
	mux.HandleFunc("DELETE /api/iam/departments/{id}",
		auth.AuthMiddleware(handlers.DeleteDepartmentHandler(q)),
	)

	return middleware.CORS(allowedOrigin)(mux)
}
