
SET search_path TO iam_schema;
INSERT INTO departments (id, name, parent_id) VALUES
  (1, 'Дирекция',            NULL),
  (2, 'Отдел разработки',    1),
  (3, 'Отдел бухгалтерии',   1),
  (4, 'Отдел кадров',        1)
ON CONFLICT DO NOTHING;

SELECT setval('departments_id_seq', (SELECT MAX(id) FROM departments));

-- Пароли (bcrypt cost=10):
--   admin123 -> $2a$10$J7tDle97dHd5eZZxdVLGDOvftAo8VE35L8i1EZXqGv7TPAUZ4Rd0a
--   head123  -> $2a$10$TqOs1TLWbHHtuOu7CmSK9eJZDTvzxMLfvkHkXJqgH2WQgXt.EMw7O
--   user123  -> $2a$10$oHYk3AqBeV3VWPgaAqxeKuSviKcGJ42RH/6rg36Jlb4BFFOCpP1U2

UPDATE users
SET department_id = NULL
WHERE login = 'admin';

INSERT INTO users (login, email, password_hash, full_name, department_id, is_head, system_role) VALUES
  ('admin',      'admin@docflow.local',      '$2a$10$J7tDle97dHd5eZZxdVLGDOvftAo8VE35L8i1EZXqGv7TPAUZ4Rd0a', 'Администратов Админ Админович', NULL, false, 'ADMIN'),

  ('head_dir_1', 'head.dir1@docflow.local',  '$2a$10$TqOs1TLWbHHtuOu7CmSK9eJZDTvzxMLfvkHkXJqgH2WQgXt.EMw7O', 'Директоров Директор Первый',      1, true,  'USER'),
  ('head_dir_2', 'head.dir2@docflow.local',  '$2a$10$TqOs1TLWbHHtuOu7CmSK9eJZDTvzxMLfvkHkXJqgH2WQgXt.EMw7O', 'Директоров Директор Второй',      1, true,  'USER'),
  ('dir_user_1', 'dir.user1@docflow.local',  '$2a$10$oHYk3AqBeV3VWPgaAqxeKuSviKcGJ42RH/6rg36Jlb4BFFOCpP1U2', 'Дирекционов Денис Денисович',      1, false, 'USER'),
  ('dir_user_2', 'dir.user2@docflow.local',  '$2a$10$oHYk3AqBeV3VWPgaAqxeKuSviKcGJ42RH/6rg36Jlb4BFFOCpP1U2', 'Дирекционова Дарья Дмитриевна',    1, false, 'USER'),

  ('head_dev',   'head.dev@docflow.local',   '$2a$10$TqOs1TLWbHHtuOu7CmSK9eJZDTvzxMLfvkHkXJqgH2WQgXt.EMw7O', 'Начальников Начальник Начальникович', 2, true,  'USER'),
  ('head_dev_2', 'head.dev2@docflow.local',  '$2a$10$TqOs1TLWbHHtuOu7CmSK9eJZDTvzxMLfvkHkXJqgH2WQgXt.EMw7O', 'Разработчиков Роман Романович',    2, true,  'USER'),
  ('user1',      'user1@docflow.local',      '$2a$10$oHYk3AqBeV3VWPgaAqxeKuSviKcGJ42RH/6rg36Jlb4BFFOCpP1U2', 'Иванов Иван Иванович',             2, false, 'USER'),
  ('dev_user_2', 'dev.user2@docflow.local',  '$2a$10$oHYk3AqBeV3VWPgaAqxeKuSviKcGJ42RH/6rg36Jlb4BFFOCpP1U2', 'Петрова Полина Петровна',          2, false, 'USER'),

  ('head_acc_1', 'head.acc1@docflow.local',  '$2a$10$TqOs1TLWbHHtuOu7CmSK9eJZDTvzxMLfvkHkXJqgH2WQgXt.EMw7O', 'Бухгалтеров Борис Борисович',      3, true,  'USER'),
  ('head_acc_2', 'head.acc2@docflow.local',  '$2a$10$TqOs1TLWbHHtuOu7CmSK9eJZDTvzxMLfvkHkXJqgH2WQgXt.EMw7O', 'Бухгалтерова Белла Борисовна',     3, true,  'USER'),
  ('user2',      'user2@docflow.local',      '$2a$10$oHYk3AqBeV3VWPgaAqxeKuSviKcGJ42RH/6rg36Jlb4BFFOCpP1U2', 'Петров Пётр Петрович',             3, false, 'USER'),
  ('acc_user_2', 'acc.user2@docflow.local',  '$2a$10$oHYk3AqBeV3VWPgaAqxeKuSviKcGJ42RH/6rg36Jlb4BFFOCpP1U2', 'Смирнова Светлана Сергеевна',      3, false, 'USER'),

  ('head_hr_1',  'head.hr1@docflow.local',   '$2a$10$TqOs1TLWbHHtuOu7CmSK9eJZDTvzxMLfvkHkXJqgH2WQgXt.EMw7O', 'Кадрова Кира Кирилловна',          4, true,  'USER'),
  ('head_hr_2',  'head.hr2@docflow.local',   '$2a$10$TqOs1TLWbHHtuOu7CmSK9eJZDTvzxMLfvkHkXJqgH2WQgXt.EMw7O', 'Кадров Андрей Андреевич',          4, true,  'USER'),
  ('hr_user_1',  'hr.user1@docflow.local',   '$2a$10$oHYk3AqBeV3VWPgaAqxeKuSviKcGJ42RH/6rg36Jlb4BFFOCpP1U2', 'Соколова София Сергеевна',         4, false, 'USER'),
  ('hr_user_2',  'hr.user2@docflow.local',   '$2a$10$oHYk3AqBeV3VWPgaAqxeKuSviKcGJ42RH/6rg36Jlb4BFFOCpP1U2', 'Орлов Олег Олегович',              4, false, 'USER')
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
