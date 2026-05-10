package service

import (
	"context"
	"database/sql"
	"errors"
	db "iam-service/db/generated"
)

func FireUser(ctx context.Context, q *db.Queries, callerLogin, callerDeptID string, callerIsHead bool, callerRole string, targetID int32) error {
	if callerRole != "ADMIN" && !callerIsHead {
		return errors.New("forbidden: only department head or admin can fire employees")
	}

	target, err := q.GetUserByID(ctx, targetID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("user not found")
		}
		return err
	}

	if target.Login == callerLogin {
		return errors.New("forbidden: you cannot fire yourself")
	}

	if callerRole != "ADMIN" {
		if !target.DepartmentID.Valid || callerDeptID == "" {
			return errors.New("target user has no department")
		}
		if target.IsHead {
			return errors.New("forbidden: department head cannot fire another head")
		}
		if target.DepartmentID.Int32 != int32(mustParseInt(callerDeptID)) {
			return errors.New("forbidden: user is not in your department")
		}
	}

	return q.FireUser(ctx, targetID)
}

func MoveUserToDepartment(ctx context.Context, q *db.Queries, userID, departmentID int32) error {
	return q.MoveUserToDepartment(ctx, db.MoveUserToDepartmentParams{
		ID:           userID,
		DepartmentID: sql.NullInt32{Int32: departmentID, Valid: true},
	})
}

func PromoteUser(ctx context.Context, q *db.Queries, userID int32) error {
	return q.PromoteUser(ctx, userID)
}

func DemoteUser(ctx context.Context, q *db.Queries, userID int32) error {
	return q.DemoteUser(ctx, userID)
}

func ListUsers(ctx context.Context, q *db.Queries) ([]db.ListUsersRow, error) {
	return q.ListUsers(ctx)
}

func ListUsersByDepartment(ctx context.Context, q *db.Queries, deptID int32) ([]db.ListUsersByDepartmentRow, error) {
	return q.ListUsersByDepartment(ctx, sql.NullInt32{Int32: deptID, Valid: true})
}

func mustParseInt(s string) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0
		}
		n = n*10 + int(c-'0')
	}
	return n
}

