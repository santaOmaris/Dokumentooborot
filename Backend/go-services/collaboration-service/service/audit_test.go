package service

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	db "collaboration-service/db/generated"

	"github.com/DATA-DOG/go-sqlmock"
)

func newCollabQueries(t *testing.T) (*db.Queries, sqlmock.Sqlmock, *sql.DB) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New error: %v", err)
	}
	return db.New(sqlDB), mock, sqlDB
}

func TestListAuditByDepartment(t *testing.T) {
	q, mock, sqlDB := newCollabQueries(t)
	defer sqlDB.Close()

	now := time.Now()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, department_id, document_id, actor_login, action, details, created_at\nFROM audit_logs\nWHERE department_id = $1\nORDER BY created_at DESC\nLIMIT $2 OFFSET $3")).
		WithArgs(sql.NullInt32{Int32: 2, Valid: true}, int32(20), int32(0)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "department_id", "document_id", "actor_login", "action", "details", "created_at"}).
			AddRow(int32(1), int32(2), int32(3), "admin", "DOC_APPROVED", "ok", now))

	logs, err := ListAuditByDepartment(context.Background(), q, 2, 20, 0)
	if err != nil {
		t.Fatalf("ListAuditByDepartment error: %v", err)
	}
	if len(logs) != 1 || logs[0].Action != "DOC_APPROVED" {
		t.Fatalf("unexpected logs: %+v", logs)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestListAuditByDocument(t *testing.T) {
	q, mock, sqlDB := newCollabQueries(t)
	defer sqlDB.Close()

	now := time.Now()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, department_id, document_id, actor_login, action, details, created_at\nFROM audit_logs\nWHERE document_id = $1\nORDER BY created_at DESC")).
		WithArgs(sql.NullInt32{Int32: 5, Valid: true}).
		WillReturnRows(sqlmock.NewRows([]string{"id", "department_id", "document_id", "actor_login", "action", "details", "created_at"}).
			AddRow(int32(2), int32(2), int32(5), "head_dev", "DOC_SENT_FOR_APPROVAL", "note", now))

	logs, err := ListAuditByDocument(context.Background(), q, 5)
	if err != nil {
		t.Fatalf("ListAuditByDocument error: %v", err)
	}
	if len(logs) != 1 || logs[0].DocumentID.Int32 != 5 {
		t.Fatalf("unexpected logs: %+v", logs)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestWriteAuditLog(t *testing.T) {
	q, mock, sqlDB := newCollabQueries(t)
	defer sqlDB.Close()

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO audit_logs (department_id, document_id, actor_login, action, details)\nVALUES ($1, $2, $3, $4, $5)")).
		WithArgs(
			sql.NullInt32{Int32: 2, Valid: true},
			sql.NullInt32{Int32: 7, Valid: true},
			sql.NullString{String: "admin", Valid: true},
			"DOC_APPROVED",
			sql.NullString{String: "ok", Valid: true},
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := WriteAuditLog(context.Background(), q, 2, 7, "admin", "DOC_APPROVED", "ok")
	if err != nil {
		t.Fatalf("WriteAuditLog error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}
