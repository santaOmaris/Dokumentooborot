#!/usr/bin/env bash

set -u

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPORTS_DIR="${ROOT_DIR}/.test-reports"
RUN_ID="$(date +%Y%m%d_%H%M%S)"
RUN_DIR="${REPORTS_DIR}/${RUN_ID}"
MERGED_COVER="${RUN_DIR}/all.cover"

mkdir -p "${RUN_DIR}"

failed=0

run_module_tests() {
  local module_name="$1"
  local module_dir="$2"
  local cover_file="$3"

  echo ""
  echo "==> ${module_name}"
  echo "    dir: ${module_dir}"

  if (cd "${ROOT_DIR}/${module_dir}" && go test ./... -covermode=atomic -coverprofile="${cover_file}"); then
    echo "[OK] ${module_name}"
  else
    echo "[FAIL] ${module_name}"
    failed=1
  fi
}

run_module_tests "pkg" "pkg" "${RUN_DIR}/pkg.cover"
run_module_tests "iam-service" "go-services/iam-service" "${RUN_DIR}/iam.cover"
run_module_tests "catalog-service" "go-services/catalog-service" "${RUN_DIR}/catalog.cover"
run_module_tests "collaboration-service" "go-services/collaboration-service" "${RUN_DIR}/collaboration.cover"
run_module_tests "orchestrator-service" "go-services/orchestrator-service" "${RUN_DIR}/orchestrator.cover"

cover_files=(
  "${RUN_DIR}/pkg.cover"
  "${RUN_DIR}/iam.cover"
  "${RUN_DIR}/catalog.cover"
  "${RUN_DIR}/collaboration.cover"
  "${RUN_DIR}/orchestrator.cover"
)

first_cover=""
for f in "${cover_files[@]}"; do
  if [[ -s "${f}" ]]; then
    first_cover="${f}"
    break
  fi
done

if [[ -n "${first_cover}" ]]; then
  head -n 1 "${first_cover}" > "${MERGED_COVER}"
  for f in "${cover_files[@]}"; do
    if [[ -s "${f}" ]]; then
      tail -n +2 "${f}" >> "${MERGED_COVER}"
    fi
  done

  echo ""
  echo "==> Total coverage"
  go tool cover -func="${MERGED_COVER}" | tail -n 1
  echo "Coverage file: ${MERGED_COVER}"
else
  echo ""
  echo "Coverage files were not created."
fi

echo ""
echo "Test report directory: ${RUN_DIR}"

if [[ "${failed}" -ne 0 ]]; then
  echo ""
  echo "At least one module failed tests."
  exit 1
fi

echo ""
echo "All module tests passed."
exit 0
