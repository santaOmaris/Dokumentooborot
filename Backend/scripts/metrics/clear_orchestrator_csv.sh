#!/usr/bin/env sh
set -eu

ROOT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)"
METRICS_DIR="$ROOT_DIR/metrics"

mkdir -p "$METRICS_DIR"

rm -f \
  "$METRICS_DIR/metrics_snapshot.csv" \
  "$METRICS_DIR/state_distribution.csv" \
  "$METRICS_DIR/transition_matrix_24h.csv" \
  "$METRICS_DIR/actor_activity_24h.csv" \
  "$METRICS_DIR/hourly_transitions_24h.csv" \
  "$METRICS_DIR/workflow_transitions_feed.csv" \
  "$METRICS_DIR/orchestrator_system_events.csv" \
  "$METRICS_DIR/orchestrator_conversions.csv"

echo "Cleared orchestrator CSV exports in: $METRICS_DIR"
