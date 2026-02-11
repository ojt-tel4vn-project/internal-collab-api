-- ============================================================================
-- DATABASE SCHEMA: HỆ THỐNG CỘNG TÁC NỘI BỘ (Internal Collaboration System)
-- Database: PostgreSQL 14+
-- Language: Golang (Models aligned)
-- ============================================================================

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================================================
-- 1. CORE TABLES - User Management
-- ============================================================================

-- Departments table
DROP TABLE IF EXISTS departments CASCADE;
CREATE TABLE departments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Roles table (HR, Manager, Employee, Admin)
DROP TABLE IF EXISTS roles CASCADE;
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(50) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Employees table (main user table)
DROP TABLE IF EXISTS employees CASCADE;
CREATE TABLE employees (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_code VARCHAR(20) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    
    -- Personal info
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    -- full_name GENERATED ALWAYS AS (first_name || ' ' || last_name) STORED, -- Leaving out generated column for GORM compatibility (handled in Go or via GORM tag ->)
    
    date_of_birth DATE NOT NULL,
    phone VARCHAR(20),
    address TEXT,
    avatar_url VARCHAR(500),
    
    -- Work info
    department_id UUID REFERENCES departments(id) ON DELETE SET NULL,
    position VARCHAR(100) NOT NULL,
    manager_id UUID REFERENCES employees(id) ON DELETE SET NULL,
    join_date DATE NOT NULL,
    leave_date DATE,
    
    -- System
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'pending')),
    last_login_at TIMESTAMP,
    
    -- Password Reset
    password_reset_token VARCHAR(255),
    password_reset_expires_at TIMESTAMP,

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Indexes
    CONSTRAINT check_join_before_leave CHECK (leave_date IS NULL OR leave_date >= join_date)
);

-- Employee roles (many-to-many)
DROP TABLE IF EXISTS employee_roles CASCADE;
CREATE TABLE employee_roles (
    employee_id UUID REFERENCES employees(id) ON DELETE CASCADE,
    role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
    assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (employee_id, role_id)
);

-- Refresh Tokens table
DROP TABLE IF EXISTS refresh_tokens CASCADE;
CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    token VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMP NOT NULL,
    revoked BOOLEAN DEFAULT false NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- 2. ATTENDANCE MANAGEMENT
-- ============================================================================

-- Attendance records (monthly)
DROP TABLE IF EXISTS attendances CASCADE;
CREATE TABLE attendances (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    month INT NOT NULL CHECK (month BETWEEN 1 AND 12),
    year INT NOT NULL CHECK (year >= 2020),
    
    -- Store daily attendance as JSONB
    -- Format: {"1": "present", "2": "absent", "3": "late", ...}
    attendance_data JSONB NOT NULL DEFAULT '{}',
    
    -- Metadata
    total_days_present INT DEFAULT 0,
    total_days_absent INT DEFAULT 0,
    total_days_late INT DEFAULT 0,
    
    -- Status tracking
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'confirmed', 'auto_confirmed')),
    confirmed_at TIMESTAMP,
    uploaded_by UUID REFERENCES employees(id),
    uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Unique constraint: one record per employee per month
    CONSTRAINT unique_employee_month UNIQUE (employee_id, month, year)
);

-- Attendance comments (for disputes/clarifications)
DROP TABLE IF EXISTS attendance_comments CASCADE;
CREATE TABLE attendance_comments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    attendance_id UUID NOT NULL REFERENCES attendances(id) ON DELETE CASCADE,
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    comment TEXT NOT NULL,
    day_number INT CHECK (day_number BETWEEN 1 AND 31),
    
    -- Status
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'reviewed', 'resolved')),
    reviewed_by UUID REFERENCES employees(id),
    reviewed_at TIMESTAMP,
    hr_response TEXT,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- 3. LEAVE MANAGEMENT
-- ============================================================================

-- Leave types
DROP TABLE IF EXISTS leave_types CASCADE;
CREATE TABLE leave_types (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(50) NOT NULL UNIQUE,
    description TEXT,
    requires_approval BOOLEAN DEFAULT true,
    is_paid BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Annual leave quota
DROP TABLE IF EXISTS leave_quotas CASCADE;
CREATE TABLE leave_quotas (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    year INT NOT NULL CHECK (year >= 2020),
    total_days DECIMAL(4,1) DEFAULT 12.0,
    used_days DECIMAL(4,1) DEFAULT 0.0,
    -- remaining_days DECIMAL(4,1) GENERATED ALWAYS AS (total_days - used_days) STORED, -- Leaving out for GORM compatibility if needed, or keeping if pure SQL logic is preferred. Let's keep it but be aware GORM might not manage it.
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT unique_employee_year UNIQUE (employee_id, year),
    CONSTRAINT check_used_days CHECK (used_days >= 0)
);

-- Leave requests
DROP TABLE IF EXISTS leave_requests CASCADE;
CREATE TABLE leave_requests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    leave_type_id UUID NOT NULL REFERENCES leave_types(id),
    
    -- Date range
    from_date DATE NOT NULL,
    to_date DATE NOT NULL,
    total_days DECIMAL(4,1) NOT NULL,
    
    -- Request details
    reason TEXT NOT NULL,
    contact_during_leave VARCHAR(100),
    
    -- Approval workflow
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected', 'cancelled')),
    submitted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Manager approval
    manager_id UUID REFERENCES employees(id),
    approved_by UUID REFERENCES employees(id),
    approved_at TIMESTAMP,
    rejection_reason TEXT,
    
    -- Metadata
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT check_date_range CHECK (to_date >= from_date),
    CONSTRAINT check_total_days CHECK (total_days > 0)
);

-- ============================================================================
-- 4. REWARD SYSTEM - Points & Stickers
-- ============================================================================

-- Points balance (annual)
DROP TABLE IF EXISTS point_balances CASCADE;
CREATE TABLE point_balances (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    year INT NOT NULL CHECK (year >= 2020),
    
    -- Points tracking
    initial_points INT DEFAULT 0,
    current_points INT DEFAULT 0,
    total_earned INT DEFAULT 0,
    total_spent INT DEFAULT 0,
    
    -- Metadata
    last_reset_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT unique_employee_year_points UNIQUE (employee_id, year),
    CONSTRAINT check_current_points CHECK (current_points >= 0)
);

-- Sticker types (catalog)
DROP TABLE IF EXISTS sticker_types CASCADE;
CREATE TABLE sticker_types (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(50) NOT NULL UNIQUE,
    description TEXT,
    icon_url VARCHAR(500),
    point_cost INT NOT NULL CHECK (point_cost > 0),
    category VARCHAR(50), -- e.g., "appreciation", "achievement", "fun"
    is_active BOOLEAN DEFAULT true,
    display_order INT DEFAULT 0,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Sticker transactions
DROP TABLE IF EXISTS sticker_transactions CASCADE;
CREATE TABLE sticker_transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Sender & Receiver
    sender_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    receiver_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    
    -- Sticker details
    sticker_type_id UUID NOT NULL REFERENCES sticker_types(id),
    point_cost INT NOT NULL,
    message TEXT,
    
    -- Public/Private
    is_public BOOLEAN DEFAULT true,
    
    -- Metadata
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT check_different_users CHECK (sender_id != receiver_id)
);

-- ============================================================================
-- 5. DOCUMENTS & KNOWLEDGE BASE
-- ============================================================================

-- Document categories
DROP TABLE IF EXISTS document_categories CASCADE;
CREATE TABLE document_categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    parent_id UUID REFERENCES document_categories(id) ON DELETE SET NULL,
    display_order INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Documents
DROP TABLE IF EXISTS documents CASCADE;
CREATE TABLE documents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    category_id UUID REFERENCES document_categories(id) ON DELETE SET NULL,
    
    -- File info
    file_name VARCHAR(255) NOT NULL,
    file_path VARCHAR(500) NOT NULL,
    file_size BIGINT, -- in bytes
    mime_type VARCHAR(100),
    
    -- Access control
    is_public BOOLEAN DEFAULT true,
    allowed_roles UUID[], -- array of role IDs
    
    -- Metadata
    uploaded_by UUID NOT NULL REFERENCES employees(id),
    version INT DEFAULT 1,
    download_count INT DEFAULT 0,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Document read tracking
DROP TABLE IF EXISTS document_reads CASCADE;
CREATE TABLE document_reads (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    employee_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    read_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT unique_doc_employee UNIQUE (document_id, employee_id)
);

-- ============================================================================
-- 6. NOTIFICATIONS
-- ============================================================================

-- Notifications
DROP TABLE IF EXISTS notifications CASCADE;
CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    
    -- Notification content
    type VARCHAR(50) NOT NULL, -- 'birthday', 'system', 'task'
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    
    -- Related entity
    entity_type VARCHAR(50), 
    entity_id UUID,
    
    -- Action URL
    action_url VARCHAR(500),
    
    -- Status
    is_read BOOLEAN DEFAULT false,
    read_at TIMESTAMP,
    
    -- Priority
    priority VARCHAR(20) DEFAULT 'normal' CHECK (priority IN ('low', 'normal', 'high', 'urgent')),
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP
);

-- Email queue (for async email sending)
DROP TABLE IF EXISTS email_queue CASCADE;
CREATE TABLE email_queue (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Recipients
    to_email VARCHAR(255) NOT NULL,
    cc_email VARCHAR(255),
    bcc_email VARCHAR(255),
    
    -- Email content
    subject VARCHAR(255) NOT NULL,
    body_html TEXT NOT NULL,
    body_text TEXT,
    
    -- Related entity
    entity_type VARCHAR(50),
    entity_id UUID,
    
    -- Status
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'sent', 'failed', 'cancelled')),
    attempts INT DEFAULT 0,
    max_attempts INT DEFAULT 3,
    last_error TEXT,
    
    -- Timestamps
    scheduled_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    sent_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- 7. AUDIT LOGS
-- ============================================================================

-- Audit logs (for compliance and tracking)
DROP TABLE IF EXISTS audit_logs CASCADE;
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Actor
    employee_id UUID REFERENCES employees(id) ON DELETE SET NULL,
    ip_address INET,
    user_agent TEXT,
    
    -- Action
    action VARCHAR(100) NOT NULL, -- 'create', 'update', 'login'
    entity_type VARCHAR(50) NOT NULL, -- 'employee', 'auth'
    entity_id UUID NOT NULL,
    
    -- Changes (store JSON for flexibility)
    old_values JSONB,
    new_values JSONB,
    
    -- Metadata
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- 8. SYSTEM CONFIGURATION
-- ============================================================================

-- System settings (key-value store)
DROP TABLE IF EXISTS system_settings CASCADE;
CREATE TABLE system_settings (
    key VARCHAR(100) PRIMARY KEY,
    value TEXT NOT NULL,
    value_type VARCHAR(20) DEFAULT 'string' CHECK (value_type IN ('string', 'integer', 'boolean', 'json')),
    description TEXT,
    is_editable BOOLEAN DEFAULT true,
    updated_by UUID REFERENCES employees(id),
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- 9. INDEXES for Performance
-- ============================================================================

-- Employees indexes
CREATE INDEX IF NOT EXISTS idx_employees_email ON employees(email);
CREATE INDEX IF NOT EXISTS idx_employees_status ON employees(status);
CREATE INDEX IF NOT EXISTS idx_employees_department ON employees(department_id);
CREATE INDEX IF NOT EXISTS idx_employees_manager ON employees(manager_id);
CREATE INDEX IF NOT EXISTS idx_employees_dob_month_day ON employees(EXTRACT(MONTH FROM date_of_birth), EXTRACT(DAY FROM date_of_birth));

-- Refresh Token indexes
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token ON refresh_tokens(token);

-- Attendances indexes
CREATE INDEX IF NOT EXISTS idx_attendances_employee ON attendances(employee_id);
CREATE INDEX IF NOT EXISTS idx_attendances_month_year ON attendances(year, month);
CREATE INDEX IF NOT EXISTS idx_attendances_status ON attendances(status);

-- Leave requests indexes
CREATE INDEX IF NOT EXISTS idx_leave_requests_employee ON leave_requests(employee_id);
CREATE INDEX IF NOT EXISTS idx_leave_requests_status ON leave_requests(status);
CREATE INDEX IF NOT EXISTS idx_leave_requests_dates ON leave_requests(from_date, to_date);
CREATE INDEX IF NOT EXISTS idx_leave_requests_manager ON leave_requests(manager_id);

-- Sticker transactions indexes
CREATE INDEX IF NOT EXISTS idx_sticker_sender ON sticker_transactions(sender_id);
CREATE INDEX IF NOT EXISTS idx_sticker_receiver ON sticker_transactions(receiver_id);
CREATE INDEX IF NOT EXISTS idx_sticker_created_at ON sticker_transactions(created_at DESC);

-- Notifications indexes
CREATE INDEX IF NOT EXISTS idx_notifications_employee ON notifications(employee_id);
CREATE INDEX IF NOT EXISTS idx_notifications_is_read ON notifications(is_read);
CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON notifications(created_at DESC);

-- Audit logs indexes
CREATE INDEX IF NOT EXISTS idx_audit_employee ON audit_logs(employee_id);
CREATE INDEX IF NOT EXISTS idx_audit_entity ON audit_logs(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_audit_created_at ON audit_logs(created_at DESC);

-- Email queue indexes
CREATE INDEX IF NOT EXISTS idx_email_queue_status ON email_queue(status);
CREATE INDEX IF NOT EXISTS idx_email_queue_scheduled ON email_queue(scheduled_at) WHERE status = 'pending';

-- ============================================================================
-- 10. TRIGGERS & FUNCTIONS
-- ============================================================================

-- Function: Update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply trigger to all tables with updated_at
DROP TRIGGER IF EXISTS update_employees_updated_at ON employees;
CREATE TRIGGER update_employees_updated_at BEFORE UPDATE ON employees
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_departments_updated_at ON departments;
CREATE TRIGGER update_departments_updated_at BEFORE UPDATE ON departments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_attendances_updated_at ON attendances;
CREATE TRIGGER update_attendances_updated_at BEFORE UPDATE ON attendances
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_attendance_comments_updated_at ON attendance_comments;
CREATE TRIGGER update_attendance_comments_updated_at BEFORE UPDATE ON attendance_comments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_leave_quotas_updated_at ON leave_quotas;
CREATE TRIGGER update_leave_quotas_updated_at BEFORE UPDATE ON leave_quotas
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_leave_requests_updated_at ON leave_requests;
CREATE TRIGGER update_leave_requests_updated_at BEFORE UPDATE ON leave_requests
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_point_balances_updated_at ON point_balances;
CREATE TRIGGER update_point_balances_updated_at BEFORE UPDATE ON point_balances
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_documents_updated_at ON documents;
CREATE TRIGGER update_documents_updated_at BEFORE UPDATE ON documents
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Function: Auto-calculate attendance totals from JSONB
CREATE OR REPLACE FUNCTION calculate_attendance_totals()
RETURNS TRIGGER AS $$
DECLARE
    day_value TEXT;
    present_count INT := 0;
    absent_count INT := 0;
    late_count INT := 0;
BEGIN
    -- Loop through JSONB object
    FOR day_value IN SELECT jsonb_object_keys(NEW.attendance_data)
    LOOP
        CASE NEW.attendance_data->>day_value
            WHEN 'present' THEN present_count := present_count + 1;
            WHEN 'absent' THEN absent_count := absent_count + 1;
            WHEN 'late' THEN late_count := late_count + 1;
        END CASE;
    END LOOP;
    
    NEW.total_days_present := present_count;
    NEW.total_days_absent := absent_count;
    NEW.total_days_late := late_count;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS calculate_attendance_totals_trigger ON attendances;
CREATE TRIGGER calculate_attendance_totals_trigger
    BEFORE INSERT OR UPDATE OF attendance_data ON attendances
    FOR EACH ROW EXECUTE FUNCTION calculate_attendance_totals();

-- Function: Update leave quota when leave is approved
CREATE OR REPLACE FUNCTION update_leave_quota_on_approval()
RETURNS TRIGGER AS $$
BEGIN
    -- Only when status changes to 'approved'
    IF NEW.status = 'approved' AND (OLD.status IS NULL OR OLD.status != 'approved') THEN
        -- Update used_days in leave_quotas
        UPDATE leave_quotas
        SET used_days = used_days + NEW.total_days
        WHERE employee_id = NEW.employee_id
          AND year = EXTRACT(YEAR FROM NEW.from_date);
    END IF;
    
    -- If rejected or cancelled, revert the days
    IF (NEW.status = 'rejected' OR NEW.status = 'cancelled') AND OLD.status = 'approved' THEN
        UPDATE leave_quotas
        SET used_days = used_days - NEW.total_days
        WHERE employee_id = NEW.employee_id
          AND year = EXTRACT(YEAR FROM NEW.from_date);
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS update_leave_quota_trigger ON leave_requests;
CREATE TRIGGER update_leave_quota_trigger
    AFTER INSERT OR UPDATE OF status ON leave_requests
    FOR EACH ROW EXECUTE FUNCTION update_leave_quota_on_approval();

-- Function: Deduct points when sticker is sent
CREATE OR REPLACE FUNCTION deduct_points_on_sticker_send()
RETURNS TRIGGER AS $$
BEGIN
    -- Deduct points from sender
    UPDATE point_balances
    SET current_points = current_points - NEW.point_cost,
        total_spent = total_spent + NEW.point_cost
    WHERE employee_id = NEW.sender_id
      AND year = EXTRACT(YEAR FROM CURRENT_DATE);
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS deduct_points_trigger ON sticker_transactions;
CREATE TRIGGER deduct_points_trigger
    AFTER INSERT ON sticker_transactions
    FOR EACH ROW EXECUTE FUNCTION deduct_points_on_sticker_send();

-- Function: Create audit log automatically
CREATE OR REPLACE FUNCTION create_audit_log()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        INSERT INTO audit_logs (employee_id, action, entity_type, entity_id, old_values)
        VALUES (
            NULLIF(current_setting('app.current_employee_id', TRUE), '')::UUID,
            'delete',
            TG_TABLE_NAME,
            OLD.id,
            to_jsonb(OLD)
        );
        RETURN OLD;
    ELSIF TG_OP = 'UPDATE' THEN
        INSERT INTO audit_logs (employee_id, action, entity_type, entity_id, old_values, new_values)
        VALUES (
            NULLIF(current_setting('app.current_employee_id', TRUE), '')::UUID,
            'update',
            TG_TABLE_NAME,
            NEW.id,
            to_jsonb(OLD),
            to_jsonb(NEW)
        );
        RETURN NEW;
    ELSIF TG_OP = 'INSERT' THEN
        INSERT INTO audit_logs (employee_id, action, entity_type, entity_id, new_values)
        VALUES (
            NULLIF(current_setting('app.current_employee_id', TRUE), '')::UUID,
            'create',
            TG_TABLE_NAME,
            NEW.id,
            to_jsonb(NEW)
        );
        RETURN NEW;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Apply audit trigger to critical tables
DROP TRIGGER IF EXISTS audit_employees ON employees;
CREATE TRIGGER audit_employees AFTER INSERT OR UPDATE OR DELETE ON employees
    FOR EACH ROW EXECUTE FUNCTION create_audit_log();

DROP TRIGGER IF EXISTS audit_leave_requests ON leave_requests;
CREATE TRIGGER audit_leave_requests AFTER INSERT OR UPDATE OR DELETE ON leave_requests
    FOR EACH ROW EXECUTE FUNCTION create_audit_log();

DROP TRIGGER IF EXISTS audit_attendances ON attendances;
CREATE TRIGGER audit_attendances AFTER INSERT OR UPDATE OR DELETE ON attendances
    FOR EACH ROW EXECUTE FUNCTION create_audit_log();

-- ============================================================================
-- 11. INITIAL DATA / SEED DATA
-- ============================================================================

-- Insert default roles
INSERT INTO roles (name, description) VALUES
    ('admin', 'System administrator'),
    ('hr', 'Human Resources'),
    ('manager', 'Team manager'),
    ('employee', 'Regular employee')
ON CONFLICT (name) DO NOTHING;

-- Insert default leave types
INSERT INTO leave_types (name, description, requires_approval, is_paid) VALUES
    ('Annual Leave', 'Paid annual vacation', true, true),
    ('Sick Leave', 'Medical leave', true, true),
    ('Personal Leave', 'Personal matters', true, false),
    ('Maternity Leave', 'Maternity leave', true, true),
    ('Paternity Leave', 'Paternity leave', true, true),
    ('Unpaid Leave', 'Leave without pay', true, false)
ON CONFLICT (name) DO NOTHING;

-- Insert default document categories
INSERT INTO document_categories (name, description, display_order) VALUES
    ('Onboarding', 'New employee orientation materials', 1),
    ('Policies', 'Company policies and regulations', 2),
    ('Procedures', 'Standard operating procedures', 3),
    ('Forms', 'Downloadable forms', 4),
    ('Training', 'Training materials', 5)
ON CONFLICT (name) DO NOTHING;

-- Insert default sticker types
INSERT INTO sticker_types (name, description, point_cost, category, display_order) VALUES
    ('Thumbs Up', 'Great work!', 10, 'appreciation', 1),
    ('Star', 'Outstanding performance', 20, 'achievement', 2),
    ('Heart', 'Thank you!', 10, 'appreciation', 3),
    ('Trophy', 'Champion!', 30, 'achievement', 4),
    ('Fire', 'On fire!', 15, 'achievement', 5),
    ('Clap', 'Well done!', 10, 'appreciation', 6)
ON CONFLICT (name) DO NOTHING;

-- Insert default system settings
INSERT INTO system_settings (key, value, value_type, description, is_editable) VALUES
    ('annual_points_per_employee', '100', 'integer', 'Points granted to each employee annually', true),
    ('points_reset_date', '01-01', 'string', 'Annual points reset date (MM-DD)', true),
    ('attendance_confirm_deadline_days', '7', 'integer', 'Days to confirm attendance after upload', true),
    ('leave_quota_annual_days', '12', 'integer', 'Default annual leave quota', true),
    ('birthday_notification_enabled', 'true', 'boolean', 'Enable birthday notifications', true),
    ('email_notification_enabled', 'true', 'boolean', 'Enable email notifications', true),
    ('company_name', 'Your Company Name', 'string', 'Company name for branding', true),
    ('support_email', 'hr@company.com', 'string', 'HR support email', true)
ON CONFLICT (key) DO NOTHING;

-- ============================================================================
-- 13. ADDITIONAL SEED DATA (For Testing)
-- ============================================================================

-- Insert Departments
INSERT INTO departments (name, description) VALUES
    ('Engineering', 'Create and maintain software products'),
    ('Human Resources', 'Manage employee relations and benefits'),
    ('Sales', 'Sell products and services')
ON CONFLICT (name) DO NOTHING;

-- Insert Test Employees & Related Data
DO $$
DECLARE
    dept_eng UUID;
    dept_hr UUID;
    dept_sales UUID;
    
    role_admin UUID;
    role_hr UUID;
    role_manager UUID;
    role_employee UUID;
    
    emp_admin UUID;
    emp_hr UUID;
    emp_manager UUID;
    emp_staff UUID;
    
    current_year INT := EXTRACT(YEAR FROM CURRENT_DATE);
    -- Hash for password '123456' generated using bcrypt
    dummy_hash VARCHAR := '$2a$10$4ptAlmlSkklAtLgq4sArP.RmBaFGhG61CUIeWlXrWdS9gImic/uIO'; 
BEGIN
    -- Get Department IDs
    SELECT id INTO dept_eng FROM departments WHERE name = 'Engineering';
    SELECT id INTO dept_hr FROM departments WHERE name = 'Human Resources';
    SELECT id INTO dept_sales FROM departments WHERE name = 'Sales';
    
    -- Get Role IDs
    SELECT id INTO role_admin FROM roles WHERE name = 'admin';
    SELECT id INTO role_hr FROM roles WHERE name = 'hr';
    SELECT id INTO role_manager FROM roles WHERE name = 'manager';
    SELECT id INTO role_employee FROM roles WHERE name = 'employee';

    -- Create Admin User (admin@company.com / 123456)
    INSERT INTO employees (
        employee_code, email, password_hash, first_name, last_name, 
        date_of_birth, phone, address, department_id, position, join_date, status
    ) VALUES (
        'EMP001', 'admin@company.com', dummy_hash,
        'Admin', 'User', '1990-01-01', '0123456789', '123 Tech Street', 
        dept_eng, 'System Administrator', '2023-01-01', 'active'
    ) ON CONFLICT (email) DO UPDATE SET email=EXCLUDED.email RETURNING id INTO emp_admin;
    
    -- Create HR User (hr@company.com / 123456)
    INSERT INTO employees (
        employee_code, email, password_hash, first_name, last_name, 
        date_of_birth, phone, address, department_id, position, join_date, status
    ) VALUES (
        'EMP002', 'hr@company.com', dummy_hash,
        'HR', 'Manager', '1992-05-15', '0987654321', '456 People Ave', 
        dept_hr, 'HR Manager', '2023-02-01', 'active'
    ) ON CONFLICT (email) DO UPDATE SET email=EXCLUDED.email RETURNING id INTO emp_hr;
    
    -- Create Manager User (manager@company.com / 123456)
    INSERT INTO employees (
        employee_code, email, password_hash, first_name, last_name, 
        date_of_birth, phone, address, department_id, position, join_date, status
    ) VALUES (
        'EMP003', 'manager@company.com', dummy_hash,
        'Sales', 'Lead', '1988-11-20', '0112233445', '789 Market Blvd', 
        dept_sales, 'Sales Manager', '2023-03-01', 'active'
    ) ON CONFLICT (email) DO UPDATE SET email=EXCLUDED.email RETURNING id INTO emp_manager;
    
    -- Create Regular Employee (staff@company.com / 123456)
    INSERT INTO employees (
        employee_code, email, password_hash, first_name, last_name, 
        date_of_birth, phone, address, department_id, position, manager_id, join_date, status
    ) VALUES (
        'EMP004', 'staff@company.com', dummy_hash,
        'John', 'Developer', '1995-08-10', '0556677889', '321 Code Lane', 
        dept_eng, 'Software Engineer', emp_manager, '2023-06-01', 'active'
    ) ON CONFLICT (email) DO UPDATE SET email=EXCLUDED.email RETURNING id INTO emp_staff;

    -- Assign Roles (avoiding duplicates)
    INSERT INTO employee_roles (employee_id, role_id) VALUES
        (emp_admin, role_admin),
        (emp_admin, role_employee),
        (emp_hr, role_hr),
        (emp_hr, role_employee),
        (emp_manager, role_manager),
        (emp_manager, role_employee),
        (emp_staff, role_employee)
    ON CONFLICT DO NOTHING;

    -- Initialize Leave Quotas for Current Year
    INSERT INTO leave_quotas (employee_id, year, total_days) VALUES
        (emp_admin, current_year, 12),
        (emp_hr, current_year, 12),
        (emp_manager, current_year, 12),
        (emp_staff, current_year, 12)
    ON CONFLICT DO NOTHING;

    -- Initialize Point Balances for Current Year
    INSERT INTO point_balances (employee_id, year, initial_points, current_points) VALUES
        (emp_admin, current_year, 100, 100),
        (emp_hr, current_year, 100, 100),
        (emp_manager, current_year, 100, 100),
        (emp_staff, current_year, 100, 100)
    ON CONFLICT DO NOTHING;

    -- [NEW] Insert Sample Notifications (for Staff)
    INSERT INTO notifications (
        employee_id, type, title, message, priority, is_read, created_at
    ) VALUES 
    (
        emp_staff, 'system', 'Welcome!', 'Welcome to the new collaboration system. Please complete your profile.', 'normal', false, NOW() - INTERVAL '2 days'
    ),
    (
        emp_staff, 'birthday', 'Happy Birthday!', 'Wishing you a fantastic day!', 'high', false, NOW() - INTERVAL '5 minutes'
    );

    -- [NEW] Insert Sample Leave Requests (Staff requests, Manager approves/pending)
    -- [NEW] Insert Sample Leave Requests (Staff requests, Manager approves/pending)
    -- Fetch IDs into variables declared at the top block if needed, or just select directly
    
    -- Request 1: Pending (Annual Leave)
    INSERT INTO leave_requests (
        employee_id, leave_type_id, from_date, to_date, total_days, reason, status, manager_id
    ) 
    SELECT emp_staff, id, CURRENT_DATE + INTERVAL '10 days', CURRENT_DATE + INTERVAL '12 days', 3, 'Family trip', 'pending', emp_manager
    FROM leave_types WHERE name = 'Annual Leave';

    -- Request 2: Approved (Sick Leave)
    INSERT INTO leave_requests (
        employee_id, leave_type_id, from_date, to_date, total_days, reason, status, manager_id, approved_by, approved_at
    ) 
    SELECT emp_staff, id, CURRENT_DATE - INTERVAL '1 month', CURRENT_DATE - INTERVAL '1 month' + INTERVAL '1 day', 2, 'Flu', 'approved', emp_manager, emp_manager, NOW() - INTERVAL '2 weeks'
    FROM leave_types WHERE name = 'Sick Leave';

    -- [NEW] Insert Sample Attendance (Last Month)
    INSERT INTO attendances (
        employee_id, month, year, attendance_data, status, total_days_present
    ) VALUES (
        emp_staff, EXTRACT(MONTH FROM CURRENT_DATE - INTERVAL '1 month')::INT, EXTRACT(YEAR FROM CURRENT_DATE - INTERVAL '1 month')::INT, 
        jsonb_build_object(
            '1', 'present', '2', 'present', '3', 'present', '4', 'present', '5', 'present',
            '8', 'present', '9', 'present', '10', 'late', '11', 'present', '12', 'present'
        ), 
        'confirmed', 10
    ) ON CONFLICT DO NOTHING;


END $$;