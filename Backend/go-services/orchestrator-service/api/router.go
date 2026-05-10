package api

import (
	"net/http"

	"docflow.local/pkg/auth"
	"docflow.local/pkg/middleware"
	"orchestrator-service/api/handlers"
	"orchestrator-service/service"
)

func NewRouter(svc *service.WorkflowService, allowedOrigin string) http.Handler {
	mux := http.NewServeMux()
	h := &handlers.WorkflowHandlers{S: svc}

	mux.HandleFunc("GET /api/orchestrator/documents/{id}/status",
		auth.AuthMiddleware(h.GetStatusHandler()))
	mux.HandleFunc("GET /api/orchestrator/documents/{id}/history",
		auth.AuthMiddleware(h.GetHistoryHandler()))
	mux.HandleFunc("GET /api/orchestrator/metrics",
		auth.AuthMiddleware(h.MetricsHandler()))
	mux.HandleFunc("POST /api/orchestrator/metrics/export-csv",
		auth.AuthMiddleware(h.ExportMetricsCSVHandler()))

	mux.HandleFunc("POST /api/orchestrator/documents/{id}/send-for-visa",
		auth.AuthMiddleware(h.SendForVisaHandler()))

	mux.HandleFunc("POST /api/orchestrator/documents/{id}/approve",
		auth.AuthMiddleware(h.ApproveHandler()))
	mux.HandleFunc("POST /api/orchestrator/documents/{id}/reject",
		auth.AuthMiddleware(h.RejectHandler()))

	mux.HandleFunc("POST /api/orchestrator/documents/{id}/delegate",
		auth.AuthMiddleware(h.DelegateHandler()))

	mux.HandleFunc("POST /api/orchestrator/documents/{id}/request-approval",
		auth.AuthMiddleware(h.RequestApprovalHandler()))

	return middleware.CORS(allowedOrigin)(mux)
}
