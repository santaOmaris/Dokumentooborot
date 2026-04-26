package service

import (
	"context"
	"database/sql"

	db "collaboration-service/db/generated"
)

func ListAuditByDepartment(ctx context.Context, q *db.Queries, departmentID int32, limit, offset int32) ([]db.AuditLog, error) {
	return q.ListAuditByDepartment(ctx, db.ListAuditByDepartmentParams{
		DepartmentID: sql.NullInt32{Int32: departmentID, Valid: true},
		Limit:        limit,
		Offset:       offset,
	})
}

func ListAuditByDocument(ctx context.Context, q *db.Queries, documentID int32) ([]db.AuditLog, error) {
	return q.ListAuditByDocument(ctx, sql.NullInt32{Int32: documentID, Valid: true})
}

func WriteAuditLog(ctx context.Context, q *db.Queries, departmentID, documentID int32, actorLogin, action, details string) error {
	return q.InsertAuditLog(ctx, db.InsertAuditLogParams{
		DepartmentID: sql.NullInt32{Int32: departmentID, Valid: departmentID != 0},
		DocumentID:   sql.NullInt32{Int32: documentID, Valid: documentID != 0},
		ActorLogin:   sql.NullString{String: actorLogin, Valid: actorLogin != ""},
		Action:       action,
		Details:      sql.NullString{String: details, Valid: details != ""},
	})
}
