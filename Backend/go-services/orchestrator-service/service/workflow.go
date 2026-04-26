package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/looplab/fsm"

	catalogpb "docflow.local/pkg/pb/catalog"
	iampb "docflow.local/pkg/pb/iam"
	"docflow.local/pkg/mq"
	db "orchestrator-service/db/generated"
	"orchestrator-service/publisher"
)

const (
	StateDraft       = "DRAFT"
	StatePendingVisa = "PENDING_VISA"
	StateApproved    = "APPROVED"
	StateRejected    = "REJECTED"
)

const (
	EventSendForVisa = "send_for_visa"
	EventApprove     = "approve"
	EventReject      = "reject"
	EventReset       = "reset"
)

type WorkflowService struct {
	q         *db.Queries
	catalog   catalogpb.CatalogServiceClient
	iam       iampb.IAMServiceClient
	publisher *publisher.Publisher
}

func New(q *db.Queries, catalog catalogpb.CatalogServiceClient, iam iampb.IAMServiceClient, pub *publisher.Publisher) *WorkflowService {
	return &WorkflowService{q: q, catalog: catalog, iam: iam, publisher: pub}
}

func (s *WorkflowService) GetStatus(ctx context.Context, documentID int32) (*db.DocumentState, error) {
	state, err := s.q.GetDocumentState(ctx, documentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &db.DocumentState{DocumentID: documentID, State: StateDraft}, nil
		}
		return nil, err
	}
	return &state, nil
}

func (s *WorkflowService) GetHistory(ctx context.Context, documentID int32) ([]db.StateTransition, error) {
	return s.q.GetTransitionHistory(ctx, documentID)
}

func (s *WorkflowService) SendForVisa(ctx context.Context, documentID int32, actorLogin, note string) error {
	cur, err := s.GetStatus(ctx, documentID)
	if err != nil {
		return err
	}
	f := newFSM(cur.State)
	if err = f.Event(ctx, EventSendForVisa); err != nil {
		return fmt.Errorf("fsm: %w", err)
	}

	docResp, err := s.catalog.GetDocument(ctx, &catalogpb.GetDocumentRequest{DocumentId: documentID})
	if err != nil {
		return fmt.Errorf("catalog.GetDocument: %w", err)
	}

	mgrsResp, err := s.iam.GetManagersByDepartment(ctx, &iampb.GetManagersByDepartmentRequest{DepartmentId: docResp.DepartmentId})
	if err != nil || len(mgrsResp.Managers) == 0 {
		return errors.New("no manager found for department")
	}
	approver := mgrsResp.Managers[0]

	if _, err = s.catalog.ChangeDocumentAssignee(ctx, &catalogpb.ChangeAssigneeRequest{
		DocumentId: documentID, NewAssigneeId: approver.Id,
	}); err != nil {
		return fmt.Errorf("catalog.ChangeDocumentAssignee: %w", err)
	}

	if _, err = s.catalog.UpdateDocumentStatus(ctx, &catalogpb.UpdateDocumentStatusRequest{
		DocumentId: documentID, Status: StatePendingVisa,
	}); err != nil {
		return fmt.Errorf("catalog.UpdateDocumentStatus: %w", err)
	}

	if err = s.upsertState(ctx, documentID, StatePendingVisa, approver.Id, ""); err != nil {
		return err
	}
	if err = s.insertTransition(ctx, documentID, actorLogin, cur.State, StatePendingVisa, note); err != nil {
		return err
	}

	auditDetails := fmt.Sprintf("назначен визировщик: %s", approver.Login)
	if note != "" {
		auditDetails = fmt.Sprintf("%s; примечание: %s", auditDetails, note)
	}
	notificationDetails := fmt.Sprintf("Документ ожидает вашего решения. Отправил: %s", actorLogin)
	if note != "" {
		notificationDetails = fmt.Sprintf("%s. Примечание: %s", notificationDetails, note)
	}

	s.publisher.PublishAudit(mq.AuditEvent{
		DepartmentID: docResp.DepartmentId,
		DocumentID:   documentID,
		ActorLogin:   actorLogin,
		Action:       mq.ActionDocSentForApproval,
		Details:      auditDetails,
	})
	s.publisher.PublishNotification(mq.NotificationEvent{
		RecipientEmail: approver.Email,
		RecipientName:  approver.FullName,
		Action:         mq.ActionDocSentForApproval,
		DocumentTitle:  docResp.Title,
		Details:        notificationDetails,
	})

	return nil
}

func (s *WorkflowService) Approve(ctx context.Context, documentID int32, actorLogin string) error {
	cur, err := s.GetStatus(ctx, documentID)
	if err != nil {
		return err
	}
	f := newFSM(cur.State)
	if err = f.Event(ctx, EventApprove); err != nil {
		return fmt.Errorf("fsm: %w", err)
	}

	docResp, err := s.catalog.GetDocument(ctx, &catalogpb.GetDocumentRequest{DocumentId: documentID})
	if err != nil {
		return fmt.Errorf("catalog.GetDocument: %w", err)
	}

	if _, err = s.catalog.UpdateDocumentStatus(ctx, &catalogpb.UpdateDocumentStatusRequest{
		DocumentId: documentID, Status: StateApproved,
	}); err != nil {
		return fmt.Errorf("catalog.UpdateDocumentStatus: %w", err)
	}

	if err = s.upsertState(ctx, documentID, StateApproved, cur.ApproverID.Int32, ""); err != nil {
		return err
	}
	if err = s.insertTransition(ctx, documentID, actorLogin, cur.State, StateApproved, ""); err != nil {
		return err
	}

	s.publisher.PublishAudit(mq.AuditEvent{
		DepartmentID: docResp.DepartmentId,
		DocumentID:   documentID,
		ActorLogin:   actorLogin,
		Action:       mq.ActionDocApproved,
		Details:      "документ успешно завизирован",
	})
	creatorEmail := s.fetchEmail(ctx, docResp.CreatedBy)
	if creatorEmail != nil {
		s.publisher.PublishNotification(mq.NotificationEvent{
			RecipientEmail: creatorEmail.email,
			RecipientName:  creatorEmail.fullName,
			Action:         mq.ActionDocApproved,
			DocumentTitle:  docResp.Title,
			Details:        fmt.Sprintf("Документ завизирован пользователем %s", actorLogin),
		})
	}

	return nil
}

func (s *WorkflowService) Reject(ctx context.Context, documentID int32, actorLogin, revisionNote string) error {
	cur, err := s.GetStatus(ctx, documentID)
	if err != nil {
		return err
	}
	f := newFSM(cur.State)
	if err = f.Event(ctx, EventReject); err != nil {
		return fmt.Errorf("fsm: %w", err)
	}

	docResp, err := s.catalog.GetDocument(ctx, &catalogpb.GetDocumentRequest{DocumentId: documentID})
	if err != nil {
		return fmt.Errorf("catalog.GetDocument: %w", err)
	}

	if _, err = s.catalog.UpdateDocumentStatus(ctx, &catalogpb.UpdateDocumentStatusRequest{
		DocumentId: documentID, Status: StateDraft,
	}); err != nil {
		return fmt.Errorf("catalog.UpdateDocumentStatus: %w", err)
	}

	if err = s.upsertState(ctx, documentID, StateDraft, 0, revisionNote); err != nil {
		return err
	}
	if err = s.insertTransition(ctx, documentID, actorLogin, cur.State, StateDraft, revisionNote); err != nil {
		return err
	}

	s.publisher.PublishAudit(mq.AuditEvent{
		DepartmentID: docResp.DepartmentId,
		DocumentID:   documentID,
		ActorLogin:   actorLogin,
		Action:       mq.ActionDocRejected,
		Details:      revisionNote,
	})
	creatorEmail := s.fetchEmail(ctx, docResp.CreatedBy)
	if creatorEmail != nil {
		s.publisher.PublishNotification(mq.NotificationEvent{
			RecipientEmail: creatorEmail.email,
			RecipientName:  creatorEmail.fullName,
			Action:         mq.ActionDocRejected,
			DocumentTitle:  docResp.Title,
			Details:        revisionNote,
		})
	}

	return nil
}

func (s *WorkflowService) DelegateDocument(ctx context.Context, documentID, newAssigneeID int32, actorLogin string) error {
	docResp, err := s.catalog.GetDocument(ctx, &catalogpb.GetDocumentRequest{DocumentId: documentID})
	if err != nil {
		return fmt.Errorf("catalog.GetDocument: %w", err)
	}

	if _, err = s.catalog.ChangeDocumentAssignee(ctx, &catalogpb.ChangeAssigneeRequest{
		DocumentId: documentID, NewAssigneeId: newAssigneeID,
	}); err != nil {
		return fmt.Errorf("catalog.ChangeDocumentAssignee: %w", err)
	}

	cur, _ := s.GetStatus(ctx, documentID)
	s.upsertState(ctx, documentID, cur.State, newAssigneeID, "") //nolint

	s.publisher.PublishAudit(mq.AuditEvent{
		DepartmentID: docResp.DepartmentId,
		DocumentID:   documentID,
		ActorLogin:   actorLogin,
		Action:       mq.ActionDocDelegated,
		Details:      fmt.Sprintf("документ делегирован пользователю id=%d", newAssigneeID),
	})

	assigneeEmail := s.fetchEmail(ctx, newAssigneeID)
	if assigneeEmail != nil {
		s.publisher.PublishNotification(mq.NotificationEvent{
			RecipientEmail: assigneeEmail.email,
			RecipientName:  assigneeEmail.fullName,
			Action:         mq.ActionDocDelegated,
			DocumentTitle:  docResp.Title,
			Details:        fmt.Sprintf("Делегирован пользователем %s", actorLogin),
		})
	}

	return nil
}

func (s *WorkflowService) RequestApproval(ctx context.Context, documentID int32, actorLogin, question string, targetDepartmentID int32) error {
	docResp, err := s.catalog.GetDocument(ctx, &catalogpb.GetDocumentRequest{DocumentId: documentID})
	if err != nil {
		return fmt.Errorf("catalog.GetDocument: %w", err)
	}

	deptID := docResp.DepartmentId
	if targetDepartmentID != 0 {
		deptID = targetDepartmentID
	}

	mgrsResp, err := s.iam.GetManagersByDepartment(ctx, &iampb.GetManagersByDepartmentRequest{DepartmentId: deptID})
	if err != nil || len(mgrsResp.Managers) == 0 {
		return errors.New("no manager found for department")
	}

	for _, mgr := range mgrsResp.Managers {
		s.publisher.PublishNotification(mq.NotificationEvent{
			RecipientEmail: mgr.Email,
			RecipientName:  mgr.FullName,
			Action:         mq.ActionApprovalRequested,
			DocumentTitle:  docResp.Title,
			Details:        question,
		})
	}

	s.publisher.PublishAudit(mq.AuditEvent{
		DepartmentID: docResp.DepartmentId,
		DocumentID:   documentID,
		ActorLogin:   actorLogin,
		Action:       mq.ActionApprovalRequested,
		Details:      question,
	})

	return nil
}


func newFSM(currentState string) *fsm.FSM {
	return fsm.NewFSM(currentState, fsm.Events{
		{Name: EventSendForVisa, Src: []string{StateDraft}, Dst: StatePendingVisa},
		{Name: EventApprove, Src: []string{StatePendingVisa}, Dst: StateApproved},
		{Name: EventReject, Src: []string{StatePendingVisa}, Dst: StateRejected},
		{Name: EventReset, Src: []string{StateRejected}, Dst: StateDraft},
	}, fsm.Callbacks{})
}

func (s *WorkflowService) upsertState(ctx context.Context, docID int32, state string, approverID int32, revisionNote string) error {
	return s.q.UpsertDocumentState(ctx, db.UpsertDocumentStateParams{
		DocumentID:   docID,
		State:        state,
		ApproverID:   sql.NullInt32{Int32: approverID, Valid: approverID != 0},
		RevisionNote: sql.NullString{String: revisionNote, Valid: revisionNote != ""},
	})
}

func (s *WorkflowService) insertTransition(ctx context.Context, docID int32, actorLogin, from, to, revisionNote string) error {
	return s.q.InsertStateTransition(ctx, db.InsertStateTransitionParams{
		DocumentID:   docID,
		ActorLogin:   actorLogin,
		FromState:    from,
		ToState:      to,
		RevisionNote: sql.NullString{String: revisionNote, Valid: revisionNote != ""},
	})
}

type emailInfo struct{ email, fullName string }

func (s *WorkflowService) fetchEmail(ctx context.Context, userID int32) *emailInfo {
	if userID == 0 {
		return nil
	}
	resp, err := s.iam.GetUserEmail(ctx, &iampb.GetUserEmailRequest{UserId: userID})
	if err != nil {
		return nil
	}
	return &emailInfo{email: resp.Email, fullName: resp.FullName}
}
