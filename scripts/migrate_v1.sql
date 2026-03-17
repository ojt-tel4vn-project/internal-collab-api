-- 1. CẬP NHẬT TRẠNG THÁI NHÂN VIÊN (inactive -> offboard)
-- Xử lý dữ liệu cũ trước (Tránh lỗi violation)
UPDATE employees SET status = 'offboard' WHERE status = 'inactive';

-- Xoá bỏ check constraint cũ và thêm check constraint mới
ALTER TABLE employees DROP CONSTRAINT IF EXISTS employees_status_check;
ALTER TABLE employees ADD CONSTRAINT employees_status_check CHECK (status IN ('active', 'offboard', 'pending'));


-- 2. CẬP NHẬT BẢNG TÀI LIỆU (documents)
-- Thêm cột roles mới
ALTER TABLE documents ADD COLUMN IF NOT EXISTS roles VARCHAR(100) DEFAULT 'employee';

-- (Tuỳ chọn) Chuyển đổi trạng thái is_public cũ qua roles
-- Giả sử lúc trước is_public = true thì cho tất cả mọi người xem ('employee,manager,hr')
UPDATE documents SET roles = 'employee,manager,hr' WHERE is_public = true;

-- Xoá cột is_public và các cột dư thừa cũ
ALTER TABLE documents DROP COLUMN IF EXISTS is_public;
ALTER TABLE documents DROP COLUMN IF EXISTS version;
ALTER TABLE documents DROP COLUMN IF EXISTS download_count;


-- 3. CẬP NHẬT BẢNG COMMENT (Nếu bạn đã có comment cũ từ Document)
-- Cột document_id ở DB cũ cần xoá và thay bằng attendance_id
ALTER TABLE comments DROP COLUMN IF EXISTS document_id;

-- Thêm thuộc tính attendance_id và is_read
ALTER TABLE comments ADD COLUMN IF NOT EXISTS attendance_id UUID REFERENCES attendances(id) ON DELETE CASCADE;
ALTER TABLE comments ADD COLUMN IF NOT EXISTS is_read BOOLEAN DEFAULT false;
