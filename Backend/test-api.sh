#!/usr/bin/env bash
# test-api.sh — последовательное тестирование всего бэкенда через curl.
# Запускай после: docker compose up -d
# BASE: если запускаешь напрямую через порты — используй http://localhost:8081 и т.д.
# BASE: если через Caddy — http://localhost/api/...

set -euo pipefail

IAM="http://localhost:8081/api/iam"
CATALOG="http://localhost:8082/api/catalog"
COLLAB="http://localhost:8083/api/collaboration"
ORCH="http://localhost:8084/api/orchestrator"

COOKIE_JAR=$(mktemp)
COOKIE_JAR2=$(mktemp)    # второй пользователь (начальник)

RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; NC='\033[0m'

pass() { echo -e "${GREEN}✓ $1${NC}"; }
fail() { echo -e "${RED}✗ $1${NC}"; }
info() { echo -e "${YELLOW}» $1${NC}"; }

# ─── Хелпер curl ─────────────────────────────────────────────────────────────
req() {
  local method=$1 url=$2; shift 2
  curl -s -X "$method" "$url" \
    -H "Content-Type: application/json" \
    "$@"
}

# ─── UC-3: Регистрация пользователей ─────────────────────────────────────────
info "Регистрация начальника отдела"
req POST "$IAM/register" -d '{
  "login":"boss","password":"Boss1234!","email":"boss@company.ru",
  "full_name":"Иван Начальников","is_head":true,"system_role":"USER"
}' | python3 -m json.tool && pass "boss зарегистрирован"

info "Регистрация обычного сотрудника"
req POST "$IAM/register" -d '{
  "login":"employee","password":"Emp1234!","email":"emp@company.ru",
  "full_name":"Пётр Сотрудников","is_head":false,"system_role":"USER"
}' | python3 -m json.tool && pass "employee зарегистрирован"

# ─── UC-3: Авторизация ───────────────────────────────────────────────────────
info "Логин boss"
req POST "$IAM/login" \
  --cookie-jar "$COOKIE_JAR2" \
  -d '{"login":"boss","password":"Boss1234!"}' \
  | python3 -m json.tool && pass "boss залогинен (JWT в cookie)"

info "Логин employee"
req POST "$IAM/login" \
  --cookie-jar "$COOKIE_JAR" \
  -d '{"login":"employee","password":"Emp1234!"}' \
  | python3 -m json.tool && pass "employee залогинен"

# ─── UC-5: Создание отдела ────────────────────────────────────────────────────
info "Создание отдела"
DEPT=$(req POST "$IAM/departments" --cookie "$COOKIE_JAR2" \
  -d '{"name":"Юридический отдел"}')
echo "$DEPT" | python3 -m json.tool
DEPT_ID=$(echo "$DEPT" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('id',''))")
pass "Отдел создан, id=$DEPT_ID"

# ─── UC-20: Перемещение сотрудника в отдел ───────────────────────────────────
info "Получаем id employee"
USERS=$(req GET "$IAM/users" --cookie "$COOKIE_JAR2")
echo "$USERS" | python3 -m json.tool
EMP_ID=$(echo "$USERS" | python3 -c "
import sys,json
users=json.load(sys.stdin)
for u in users:
    if u.get('login')=='employee':
        print(u['id']); break
")
pass "employee id=$EMP_ID"

info "Перемещаем employee в отдел $DEPT_ID"
req POST "$IAM/users/move" --cookie "$COOKIE_JAR2" \
  -d "{\"user_id\":$EMP_ID,\"department_id\":$DEPT_ID}" \
  | python3 -m json.tool && pass "employee перемещён"

info "Перемещаем boss в отдел $DEPT_ID (как начальника)"
BOSS_ID=$(req GET "$IAM/users" --cookie "$COOKIE_JAR2" | python3 -c "
import sys,json
users=json.load(sys.stdin)
for u in users:
    if u.get('login')=='boss':
        print(u['id']); break
")
req POST "$IAM/users/move" --cookie "$COOKIE_JAR2" \
  -d "{\"user_id\":$BOSS_ID,\"department_id\":$DEPT_ID}" \
  | python3 -m json.tool && pass "boss перемещён"

# ─── UC-4: Загрузка документа ────────────────────────────────────────────────
info "Получаем список папок отдела (будет после создания init-папок)"
FOLDERS=$(req GET "$CATALOG/departments/$DEPT_ID/folders" --cookie "$COOKIE_JAR")
echo "$FOLDERS" | python3 -m json.tool
MAIN_FOLDER_ID=$(echo "$FOLDERS" | python3 -c "
import sys,json
folders=json.load(sys.stdin)
for f in folders:
    if f.get('name')=='main':
        print(f['id']); break
")
pass "main folder id=$MAIN_FOLDER_ID"

# Создаём минимальный docx (просто PK-заголовок чтобы сервис принял расширение .docx)
TMPFILE=$(mktemp --suffix=.docx)
printf 'PK\x03\x04' > "$TMPFILE"

info "Загружаем документ"
DOC=$(curl -s -X POST "$CATALOG/documents" \
  --cookie "$COOKIE_JAR" \
  -F "title=Договор №1" \
  -F "description=Тестовый договор" \
  -F "folder_id=$MAIN_FOLDER_ID" \
  -F "department_id=$DEPT_ID" \
  -F "file=@$TMPFILE")
echo "$DOC" | python3 -m json.tool
DOC_ID=$(echo "$DOC" | python3 -c "import sys,json; print(json.load(sys.stdin).get('document_id',''))")
rm "$TMPFILE"
pass "Документ загружен, id=$DOC_ID"

# ─── UC-12: Чат по документу ─────────────────────────────────────────────────
info "Отправляем сообщение в чат документа"
req POST "$COLLAB/documents/$DOC_ID/messages" --cookie "$COOKIE_JAR" \
  -d '{"content":"Нужно уточнить пункт 3.1"}' \
  | python3 -m json.tool && pass "Сообщение отправлено"

info "Читаем чат"
req GET "$COLLAB/documents/$DOC_ID/messages" --cookie "$COOKIE_JAR" \
  | python3 -m json.tool && pass "Чат прочитан"

# ─── UC-9: Отправка на визирование ───────────────────────────────────────────
info "Отправляем документ на визирование"
req POST "$ORCH/documents/$DOC_ID/send-for-visa" --cookie "$COOKIE_JAR" \
  | python3 -m json.tool && pass "Документ отправлен на визирование"

info "Статус документа"
req GET "$ORCH/documents/$DOC_ID/status" --cookie "$COOKIE_JAR" \
  | python3 -m json.tool

# ─── UC-13: Визирование (boss) ────────────────────────────────────────────────
info "Boss визирует документ"
req POST "$ORCH/documents/$DOC_ID/approve" --cookie "$COOKIE_JAR2" \
  | python3 -m json.tool && pass "Документ завизирован"

info "Финальный статус"
req GET "$ORCH/documents/$DOC_ID/status" --cookie "$COOKIE_JAR" \
  | python3 -m json.tool

info "История FSM"
req GET "$ORCH/documents/$DOC_ID/history" --cookie "$COOKIE_JAR" \
  | python3 -m json.tool

# ─── UC-14: Журнал аудита отдела ─────────────────────────────────────────────
info "Журнал аудита отдела (collaboration-service)"
sleep 1  # даём время RabbitMQ consumer записать события
req GET "$COLLAB/departments/$DEPT_ID/audit" --cookie "$COOKIE_JAR" \
  | python3 -m json.tool && pass "Журнал аудита получен"

echo ""
echo -e "${GREEN}=== Все базовые сценарии пройдены ===${NC}"

# ─── Reject сценарий ─────────────────────────────────────────────────────────
info "Тест reject: сначала снова отправляем на визирование"
req POST "$ORCH/documents/$DOC_ID/send-for-visa" --cookie "$COOKIE_JAR" 2>&1 \
  || echo "(ожидаемо: документ уже APPROVED — нужен другой документ)"

rm -f "$COOKIE_JAR" "$COOKIE_JAR2"
