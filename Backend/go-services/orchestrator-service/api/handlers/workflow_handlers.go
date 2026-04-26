package handlers

import (
	"net/http"

	"docflow.local/pkg/auth"
	"docflow.local/pkg/httputil"
	"orchestrator-service/service"
)

type WorkflowHandlers struct {
	S *service.WorkflowService
}

func (h *WorkflowHandlers) GetStatusHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		docID, err := pathInt32(r, "id")
		if err != nil {
			httputil.Error(w, http.StatusBadRequest, "invalid id")
			return
		}
		state, err := h.S.GetStatus(r.Context(), docID)
		if err != nil {
			httputil.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		httputil.OK(w, state)
	}
}

func (h *WorkflowHandlers) GetHistoryHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		docID, err := pathInt32(r, "id")
		if err != nil {
			httputil.Error(w, http.StatusBadRequest, "invalid id")
			return
		}
		hist, err := h.S.GetHistory(r.Context(), docID)
		if err != nil {
			httputil.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		httputil.OK(w, hist)
	}
}

func (h *WorkflowHandlers) SendForVisaHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		docID, err := pathInt32(r, "id")
		if err != nil {
			httputil.Error(w, http.StatusBadRequest, "invalid id")
			return
		}
		var body struct {
			Note string `json:"note"`
		}
		if r.ContentLength > 0 {
			if err = decodeJSON(r, &body); err != nil {
				httputil.Error(w, http.StatusBadRequest, "invalid json")
				return
			}
		}
		login, _ := auth.UserLoginFromContext(r.Context())
		if err = h.S.SendForVisa(r.Context(), docID, login, body.Note); err != nil {
			httputil.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		httputil.OK(w, map[string]string{"status": "PENDING_VISA"})
	}
}

func (h *WorkflowHandlers) ApproveHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		docID, err := pathInt32(r, "id")
		if err != nil {
			httputil.Error(w, http.StatusBadRequest, "invalid id")
			return
		}
		login, _ := auth.UserLoginFromContext(r.Context())
		if err = h.S.Approve(r.Context(), docID, login); err != nil {
			httputil.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		httputil.OK(w, map[string]string{"status": "APPROVED"})
	}
}

func (h *WorkflowHandlers) RejectHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		docID, err := pathInt32(r, "id")
		if err != nil {
			httputil.Error(w, http.StatusBadRequest, "invalid id")
			return
		}
		var body struct {
			RevisionNote string `json:"revision_note"`
		}
		if err = decodeJSON(r, &body); err != nil {
			httputil.Error(w, http.StatusBadRequest, "invalid json")
			return
		}
		login, _ := auth.UserLoginFromContext(r.Context())
		if err = h.S.Reject(r.Context(), docID, login, body.RevisionNote); err != nil {
			httputil.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		httputil.OK(w, map[string]string{"status": "DRAFT"})
	}
}

func (h *WorkflowHandlers) DelegateHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		isAdmin := auth.HasRole(r.Context(), "ADMIN")
		isHead, _ := auth.IsHeadFromContext(r.Context())
		if !isAdmin && !isHead {
			httputil.Error(w, http.StatusForbidden, "forbidden: only department head or admin can delegate")
			return
		}

		docID, err := pathInt32(r, "id")
		if err != nil {
			httputil.Error(w, http.StatusBadRequest, "invalid id")
			return
		}
		var body struct {
			AssigneeID int32 `json:"assignee_id"`
		}
		if err = decodeJSON(r, &body); err != nil || body.AssigneeID == 0 {
			httputil.Error(w, http.StatusBadRequest, "assignee_id is required")
			return
		}
		login, _ := auth.UserLoginFromContext(r.Context())
		if err = h.S.DelegateDocument(r.Context(), docID, body.AssigneeID, login); err != nil {
			httputil.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		httputil.OK(w, map[string]string{"ok": "delegated"})
	}
}

func (h *WorkflowHandlers) RequestApprovalHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		docID, err := pathInt32(r, "id")
		if err != nil {
			httputil.Error(w, http.StatusBadRequest, "invalid id")
			return
		}
		var body struct {
			Question           string `json:"question"`
			TargetDepartmentID int32  `json:"target_department_id"`
		}
		if err = decodeJSON(r, &body); err != nil {
			httputil.Error(w, http.StatusBadRequest, "invalid json")
			return
		}
		login, _ := auth.UserLoginFromContext(r.Context())
		if err = h.S.RequestApproval(r.Context(), docID, login, body.Question, body.TargetDepartmentID); err != nil {
			httputil.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		httputil.OK(w, map[string]string{"ok": "approval requested"})
	}
}
