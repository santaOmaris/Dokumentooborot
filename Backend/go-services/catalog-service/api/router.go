package api

import (
	"net/http"

	"docflow.local/pkg/auth"
	"docflow.local/pkg/middleware"
	db "catalog-service/db/generated"
	"catalog-service/api/handlers"

	"google.golang.org/grpc"
)

func NewRouter(q *db.Queries, fileConn *grpc.ClientConn, allowedOrigin string) http.Handler {
	mux := http.NewServeMux()
	dh := handlers.NewDocumentHandlers(q, fileConn)

	mux.HandleFunc("GET /api/catalog/types",
		auth.AuthMiddleware(handlers.ListDocumentTypesHandler(q)))
	mux.HandleFunc("POST /api/catalog/types",
		auth.AuthMiddleware(handlers.CreateDocumentTypeHandler(q)))
	mux.HandleFunc("DELETE /api/catalog/types/{id}",
		auth.AuthMiddleware(handlers.DeleteDocumentTypeHandler(q)))

	mux.HandleFunc("GET /api/catalog/departments/{dept_id}/folders",
		auth.AuthMiddleware(handlers.ListFoldersHandler(q)))
	mux.HandleFunc("POST /api/catalog/departments/{dept_id}/folders",
		auth.AuthMiddleware(handlers.CreateFolderHandler(q)))
	mux.HandleFunc("POST /api/catalog/departments/{dept_id}/init-folders",
		auth.AuthMiddleware(handlers.InitDepartmentFoldersHandler(q)))
	mux.HandleFunc("DELETE /api/catalog/folders/{id}",
		auth.AuthMiddleware(handlers.DeleteFolderHandler(q)))
	mux.HandleFunc("PATCH /api/catalog/folders/{id}/rename",
		auth.AuthMiddleware(handlers.RenameFolderHandler(q)))
	mux.HandleFunc("PATCH /api/catalog/folders/{id}/move",
		auth.AuthMiddleware(handlers.MoveFolderHandler(q)))

	mux.HandleFunc("GET /api/catalog/folders/{id}/documents",
		auth.AuthMiddleware(handlers.ListDocumentsByFolderHandler(q)))

	mux.HandleFunc("POST /api/catalog/documents",
		auth.AuthMiddleware(dh.UploadDocumentHandler()))

	mux.HandleFunc("GET /api/catalog/documents/{id}/download",
		auth.AuthMiddleware(dh.DownloadDocumentHandler()))

	mux.HandleFunc("GET /api/catalog/documents/{id}",
		auth.AuthMiddleware(handlers.GetDocumentHandler(q)))
	mux.HandleFunc("GET /api/catalog/documents/{id}/history",
		auth.AuthMiddleware(handlers.GetDocumentHistoryHandler(q)))

	mux.HandleFunc("PATCH /api/catalog/documents/{id}/move",
		auth.AuthMiddleware(handlers.MoveDocumentHandler(q)))

	mux.HandleFunc("POST /api/catalog/documents/{id}/hide",
		auth.AuthMiddleware(dh.HideDocumentHandler()))
	mux.HandleFunc("POST /api/catalog/documents/{id}/unhide",
		auth.AuthMiddleware(handlers.UnhideDocumentHandler(q)))

	mux.HandleFunc("GET /api/catalog/departments/{dept_id}/search",
		auth.AuthMiddleware(handlers.SearchDocumentsHandler(q)))

	return middleware.CORS(allowedOrigin)(mux)
}
