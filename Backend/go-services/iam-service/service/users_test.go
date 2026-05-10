package service

import (
	"context"
	"database/sql"
	"regexp"
	"strings"
	"testing"

	db "iam-service/db/generated"

	"github.com/DATA-DOG/go-sqlmock"
)

func newIAMQueries(t *testing.T) (*db.Queries, sqlmock.Sqlmock, *sql.DB) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New error: %v", err)
	}
	return db.New(sqlDB), mock, sqlDB
}

func TestFireUserForbiddenForRegularUser(t *testing.T) {
	q, _, sqlDB := newIAMQueries(t)
	defer sqlDB.Close()

	err := FireUser(context.Background(), q, "user1", "2", false, "USER", 10)
	if err == nil || !strings.Contains(err.Error(), "forbidden") {
		t.Fatalf("expected forbidden error, got: %v", err)
	}
}

func TestFireUserAdminSuccess(t *testing.T) {
	q, mock, sqlDB := newIAMQueries(t)
	defer sqlDB.Close()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, login, email, full_name, department_id, is_head, system_role FROM users WHERE id = $1")).
		WithArgs(int32(10)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "login", "email", "full_name", "department_id", "is_head", "system_role"}).
			AddRow(int32(10), "u", "u@x", "U", int32(2), false, "USER"))

	mock.ExpectExec(regexp.QuoteMeta("UPDATE users SET department_id = NULL WHERE id = $1")).
		WithArgs(int32(10)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := FireUser(context.Background(), q, "admin", "", false, "ADMIN", 10)
	if err != nil {
		t.Fatalf("FireUser error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestFireUserHeadCannotFireAnotherHead(t *testing.T) {
	q, mock, sqlDB := newIAMQueries(t)
	defer sqlDB.Close()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, login, email, full_name, department_id, is_head, system_role FROM users WHERE id = $1")).
		WithArgs(int32(10)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "login", "email", "full_name", "department_id", "is_head", "system_role"}).
			AddRow(int32(10), "head_x", "h@x", "Head X", int32(2), true, "USER"))

	err := FireUser(context.Background(), q, "head_dev", "2", true, "USER", 10)
	if err == nil || !strings.Contains(err.Error(), "cannot fire another head") {
		t.Fatalf("expected head-to-head forbidden error, got: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestFireUserHeadSameDepartment(t *testing.T) {
	q, mock, sqlDB := newIAMQueries(t)
	defer sqlDB.Close()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, login, email, full_name, department_id, is_head, system_role FROM users WHERE id = $1")).
		WithArgs(int32(10)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "login", "email", "full_name", "department_id", "is_head", "system_role"}).
			AddRow(int32(10), "u", "u@x", "U", int32(2), false, "USER"))

	mock.ExpectExec(regexp.QuoteMeta("UPDATE users SET department_id = NULL WHERE id = $1")).
		WithArgs(int32(10)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := FireUser(context.Background(), q, "head_dev", "2", true, "USER", 10)
	if err != nil {
		t.Fatalf("FireUser error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestFireUserHeadForeignDepartment(t *testing.T) {
	q, mock, sqlDB := newIAMQueries(t)
	defer sqlDB.Close()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, login, email, full_name, department_id, is_head, system_role FROM users WHERE id = $1")).
		WithArgs(int32(10)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "login", "email", "full_name", "department_id", "is_head", "system_role"}).
			AddRow(int32(10), "u", "u@x", "U", int32(3), false, "USER"))

	err := FireUser(context.Background(), q, "head_dev", "2", true, "USER", 10)
	if err == nil || !strings.Contains(err.Error(), "forbidden") {
		t.Fatalf("expected forbidden error, got: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestFireUserHeadTargetWithoutDepartment(t *testing.T) {
	q, mock, sqlDB := newIAMQueries(t)
	defer sqlDB.Close()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, login, email, full_name, department_id, is_head, system_role FROM users WHERE id = $1")).
		WithArgs(int32(10)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "login", "email", "full_name", "department_id", "is_head", "system_role"}).
			AddRow(int32(10), "u", "u@x", "U", nil, false, "USER"))

	err := FireUser(context.Background(), q, "head_dev", "2", true, "USER", 10)
	if err == nil || !strings.Contains(err.Error(), "target user has no department") {
		t.Fatalf("expected target user has no department, got: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestMovePromoteDemoteAndListWrappers(t *testing.T) {
	q, mock, sqlDB := newIAMQueries(t)
	defer sqlDB.Close()

	mock.ExpectExec(regexp.QuoteMeta("UPDATE users SET department_id = $2 WHERE id = $1")).
		WithArgs(int32(5), sql.NullInt32{Int32: 3, Valid: true}).
		WillReturnResult(sqlmock.NewResult(0, 1))
	if err := MoveUserToDepartment(context.Background(), q, 5, 3); err != nil {
		t.Fatalf("MoveUserToDepartment error: %v", err)
	}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE users SET is_head = true WHERE id = $1")).
		WithArgs(int32(5)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	if err := PromoteUser(context.Background(), q, 5); err != nil {
		t.Fatalf("PromoteUser error: %v", err)
	}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE users SET is_head = false WHERE id = $1")).
		WithArgs(int32(5)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	if err := DemoteUser(context.Background(), q, 5); err != nil {
		t.Fatalf("DemoteUser error: %v", err)
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, full_name, login, is_head, department_id FROM users")).
		WillReturnRows(sqlmock.NewRows([]string{"id", "full_name", "login", "is_head", "department_id"}).
			AddRow(int32(1), "Admin", "admin", false, int32(1)))
	users, err := ListUsers(context.Background(), q)
	if err != nil {
		t.Fatalf("ListUsers error: %v", err)
	}
	if len(users) != 1 || users[0].Login != "admin" {
		t.Fatalf("unexpected users list: %+v", users)
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, login, full_name, is_head FROM users WHERE department_id = $1")).
		WithArgs(sql.NullInt32{Int32: 3, Valid: true}).
		WillReturnRows(sqlmock.NewRows([]string{"id", "login", "full_name", "is_head"}).
			AddRow(int32(2), "user2", "User Two", false))
	deptUsers, err := ListUsersByDepartment(context.Background(), q, 3)
	if err != nil {
		t.Fatalf("ListUsersByDepartment error: %v", err)
	}
	if len(deptUsers) != 1 || deptUsers[0].Login != "user2" {
		t.Fatalf("unexpected department users list: %+v", deptUsers)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestMustParseInt(t *testing.T) {
	if mustParseInt("123") != 123 {
		t.Fatalf("unexpected parse for numeric string")
	}
	if mustParseInt("12a") != 0 {
		t.Fatalf("unexpected parse for invalid string")
	}
	if mustParseInt("") != 0 {
		t.Fatalf("unexpected parse for empty string")
	}
}
