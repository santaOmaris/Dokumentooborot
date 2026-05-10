package dto


type CreateDocumentTypeRequest struct {
	Name string `json:"name"`
}


type CreateFolderRequest struct {
	ParentID *int32 `json:"parent_id,omitempty"`
	Name     string `json:"name"`
}

type RenameFolderRequest struct {
	Name string `json:"name"`
}

type MoveFolderRequest struct {
	ParentID *int32 `json:"parent_id"`
}


type UploadDocumentRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	TypeID      *int32 `json:"type_id,omitempty"`
	FolderID    int32  `json:"folder_id"`
}

type MoveDocumentRequest struct {
	FolderID int32 `json:"folder_id"`
}

type AssigneeRequest struct {
	AssigneeID int32 `json:"assignee_id"`
}

type SearchRequest struct {
	Query string `json:"query"`
}
