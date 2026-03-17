-- =============================================================================
-- assign_role.sql  —  Assign role cho một employee hiện có
-- Chạy trong: Supabase SQL Editor  hoặc  psql
-- =============================================================================

-- ---------------------------------------------------------------------------
-- 1. Xem danh sách roles hiện có
-- ---------------------------------------------------------------------------
SELECT id, name, description FROM roles ORDER BY name;

-- ---------------------------------------------------------------------------
-- 2. Xem danh sách employees (kèm role hiện tại)
-- ---------------------------------------------------------------------------
SELECT
    e.employee_code,
    e.email,
    e.full_name,
    r.name AS current_role,
    e.status
FROM employees e
LEFT JOIN roles r ON r.id = e.role_id
ORDER BY e.employee_code;

-- ---------------------------------------------------------------------------
-- 3. Assign role cho employee  ← SỬA 2 dòng này rồi chạy
-- ---------------------------------------------------------------------------

-- Cách A: Dùng email + tên role (dễ nhất)
UPDATE employees
SET role_id = (SELECT id FROM roles WHERE name = 'admin')   -- ← đổi tên role: admin | manager | hr | employee
WHERE email = 'your@email.com';                             -- ← đổi email employee

-- ---------------------------------------------------------------------------
-- Cách B: Nếu muốn dùng UUID trực tiếp
-- ---------------------------------------------------------------------------
-- UPDATE employees
-- SET role_id = '<role-uuid-here>'
-- WHERE id = '<employee-uuid-here>';

-- ---------------------------------------------------------------------------
-- 4. Kiểm tra lại sau khi update
-- ---------------------------------------------------------------------------
SELECT
    e.employee_code,
    e.email,
    e.full_name,
    r.name AS role,
    e.status
FROM employees e
LEFT JOIN roles r ON r.id = e.role_id
WHERE e.email = 'your@email.com';   -- ← đổi email để verify
