package mq


const (
	ActionDocApproved        = "DOC_APPROVED"
	ActionDocRejected        = "DOC_REJECTED"
	ActionDocDelegated       = "DOC_DELEGATED"
	ActionApprovalRequested  = "DOC_APPROVAL_REQUESTED"
	ActionDocSentForApproval = "DOC_SENT_FOR_APPROVAL"
)


type AuditEvent struct {
	DepartmentID int32  `json:"department_id"`
	DocumentID   int32  `json:"document_id"`
	ActorLogin   string `json:"actor_login"`
	Action       string `json:"action"`
	Details      string `json:"details"`
}


type NotificationEvent struct {
	RecipientEmail string `json:"recipient_email"`
	RecipientName  string `json:"recipient_name"`
	Action         string `json:"action"`
	DocumentTitle  string `json:"document_title"`
	Details        string `json:"details"`
}
