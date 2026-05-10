package api

import (
	"net/http"

	"docflow.local/pkg/auth"
	"docflow.local/pkg/middleware"
	db "collaboration-service/db/generated"
	"collaboration-service/api/handlers"
)

func NewRouter(q *db.Queries, allowedOrigin string) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/collaboration/documents/{id}/messages",
		auth.AuthMiddleware(handlers.ListMessagesHandler(q)))
	mux.HandleFunc("POST /api/collaboration/documents/{id}/messages",
		auth.AuthMiddleware(handlers.SendMessageHandler(q)))

	mux.HandleFunc("GET /api/collaboration/departments/{dept_id}/audit",
		auth.AuthMiddleware(handlers.ListAuditByDepartmentHandler(q)))

	mux.HandleFunc("GET /api/collaboration/documents/{id}/audit",
		auth.AuthMiddleware(handlers.ListAuditByDocumentHandler(q)))

	return middleware.CORS(allowedOrigin)(mux)
}
