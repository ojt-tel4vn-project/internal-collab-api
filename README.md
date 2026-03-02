# Internal Collaboration API

RESTful API cho hệ thống quản lý nhân viên nội bộ với authentication và authorization.

## 🚀 Features

### ✅ Đã hoàn thành

- **HR-Driven Employee Management**
  - HR tạo account cho nhân viên mới
  - Tự động generate temporary password
  - Email thông báo với Brevo (SendinBlue) ✅

- **Authentication & Authorization**
  - JWT-based authentication
  - Role-based access control (RBAC)
  - First-time password setup
  - Change password

- **Employee Management**
  - Create, Read, Update, Delete employees
  - Role assignment
  - Profile management

- **Email Service** 📧
  - Brevo (SendinBlue) integration
  - Welcome email with temporary password
  - Beautiful HTML templates
  - Password reset & change notifications

## 🏗️ Architecture

### **Database Schema**
- **Single Table Design**: `employees` table chứa cả auth và business data
- **RBAC**: `roles` và `employee_roles` tables
- **Status Lifecycle**: `pending` → `active` → `inactive`

### **Authentication Flow**
```
1. HR creates employee → status='pending', temp password
2. Employee receives email with temp password
3. Employee first-time setup → status='active'
4. Employee can login normally
```

## 📋 API Endpoints

### **Public (No Auth)**
```
POST /api/v1/auth/login
POST /api/v1/auth/first-time-setup?email={email}
```

### **Protected (Require JWT)**
```
POST /api/v1/auth/change-password
```

### **HR Only (Require JWT + HR Role)**
```
POST   /api/v1/hr/employees
GET    /api/v1/hr/employees
GET    /api/v1/hr/employees/{id}
PUT    /api/v1/hr/employees/{id}
DELETE /api/v1/hr/employees/{id}
```

## 🛠️ Tech Stack

- **Framework**: Go 1.21+
- **API**: Huma v2 (OpenAPI 3.1)
- **Router**: Gin v1
- **Database**: PostgreSQL
- **ORM**: GORM
- **Auth**: JWT (golang-jwt/jwt)
- **Password**: bcrypt
- **Email**: Brevo (SendinBlue)
- **Hot Reload**: Air

## 📦 Project Structure

```
internal-collab-api/
├── cmd/
│   └── main.go              # Application entry point
├── dtos/
│   ├── auth/                # Auth DTOs
│   └── employee/            # Employee DTOs
├── handlers/
│   ├── auth_handler.go      # Auth endpoints
│   └── employee_handler.go  # Employee endpoints
├── services/
│   ├── auth_service.go      # Auth business logic
│   └── employee_service.go  # Employee business logic
├── repository/
│   └── employee_repository.go
├── models/
│   ├── employee.go
│   └── role.go
├── middleware/
│   ├── jwt_auth.go          # JWT helper functions
│   └── role_auth.go         # Role helper functions
├── pkg/
│   ├── auth/                # Auth utilities
│   ├── crypto/              # JWT & password
│   ├── email/               # Brevo email service
│   ├── logger/              # Zap logger
│   └── response/            # Response helpers
├── internal/
│   ├── config/              # Configuration
│   ├── database/            # DB connection
│   └── validators/          # Validators
├── routes/
│   └── routes.go            # Route setup
└── docs/
    └── MIDDLEWARE_FIXED.md  # Documentation
```

## 🚀 Getting Started

### **Prerequisites**
- Go 1.21+
- PostgreSQL 14+
- Air (for hot reload)

### **Installation**

1. Clone repository
```bash
git clone <repo-url>
cd internal-collab-api
```

2. Install dependencies
```bash
go mod download
```

3. Setup environment
```bash
cp .env.example .env
# Edit .env with your configuration
```

4. Run database migrations
```bash
# Create database first
createdb internal_collab

# Run migrations
go run cmd/main.go
```

5. Install Air (optional, for hot reload)
```bash
go install github.com/air-verse/air@latest
```

### **Run Development Server**

With Air (hot reload):
```bash
air
```

Without Air:
```bash
go run cmd/main.go
```

Server will start on `http://localhost:8080`

### **API Documentation**
Visit `http://localhost:8080/docs` for interactive API documentation (Swagger UI)

## 🧪 Testing

### **Test Login**
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@company.com",
    "password": "password123"
  }'
```

### **Test Protected Endpoint**
```bash
TOKEN="your_access_token"

curl -X POST http://localhost:8080/api/v1/auth/change-password \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "old_password": "password123",
    "new_password": "newPassword456"
  }'
```

### **Test HR Endpoint**
```bash
curl -X POST http://localhost:8080/api/v1/hr/employees \
  -H "Authorization: Bearer $HR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "newemployee@company.com",
    "first_name": "John",
    "last_name": "Doe",
    "date_of_birth": "1995-05-15",
    "position": "Developer"
  }'
```

## 🔐 Security

- **Password Hashing**: bcrypt with cost 10
- **JWT Expiration**: 24 hours
- **Role-Based Access**: HR/Admin roles for employee management
- **Status Validation**: Pending users cannot access protected resources

## 📝 Environment Variables

```env
# Server
SERVER_PORT=8080

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=internal_collab
DB_SSLMODE=disable

# JWT
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production

# Email (Brevo/SendinBlue)
BREVO_API_KEY=your-brevo-api-key-here
EMAIL_FROM=noreply@company.com
EMAIL_FROM_NAME=Internal Collaboration System
```

## 📚 Documentation

- `docs/BREVO_EMAIL_SETUP.md` - Email service setup guide
- API Docs available at `/docs` endpoint

## 🗺️ Roadmap

### **Phase 1: Core Features** ✅
- [x] HR-driven employee creation
- [x] JWT authentication
- [x] Role-based authorization
- [x] Password management
- [x] Email service (Brevo integration)

### **Phase 2: Enhancements** (TODO)
- [x] Token refresh endpoint
- [x] Audit logging
- [x] Rate limiting
- [ ] Profile picture upload
- [x] Email templates for more events

### **Phase 3: Advanced Features** (TODO)
- [ ] Multi-factor authentication (MFA)
- [ ] SSO integration
- [ ] Advanced reporting
- [ ] Notification system

## 🤝 Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to branch (`git push origin feature/AmazingFeature`)
5. Open Pull Request

## 📄 License

This project is licensed under the MIT License.

## 👥 Authors

- Your Name - Initial work

## 🙏 Acknowledgments

- Huma v2 for excellent OpenAPI support
- GORM for powerful ORM
- Gin for fast and resilient web framework
