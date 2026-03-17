-- =============================================================================
-- reset_and_seed.sql  —  Chạy toàn bộ file này trong Supabase SQL Editor
-- Tạo lại toàn bộ bảng (khớp với GORM models) + seed data mẫu
-- Mật khẩu mẫu : 123456  (bcrypt hash bên dưới)
-- =============================================================================

-- ---------------------------------------------------------------------------
-- 0. Extensions
-- ---------------------------------------------------------------------------
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ---------------------------------------------------------------------------
-- 1. DROP tất cả bảng (theo thứ tự FK ngược lại)
-- ---------------------------------------------------------------------------
DROP TABLE IF EXISTS leave_requests       CASCADE;
DROP TABLE IF EXISTS leave_quotas         CASCADE;
DROP TABLE IF EXISTS leave_types          CASCADE;
DROP TABLE IF EXISTS notifications        CASCADE;
DROP TABLE IF EXISTS audit_logs           CASCADE;
DROP TABLE IF EXISTS refresh_tokens       CASCADE;
DROP TABLE IF EXISTS comments             CASCADE;
DROP TABLE IF EXISTS employees            CASCADE;
DROP TABLE IF EXISTS roles                CASCADE;
DROP TABLE IF EXISTS departments          CASCADE;
DROP TABLE IF EXISTS document_reads       CASCADE;
DROP TABLE IF EXISTS documents            CASCADE;
DROP TABLE IF EXISTS document_categories  CASCADE;

-- ---------------------------------------------------------------------------
-- 2. CREATE TABLES  (khớp với GORM models)
-- ---------------------------------------------------------------------------

-- departments
CREATE TABLE departments (
    id          UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    name        VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    created_at  TIMESTAMP   DEFAULT NOW(),
    updated_at  TIMESTAMP   DEFAULT NOW()
);

-- roles
CREATE TABLE roles (
    id          UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    name        VARCHAR(50) NOT NULL UNIQUE,
    description TEXT,
    created_at  TIMESTAMP   DEFAULT NOW()
);

-- employees
CREATE TABLE employees (
    id                          UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_code               VARCHAR(20)  NOT NULL UNIQUE,
    email                       VARCHAR(255) NOT NULL UNIQUE,
    password_hash               VARCHAR(255) NOT NULL,
    first_name                  VARCHAR(100) NOT NULL,
    last_name                   VARCHAR(100) NOT NULL,
    full_name                   VARCHAR(255) GENERATED ALWAYS AS (first_name || ' ' || last_name) STORED,
    date_of_birth               DATE         NOT NULL,
    phone                       VARCHAR(20),
    address                     TEXT,
    avatar_url                  VARCHAR(500),
    department_id               UUID         REFERENCES departments(id) ON DELETE SET NULL,
    position                    VARCHAR(100) NOT NULL,
    manager_id                  UUID         REFERENCES employees(id)  ON DELETE SET NULL,
    join_date                   DATE         NOT NULL,
    leave_date                  DATE,
    status                      VARCHAR(20)  DEFAULT 'active' CHECK (status IN ('active','offboard','pending')),
    last_login_at               TIMESTAMP,
    password_reset_token        VARCHAR(255),
    password_reset_expires_at   TIMESTAMP,
    -- Single role FK (replaces many-to-many employee_roles)
    role_id                     UUID         REFERENCES roles(id) ON DELETE SET NULL,
    created_at                  TIMESTAMP    DEFAULT NOW(),
    updated_at                  TIMESTAMP    DEFAULT NOW()
);

-- refresh_tokens
CREATE TABLE refresh_tokens (
    id         UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id    UUID         NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    token      VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMP    NOT NULL,
    revoked    BOOLEAN      NOT NULL DEFAULT false,
    created_at TIMESTAMP    DEFAULT NOW()
);

-- audit_logs
CREATE TABLE audit_logs (
    id          UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id UUID         REFERENCES employees(id) ON DELETE SET NULL,
    ip_address  VARCHAR(45),
    user_agent  TEXT,
    action      VARCHAR(100) NOT NULL,
    entity_type VARCHAR(50)  NOT NULL,
    entity_id   UUID         NOT NULL,
    old_values  JSONB,
    new_values  JSONB,
    description TEXT,
    created_at  TIMESTAMP    DEFAULT NOW()
);

-- notifications
CREATE TABLE notifications (
    id          UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id UUID         NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    type        VARCHAR(50)  NOT NULL,
    title       VARCHAR(255) NOT NULL,
    message     TEXT         NOT NULL,
    entity_type VARCHAR(50),
    entity_id   UUID,
    action_url  VARCHAR(500),
    is_read     BOOLEAN      DEFAULT false,
    read_at     TIMESTAMP,
    priority    VARCHAR(20)  DEFAULT 'normal',
    created_at  TIMESTAMP    DEFAULT NOW(),
    expires_at  TIMESTAMP
);

-- document_categories
CREATE TABLE document_categories (
    id           UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    name         VARCHAR(100) NOT NULL UNIQUE,
    description  TEXT,
    parent_id    UUID         REFERENCES document_categories(id) ON DELETE SET NULL,
    display_order INT         DEFAULT 0,
    created_at   TIMESTAMP    DEFAULT NOW()
);

-- documents
CREATE TABLE documents (
    id             UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    title          VARCHAR(255) NOT NULL,
    description    TEXT,
    category_id    UUID         REFERENCES document_categories(id) ON DELETE SET NULL,
    file_name      VARCHAR(255) NOT NULL,
    file_path      VARCHAR(500) NOT NULL,
    file_size      BIGINT,
    mime_type      VARCHAR(100),
    roles          VARCHAR(100) DEFAULT 'employee',
    uploaded_by    UUID         NOT NULL REFERENCES employees(id),
    created_at     TIMESTAMP    DEFAULT NOW(),
    updated_at     TIMESTAMP    DEFAULT NOW()
);

-- document_reads
CREATE TABLE document_reads (
    id          UUID  PRIMARY KEY DEFAULT uuid_generate_v4(),
    document_id UUID  NOT NULL REFERENCES documents(id)  ON DELETE CASCADE,
    employee_id UUID  NOT NULL REFERENCES employees(id)  ON DELETE CASCADE,
    read_at     TIMESTAMP DEFAULT NOW(),
    CONSTRAINT unique_doc_employee UNIQUE (document_id, employee_id)
);

-- leave_types  (khớp với models.LeaveType — KHÔNG có uniqueIndex trên name)
CREATE TABLE leave_types (
    id          UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    name        VARCHAR(100) NOT NULL,
    description TEXT,
    created_at  TIMESTAMP    DEFAULT NOW(),
    updated_at  TIMESTAMP    DEFAULT NOW()
);

-- leave_quotas  (khớp với models.LeaveQuota — có composite uniqueIndex idx_emp_type_year)
CREATE TABLE leave_quotas (
    id              UUID    PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id     UUID    NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    leave_type_id   UUID    NOT NULL REFERENCES leave_types(id),
    year            INT     NOT NULL,
    total_days      DECIMAL(5,1) NOT NULL,
    used_days       DECIMAL(5,1) NOT NULL DEFAULT 0,
    created_at      TIMESTAMP    DEFAULT NOW(),
    updated_at      TIMESTAMP    DEFAULT NOW(),
    CONSTRAINT idx_emp_type_year UNIQUE (employee_id, leave_type_id, year)
);

-- leave_requests  (khớp với models.LeaveRequest)
CREATE TABLE leave_requests (
    id                    UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id           UUID         NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    leave_type_id         UUID         NOT NULL REFERENCES leave_types(id),
    from_date             DATE         NOT NULL,
    to_date               DATE         NOT NULL,
    total_days            DECIMAL(5,1) NOT NULL,
    reason                TEXT         NOT NULL,
    contact_during_leave  VARCHAR(255),
    status                VARCHAR(20)  DEFAULT 'pending' CHECK (status IN ('pending','approved','rejected','canceled')),
    approver_id           UUID         REFERENCES employees(id),
    approver_comment      TEXT,
    -- action_token: uniqueIndex managed by GORM, do NOT add inline UNIQUE here.
    -- Instead create the index with GORM's naming convention below.
    action_token          VARCHAR(255),
    submitted_at          TIMESTAMP    DEFAULT NOW(),
    updated_at            TIMESTAMP    DEFAULT NOW()
);
-- GORM naming convention: uni_<table>_<column>
CREATE UNIQUE INDEX "uni_leave_requests_action_token" ON leave_requests(action_token)
    WHERE action_token IS NOT NULL;

-- comments (employee opinions/disputes about their attendance records)
CREATE TABLE comments (
    id            UUID      PRIMARY KEY DEFAULT uuid_generate_v4(),
    attendance_id UUID      NOT NULL REFERENCES attendances(id) ON DELETE CASCADE,
    author_id     UUID      NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    content       TEXT      NOT NULL,
    is_read       BOOLEAN   NOT NULL DEFAULT false,
    parent_id     UUID      REFERENCES comments(id) ON DELETE SET NULL,
    created_at    TIMESTAMP DEFAULT NOW(),
    updated_at    TIMESTAMP DEFAULT NOW()
);
CREATE INDEX idx_comments_attendance_id ON comments(attendance_id);

-- ---------------------------------------------------------------------------
-- 3. SEED: Departments
-- ---------------------------------------------------------------------------
INSERT INTO departments (name, description) VALUES
    ('Engineering',       'Software development and infrastructure'),
    ('Product',           'Product management and design'),
    ('Human Resources',   'People operations and recruitment'),
    ('Finance',           'Accounting, budgeting and financial planning'),
    ('Marketing',         'Brand, growth and communications'),
    ('Customer Success',  'Client support and onboarding');

-- ---------------------------------------------------------------------------
-- 4. SEED: Roles (4 canonical roles)
-- ---------------------------------------------------------------------------
INSERT INTO roles (name, description) VALUES
    ('admin',    'Full system access'),
    ('manager',  'Team management access'),
    ('hr',       'Human Resources operations'),
    ('employee', 'Standard employee access');

-- ---------------------------------------------------------------------------
-- 5. SEED: Employees + Roles + Leave data  (dùng DO $$ để dùng biến)
-- ---------------------------------------------------------------------------
DO $$
DECLARE
    dept_eng  UUID;
    dept_hr   UUID;
    dept_prod UUID;

    role_admin    UUID;
    role_manager  UUID;
    role_hr       UUID;
    role_employee UUID;

    emp_admin   UUID;
    emp_engmgr  UUID;
    emp_dev1    UUID;
    emp_dev2    UUID;
    emp_pm      UUID;

    lt_annual UUID;
    lt_sick   UUID;
    lt_comp   UUID;

    -- bcrypt hash của "123456"
    pw_hash VARCHAR := '$2a$10$4ptAlmlSkklAtLgq4sArP.RmBaFGhG61CUIeWlXrWdS9gImic/uIO';
    cur_year INT := EXTRACT(YEAR FROM CURRENT_DATE)::INT;
BEGIN

    -- Lấy ID departments
    SELECT id INTO dept_eng  FROM departments WHERE name = 'Engineering';
    SELECT id INTO dept_hr   FROM departments WHERE name = 'Human Resources';
    SELECT id INTO dept_prod FROM departments WHERE name = 'Product';

    -- Lấy ID roles
    SELECT id INTO role_admin    FROM roles WHERE name = 'admin';
    SELECT id INTO role_manager  FROM roles WHERE name = 'manager';
    SELECT id INTO role_hr       FROM roles WHERE name = 'hr';
    SELECT id INTO role_employee FROM roles WHERE name = 'employee';

    -- ---- Employees ----

    -- EMP001 – HR Admin
    INSERT INTO employees (employee_code, email, password_hash, first_name, last_name,
        date_of_birth, phone, address, department_id, position, join_date, status, role_id)
    VALUES ('EMP001', 'admin@company.com', pw_hash,
        'An', 'Nguyen', '1990-05-15', '0901234567', '123 Le Loi, Q1, HCMC',
        dept_hr, 'HR Manager', '2020-01-15', 'active', role_admin)
    RETURNING id INTO emp_admin;

    -- EMP002 – Engineering Manager
    INSERT INTO employees (employee_code, email, password_hash, first_name, last_name,
        date_of_birth, phone, address, department_id, position, join_date, status, role_id)
    VALUES ('EMP002', 'eng.manager@company.com', pw_hash,
        'Binh', 'Tran', '1988-11-22', '0912345678', '456 Nguyen Hue, Q1, HCMC',
        dept_eng, 'Engineering Manager', '2019-03-01', 'active', role_manager)
    RETURNING id INTO emp_engmgr;

    -- EMP003 – Software Engineer
    INSERT INTO employees (employee_code, email, password_hash, first_name, last_name,
        date_of_birth, phone, address, department_id, position, manager_id, join_date, status, role_id)
    VALUES ('EMP003', 'dev1@company.com', pw_hash,
        'Cuong', 'Le', '1995-07-30', '0923456789', '789 Tran Hung Dao, Q5, HCMC',
        dept_eng, 'Software Engineer', emp_engmgr, '2021-06-01', 'active', role_employee)
    RETURNING id INTO emp_dev1;

    -- EMP004 – Junior Software Engineer
    INSERT INTO employees (employee_code, email, password_hash, first_name, last_name,
        date_of_birth, phone, address, department_id, position, manager_id, join_date, status, role_id)
    VALUES ('EMP004', 'dev2@company.com', pw_hash,
        'Dung', 'Pham', '1997-02-14', '0934567890', '22 Vo Van Tan, Q3, HCMC',
        dept_eng, 'Junior Software Engineer', emp_engmgr, '2023-02-01', 'active', role_employee)
    RETURNING id INTO emp_dev2;

    -- EMP005 – Product Manager
    INSERT INTO employees (employee_code, email, password_hash, first_name, last_name,
        date_of_birth, phone, address, department_id, position, join_date, status, role_id)
    VALUES ('EMP005', 'pm@company.com', pw_hash,
        'Em', 'Hoang', '1992-09-08', '0945678901', '10 Dien Bien Phu, Binh Thanh, HCMC',
        dept_prod, 'Product Manager', '2020-09-15', 'active', role_manager)
    RETURNING id INTO emp_pm;
    -- (No employee_roles inserts needed — single role assigned above)

    -- ---- Leave Types ----
    INSERT INTO leave_types (name, description) VALUES
        ('Annual Leave',    'Paid vacation days allocated per year'),
        ('Sick Leave',      'Leave for illness or medical appointments'),
        ('Maternity Leave', 'Leave for employees expecting a newborn'),
        ('Paternity Leave', 'Leave for fathers following birth of a child'),
        ('Unpaid Leave',    'Leave without pay'),
        ('Compassionate',   'Bereavement or family emergency leave');

    SELECT id INTO lt_annual FROM leave_types WHERE name = 'Annual Leave';
    SELECT id INTO lt_sick   FROM leave_types WHERE name = 'Sick Leave';
    SELECT id INTO lt_comp   FROM leave_types WHERE name = 'Compassionate';

    -- ---- Leave Quotas (năm hiện tại) ----
    -- Annual Leave: 12 ngày
    INSERT INTO leave_quotas (employee_id, leave_type_id, year, total_days, used_days)
    SELECT e.id, lt_annual, cur_year, 12, 0
    FROM (VALUES (emp_admin),(emp_engmgr),(emp_dev1),(emp_dev2),(emp_pm)) AS e(id);

    -- Sick Leave: 10 ngày
    INSERT INTO leave_quotas (employee_id, leave_type_id, year, total_days, used_days)
    SELECT e.id, lt_sick, cur_year, 10, 0
    FROM (VALUES (emp_admin),(emp_engmgr),(emp_dev1),(emp_dev2),(emp_pm)) AS e(id);

    -- Compassionate: 3 ngày
    INSERT INTO leave_quotas (employee_id, leave_type_id, year, total_days, used_days)
    SELECT e.id, lt_comp, cur_year, 3, 0
    FROM (VALUES (emp_admin),(emp_engmgr),(emp_dev1),(emp_dev2),(emp_pm)) AS e(id);

    -- ---- Leave Requests ----
    -- EMP003: approved annual leave (3 ngày)
    INSERT INTO leave_requests (employee_id, leave_type_id, from_date, to_date, total_days,
        reason, contact_during_leave, status, approver_id, approver_comment)
    VALUES (emp_dev1, lt_annual,
        CURRENT_DATE + 10, CURRENT_DATE + 12, 3,
        'Family vacation to Da Nang', '0923456789',
        'approved', emp_engmgr, 'Approved. Enjoy your vacation!');

    -- Cập nhật used_days cho EMP003 annual quota
    UPDATE leave_quotas SET used_days = 3
    WHERE employee_id = emp_dev1 AND leave_type_id = lt_annual AND year = cur_year;

    -- EMP004: pending sick leave (2 ngày)
    INSERT INTO leave_requests (employee_id, leave_type_id, from_date, to_date, total_days,
        reason, contact_during_leave, status)
    VALUES (emp_dev2, lt_sick,
        CURRENT_DATE + 3, CURRENT_DATE + 4, 2,
        'Fever and cold', '0934567890', 'pending');

    -- EMP005: pending annual leave (5 ngày)
    INSERT INTO leave_requests (employee_id, leave_type_id, from_date, to_date, total_days,
        reason, contact_during_leave, status)
    VALUES (emp_pm, lt_annual,
        CURRENT_DATE + 20, CURRENT_DATE + 24, 5,
        'Conference trip to Hanoi', '0945678901', 'pending');

    -- ---- Document Categories ----
    INSERT INTO document_categories (name, description, display_order) VALUES
        ('Onboarding', 'New employee orientation materials', 1),
        ('Policies',   'Company policies and regulations',   2),
        ('Procedures', 'Standard operating procedures',      3),
        ('Forms',      'Downloadable forms',                 4),
        ('Training',   'Training materials',                 5)
    ON CONFLICT DO NOTHING;

    -- ---- Notifications mẫu ----
    INSERT INTO notifications (employee_id, type, title, message, priority)
    VALUES
        (emp_dev1, 'system',  'Welcome!', 'Welcome to the internal collaboration system. Please complete your profile.', 'normal'),
        (emp_dev2, 'system',  'Welcome!', 'Welcome to the internal collaboration system. Please complete your profile.', 'normal'),
        (emp_admin, 'system', 'New leave request pending', 'EMP004 has submitted a sick leave request for review.', 'high');

END $$;

-- ---------------------------------------------------------------------------
-- Kiểm tra dữ liệu:
-- ---------------------------------------------------------------------------
-- SELECT employee_code, email, first_name || ' ' || last_name AS name, position FROM employees ORDER BY employee_code;
-- SELECT lt.name, lq.total_days, lq.used_days, e.employee_code FROM leave_quotas lq JOIN leave_types lt ON lt.id = lq.leave_type_id JOIN employees e ON e.id = lq.employee_id ORDER BY e.employee_code, lt.name;
-- SELECT lr.status, lt.name AS leave_type, lr.from_date, lr.to_date, e.employee_code FROM leave_requests lr JOIN leave_types lt ON lt.id = lr.leave_type_id JOIN employees e ON e.id = lr.employee_id;
