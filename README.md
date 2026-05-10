# Описание проекта

Cистема электронного документооборота (СЭД) для внутренней работы отделов. 
Проект покрывает полный жизненный цикл документа:
- создание/загрузка шаблонов документов,
- создание/загрузка документов,
- маршрутизация между сотрудниками и отделами,
- визирование документа,
- обсуждение документа,
- аудит всех действий и сбор метрик,
- уведомления (по email).
### Предварительные требования

На машине должны быть:
1. Docker Desktop (или Docker Engine + Compose plugin).
2. Node.js 20+ и npm (для локальной разработки Frontend).
3. Порты должны быть свободны:
- 80 (Caddy),
- 8081, 8082, 8083, 8084 (Go HTTP),
- 9081, 9082, 9091 (gRPC),
- 5432 (Postgres),
- 5672, 15672 (RabbitMQ),
- 9000, 9001 (MinIO).
# Архитектура

## Описание архитектуры

| Технология                                                                                            | Задача     |
| ----------------------------------------------------------------------------------------------------- | ---------- |
| Бизнес-логика                                                                                         | Golang     |
| Инфраструктурные задачи (транспортировка файлов и отправка уведомлений)                               | Java       |
| один PostgreSQL-кластер, но с разделением по схемам (`database-per-service через schema-per-service`) | PostgreSQL |
| Файловое хранилище                                                                                    | MinIO (S3) |
| Синхронное взаимодействие между сервисамих                                                            | HTTP/gRPC  |
| Асинхронное взаимодействие между сервисами (Orcestrator)                                              | RabbitMQ   |
| Маршрутизация запросов                                                                                | Caddy      |
| Контейнеризация и развертывание                                                                       | Docker     |
## Архитектурные принципы

1. Разделение ответственности по сервисам.
2. Изоляция данных по схемам БД.
3. Явный API-контракт между сервисами.
4. Безопасность на двух уровнях:
	аутентификация через JWT,
	авторизация через роль/признак начальника/отдел.
5. Дублирование критичных ограничений на backend-уровне (не только в UI).
6. Наблюдаемость через аудит событий и структурированные логи.

## Состав системы 

Frontend: [[React]] + [[TypeScript]] + Vite + Ant Design.
API gateway/reverse proxy: [[Caddy]].
[[Go]]-сервисы:
  - [[IAM Service]] (Авторизация, управление пользователями и отделами, выдача данных другим сервисам по gRPC (руководители, email),
  - [[Catalog Service]] (CRUD операции с  документами и шаблонами),
  - [[Orchestrator Service ]](Хранение истории взаимодействия с файлами и чатов, координация действий между IAM и Catalog),
  - [[Collaboration Service]] (Сложные операции (визирование, делегирование), запись данных в RabbitMQ).
Java-сервисы [[STSR]]:
  - [[ File Service ]](Прокси к MinIO),
  -  [[Notification Service ]](Cлушает определенные события RebbitMQ).
  Инфраструктура:
  - [[PostgreSQL]],
  - [[RabbitMQ]],
  - MinIO (S3).
### Где проверяются права

Запросы проверяются в middleware который находится в папке pkg, к которой имеют доступ все go-сервисы в проекте. 
Middleware имеет 2 функции:
 - Generate JWT. Вызывается при авторизации пользователя, как обычная функция, без сетевой составляющей 
 - Aurhuser. 
##  Маршруты запросов
### Синхронные потоки:

Browser -> Caddy -> Go HTTP API.
[[Orchestrator]] -> IAM/Catalog по gRPC.
Catalog -> File Service по gRPC.
### Асинхронные потоки:

Orchestrator публикует события в RabbitMQ exchanges:
- audit,
- notifications.
Collaboration читает audit и пишет журнал действий.
- Notification Service читает notifications и отправляет e-mail (или пишет в лог в stub-режиме).
### Caddy

- /api/iam/* -> iam-service:8081
- /api/catalog/* -> catalog-service:8082
- /api/orchestrator/* -> orchestrator-service:8084
- /api/collaboration/* -> collaboration-service:8083
- остальное -> frontend:4173
## Базы данных

### Схемы PostgreSQL

Используются схемы:
- iam_schema,
- catalog_schema,
- collaboration_schema,
- orchestrator_schema.

Каждый сервис подключается своим DB-пользователем с нужным `search_path.`
### Архитектура PostgreSQL

```
Cхема БД (например, iam_schema) - это как отдельная БД
	Таблица (например, users)
		Поля (например для users: ID, ФИО, Email, пароль)
```

iam_schema:
  - departments,
  - users.
catalog_schema:
  - document_types,
  - folders,
  - documents,
  - document_history.
collaboration_schema:
  - messages,
  - audit_logs.
orchestrator_schema:
  - document_states,
  - state_transitions.
### Cвязи
 
Cервис владеет своей схемой,
Согласованность достигаются через API-вызовы между сервисами.
## USER CASE

| UC    | Название                                         | Где реализовано              | Как работает сейчас                                                                                                  | Статус      |
| ----- | ------------------------------------------------ | ---------------------------- | -------------------------------------------------------------------------------------------------------------------- | ----------- |
| UC-1  | Загрузить шаблон документа из папки              | Catalog + UI папок           | Отдельного реестра шаблонов нет; шаблоны ведутся как документы в системной папке templates                           | Частично    |
| UC-2  | Изменить список шаблонов                         | Catalog folders/documents    | Управляется через операции с папками/документами и системными папками                                                | Реализовано |
| UC-3  | Просмотреть список авторизованных сотрудников    | IAM                          | GET /api/iam/users и GET /api/iam/departments/{id}/users                                                             | Реализовано |
| UC-4  | [[Загрузить документ в систему]]                 | Catalog + File Service       | POST /api/catalog/documents -> gRPC UploadFile -> MinIO -> metadata в БД                                             | Реализовано |
| UC-5  | [[Авторизироваться]]                             | IAM                          | POST /api/iam/auth/login, cookie jwt_token, middleware проверка в каждом сервисе                                     | Реализовано |
| UC-6  | Просмотреть список загруженных документов отдела | Catalog                      | Документы по папкам отдела + UI страницы документов                                                                  | Реализовано |
| UC-7  | Скрыть документ                                  | Catalog                      | Head/Admin могут скрывать; unhide только Admin                                                                       | Реализовано |
| UC-8  | [[Делегировать документ]]                        | Orchestrator                 | POST /delegate, доступ только ADMIN/head, смена assignee                                                             | Реализовано |
| UC-9  | [[Прикрепить на визирование начальнику]]         | Orchestrator + IAM + Catalog | POST /send-for-visa, подбор руководителя отдела, смена статуса/assignee                                              | Частично    |
| UC-10 | Отправить документ другому отделу                | Orchestrator                 | Сотруднику: /delegate; в отдел: /request-approval с target_department_id                                             | Частично    |
| UC-11 | Найти документ из архива отдела                  | Catalog                      | GET /api/catalog/departments/{dept_id}/search?q=...                                                                  | Реализовано |
| UC-12 | Обсудить документ                                | Collaboration                | GET/POST messages по документу                                                                                       | Реализовано |
| UC-13 | Завизировать документ                            | Orchestrator                 | POST /approve и POST /reject, FSM и аудит                                                                            | Частично    |
| UC-14 | Просмотреть журнал действий отдела               | Collaboration                | GET /departments/{dept_id}/audit и /documents/{id}/audit                                                             | Реализовано |
| UC-15 | Изменить статус отдела в иерархии                | IAM                          | PATCH /api/iam/departments/{id}/parent                                                                               | Реализовано |
| UC-16 | Получить уведомление о событии                   | Orchestrator + Notification  | Publish в exchange notifications, consumer отправляет e-mail/лог                                                     | Частично    |
| UC-17 | Выгрузить документ для локального редактирования | Catalog + File Service       | GET /api/catalog/documents/{id}/download                                                                             | Реализовано |
| UC-18 | Выгрузить внутренние метрики системы             | Нет endpoint                 | В текущем backend нет выделенного API/задачи экспорта метрик                                                         | Реализовано |
| UC-19 | Настроить тип документа                          | Catalog                      | CRUD по /api/catalog/types                                                                                           | Реализовано |
| UC-20 | Управления организационной структурой            | IAM                          | Move/Fire пользователя, переводы между отделами                                                                      | Реализовано |
| UC-21 | Просмотр данных аудита                           | Orchestrator                 | Запрос в RebbitMQ, если текущий флаг admin - просмотр общего аудита, если текущий флаг head - просмотр аудита отдела | Реализовано |
