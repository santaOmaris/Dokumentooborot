package dto


type RegisterRequest struct {
	Login      string `json:"login"`
	Email      string `json:"email"`
	Password   string `json:"password"`
	FullName   string `json:"full_name"`
	IsHead     bool   `json:"is_head"`
	SystemRole string `json:"system_role"`
}

type RegisterResponse struct {
	UserID int32 `json:"user_id"`
}

type AuthRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}


type FireRequest struct {
	UserID int32 `json:"user_id"`
}

type MoveUserRequest struct {
	UserID       int32 `json:"user_id"`
	DepartmentID int32 `json:"department_id"`
}


type CreateDepartmentRequest struct {
	Name     string `json:"name"`
	ParentID *int32 `json:"parent_id,omitempty"`
}

type SetParentRequest struct {
	ParentID *int32 `json:"parent_id"`
}
