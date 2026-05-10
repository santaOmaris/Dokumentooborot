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

func newIAMDeptQueries(t *testing.T) (*db.Queries, sqlmock.Sqlmock, *sql.DB) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New error: %v", err)
	}
	return db.New(sqlDB), mock, sqlDB
}

func TestCreateDepartment(t *testing.T) {
	q, mock, sqlDB := newIAMDeptQueries(t)
	defer sqlDB.Close()

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO departments (name, parent_id) VALUES ($1, $2) RETURNING id")).
		WithArgs("Новый отдел", sql.NullInt32{Int32: 1, Valid: true}).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int32(7)))

	parent := int32(1)
	id, err := CreateDepartment(context.Background(), q, "Новый отдел", &parent)
	if err != nil {
		t.Fatalf("CreateDepartment error: %v", err)
	}
	if id != 7 {
		t.Fatalf("unexpected department id: %d", id)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestSetDepartmentParentRejectsSelfReference(t *testing.T) {
	q, _, sqlDB := newIAMDeptQueries(t)
	defer sqlDB.Close()

	parent := int32(5)
	err := SetDepartmentParent(context.Background(), q, 5, &parent)
	if err == nil || !strings.Contains(err.Error(), "cannot be its own parent") {
		t.Fatalf("expected self-reference error, got: %v", err)
	}
}

func TestSetDepartmentParentAndDelete(t *testing.T) {
	q, mock, sqlDB := newIAMDeptQueries(t)
	defer sqlDB.Close()

	mock.ExpectExec(regexp.QuoteMeta("UPDATE departments SET parent_id = $2 WHERE id = $1")).
		WithArgs(int32(5), sql.NullInt32{Int32: 1, Valid: true}).
		WillReturnResult(sqlmock.NewResult(0, 1))

	parent := int32(1)
	if err := SetDepartmentParent(context.Background(), q, 5, &parent); err != nil {
		t.Fatalf("SetDepartmentParent error: %v", err)
	}

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM departments WHERE id = $1")).
		WithArgs(int32(5)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := DeleteDepartment(context.Background(), q, 5); err != nil {
		t.Fatalf("DeleteDepartment error: %v", err)
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, parent_id FROM departments ORDER BY name ASC")).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "parent_id"}).
			AddRow(int32(1), "Дирекция", nil))
	departments, err := ListDepartments(context.Background(), q)
	if err != nil {
		t.Fatalf("ListDepartments error: %v", err)
	}
	if len(departments) != 1 || departments[0].Name != "Дирекция" {
		t.Fatalf("unexpected departments list: %+v", departments)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestToNullInt32(t *testing.T) {
	if got := toNullInt32(nil); got.Valid {
		t.Fatalf("expected invalid null int32, got %+v", got)
	}
	v := int32(8)
	got := toNullInt32(&v)
	if !got.Valid || got.Int32 != 8 {
		t.Fatalf("unexpected null int32: %+v", got)
	}
}
