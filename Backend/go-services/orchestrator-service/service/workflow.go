package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/looplab/fsm"

	"docflow.local/pkg/mq"
	catalogpb "docflow.local/pkg/pb/catalog"
	iampb "docflow.local/pkg/pb/iam"
	db "orchestrator-service/db/generated"
	"orchestrator-service/publisher"
)

const (
	StateDraft       = "DRAFT"
	StatePendingBoss = "PENDING_BOSS"
	StatePendingVisa = "PENDING_VISA"
	StateApproved    = "APPROVED"
	StateRejected    = "REJECTED"
)

const (
	EventSendForVisa     = "send_for_visa"
	EventRequestApproval = "request_approval"
	EventApprove         = "approve"
	EventReject          = "reject"
	EventReset           = "reset"
)

type WorkflowService struct {
	q         *db.Queries
	catalog   catalogpb.CatalogServiceClient
	iam       iampb.IAMServiceClient
	publisher *publisher.Publisher
	startedAt time.Time
	csvExport *metricsCSVExporter
}

func New(q *db.Queries, catalog catalogpb.CatalogServiceClient, iam iampb.IAMServiceClient, pub *publisher.Publisher, metricsCSVDir string) *WorkflowService {
	return &WorkflowService{
		q:         q,
		catalog:   catalog,
		iam:       iam,
		publisher: pub,
		startedAt: time.Now(),
		csvExport: newMetricsCSVExporter(metricsCSVDir),
	}
}

func (s *WorkflowService) EnsureHeadCanReview(ctx context.Context, documentID int32, actorDeptID int32) error {
	if actorDeptID == 0 {
		return errors.New("forbidden: department is required")
	}
	docResp, err := s.catalog.GetDocument(ctx, &catalogpb.GetDocumentRequest{DocumentId: documentID})
	if err != nil {
		return fmt.Errorf("catalog.GetDocument: %w", err)
	}
	if docResp.DepartmentId == 0 || docResp.DepartmentId != actorDeptID {
		return errors.New("forbidden: only head of target department can approve or reject")
	}
	return nil
}

func (s *WorkflowService) GetMetrics(ctx context.Context) (map[string]any, error) {
	statesCount, err := s.q.CountDocumentStates(ctx)
	if err != nil {
		return nil, err
	}
	totalTransitions, err := s.q.CountTransitionsTotal(ctx)
	if err != nil {
		return nil, err
	}
	transitions24h, err := s.q.CountTransitionsLast24h(ctx)
	if err != nil {
		return nil, err
	}
	pendingVisa, err := s.q.CountDocumentsByState(ctx, StatePendingVisa)
	if err != nil {
		return nil, err
	}
	pendingBoss, err := s.q.CountDocumentsByState(ctx, StatePendingBoss)
	if err != nil {
		return nil, err
	}
	approved, err := s.q.CountDocumentsByState(ctx, StateApproved)
	if err != nil {
		return nil, err
	}
	rejected, err := s.q.CountDocumentsByState(ctx, StateRejected)
	if err != nil {
		return nil, err
	}
	draft, err := s.q.CountDocumentsByState(ctx, StateDraft)
	if err != nil {
		return nil, err
	}
	updatedLast24h, err := s.q.CountDocumentsUpdatedLast24h(ctx)
	if err != nil {
		return nil, err
	}
	activeActors24h, err := s.q.CountDistinctActorsLast24h(ctx)
	if err != nil {
		return nil, err
	}

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	metrics := map[string]any{
		"service_uptime_seconds":     int(time.Since(s.startedAt).Seconds()),
		"go_goroutines":              runtime.NumGoroutine(),
		"go_heap_alloc_bytes":        mem.HeapAlloc,
		"go_heap_objects":            mem.HeapObjects,
		"workflow_documents_total":   statesCount,
		"workflow_transitions_total": totalTransitions,
		"transitions_last_24h":       transitions24h,
		"documents_pending_visa":     pendingVisa,
		"documents_pending_boss":     pendingBoss,
		"documents_approved":         approved,
		"documents_rejected":         rejected,
		"documents_draft":            draft,
		"documents_updated_last_24h": updatedLast24h,
		"active_actors_last_24h":     activeActors24h,
	}

	return metrics, nil
}

func (s *WorkflowService) LogMetricsSnapshot(ctx context.Context) error {
	metrics, err := s.GetMetrics(ctx)
	if err != nil {
		return err
	}

	timestamp := time.Now().Format(time.RFC3339)
	metricKeys := []string{
		"service_uptime_seconds",
		"go_goroutines",
		"go_heap_alloc_bytes",
		"go_heap_objects",
		"workflow_documents_total",
		"documents_draft",
		"workflow_transitions_total",
		"transitions_last_24h",
		"documents_pending_visa",
		"documents_pending_boss",
		"documents_approved",
		"documents_rejected",
		"documents_updated_last_24h",
		"active_actors_last_24h",
	}

	for _, key := range metricKeys {
		log.Printf("metric_log ts=%s metric=%s value=%v", timestamp, key, metrics[key])
	}

	if s.csvExport != nil {
		if err = s.csvExport.ExportSnapshot(ctx, s.q, metrics); err != nil {
			return fmt.Errorf("metrics csv export: %w", err)
		}
		log.Printf("metrics_csv_export ts=%s dir=%s", timestamp, s.csvExport.dir)
	}
	return nil
}

func (s *WorkflowService) MetricsCSVDir() string {
	if s.csvExport == nil {
		return ""
	}
	return s.csvExport.dir
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

func (s *WorkflowService) SendForVisa(ctx context.Context, documentID int32, actorLogin, note string, approverID int32) error {
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
	if approverID != 0 {
		matched := false
		for _, m := range mgrsResp.Managers {
			if m.Id == approverID {
				approver = m
				matched = true
				break
			}
		}
		if !matched {
			return errors.New("selected approver is not a manager of this department")
		}
	}

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

	if cur.State == StatePendingBoss {
		archiveFolder, folderErr := s.catalog.GetSystemFolder(ctx, &catalogpb.GetSystemFolderRequest{
			DepartmentId: docResp.DepartmentId,
			Name:         "archived",
		})
		if folderErr != nil {
			return fmt.Errorf("catalog.GetSystemFolder(archived): %w", folderErr)
		}
		if _, folderErr = s.catalog.MoveDocument(ctx, &catalogpb.MoveDocumentRequest{
			DocumentId: documentID,
			FolderId:   archiveFolder.Id,
			ActorLogin: actorLogin,
		}); folderErr != nil {
			return fmt.Errorf("catalog.MoveDocument(to archived): %w", folderErr)
		}
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

	if cur.State == StatePendingBoss {
		mainFolder, folderErr := s.catalog.GetSystemFolder(ctx, &catalogpb.GetSystemFolderRequest{
			DepartmentId: docResp.DepartmentId,
			Name:         "main",
		})
		if folderErr != nil {
			return fmt.Errorf("catalog.GetSystemFolder(main): %w", folderErr)
		}
		if _, folderErr = s.catalog.MoveDocument(ctx, &catalogpb.MoveDocumentRequest{
			DocumentId: documentID,
			FolderId:   mainFolder.Id,
			ActorLogin: actorLogin,
		}); folderErr != nil {
			return fmt.Errorf("catalog.MoveDocument(to main): %w", folderErr)
		}
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
	cur, err := s.GetStatus(ctx, documentID)
	if err != nil {
		return err
	}
	f := newFSM(cur.State)
	if err = f.Event(ctx, EventRequestApproval); err != nil {
		return fmt.Errorf("fsm: %w", err)
	}

	docResp, err := s.catalog.GetDocument(ctx, &catalogpb.GetDocumentRequest{DocumentId: documentID})
	if err != nil {
		return fmt.Errorf("catalog.GetDocument: %w", err)
	}

	deptID := docResp.DepartmentId
	if targetDepartmentID != 0 {
		deptID = targetDepartmentID
	}
	if deptID == 0 {
		return errors.New("target department is required")
	}

	collabFolder, err := s.catalog.GetSystemFolder(ctx, &catalogpb.GetSystemFolderRequest{
		DepartmentId: deptID,
		Name:         "collaborations",
	})
	if err != nil {
		return fmt.Errorf("catalog.GetSystemFolder(collaborations): %w", err)
	}

	if _, err = s.catalog.MoveDocument(ctx, &catalogpb.MoveDocumentRequest{
		DocumentId: documentID,
		FolderId:   collabFolder.Id,
		ActorLogin: actorLogin,
	}); err != nil {
		return fmt.Errorf("catalog.MoveDocument(to collaborations): %w", err)
	}

	if _, err = s.catalog.UpdateDocumentStatus(ctx, &catalogpb.UpdateDocumentStatusRequest{
		DocumentId: documentID,
		Status:     StatePendingBoss,
	}); err != nil {
		return fmt.Errorf("catalog.UpdateDocumentStatus: %w", err)
	}

	if err = s.upsertState(ctx, documentID, StatePendingBoss, 0, question); err != nil {
		return err
	}
	if err = s.insertTransition(ctx, documentID, actorLogin, cur.State, StatePendingBoss, question); err != nil {
		return err
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
		DepartmentID: deptID,
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
		{Name: EventRequestApproval, Src: []string{StateDraft, StatePendingVisa}, Dst: StatePendingBoss},
		{Name: EventApprove, Src: []string{StatePendingVisa, StatePendingBoss}, Dst: StateApproved},
		{Name: EventReject, Src: []string{StatePendingVisa, StatePendingBoss}, Dst: StateRejected},
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
