package grpcserver

import (
	"context"

	catalogpb "docflow.local/pkg/pb/catalog"
	db "catalog-service/db/generated"
	"catalog-service/service"
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
