# Test Attendance CSV Import

## Thay đổi chính

### 1. Model Update (models/attendance.go)
- Thêm struct `DayAttendanceDetail` để lưu thông tin chi tiết:
  - `Status`: trạng thái (present/absent/late/leave)
  - `CheckInTime`: giờ vào (HH:MM:SS)
  - `CheckOutTime`: giờ ra (HH:MM:SS)
  - `WorkHours`: tổng giờ làm việc

- `AttendanceData` giờ là `map[string]DayAttendanceDetail` thay vì `map[string]DayStatus`

### 2. Service Update (services/attendance_service.go)
- Update `UploadAttendance()` để parse format CSV mới:
  - Đọc từng dòng với format: ID Người, Tên, Bộ phận, Ngày, Thời gian biểu, Tình trạng chuyên cần, Vào, Ra
  - Skip header và các dòng chi tiết (bắt đầu với "Thời gian vào:")
  - Map trạng thái tiếng Việt sang English:
    - "bình thường" → present
    - "Muộn" → late
    - "Vắng mặt" → absent
    - "Về sớm" → present
  - Tính toán work_hours từ check-in và check-out time
  - Group theo employee và tạo/update attendance record

### 3. Database Migration (scripts/migrate_attendance_detail.sql)
- Update trigger `calculate_attendance_totals()` để support cả format cũ và mới
- Thêm comment cho column documentation
- Script migration optional để convert data cũ sang format mới

## Format CSV mới

```csv
Chi Tiết Chấm Công,,,,,,,
ID Người,Tên,Bộ phận,Ngày,Thời gian biểu,Tình trạng chuyên cần,Vào,Ra
22,Lê Văn A,P. KY THUAT,2025-07-28,Ca Chung(08:00:00-17:30:00),bình thường,08:19:35,19:26:14
Thời gian vào: 2025-07-28 08:19:35,Thời gian ra: 2025-07-28 19:26:14,Thời lượng theo dõi:11 Giờ00 phút,,,,,
```

## Cấu trúc JSONB mới trong database

```json
{
  "1": {
    "status": "present",
    "check_in_time": "08:19:35",
    "check_out_time": "19:26:14",
    "work_hours": 11.0
  },
  "2": {
    "status": "late",
    "check_in_time": "08:46:10",
    "check_out_time": "18:13:13",
    "work_hours": 9.5
  },
  "3": {
    "status": "absent",
    "check_in_time": "-",
    "check_out_time": "-",
    "work_hours": 0
  }
}
```

## Các bước test

### 1. Chạy migration
```bash
psql -U your_user -d your_database -f scripts/migrate_attendance_detail.sql
```

### 2. Test API upload
```bash
# Đọc file CSV
$csvContent = Get-Content "docs/Chấm Công (Mẫu).csv" -Raw

# Upload attendance
curl -X POST http://localhost:8080/api/v1/attendances/upload \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "month": 7,
    "year": 2025,
    "csv_content": "'"$csvContent"'"
  }'
```

### 3. Verify data
```sql
-- Check attendance data structure
SELECT 
    e.employee_code,
    e.first_name,
    a.month,
    a.year,
    a.attendance_data,
    a.total_days_present,
    a.total_days_absent,
    a.total_days_late
FROM attendances a
JOIN employees e ON a.employee_id = e.id
WHERE a.month = 7 AND a.year = 2025;

-- Check specific day detail
SELECT 
    e.employee_code,
    a.attendance_data->'28' as day_28_detail
FROM attendances a
JOIN employees e ON a.employee_id = e.id
WHERE a.month = 7 AND a.year = 2025;
```

## Lưu ý

1. **Employee Code**: File CSV sử dụng ID Người (employee_code) để map với employees table
2. **Date Format**: Ngày trong CSV phải có format YYYY-MM-DD
3. **Status Mapping**: 
   - "bình thường" và "Về sớm" đều map sang "present"
   - "Muộn" → "late"
   - "Vắng mặt" → "absent"
4. **Work Hours**: Tự động tính từ check-in và check-out time
5. **Skip Rows**: Các dòng chi tiết (bắt đầu với "Thời gian vào:") sẽ bị skip

## Backward Compatibility

Code vẫn support format cũ (simple string status) nhờ trigger function đã được update để handle cả 2 format.
