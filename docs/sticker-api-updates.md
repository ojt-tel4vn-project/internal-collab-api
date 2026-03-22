# Sticker API Updates

## Thay đổi chính

### 1. Send Sticker - Search by Email/Employee Code

**Trước đây:**
```json
POST /api/v1/stickers/send
{
  "receiver_id": "uuid-here",
  "sticker_type_id": "uuid-here",
  "message": "Great job!"
}
```

**Bây giờ:**
```json
POST /api/v1/stickers/send
{
  "receiver_email": "john.doe@company.com",
  "sticker_type_id": "uuid-here",
  "message": "Great job!"
}
```

Hoặc:
```json
POST /api/v1/stickers/send
{
  "receiver_employee_code": "EMP001",
  "sticker_type_id": "uuid-here",
  "message": "Great job!"
}
```

**Lưu ý:**
- Phải cung cấp ít nhất 1 trong 2: `receiver_email` hoặc `receiver_employee_code`
- Nếu cung cấp cả 2, hệ thống sẽ ưu tiên `receiver_email`
- API sẽ tự động tìm employee và validate

### 2. Leaderboard - Bao gồm tất cả employees

**Trước đây:**
- Chỉ trả về employees có nhận sticker
- Thiếu thông tin chi tiết về employee

**Bây giờ:**
```json
GET /api/v1/stickers/leaderboard?limit=20

Response:
{
  "data": [
    {
      "employee_id": "uuid",
      "employee_code": "EMP001",
      "full_name": "John Doe",
      "email": "john.doe@company.com",
      "position": "Senior Developer",
      "avatar_url": "https://...",
      "total": 15,
      "department": "Engineering",
      "department_id": "uuid"
    },
    {
      "employee_id": "uuid",
      "employee_code": "EMP002",
      "full_name": "Jane Smith",
      "email": "jane.smith@company.com",
      "position": "HR Manager",
      "avatar_url": "https://...",
      "total": 0,  // Chưa nhận sticker nào
      "department": "Human Resources",
      "department_id": "uuid"
    }
  ]
}
```

**Cải tiến:**
- Trả về tất cả active employees (kể cả chưa có sticker)
- Thêm thông tin: employee_code, email, position, avatar_url, department_id
- Employees chưa có sticker sẽ có `total: 0`
- Sắp xếp theo total DESC, sau đó theo full_name ASC
- Chỉ hiển thị employees có status = 'active'

## API Endpoints

### Send Sticker
```bash
POST /api/v1/stickers/send
Authorization: Bearer <token>
Content-Type: application/json

{
  "receiver_email": "receiver@company.com",
  "sticker_type_id": "sticker-uuid",
  "message": "Optional message"
}
```

### Get Leaderboard
```bash
GET /api/v1/stickers/leaderboard?limit=20&start_date=2025-01-01&end_date=2025-12-31&department_id=dept-uuid
Authorization: Bearer <token>
```

**Query Parameters:**
- `limit` (optional): Số lượng kết quả (default: 10, max: 100)
- `start_date` (optional): Lọc sticker từ ngày (YYYY-MM-DD)
- `end_date` (optional): Lọc sticker đến ngày (YYYY-MM-DD)
- `department_id` (optional): Lọc theo department

## Testing

### Test Send Sticker by Email
```bash
curl -X POST http://localhost:8080/api/v1/stickers/send \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "receiver_email": "hr@company.com",
    "sticker_type_id": "sticker-uuid",
    "message": "Thank you!"
  }'
```

### Test Send Sticker by Employee Code
```bash
curl -X POST http://localhost:8080/api/v1/stickers/send \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "receiver_employee_code": "EMP002",
    "sticker_type_id": "sticker-uuid",
    "message": "Great work!"
  }'
```

### Test Leaderboard
```bash
curl -X GET "http://localhost:8080/api/v1/stickers/leaderboard?limit=20" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## Migration Notes

### Frontend Changes Required

1. **Send Sticker Form:**
   - Thay đổi từ dropdown/select employee ID
   - Sang search box với email hoặc employee code
   - Có thể implement autocomplete search

2. **Leaderboard Display:**
   - Hiển thị thêm thông tin: employee_code, email, position, avatar
   - Hiển thị cả employees có total = 0
   - Có thể filter theo department

### Database Impact

- Không cần migration database
- Query leaderboard giờ sử dụng LEFT JOIN thay vì JOIN
- Performance tốt hơn với index trên employees.status

## Benefits

1. **UX tốt hơn:** FE không cần load toàn bộ danh sách employees để lấy ID
2. **Search linh hoạt:** Có thể search bằng email hoặc employee code
3. **Leaderboard đầy đủ:** Hiển thị tất cả employees, dễ tracking
4. **Thông tin phong phú:** Có đủ data để hiển thị profile card
5. **Backward compatible:** API vẫn hoạt động với existing data
