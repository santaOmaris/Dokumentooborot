package service

import (
	"context"
	"errors"

	db "collaboration-service/db/generated"
)

func ListMessages(ctx context.Context, q *db.Queries, documentID int32) ([]db.Message, error) {
	return q.ListMessagesByDocument(ctx, documentID)
}

func SendMessage(ctx context.Context, q *db.Queries, documentID int32, senderLogin, content string) (db.Message, error) {
	if content == "" {
		return db.Message{}, errors.New("content is required")
	}
	return q.CreateMessage(ctx, db.CreateMessageParams{
		DocumentID:  documentID,
		SenderLogin: senderLogin,
		Content:     content,
	})
}
