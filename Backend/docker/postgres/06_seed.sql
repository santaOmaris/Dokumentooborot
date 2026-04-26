
SET search_path TO iam_schema;
INSERT INTO departments (id, name, parent_id) VALUES
  (1, 'Дирекция',            NULL),
  (2, 'Отдел разработки',    1),
  (3, 'Отдел бухгалтерии',   1),
  (4, 'Отдел кадров',        1)
ON CONFLICT DO NOTHING;

SELECT setval('departments_id_seq', (SELECT MAX(id) FROM departments));

-- Пароли (bcrypt cost=10):
--   admin    → admin123
--   head_dev → head123
--   user1    → user123
--   user2    → user123

INSERT INTO users (login, email, password_hash, full_name, department_id, is_head, system_role) VALUES
  (
    'admin',
    'admin@docflow.local',
    '$2a$10$J7tDle97dHd5eZZxdVLGDOvftAo8VE35L8i1EZXqGv7TPAUZ4Rd0a',
    'Администратов Админ Админович',
    1,
    false,
    'ADMIN'
  ),
  (
    'head_dev',
    'head.dev@docflow.local',
    '$2a$10$TqOs1TLWbHHtuOu7CmSK9eJZDTvzxMLfvkHkXJqgH2WQgXt.EMw7O',
    'Начальников Начальник Начальникович',
    2,
    true,
    'USER'
  ),
  (
    'user1',
    'user1@docflow.local',
    '$2a$10$oHYk3AqBeV3VWPgaAqxeKuSviKcGJ42RH/6rg36Jlb4BFFOCpP1U2',
    'Иванов Иван Иванович',
    2,
    false,
    'USER'
  ),
  (
    'user2',
    'user2@docflow.local',
    '$2a$10$oHYk3AqBeV3VWPgaAqxeKuSviKcGJ42RH/6rg36Jlb4BFFOCpP1U2',
    'Петров Пётр Петрович',
    3,
    false,
    'USER'
  )
ON CONFLICT DO NOTHING;

SET search_path TO catalog_schema;

INSERT INTO document_types (id, name) VALUES
  (1, 'Приказ'),
  (2, 'Служебная записка'),
  (3, 'Договор'),
  (4, 'Акт')
ON CONFLICT DO NOTHING;

SELECT setval('document_types_id_seq', (SELECT MAX(id) FROM document_types));

INSERT INTO folders (id, name, department_id, parent_id) VALUES
  (1, 'Входящие',  2, NULL),
  (2, 'Исходящие', 2, NULL),
  (3, 'Архив',     2, NULL)
ON CONFLICT DO NOTHING;

SELECT setval('folders_id_seq', (SELECT MAX(id) FROM folders));
