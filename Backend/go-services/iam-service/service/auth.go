package service

import (
	"docflow.local/pkg/auth"
	"context"
	"database/sql"
	"errors"
	"fmt"
	db "iam-service/db/generated"

	"golang.org/x/crypto/bcrypt"
)

func RegisterUser(ctx context.Context, q *db.Queries, login, password, email, fullname string, isHead bool, systemRole string) (int32, error) {
	_, err := q.IsLoginExistInDB(ctx, login)
	if err == nil {
		return 0, errors.New("login already exists")
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	return q.CreateUser(ctx, db.CreateUserParams{
		Login:        login,
		PasswordHash: string(hashed),
		Email:        email,
		FullName:     fullname,
		IsHead:       isHead,
		SystemRole:   systemRole,
		DepartmentID: sql.NullInt32{Valid: false},
	})
}

func AuthUser(ctx context.Context, q *db.Queries, login, password string) (string, error) {
	row, err := q.GetUserForAuth(ctx, login)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", errors.New("invalid credentials")
		}
		return "", err
	}

	if err = bcrypt.CompareHashAndPassword([]byte(row.PasswordHash), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	deptStr := ""
	if row.DepartmentID.Valid {
		deptStr = fmt.Sprintf("%d", row.DepartmentID.Int32)
	}

	return auth.GenerateToken(row.ID, login, row.SystemRole, row.IsHead, deptStr)
}
