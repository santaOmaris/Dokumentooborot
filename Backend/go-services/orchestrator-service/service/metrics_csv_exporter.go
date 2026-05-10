package service

import (
	"context"
	"encoding/csv"
	"fmt"
	"hash/fnv"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	db "orchestrator-service/db/generated"
)

type metricsCSVExporter struct {
	dir                        string
	mu                         sync.Mutex
	initialized                bool
	lastTransitionID           int32
	lastStateDistributionHash  string
	lastTransitionMatrixHash   string
	lastActorActivityHash      string
	lastHourlyTransitionsHash  string
	lastConversionsHash        string
}

func newMetricsCSVExporter(dir string) *metricsCSVExporter {
	if dir == "" {
		return nil
	}
	return &metricsCSVExporter{dir: dir}
}

func (e *metricsCSVExporter) ExportSnapshot(ctx context.Context, q *db.Queries, metrics map[string]any) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if err := os.MkdirAll(e.dir, 0o755); err != nil {
		return fmt.Errorf("mkdir metrics dir: %w", err)
	}
	if err := e.initFromDisk(); err != nil {
		return fmt.Errorf("init metrics exporter state: %w", err)
	}

	loggedAt := time.Now().Format(time.RFC3339)

	stateDistribution, err := q.GetStateDistribution(ctx)
	if err != nil {
		return err
	}
	transitionMatrix, err := q.GetTransitionMatrixLast24h(ctx)
	if err != nil {
		return err
	}
	actorActivity, err := q.GetActorActivityLast24h(ctx)
	if err != nil {
		return err
	}
	hourlyTransitions, err := q.GetHourlyTransitionsLast24h(ctx)
	if err != nil {
		return err
	}
	newTransitions, err := q.GetTransitionsSinceID(ctx, e.lastTransitionID)
	if err != nil {
		return err
	}

	if err = appendCSVRows(filepath.Join(e.dir, "metrics_snapshot.csv"),
		[]string{
			"logged_at",
			"service_uptime_seconds",
			"go_goroutines",
			"go_heap_alloc_bytes",
			"go_heap_objects",
			"workflow_documents_total",
			"workflow_transitions_total",
			"transitions_last_24h",
			"documents_draft",
			"documents_pending_visa",
			"documents_pending_boss",
			"documents_approved",
			"documents_rejected",
			"documents_updated_last_24h",
			"active_actors_last_24h",
		},
		[][]string{{
			loggedAt,
			metricString(metrics, "service_uptime_seconds"),
			metricString(metrics, "go_goroutines"),
			metricString(metrics, "go_heap_alloc_bytes"),
			metricString(metrics, "go_heap_objects"),
			metricString(metrics, "workflow_documents_total"),
			metricString(metrics, "workflow_transitions_total"),
			metricString(metrics, "transitions_last_24h"),
			metricString(metrics, "documents_draft"),
			metricString(metrics, "documents_pending_visa"),
			metricString(metrics, "documents_pending_boss"),
			metricString(metrics, "documents_approved"),
			metricString(metrics, "documents_rejected"),
			metricString(metrics, "documents_updated_last_24h"),
			metricString(metrics, "active_actors_last_24h"),
		}},
	); err != nil {
		return err
	}

	stateRows := make([][]string, 0, len(stateDistribution))
	stateHashRows := make([][]string, 0, len(stateDistribution))
	for _, row := range stateDistribution {
		documentsCount := strconv.FormatInt(row.DocumentsCount, 10)
		stateRows = append(stateRows, []string{loggedAt, row.State, documentsCount})
		stateHashRows = append(stateHashRows, []string{row.State, documentsCount})
	}
	if err = appendCSVRowsIfChanged(
		filepath.Join(e.dir, "state_distribution.csv"),
		[]string{"logged_at", "state", "documents_count"},
		stateRows,
		rowsHash(stateHashRows),
		&e.lastStateDistributionHash,
	); err != nil {
		return err
	}

	matrixRows := make([][]string, 0, len(transitionMatrix))
	matrixHashRows := make([][]string, 0, len(transitionMatrix))
	for _, row := range transitionMatrix {
		transitionsCount := strconv.FormatInt(row.TransitionsCount, 10)
		matrixRows = append(matrixRows, []string{loggedAt, row.FromState, row.ToState, transitionsCount})
		matrixHashRows = append(matrixHashRows, []string{row.FromState, row.ToState, transitionsCount})
	}
	if err = appendCSVRowsIfChanged(
		filepath.Join(e.dir, "transition_matrix_24h.csv"),
		[]string{"logged_at", "from_state", "to_state", "transitions_count_24h"},
		matrixRows,
		rowsHash(matrixHashRows),
		&e.lastTransitionMatrixHash,
	); err != nil {
		return err
	}

	actorRows := make([][]string, 0, len(actorActivity))
	actorHashRows := make([][]string, 0, len(actorActivity))
	for _, row := range actorActivity {
		transitionsCount := strconv.FormatInt(row.TransitionsCount, 10)
		actorRows = append(actorRows, []string{loggedAt, row.ActorLogin, transitionsCount})
		actorHashRows = append(actorHashRows, []string{row.ActorLogin, transitionsCount})
	}
	if err = appendCSVRowsIfChanged(
		filepath.Join(e.dir, "actor_activity_24h.csv"),
		[]string{"logged_at", "actor_login", "transitions_count_24h"},
		actorRows,
		rowsHash(actorHashRows),
		&e.lastActorActivityHash,
	); err != nil {
		return err
	}

	hourlyRows := make([][]string, 0, len(hourlyTransitions))
	hourlyHashRows := make([][]string, 0, len(hourlyTransitions))
	for _, row := range hourlyTransitions {
		transitionsCount := strconv.FormatInt(row.TransitionsCount, 10)
		hourlyRows = append(hourlyRows, []string{loggedAt, row.HourBucket, transitionsCount})
		hourlyHashRows = append(hourlyHashRows, []string{row.HourBucket, transitionsCount})
	}
	if err = appendCSVRowsIfChanged(
		filepath.Join(e.dir, "hourly_transitions_24h.csv"),
		[]string{"logged_at", "hour_bucket", "transitions_count"},
		hourlyRows,
		rowsHash(hourlyHashRows),
		&e.lastHourlyTransitionsHash,
	); err != nil {
		return err
	}

	feedRows := make([][]string, 0, len(newTransitions))
	maxTransitionID := e.lastTransitionID
	for _, row := range newTransitions {
		revisionNote := ""
		if row.RevisionNote.Valid {
			revisionNote = row.RevisionNote.String
		}
		feedRows = append(feedRows, []string{
			strconv.FormatInt(int64(row.ID), 10),
			strconv.FormatInt(int64(row.DocumentID), 10),
			row.ActorLogin,
			row.FromState,
			row.ToState,
			revisionNote,
			row.CreatedAt.Format(time.RFC3339),
		})
		if row.ID > maxTransitionID {
			maxTransitionID = row.ID
		}
	}
	if err = appendCSVRows(filepath.Join(e.dir, "workflow_transitions_feed.csv"),
		[]string{"transition_id", "document_id", "actor_login", "from_state", "to_state", "revision_note", "created_at"}, feedRows,
	); err != nil {
		return err
	}
	e.lastTransitionID = maxTransitionID

	if len(newTransitions) > 0 {
		eventRows := [][]string{{
			strconv.FormatInt(time.Now().UnixNano(), 10),
			"metrics_export_completed",
			"info",
			"orchestrator-service",
			fmt.Sprintf("snapshot exported; new_transitions=%d", len(newTransitions)),
			strconv.Itoa(len(newTransitions)),
			loggedAt,
		}}
		if err = appendCSVRows(filepath.Join(e.dir, "orchestrator_system_events.csv"),
			[]string{"event_id", "event_type", "severity", "source", "details", "event_value", "logged_at"}, eventRows,
		); err != nil {
			return err
		}
	} else {
		if err = appendCSVRows(filepath.Join(e.dir, "orchestrator_system_events.csv"),
			[]string{"event_id", "event_type", "severity", "source", "details", "event_value", "logged_at"}, nil,
		); err != nil {
			return err
		}
	}

	conversionRows := [][]string{
		{loggedAt, "approval_rate_24h_percent", approvalRate(transitionMatrix, StateApproved)},
		{loggedAt, "rejection_rate_24h_percent", approvalRate(transitionMatrix, StateRejected)},
	}
	conversionHashRows := [][]string{
		{"approval_rate_24h_percent", conversionRows[0][2]},
		{"rejection_rate_24h_percent", conversionRows[1][2]},
	}
	if err = appendCSVRowsIfChanged(
		filepath.Join(e.dir, "orchestrator_conversions.csv"),
		[]string{"logged_at", "metric_name", "metric_value"},
		conversionRows,
		rowsHash(conversionHashRows),
		&e.lastConversionsHash,
	); err != nil {
		return err
	}

	return nil
}

func (e *metricsCSVExporter) initFromDisk() error {
	if e.initialized {
		return nil
	}
	e.initialized = true

	path := filepath.Join(e.dir, "workflow_transitions_feed.csv")
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer f.Close()

	r := csv.NewReader(f)
	for {
		rec, readErr := r.Read()
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return readErr
		}
		if len(rec) == 0 || rec[0] == "transition_id" {
			continue
		}
		id, parseErr := strconv.ParseInt(strings.TrimSpace(rec[0]), 10, 32)
		if parseErr != nil {
			continue
		}
		if int32(id) > e.lastTransitionID {
			e.lastTransitionID = int32(id)
		}
	}

	return nil
}

func approvalRate(matrix []db.GetTransitionMatrixLast24hRow, state string) string {
	var total int64
	var matched int64
	for _, row := range matrix {
		total += row.TransitionsCount
		if row.ToState == state {
			matched += row.TransitionsCount
		}
	}
	if total == 0 {
		return "0"
	}
	rate := (float64(matched) / float64(total)) * 100
	return fmt.Sprintf("%.2f", rate)
}

func metricString(metrics map[string]any, key string) string {
	value, ok := metrics[key]
	if !ok {
		return "0"
	}
	switch v := value.(type) {
	case int:
		return strconv.Itoa(v)
	case int8:
		return strconv.FormatInt(int64(v), 10)
	case int16:
		return strconv.FormatInt(int64(v), 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint8:
		return strconv.FormatUint(uint64(v), 10)
	case uint16:
		return strconv.FormatUint(uint64(v), 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case float32:
		return fmt.Sprintf("%.2f", v)
	case float64:
		return fmt.Sprintf("%.2f", v)
	case string:
		return v
	default:
		return fmt.Sprint(v)
	}
}

func rowsHash(rows [][]string) string {
	h := fnv.New64a()
	for _, row := range rows {
		for i, col := range row {
			_, _ = h.Write([]byte(col))
			if i < len(row)-1 {
				_, _ = h.Write([]byte{0})
			}
		}
		_, _ = h.Write([]byte{'\n'})
	}
	return strconv.FormatUint(h.Sum64(), 16)
}

func appendCSVRowsIfChanged(path string, header []string, rows [][]string, nextHash string, lastHash *string) error {
	if *lastHash == nextHash {
		return appendCSVRows(path, header, nil)
	}
	if err := appendCSVRows(path, header, rows); err != nil {
		return err
	}
	*lastHash = nextHash
	return nil
}

func appendCSVRows(path string, header []string, rows [][]string) error {
	if len(rows) == 0 {
		return ensureCSVHeader(path, header)
	}

	isNewFile := false
	st, err := os.Stat(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		isNewFile = true
	} else if st.Size() == 0 {
		isNewFile = true
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	if isNewFile {
		if err = w.Write(header); err != nil {
			return err
		}
	}
	for _, row := range rows {
		if err = w.Write(row); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

func ensureCSVHeader(path string, header []string) error {
	st, err := os.Stat(path)
	if err == nil && st.Size() > 0 {
		return nil
	}
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	if err = w.Write(header); err != nil {
		return err
	}
	w.Flush()
	return w.Error()
}
