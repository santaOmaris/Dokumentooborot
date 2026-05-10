package handlers

import (
	"net/http"

	"docflow.local/pkg/auth"
	"docflow.local/pkg/httputil"
	db "collaboration-service/db/generated"
	"collaboration-service/service"
)

func ListMessagesHandler(q *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		docID, err := pathInt32(r, "id")
		if err != nil {
			httputil.Error(w, http.StatusBadRequest, "invalid document id")
			return
		}
		msgs, err := service.ListMessages(r.Context(), q, docID)
		if err != nil {
			httputil.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		httputil.OK(w, msgs)
	}
}

func SendMessageHandler(q *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		docID, err := pathInt32(r, "id")
		if err != nil {
			httputil.Error(w, http.StatusBadRequest, "invalid document id")
			return
		}

		var body struct {
			Content string `json:"content"`
		}
		if err := decodeJSON(r, &body); err != nil {
			httputil.Error(w, http.StatusBadRequest, "invalid json")
			return
		}

		login, _ := auth.UserLoginFromContext(r.Context())
		msg, err := service.SendMessage(r.Context(), q, docID, login, body.Content)
		if err != nil {
			httputil.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		httputil.Created(w, msg)
	}
}
