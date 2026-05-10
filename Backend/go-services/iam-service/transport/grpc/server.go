package grpcserver

import (
	"context"
	"database/sql"

	iampb "docflow.local/pkg/pb/iam"
	db "iam-service/db/generated"
)

type IAMGRPCServer struct {
	iampb.UnimplementedIAMServiceServer
	Q *db.Queries
}

func New(q *db.Queries) *IAMGRPCServer {
	return &IAMGRPCServer{Q: q}
}

func (s *IAMGRPCServer) GetUserByLogin(ctx context.Context, req *iampb.GetUserByLoginRequest) (*iampb.UserResponse, error) {
	row, err := s.Q.GetUserByLogin(ctx, req.Login)
	if err != nil {
		return nil, err
	}
	resp := &iampb.UserResponse{
		Id:         row.ID,
		Login:      row.Login,
		Email:      row.Email,
		FullName:   row.FullName,
		IsHead:     row.IsHead,
		SystemRole: row.SystemRole,
	}
	if row.DepartmentID.Valid {
		resp.DepartmentId = row.DepartmentID.Int32
	}
	return resp, nil
}

func (s *IAMGRPCServer) GetManagersByDepartment(ctx context.Context, req *iampb.GetManagersByDepartmentRequest) (*iampb.ManagersResponse, error) {
	rows, err := s.Q.GetManagersByDepartment(ctx, sql.NullInt32{Int32: req.DepartmentId, Valid: true})
	if err != nil {
		return nil, err
	}
	managers := make([]*iampb.UserResponse, 0, len(rows))
	for _, r := range rows {
		managers = append(managers, &iampb.UserResponse{
			Id:       r.ID,
			Login:    r.Login,
			Email:    r.Email,
			FullName: r.FullName,
		})
	}
	return &iampb.ManagersResponse{Managers: managers}, nil
}

func (s *IAMGRPCServer) GetUserEmail(ctx context.Context, req *iampb.GetUserEmailRequest) (*iampb.GetUserEmailResponse, error) {
	row, err := s.Q.GetUserByID(ctx, req.UserId)
	if err != nil {
		return nil, err
	}
	return &iampb.GetUserEmailResponse{
		Email:    row.Email,
		FullName: row.FullName,
	}, nil
}
