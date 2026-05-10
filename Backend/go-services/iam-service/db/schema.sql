CREATE SCHEMA IF NOT EXISTS iam_schema;
SET search_path TO iam_schema;


--ТАБЛИЦА ОТДЕЛОВ UC-15: Иерархия отделов
CREATE TABLE departments (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    -- parent_id: Самоссылающийся внешний ключ.
    -- Если NULL -> это корень (например, Совет Директоров или CEO).
    -- Если указан ID -> этот отдел подчиняется другому отделу.
    -- ON DELETE SET NULL защищает от каскадного удаления всей ветки.
    parent_id INT REFERENCES departments(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Индекс для быстрого поиска дочерних отделов
CREATE INDEX idx_departments_parent_id ON departments(parent_id);

-- 2. ТАБЛИЦА ПОЛЬЗОВАТЕЛЕЙ UC-3, UC-5, UC-20, UC-21, UC-22

CREATE TABLE users (
    id SERIAL PRIMARY KEY,

    -- UC5 Авторизация: Логин и хэш пароля (bcrypt)
    login VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,

    -- UC3 Справочник: ФИО сотрудника
    full_name VARCHAR(255) NOT NULL,

    -- UC-20 Состав отдела: Привязка к отделу.
    -- Может быть NULL, если сотрудник еще не распределен (на испытательном сроке)
    department_id INT REFERENCES departments(id) ON DELETE SET NULL,

    -- UC-21 (Изменение должности): Бизнес-роль
    -- true = Начальник отдела (имеет право визировать документы своего отдела)
    is_head BOOLEAN NOT NULL DEFAULT FALSE,

    -- UC-22 (Создание юзера): Системная роль (RBAC)
    -- Возможные значения: 'USER', 'ADMIN'.
    -- Определяет глобальные права (настройки типов, скрытие любых документов)
    system_role VARCHAR(50) NOT NULL DEFAULT 'USER',

    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Индексы для ускорения поиска (sqlc будет генерировать эффективный код с их учетом)
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_department_id ON users(department_id);
