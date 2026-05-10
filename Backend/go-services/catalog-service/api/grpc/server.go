package grpcserver

import (
	"context"

	db "catalog-service/db/generated"
	"catalog-service/service"
	catalogpb "docflow.local/pkg/pb/catalog"
)

type CatalogGRPCServer struct {
	catalogpb.UnimplementedCatalogServiceServer
	Q *db.Queries
}

func New(q *db.Queries) *CatalogGRPCServer {
	return &CatalogGRPCServer{Q: q}
}

func (s *CatalogGRPCServer) GetDocument(ctx context.Context, req *catalogpb.GetDocumentRequest) (*catalogpb.DocumentResponse, error) {
	doc, err := service.GetDocument(ctx, s.Q, req.DocumentId)
	if err != nil {
		return nil, err
	}
	resp := &catalogpb.DocumentResponse{
		Id:           doc.ID,
		Title:        doc.Title,
		FilePath:     doc.FilePath,
		DepartmentId: doc.DepartmentID,
		CreatedBy:    doc.CreatedBy,
		Status:       doc.Status,
	}
	if doc.AssigneeID.Valid {
		resp.AssigneeId = doc.AssigneeID.Int32
	}
	return resp, nil
}

func (s *CatalogGRPCServer) ChangeDocumentAssignee(ctx context.Context, req *catalogpb.ChangeAssigneeRequest) (*catalogpb.ChangeAssigneeResponse, error) {
	err := service.ChangeDocumentAssignee(ctx, s.Q, req.DocumentId, req.NewAssigneeId)
	return &catalogpb.ChangeAssigneeResponse{Success: err == nil}, err
}

func (s *CatalogGRPCServer) UpdateDocumentStatus(ctx context.Context, req *catalogpb.UpdateDocumentStatusRequest) (*catalogpb.UpdateDocumentStatusResponse, error) {
	err := service.UpdateDocumentStatus(ctx, s.Q, req.DocumentId, req.Status)
	return &catalogpb.UpdateDocumentStatusResponse{Success: err == nil}, err
}

func (s *CatalogGRPCServer) GetSystemFolder(ctx context.Context, req *catalogpb.GetSystemFolderRequest) (*catalogpb.FolderResponse, error) {
	folder, err := service.FindSystemFolder(ctx, s.Q, req.DepartmentId, req.Name)
	if err != nil {
		return nil, err
	}
	return &catalogpb.FolderResponse{
		Id:           folder.ID,
		DepartmentId: req.DepartmentId,
		Name:         folder.Name,
		IsSystem:     folder.IsSystem,
	}, nil
}

func (s *CatalogGRPCServer) MoveDocument(ctx context.Context, req *catalogpb.MoveDocumentRequest) (*catalogpb.MoveDocumentResponse, error) {
	err := service.MoveDocument(ctx, s.Q, req.DocumentId, req.FolderId, req.ActorLogin)
	return &catalogpb.MoveDocumentResponse{Success: err == nil}, err
}
