package service

import (
	"context"
	"database/sql"
	"errors"

	db "catalog-service/db/generated"
)

func ListDocumentTypes(ctx context.Context, q *db.Queries) ([]db.ListDocumentTypesRow, error) {
	return q.ListDocumentTypes(ctx)
}

func CreateDocumentType(ctx context.Context, q *db.Queries, name string) (db.CreateDocumentTypeRow, error) {
	return q.CreateDocumentType(ctx, name)
}

func DeleteDocumentType(ctx context.Context, q *db.Queries, id int32) error {
	return q.DeleteDocumentType(ctx, id)
}

func ListFoldersByDepartment(ctx context.Context, q *db.Queries, deptID int32) ([]db.ListFoldersByDepartmentRow, error) {
	return q.ListFoldersByDepartment(ctx, deptID)
}

func CreateFolder(ctx context.Context, q *db.Queries, deptID int32, parentID *int32, name string) (db.CreateFolderRow, error) {
	return q.CreateFolder(ctx, db.CreateFolderParams{
		DepartmentID: deptID,
		ParentID:     toNullInt32(parentID),
		Name:         name,
		IsSystem:     false,
	})
}

func InitDepartmentFolders(ctx context.Context, q *db.Queries, deptID int32) error {
	existing, err := q.ListFoldersByDepartment(ctx, deptID)
	if err != nil {
		return err
	}

	existingByName := make(map[string]bool, len(existing))
	for _, f := range existing {
		existingByName[f.Name] = true
	}

	systemFolders := []string{"main", "templates", "archived", "head_only", "collaborations"}
	for _, name := range systemFolders {
		if existingByName[name] {
			continue
		}
		_, err := q.CreateFolder(ctx, db.CreateFolderParams{
			DepartmentID: deptID,
			ParentID:     toNullInt32(nil),
			Name:         name,
			IsSystem:     true,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func DeleteFolder(ctx context.Context, q *db.Queries, folderID int32) error {
	return q.DeleteFolder(ctx, folderID)
}

func RenameFolder(ctx context.Context, q *db.Queries, folderID int32, newName string) error {
	return q.RenameFolder(ctx, db.RenameFolderParams{ID: folderID, Name: newName})
}

func MoveFolder(ctx context.Context, q *db.Queries, folderID int32, newParentID *int32) error {
	return q.MoveFolder(ctx, db.MoveFolderParams{ID: folderID, ParentID: toNullInt32(newParentID)})
}

func GetDocument(ctx context.Context, q *db.Queries, id int32) (db.Document, error) {
	return q.GetDocument(ctx, id)
}

func ListDocumentsByFolder(ctx context.Context, q *db.Queries, folderID int32) ([]db.ListDocumentsByFolderRow, error) {
	return q.ListDocumentsByFolder(ctx, folderID)
}

func SearchDocuments(ctx context.Context, q *db.Queries, deptID int32, query string) ([]db.SearchDocumentsByTitleRow, error) {
	return q.SearchDocumentsByTitle(ctx, db.SearchDocumentsByTitleParams{
		DepartmentID: deptID,
		PlaintoTsquery: query,
	})
}

type UploadDocumentParams struct {
	Title        string
	Description  string
	TypeID       *int32
	FolderID     int32
	FilePath     string
	OriginalName string
	DepartmentID int32
	CreatedBy    int32
}

func CreateDocument(ctx context.Context, q *db.Queries, p UploadDocumentParams) (int32, error) {
	return q.CreateDocument(ctx, db.CreateDocumentParams{
		Title:        p.Title,
		Description:  sql.NullString{String: p.Description, Valid: p.Description != ""},
		TypeID:       toNullInt32(p.TypeID),
		FolderID:     p.FolderID,
		FilePath:     p.FilePath,
		OriginalName: p.OriginalName,
		DepartmentID: p.DepartmentID,
		CreatedBy:    p.CreatedBy,
		AssigneeID:   toNullInt32(nil),
	})
}

func MoveDocument(ctx context.Context, q *db.Queries, docID, folderID int32, actorLogin string) error {
	if err := q.MoveDocument(ctx, db.MoveDocumentParams{ID: docID, FolderID: folderID}); err != nil {
		return err
	}
	return q.AddDocumentHistory(ctx, db.AddDocumentHistoryParams{
		DocumentID: docID,
		ActorLogin: actorLogin,
		Action:     "MOVED",
		Details:    sql.NullString{String: "moved to folder", Valid: true},
	})
}


func MoveToHeadOnly(ctx context.Context, q *db.Queries, docID int32, deptID int32, actorLogin string) error {
	folder, err := findSystemFolder(ctx, q, deptID, "head_only")
	if err != nil {
		return err
	}
	return MoveDocument(ctx, q, docID, folder.ID, actorLogin)
}

func HideDocument(ctx context.Context, q *db.Queries, docID int32) error {
	return q.HideDocument(ctx, docID)
}

func UnhideDocument(ctx context.Context, q *db.Queries, docID int32) error {
	return q.UnhideDocument(ctx, docID)
}

func ChangeDocumentAssignee(ctx context.Context, q *db.Queries, docID, assigneeID int32) error {
	return q.ChangeDocumentAssignee(ctx, db.ChangeDocumentAssigneeParams{
		ID:         docID,
		AssigneeID: toNullInt32(&assigneeID),
	})
}

func UpdateDocumentStatus(ctx context.Context, q *db.Queries, docID int32, status string) error {
	return q.UpdateDocumentStatus(ctx, db.UpdateDocumentStatusParams{
		ID:     docID,
		Status: status,
	})
}

func DeleteDocument(ctx context.Context, q *db.Queries, id int32) error {
	return q.DeleteDocument(ctx, id)
}

func GetDocumentHistory(ctx context.Context, q *db.Queries, docID int32) ([]db.GetDocumentHistoryRow, error) {
	return q.GetDocumentHistory(ctx, docID)
}

func AddHistory(ctx context.Context, q *db.Queries, docID int32, actor, action, details string) error {
	return q.AddDocumentHistory(ctx, db.AddDocumentHistoryParams{
		DocumentID: docID,
		ActorLogin: actor,
		Action:     action,
		Details:    sql.NullString{String: details, Valid: details != ""},
	})
}



func toNullInt32(v *int32) sql.NullInt32 {
	if v == nil {
		return sql.NullInt32{Valid: false}
	}
	return sql.NullInt32{Int32: *v, Valid: true}
}

func findSystemFolder(ctx context.Context, q *db.Queries, deptID int32, name string) (db.ListFoldersByDepartmentRow, error) {
	folders, err := q.ListFoldersByDepartment(ctx, deptID)
	if err != nil {
		return db.ListFoldersByDepartmentRow{}, err
	}
	for _, f := range folders {
		if f.Name == name && f.IsSystem {
			return f, nil
		}
	}
	return db.ListFoldersByDepartmentRow{}, errors.New("system folder not found: " + name)
}
