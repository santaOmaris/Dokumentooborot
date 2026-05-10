package service

import (
	"context"
	"database/sql"
	"errors"
	db "iam-service/db/generated"
)

func ListDepartments(ctx context.Context, q *db.Queries) ([]db.ListDepartmentsRow, error) {
	return q.ListDepartments(ctx)
}

func CreateDepartment(ctx context.Context, q *db.Queries, name string, parentID *int32) (int32, error) {
	return q.CreateDepartment(ctx, db.CreateDepartmentParams{
		Name:     name,
		ParentID: toNullInt32(parentID),
	})
}

func SetDepartmentParent(ctx context.Context, q *db.Queries, deptID int32, parentID *int32) error {
	if parentID != nil && *parentID == deptID {
		return errors.New("department cannot be its own parent")
	}
	return q.UpdateDepartmentParent(ctx, db.UpdateDepartmentParentParams{
		ID:       deptID,
		ParentID: toNullInt32(parentID),
	})
}

func DeleteDepartment(ctx context.Context, q *db.Queries, deptID int32) error {
	return q.DeleteDepartment(ctx, deptID)
}

func toNullInt32(v *int32) sql.NullInt32 {
	if v == nil {
		return sql.NullInt32{Valid: false}
	}
	return sql.NullInt32{Int32: *v, Valid: true}
}
