# API Documentation - Internal Collaboration System

## 📚 Table of Contents

- [Overview](#overview)
- [Base URL](#base-url)
- [Authentication](#authentication)
- [Request/Response Format](#requestresponse-format)
- [Error Handling](#error-handling)
- [Rate Limiting](#rate-limiting)
- [API Endpoints](#api-endpoints)
- [Code Examples](#code-examples)
- [Postman Collection](#postman-collection)

---

## Overview

RESTful API for Internal Collaboration System built with Golang and PostgreSQL.

**Version:** 1.0.0  
**API Spec:** OpenAPI 3.0.3

### Features
Phúc
- 🔐 JWT Authentication (Login, Refresh Token, Password Reset)
- 👥 Employee Management (Onboard, Offboard, Profile, Subordinates)
- 🔔 Notifications (In-app, Email, Preferences)
- 👤 Admin & RBAC (User Management, Roles, Permissions)
- 📝 Audit Logs (Action Tracking, Export)
- ⚙️ Background Jobs (Birthday, Points Reset, Auto-confirm)

Trung
- 📅 Attendance Tracking (Upload, Confirm, Dispute, Auto-confirm)
- 🏖️ Leave Management (Request, Approve via System/Email, Quota)
- 🎁 Reward System (Points, Stickers, Leaderboard)
- 📚 Document Management (Categories, Upload, Read Tracking)


---

## Base URL

```
Local:      http://localhost:8080/api/v1
Staging:    https://api-staging.yourcompany.com/api/v1
Production: https://api.yourcompany.com/api/v1
```

---

## Authentication

### Login Flow

1. **Login** to get JWT token:

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john.doe@company.com",
    "password": "password123"
  }'
```

Response:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2024-11-01T10:00:00Z",
  "user": {
    "id": 1,
    "email": "john.doe@company.com",
    "full_name": "John Doe",
    "roles": ["employee"]
  }
}
```

2. **Use token** in subsequent requests:

```bash
curl -X GET http://localhost:8080/api/v1/employees/me \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

### Token Refresh

When token expires, use refresh token to get new access token:

```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "YOUR_REFRESH_TOKEN"
  }'
```

---

## Request/Response Format

### Request Headers

```
Content-Type: application/json
Authorization: Bearer <token>
Accept: application/json
```

### Success Response

```json
{
  "data": {
    // Resource data
  },
  "meta": {
    // Pagination metadata (for list endpoints)
    "page": 1,
    "limit": 20,
    "total": 100,
    "total_pages": 5
  }
}
```

### Error Response

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable error message",
    "details": {
      // Additional error context
    }
  }
}
```

---

## Error Handling

### HTTP Status Codes

| Code | Description |
|------|-------------|
| 200 | Success |
| 201 | Created |
| 400 | Bad Request - Invalid input |
| 401 | Unauthorized - Invalid or missing token |
| 403 | Forbidden - Insufficient permissions |
| 404 | Not Found |
| 422 | Unprocessable Entity - Validation failed |
| 429 | Too Many Requests - Rate limit exceeded |
| 500 | Internal Server Error |

### Error Codes

| Code | Description |
|------|-------------|
| `VALIDATION_ERROR` | Request validation failed |
| `AUTHENTICATION_FAILED` | Invalid credentials |
| `TOKEN_EXPIRED` | JWT token expired |
| `PERMISSION_DENIED` | User doesn't have required permissions |
| `RESOURCE_NOT_FOUND` | Requested resource doesn't exist |
| `DUPLICATE_ENTRY` | Resource already exists |
| `INSUFFICIENT_POINTS` | Not enough points to send sticker |
| `INVALID_STATUS_TRANSITION` | Invalid status change |

### Example Error Response

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "details": {
      "email": "Invalid email format",
      "date_of_birth": "Must be at least 18 years old"
    }
  }
}
```

---

## Rate Limiting

### Limits

- **Authentication endpoints:** 5 requests per minute per IP
- **General API:** 100 requests per minute per user
- **File uploads:** 10 requests per hour per user

### Response Headers

```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1635724800
```

### Rate Limit Exceeded Response

```json
{
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Too many requests. Please try again later.",
    "details": {
      "retry_after": 60
    }
  }
}
```

---

## API Endpoints

### 1. Authentication

#### Login
```http
POST /auth/login
```

Request:
```json
{
  "email": "john.doe@company.com",
  "password": "password123"
}
```

#### Change Password
```http
POST /auth/change-password
Authorization: Bearer <token>
```

Request:
```json
{
  "current_password": "old_password",
  "new_password": "new_password"
}
```

#### Refresh Token
```http
POST /auth/refresh
```

Request:
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

Response:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2024-11-01T11:00:00Z"
}
```

#### Forgot Password
```http
POST /auth/forgot-password
```

Request:
```json
{
  "email": "john.doe@company.com"
}
```

Response:
```json
{
  "message": "Password reset email sent if account exists"
}
```

#### Reset Password
```http
POST /auth/reset-password
```

Request:
```json
{
  "token": "reset_token_from_email",
  "new_password": "new_password123"
}
```

---

### 2. Employees

#### List Employees
```http
GET /employees?page=1&limit=20&status=active&department_id=1&search=john
Authorization: Bearer <token>
```

#### Create Employee (Onboard)
```http
POST /employees
Authorization: Bearer <token>
```

Request:
```json
{
  "email": "jane.smith@company.com",
  "employee_code": "EMP002",
  "first_name": "Jane",
  "last_name": "Smith",
  "date_of_birth": "1992-03-20",
  "phone": "+84901234567",
  "position": "Frontend Developer",
  "department_id": 1,
  "manager_id": 5,
  "join_date": "2024-01-15",
  "roles": ["employee"]
}
```

#### Get Employee Details
```http
GET /employees/{id}
Authorization: Bearer <token>
```

#### Get Current User Profile
```http
GET /employees/me
Authorization: Bearer <token>
```

#### Update Own Profile
```http
PUT /employees/me
Authorization: Bearer <token>
```

Request:
```json
{
  "phone": "+84909876543",
  "address": "123 Main St, HCMC",
  "avatar_url": "https://cdn.company.com/avatars/john-new.jpg"
}
```

#### Update Employee (HR only)
```http
PUT /employees/{id}
Authorization: Bearer <token>
```

Request:
```json
{
  "position": "Senior Developer",
  "department_id": 2,
  "manager_id": 10,
  "status": "active"
}
```

#### Get Manager's Subordinates
```http
GET /employees/subordinates
Authorization: Bearer <token>
```

Response:
```json
{
  "data": [
    {
      "id": 5,
      "full_name": "Jane Smith",
      "position": "Frontend Developer",
      "email": "jane.smith@company.com"
    }
  ]
}
```

#### Offboard Employee
```http
DELETE /employees/{id}
Authorization: Bearer <token>
```

#### Get Birthdays
```http
GET /employees/birthdays?period=today
Authorization: Bearer <token>
```

#### Get Birthday Notifications Config
```http
GET /employees/birthdays/config
Authorization: Bearer <token>
```

#### Update Birthday Notifications Config (HR only)
```http
PUT /employees/birthdays/config
Authorization: Bearer <token>
```

Request:
```json
{
  "enabled": true,
  "notification_time": "09:00",
  "channels": ["in_app", "email"]
}
```

---

### 3. Attendance

#### List Attendances
```http
GET /attendances?year=2024&month=10&status=pending
Authorization: Bearer <token>
```

#### Upload Attendance (HR only)
```http
POST /attendances
Authorization: Bearer <token>
Content-Type: multipart/form-data
```

Form Data:
```
month: 10
year: 2024
file: <excel/csv file>
```

#### Get Attendance Details
```http
GET /attendances/{id}
Authorization: Bearer <token>
```

Response:
```json
{
  "data": {
    "id": 1,
    "employee": {
      "id": 1,
      "full_name": "John Doe"
    },
    "month": 10,
    "year": 2024,
    "attendance_data": {
      "1": "present",
      "2": "present",
      "3": "late",
      "4": "absent",
      "5": "leave"
    },
    "total_days_present": 20,
    "total_days_absent": 1,
    "total_days_late": 2,
    "status": "pending"
  }
}
```

#### Confirm Attendance
```http
POST /attendances/{id}/confirm
Authorization: Bearer <token>
```

Request:
```json
{
  "status": "confirmed",  // status: "confirmed", "disputed", "auto_confirmed"
  "comment": "I was working from home on day 4"
}
```

#### Add Comment to Attendance
```http
POST /attendances/{id}/comments
Authorization: Bearer <token>
```

Request:
```json
{
  "comment": "I forgot to check in on day 15, but I was at the office",
  "day_number": 15
}
```

#### Review Comment (HR only)
```http
POST /attendances/comments/{comment_id}/review
Authorization: Bearer <token>
```

Request:
```json
{
  "hr_response": "We've updated your record for that day",
  "status": "resolved"
}
```

#### Get Attendance Config (HR only)
```http
GET /attendances/config
Authorization: Bearer <token>
```

Response:
```json
{
  "data": {
    "confirmation_deadline_days": 5,
    "auto_confirm_enabled": true,
    "reminder_before_deadline_days": 2
  }
}
```

#### Update Attendance Config (HR only)
```http
PUT /attendances/config
Authorization: Bearer <token>
```

Request:
```json
{
  "confirmation_deadline_days": 7,
  "auto_confirm_enabled": true,
  "reminder_before_deadline_days": 3
}
```

#### Get Attendance Summary (HR only)
```http
GET /attendances/summary?year=2024&month=10
Authorization: Bearer <token>
```

Response:
```json
{
  "data": {
    "total_employees": 50,
    "confirmed": 40,
    "pending": 8,
    "disputed": 2,
    "auto_confirmed": 5
  }
}
```

---

### 4. Leave Management

#### List Leave Types
```http
GET /leave-types
Authorization: Bearer <token>
```

#### Get Leave Quotas
```http
GET /leave-quotas?year=2024
Authorization: Bearer <token>
```

Response:
```json
{
  "data": [
    {
      "id": 1,
      "employee_id": 1,
      "year": 2024,
      "total_days": 12.0,
      "used_days": 5.5,
      "remaining_days": 6.5
    }
  ]
}
```

#### Update Leave Quota (HR only)
```http
PUT /leave-quotas/{id}
Authorization: Bearer <token>
```

Request:
```json
{
  "total_days": 15.0,
  "reason": "Promoted to senior level"
}
```

#### Bulk Import Leave Quotas (HR only)
```http
POST /leave-quotas/import
Authorization: Bearer <token>
Content-Type: multipart/form-data
```

Form Data:
```
year: 2024
file: <excel/csv file>
```

#### List Leave Requests
```http
GET /leave-requests?status=pending&page=1&limit=20
Authorization: Bearer <token>
```

#### Create Leave Request
```http
POST /leave-requests
Authorization: Bearer <token>
```

Request:
```json
{
  "leave_type_id": 1,
  "from_date": "2024-11-15",
  "to_date": "2024-11-17",
  "reason": "Family vacation",
  "contact_during_leave": "+84901234567"
}
```

Response:
```json
{
  "data": {
    "id": 10,
    "employee": {
      "id": 1,
      "full_name": "John Doe"
    },
    "leave_type": {
      "id": 1,
      "name": "Annual Leave"
    },
    "from_date": "2024-11-15",
    "to_date": "2024-11-17",
    "total_days": 3.0,
    "reason": "Family vacation",
    "status": "pending",
    "submitted_at": "2024-10-20T10:30:00Z"
  },
  "warning": {
    "type": "QUOTA_EXCEEDED",
    "message": "This request exceeds your remaining leave quota by 2 days",
    "remaining_days": 1.5,
    "requested_days": 3.0
  }
}
```

#### Approve/Reject Leave Request (Manager only)
```http
POST /leave-requests/{id}/approve
Authorization: Bearer <token>
```

Approve:
```json
{
  "action": "approve"
}
```

Reject:
```json
{
  "action": "reject",
  "comment": "We need you during this period for important project deadline"
}
```

#### Cancel Leave Request
```http
DELETE /leave-requests/{id}
Authorization: Bearer <token>
```

#### Approve Leave via Email Link
```http
POST /leave-requests/{id}/email-action
```

Request:
```json
{
  "token": "signed_one_time_token_from_email",
  "action": "approve"
}
```

Note: This endpoint does not require Authorization header. The `token` is a signed, one-time use token sent via email to the manager.

#### Get Leave Requests for Manager
```http
GET /leave-requests/pending-approval?page=1&limit=20
Authorization: Bearer <token>
```

Response:
```json
{
  "data": [
    {
      "id": 10,
      "employee": {
        "id": 5,
        "full_name": "Jane Smith"
      },
      "leave_type": {
        "id": 1,
        "name": "Annual Leave"
      },
      "from_date": "2024-11-15",
      "to_date": "2024-11-17",
      "total_days": 3.0,
      "reason": "Family vacation",
      "status": "pending",
      "submitted_at": "2024-10-20T10:30:00Z"
    }
  ]
}
```

#### Get Company Leave Overview (HR only)
```http
GET /leave-requests/overview?year=2024&month=11
Authorization: Bearer <token>
```

Response:
```json
{
  "data": {
    "total_requests": 25,
    "pending": 5,
    "approved": 18,
    "rejected": 2,
    "employees_on_leave_today": 3,
    "upcoming_leaves": [
      {
        "employee": "John Doe",
        "from_date": "2024-11-20",
        "to_date": "2024-11-22"
      }
    ]
  }
}
```

---

### 5. Rewards

#### Get Point Balance
```http
GET /rewards/points?year=2024
Authorization: Bearer <token>
```

Response:
```json
{
  "data": {
    "id": 1,
    "employee_id": 1,
    "year": 2024,
    "initial_points": 100,
    "current_points": 75,
    "total_earned": 100,
    "total_spent": 25
  }
}
```

#### Grant Annual Points (HR/Admin only)
```http
POST /rewards/points/grant
Authorization: Bearer <token>
```

Request:
```json
{
  "year": 2025,
  "points": 100,
  "scope": "all",
  "department_id": null
}
```

Note: `scope` can be "all", "department", or "individual". If "department", provide `department_id`. If "individual", provide `employee_ids` array.

#### Get Points Grant History (HR/Admin only)
```http
GET /rewards/points/grants?year=2024
Authorization: Bearer <token>
```

Response:
```json
{
  "data": [
    {
      "id": 1,
      "year": 2024,
      "points": 100,
      "scope": "all",
      "granted_by": {
        "id": 1,
        "full_name": "Admin User"
      },
      "granted_at": "2024-01-01T00:00:00Z",
      "employees_affected": 50
    }
  ]
}
```

#### Get Reward Config (HR/Admin only)
```http
GET /rewards/config
Authorization: Bearer <token>
```

Response:
```json
{
  "data": {
    "annual_points": 100,
    "reset_date": "01-01",
    "reset_type": "yearly",
    "allow_negative_balance": false
  }
}
```

#### Update Reward Config (HR/Admin only)
```http
PUT /rewards/config
Authorization: Bearer <token>
```

Request:
```json
{
  "annual_points": 120,
  "reset_date": "01-01",
  "reset_type": "yearly",
  "allow_negative_balance": false
}
```

#### List Sticker Types
```http
GET /rewards/sticker-types
Authorization: Bearer <token>
```

Response:
```json
{
  "data": [
    {
      "id": 1,
      "name": "Thumbs Up",
      "description": "Great work!",
      "icon_url": "https://cdn.company.com/stickers/thumbs-up.png",
      "point_cost": 10,
      "category": "appreciation",
      "is_active": true
    }
  ]
}
```

#### Create Sticker Type (HR/Admin only)
```http
POST /rewards/sticker-types
Authorization: Bearer <token>
```

Request:
```json
{
  "name": "Star Performer",
  "description": "For outstanding performance",
  "icon_url": "https://cdn.company.com/stickers/star.png",
  "point_cost": 25,
  "category": "recognition",
  "is_active": true
}
```

#### Update Sticker Type (HR/Admin only)
```http
PUT /rewards/sticker-types/{id}
Authorization: Bearer <token>
```

Request:
```json
{
  "point_cost": 30,
  "is_active": false
}
```

#### Delete Sticker Type (HR/Admin only)
```http
DELETE /rewards/sticker-types/{id}
Authorization: Bearer <token>
```

#### Send Sticker
```http
POST /rewards/stickers
Authorization: Bearer <token>
```

Request:
```json
{
  "receiver_id": 5,
  "sticker_type_id": 1,
  "message": "Thanks for helping me with the bug fix!",
  "is_public": true
}
```

Response:
```json
{
  "data": {
    "id": 42,
    "sender": {
      "id": 1,
      "full_name": "John Doe"
    },
    "receiver": {
      "id": 5,
      "full_name": "Jane Smith"
    },
    "sticker_type": {
      "id": 1,
      "name": "Thumbs Up",
      "icon_url": "https://..."
    },
    "point_cost": 10,
    "message": "Thanks for helping me with the bug fix!",
    "created_at": "2024-10-20T15:30:00Z"
  }
}
```

#### Get Sticker Transactions
```http
GET /rewards/stickers?sent=true&received=false&page=1&limit=20
Authorization: Bearer <token>
```

#### Get Leaderboard
```http
GET /rewards/leaderboard?period=this_month&department_id=1&limit=10
Authorization: Bearer <token>
```

Response:
```json
{
  "data": [
    {
      "employee": {
        "id": 5,
        "full_name": "Jane Smith",
        "avatar_url": "https://..."
      },
      "department": "Engineering",
      "total_stickers_received": 45,
      "total_points_value": 650,
      "rank": 1
    },
    {
      "employee": {
        "id": 3,
        "full_name": "Bob Johnson"
      },
      "department": "Engineering",
      "total_stickers_received": 38,
      "total_points_value": 520,
      "rank": 2
    }
  ]
}
```

---

### 6. Documents

#### List Document Categories
```http
GET /documents/categories
Authorization: Bearer <token>
```

#### Create Document Category (HR only)
```http
POST /documents/categories
Authorization: Bearer <token>
```

Request:
```json
{
  "name": "Training Materials",
  "description": "Employee training documents",
  "parent_id": null
}
```

#### Update Document Category (HR only)
```http
PUT /documents/categories/{id}
Authorization: Bearer <token>
```

Request:
```json
{
  "name": "Updated Category Name",
  "description": "Updated description"
}
```

#### Delete Document Category (HR only)
```http
DELETE /documents/categories/{id}
Authorization: Bearer <token>
```

Note: Category can only be deleted if it has no documents.

#### List Documents
```http
GET /documents?category_id=1&search=handbook&page=1&limit=20
Authorization: Bearer <token>
```

Request:
```json
{
  "name": "Training Materials",
  "description": "Employee training documents",
  "parent_id": null
}
```

#### Update Document Category (HR only)
```http
PUT /documents/categories/{id}
Authorization: Bearer <token>
```

Request:
```json
{
  "name": "Updated Category Name",
  "description": "Updated description"
}
```

#### Delete Document Category (HR only)
```http
DELETE /documents/categories/{id}
Authorization: Bearer <token>
```

Note: Category can only be deleted if it has no documents.

#### List Documents
```http
GET /documents?category_id=1&search=handbook&page=1&limit=20
Authorization: Bearer <token>
```

#### Upload Document (HR only)
```http
POST /documents
Authorization: Bearer <token>
Content-Type: multipart/form-data
```

Form Data:
```
title: Employee Handbook 2024
description: Updated employee handbook
category_id: 1
is_public: true
file: <file>
```

#### Get Document Details
```http
GET /documents/{id}
Authorization: Bearer <token>
```

#### Download Document
```http
GET /documents/{id}/download
Authorization: Bearer <token>
```

#### Mark Document as Read
```http
POST /documents/{id}/mark-read
Authorization: Bearer <token>
```

#### Update Document (HR only)
```http
PUT /documents/{id}
Authorization: Bearer <token>
```

Request:
```json
{
  "title": "Updated Employee Handbook 2024",
  "description": "Revised version with policy updates",
  "category_id": 2,
  "is_public": true
}
```

#### Delete Document (HR only)
```http
DELETE /documents/{id}
Authorization: Bearer <token>
```

#### Get Document Read Status (HR only)
```http
GET /documents/{id}/readers
Authorization: Bearer <token>
```

Response:
```json
{
  "data": {
    "total_employees": 50,
    "read_count": 35,
    "unread_count": 15,
    "readers": [
      {
        "employee_id": 1,
        "full_name": "John Doe",
        "read_at": "2024-10-20T10:30:00Z"
      }
    ]
  }
}
```

---

### 7. Notifications

#### List Notifications
```http
GET /notifications?unread_only=true&page=1&limit=20
Authorization: Bearer <token>
```

Response:
```json
{
  "data": [
    {
      "id": 1,
      "type": "leave_request",
      "title": "New Leave Request",
      "message": "John Doe has submitted a leave request",
      "action_url": "/leave-requests/10",
      "is_read": false,
      "priority": "normal",
      "created_at": "2024-10-20T10:30:00Z"
    }
  ],
  "meta": {
    "page": 1,
    "limit": 20,
    "total": 5
  },
  "unread_count": 3
}
```

#### Mark Notification as Read
```http
POST /notifications/{id}/read
Authorization: Bearer <token>
```

#### Mark All as Read
```http
POST /notifications/read-all
Authorization: Bearer <token>
```

#### Get Notification Preferences
```http
GET /notifications/preferences
Authorization: Bearer <token>
```

Response:
```json
{
  "data": {
    "in_app": {
      "leave_request": true,
      "leave_approval": true,
      "attendance_reminder": true,
      "sticker_received": true,
      "birthday": true,
      "document_new": true
    },
    "email": {
      "leave_request": true,
      "leave_approval": true,
      "attendance_reminder": false,
      "sticker_received": false,
      "birthday": true,
      "document_new": false
    }
  }
}
```

#### Update Notification Preferences
```http
PUT /notifications/preferences
Authorization: Bearer <token>
```

Request:
```json
{
  "email": {
    "leave_approval": true,
    "sticker_received": true
  }
}
```

#### Get Notification Types
```http
GET /notifications/types
Authorization: Bearer <token>
```

Response:
```json
{
  "data": [
    {
      "type": "leave_request",
      "name": "Leave Request",
      "description": "When someone submits a leave request for your approval",
      "channels": ["in_app", "email"]
    },
    {
      "type": "leave_approval",
      "name": "Leave Approval/Rejection",
      "description": "When your leave request is approved or rejected",
      "channels": ["in_app", "email"]
    },
    {
      "type": "attendance_reminder",
      "name": "Attendance Confirmation Reminder",
      "description": "Reminder to confirm your monthly attendance",
      "channels": ["in_app", "email"]
    },
    {
      "type": "sticker_received",
      "name": "Sticker Received",
      "description": "When someone sends you a sticker",
      "channels": ["in_app", "email"]
    },
    {
      "type": "birthday",
      "name": "Birthday Notification",
      "description": "Daily birthday announcements",
      "channels": ["in_app", "email"]
    },
    {
      "type": "document_new",
      "name": "New Document",
      "description": "When a new document is uploaded",
      "channels": ["in_app", "email"]
    }
  ]
}
```

---

### 8. System

#### Health Check
```http
GET /system/health
```

Response:
```json
{
  "status": "healthy",
  "database": "connected",
  "version": "1.0.0",
  "timestamp": "2024-10-20T10:00:00Z"
}
```

#### Get System Settings (Admin only)
```http
GET /system/settings
Authorization: Bearer <token>
```

#### Update System Settings (Admin only)
```http
PUT /system/settings
Authorization: Bearer <token>
```

Request:
```json
{
  "company_name": "My Company",
  "timezone": "Asia/Ho_Chi_Minh",
  "date_format": "DD/MM/YYYY",
  "working_days": ["monday", "tuesday", "wednesday", "thursday", "friday"]
}
```

#### List Departments
```http
GET /departments
Authorization: Bearer <token>
```

---

### 9. Audit Logs (Admin/HR only)

#### List Audit Logs
```http
GET /audit-logs?page=1&limit=50&action=leave.approve&actor_id=1&from=2024-10-01&to=2024-10-31
Authorization: Bearer <token>
```

Query Parameters:
- `action`: Filter by action type (e.g., `leave.approve`, `leave.reject`, `sticker.send`, `employee.create`, `employee.offboard`, `attendance.confirm`)
- `actor_id`: Filter by user who performed the action
- `target_type`: Filter by target entity type (e.g., `employee`, `leave_request`, `sticker`)
- `target_id`: Filter by specific target entity ID
- `from`: Start date (ISO 8601)
- `to`: End date (ISO 8601)

Response:
```json
{
  "data": [
    {
      "id": 1,
      "action": "leave.approve",
      "actor": {
        "id": 5,
        "full_name": "Manager Name",
        "email": "manager@company.com"
      },
      "target_type": "leave_request",
      "target_id": 10,
      "details": {
        "leave_type": "Annual Leave",
        "from_date": "2024-11-15",
        "to_date": "2024-11-17",
        "employee_name": "John Doe"
      },
      "ip_address": "192.168.1.100",
      "user_agent": "Mozilla/5.0...",
      "created_at": "2024-10-20T14:30:00Z"
    },
    {
      "id": 2,
      "action": "sticker.send",
      "actor": {
        "id": 1,
        "full_name": "John Doe",
        "email": "john.doe@company.com"
      },
      "target_type": "sticker",
      "target_id": 42,
      "details": {
        "receiver_name": "Jane Smith",
        "sticker_type": "Thumbs Up",
        "point_cost": 10
      },
      "ip_address": "192.168.1.50",
      "user_agent": "Mozilla/5.0...",
      "created_at": "2024-10-20T15:30:00Z"
    }
  ],
  "meta": {
    "page": 1,
    "limit": 50,
    "total": 234,
    "total_pages": 5
  }
}
```

#### Get Audit Log Detail
```http
GET /audit-logs/{id}
Authorization: Bearer <token>
```

#### Get Audit Log Actions
```http
GET /audit-logs/actions
Authorization: Bearer <token>
```

Response:
```json
{
  "data": [
    {
      "action": "leave.approve",
      "display_name": "Leave Request Approved",
      "category": "leave"
    },
    {
      "action": "leave.reject",
      "display_name": "Leave Request Rejected",
      "category": "leave"
    },
    {
      "action": "sticker.send",
      "display_name": "Sticker Sent",
      "category": "rewards"
    },
    {
      "action": "employee.create",
      "display_name": "Employee Created (Onboard)",
      "category": "employee"
    },
    {
      "action": "employee.offboard",
      "display_name": "Employee Offboarded",
      "category": "employee"
    },
    {
      "action": "attendance.confirm",
      "display_name": "Attendance Confirmed",
      "category": "attendance"
    },
    {
      "action": "attendance.auto_confirm",
      "display_name": "Attendance Auto-Confirmed",
      "category": "attendance"
    },
    {
      "action": "user.login",
      "display_name": "User Login",
      "category": "auth"
    },
    {
      "action": "user.password_change",
      "display_name": "Password Changed",
      "category": "auth"
    }
  ]
}
```

#### Export Audit Logs (Admin only)
```http
GET /audit-logs/export?format=csv&from=2024-10-01&to=2024-10-31
Authorization: Bearer <token>
```

Response: CSV file download

---

### 10. Admin (Admin only)

#### List Users
```http
GET /admin/users?page=1&limit=20&status=active&role=employee
Authorization: Bearer <token>
```

Response:
```json
{
  "data": [
    {
      "id": 1,
      "email": "john.doe@company.com",
      "full_name": "John Doe",
      "roles": ["employee"],
      "status": "active",
      "last_login_at": "2024-10-20T10:00:00Z",
      "created_at": "2024-01-15T00:00:00Z"
    }
  ]
}
```

#### Create User
```http
POST /admin/users
Authorization: Bearer <token>
```

Request:
```json
{
  "email": "new.user@company.com",
  "password": "temporary_password",
  "roles": ["employee"],
  "employee_id": 10
}
```

#### Update User
```http
PUT /admin/users/{id}
Authorization: Bearer <token>
```

Request:
```json
{
  "roles": ["employee", "manager"],
  "status": "active"
}
```

#### Disable User
```http
POST /admin/users/{id}/disable
Authorization: Bearer <token>
```

Request:
```json
{
  "reason": "Employee offboarded"
}
```

#### Enable User
```http
POST /admin/users/{id}/enable
Authorization: Bearer <token>
```

#### Reset User Password
```http
POST /admin/users/{id}/reset-password
Authorization: Bearer <token>
```

Response:
```json
{
  "temporary_password": "auto_generated_temp_password",
  "message": "User will be required to change password on next login"
}
```

#### List Roles
```http
GET /admin/roles
Authorization: Bearer <token>
```

Response:
```json
{
  "data": [
    {
      "id": 1,
      "name": "admin",
      "display_name": "Administrator",
      "description": "Full system access",
      "permissions": ["users.manage", "settings.manage", "all"]
    },
    {
      "id": 2,
      "name": "hr",
      "display_name": "Human Resources",
      "description": "HR operations access",
      "permissions": ["employees.manage", "attendance.manage", "documents.manage"]
    },
    {
      "id": 3,
      "name": "manager",
      "display_name": "Manager",
      "description": "Team management access",
      "permissions": ["leave.approve", "employees.view_subordinates"]
    },
    {
      "id": 4,
      "name": "employee",
      "display_name": "Employee",
      "description": "Basic employee access",
      "permissions": ["profile.manage", "leave.request", "stickers.send"]
    }
  ]
}
```

#### Assign Roles to User
```http
POST /admin/users/{id}/roles
Authorization: Bearer <token>
```

Request:
```json
{
  "roles": ["employee", "manager"]
}
```

---

### 11. Background Jobs (Admin only)

#### List Scheduled Jobs
```http
GET /system/jobs
Authorization: Bearer <token>
```

Response:
```json
{
  "data": [
    {
      "id": "birthday_notification",
      "name": "Birthday Notification",
      "description": "Send daily birthday notifications",
      "schedule": "0 9 * * *",
      "next_run": "2024-10-21T09:00:00Z",
      "last_run": "2024-10-20T09:00:00Z",
      "last_status": "success",
      "enabled": true
    },
    {
      "id": "points_reset",
      "name": "Annual Points Reset",
      "description": "Reset employee points at the start of year",
      "schedule": "0 0 1 1 *",
      "next_run": "2025-01-01T00:00:00Z",
      "last_run": "2024-01-01T00:00:00Z",
      "last_status": "success",
      "enabled": true
    },
    {
      "id": "attendance_auto_confirm",
      "name": "Attendance Auto-Confirm",
      "description": "Auto-confirm pending attendance after deadline",
      "schedule": "0 0 * * *",
      "next_run": "2024-10-21T00:00:00Z",
      "last_run": "2024-10-20T00:00:00Z",
      "last_status": "success",
      "enabled": true
    },
    {
      "id": "attendance_reminder",
      "name": "Attendance Confirmation Reminder",
      "description": "Send reminder before confirmation deadline",
      "schedule": "0 10 * * *",
      "next_run": "2024-10-21T10:00:00Z",
      "last_run": "2024-10-20T10:00:00Z",
      "last_status": "success",
      "enabled": true
    }
  ]
}
```

#### Get Job Details
```http
GET /system/jobs/{id}
Authorization: Bearer <token>
```

#### Update Job Schedule
```http
PUT /system/jobs/{id}
Authorization: Bearer <token>
```

Request:
```json
{
  "schedule": "0 8 * * *",
  "enabled": true
}
```

#### Get Job History
```http
GET /system/jobs/{id}/history?page=1&limit=20
Authorization: Bearer <token>
```

Response:
```json
{
  "data": [
    {
      "id": 1,
      "job_id": "birthday_notification",
      "started_at": "2024-10-20T09:00:00Z",
      "completed_at": "2024-10-20T09:00:05Z",
      "status": "success",
      "details": {
        "employees_notified": 3,
        "emails_sent": 50
      }
    },
    {
      "id": 2,
      "job_id": "birthday_notification",
      "started_at": "2024-10-19T09:00:00Z",
      "completed_at": "2024-10-19T09:00:03Z",
      "status": "success",
      "details": {
        "employees_notified": 1,
        "emails_sent": 50
      }
    }
  ]
}
```

#### Trigger Job Manually
```http
POST /system/jobs/{id}/run
Authorization: Bearer <token>
```

Response:
```json
{
  "message": "Job triggered successfully",
  "execution_id": "exec_123456"
}
```

---

## Code Examples

### JavaScript (Fetch API)

```javascript
// Login
async function login(email, password) {
  const response = await fetch('http://localhost:8080/api/v1/auth/login', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ email, password }),
  });
  
  const data = await response.json();
  
  if (!response.ok) {
    throw new Error(data.error.message);
  }
  
  // Store token
  localStorage.setItem('token', data.token);
  return data;
}

// Get employees with token
async function getEmployees(page = 1, limit = 20) {
  const token = localStorage.getItem('token');
  
  const response = await fetch(
    `http://localhost:8080/api/v1/employees?page=${page}&limit=${limit}`,
    {
      headers: {
        'Authorization': `Bearer ${token}`,
      },
    }
  );
  
  return response.json();
}

// Create leave request
async function createLeaveRequest(data) {
  const token = localStorage.getItem('token');
  
  const response = await fetch('http://localhost:8080/api/v1/leave-requests', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
    },
    body: JSON.stringify(data),
  });
  
  return response.json();
}

// Send sticker
async function sendSticker(receiverId, stickerTypeId, message) {
  const token = localStorage.getItem('token');
  
  const response = await fetch('http://localhost:8080/api/v1/rewards/stickers', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
    },
    body: JSON.stringify({
      receiver_id: receiverId,
      sticker_type_id: stickerTypeId,
      message,
      is_public: true,
    }),
  });
  
  return response.json();
}
```

### React Hook Example

```javascript
import { useState, useEffect } from 'react';

function useApi(endpoint) {
  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  
  useEffect(() => {
    const fetchData = async () => {
      try {
        const token = localStorage.getItem('token');
        const response = await fetch(`http://localhost:8080/api/v1${endpoint}`, {
          headers: {
            'Authorization': `Bearer ${token}`,
          },
        });
        
        const json = await response.json();
        
        if (!response.ok) {
          throw new Error(json.error.message);
        }
        
        setData(json.data);
      } catch (err) {
        setError(err.message);
      } finally {
        setLoading(false);
      }
    };
    
    fetchData();
  }, [endpoint]);
  
  return { data, loading, error };
}

// Usage in component
function EmployeeList() {
  const { data: employees, loading, error } = useApi('/employees');
  
  if (loading) return <div>Loading...</div>;
  if (error) return <div>Error: {error}</div>;
  
  return (
    <ul>
      {employees.map(emp => (
        <li key={emp.id}>{emp.full_name}</li>
      ))}
    </ul>
  );
}
```

### Golang Client Example

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
)

const baseURL = "http://localhost:8080/api/v1"

type Client struct {
    token      string
    httpClient *http.Client
}

func NewClient(token string) *Client {
    return &Client{
        token:      token,
        httpClient: &http.Client{},
    }
}

func (c *Client) Login(email, password string) (string, error) {
    body := map[string]string{
        "email":    email,
        "password": password,
    }
    
    jsonBody, _ := json.Marshal(body)
    
    resp, err := c.httpClient.Post(
        baseURL+"/auth/login",
        "application/json",
        bytes.NewBuffer(jsonBody),
    )
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()
    
    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)
    
    token := result["token"].(string)
    c.token = token
    
    return token, nil
}

func (c *Client) GetEmployees() ([]map[string]interface{}, error) {
    req, _ := http.NewRequest("GET", baseURL+"/employees", nil)
    req.Header.Set("Authorization", "Bearer "+c.token)
    
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var result struct {
        Data []map[string]interface{} `json:"data"`
    }
    
    json.NewDecoder(resp.Body).Decode(&result)
    
    return result.Data, nil
}

func main() {
    client := NewClient("")
    
    // Login
    token, err := client.Login("john.doe@company.com", "password123")
    if err != nil {
        panic(err)
    }
    
    fmt.Println("Token:", token)
    
    // Get employees
    employees, err := client.GetEmployees()
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Found %d employees\n", len(employees))
}
```

---

## Postman Collection

### Import OpenAPI Spec to Postman

1. Open Postman
2. Click **Import** button
3. Select `api-documentation.yaml`
4. Postman will auto-generate collection

### Set up Environment Variables

Create a Postman environment with:

```
base_url: http://localhost:8080/api/v1
token: <will be set after login>
```

### Pre-request Script for Authentication

Add this to collection-level pre-request script:

```javascript
// Check if token exists
const token = pm.environment.get("token");

if (token) {
    pm.request.headers.add({
        key: 'Authorization',
        value: 'Bearer ' + token
    });
}
```

### Test Script Example

```javascript
// Save token after login
if (pm.response.code === 200 && pm.response.json().token) {
    pm.environment.set("token", pm.response.json().token);
}

// Test response
pm.test("Status code is 200", function () {
    pm.response.to.have.status(200);
});

pm.test("Response has data", function () {
    const jsonData = pm.response.json();
    pm.expect(jsonData).to.have.property('data');
});
```

---

## Best Practices

### 1. Always Handle Errors

```javascript
async function apiCall() {
  try {
    const response = await fetch(url, options);
    const data = await response.json();
    
    if (!response.ok) {
      // Handle API error
      throw new Error(data.error.message);
    }
    
    return data;
  } catch (error) {
    // Handle network error
    console.error('API Error:', error);
    throw error;
  }
}
```

### 2. Use Interceptors (Axios Example)

```javascript
import axios from 'axios';

const api = axios.create({
  baseURL: 'http://localhost:8080/api/v1',
});

// Request interceptor - add token
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

// Response interceptor - handle errors
api.interceptors.response.use(
  (response) => response.data,
  (error) => {
    if (error.response?.status === 401) {
      // Token expired - redirect to login
      localStorage.removeItem('token');
      window.location.href = '/login';
    }
    return Promise.reject(error.response?.data?.error || error);
  }
);

export default api;
```

### 3. Implement Retry Logic

```javascript
async function fetchWithRetry(url, options, retries = 3) {
  for (let i = 0; i < retries; i++) {
    try {
      const response = await fetch(url, options);
      return response;
    } catch (error) {
      if (i === retries - 1) throw error;
      await new Promise(r => setTimeout(r, 1000 * (i + 1)));
    }
  }
}
```

### 4. Cache Responses (React Query Example)

```javascript
import { useQuery } from '@tanstack/react-query';

function useEmployees() {
  return useQuery({
    queryKey: ['employees'],
    queryFn: async () => {
      const response = await api.get('/employees');
      return response.data;
    },
    staleTime: 5 * 60 * 1000, // 5 minutes
  });
}
```

---

## Testing

### Unit Tests Example (Go)

```go
func TestLoginAPI(t *testing.T) {
    router := setupRouter()
    
    w := httptest.NewRecorder()
    body := `{"email":"test@company.com","password":"password123"}`
    req, _ := http.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    
    router.ServeHTTP(w, req)
    
    assert.Equal(t, 200, w.Code)
    
    var response map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)
    
    assert.NotEmpty(t, response["token"])
}
```

### Integration Tests

```bash
# Using curl
./test-api.sh
```

```bash
#!/bin/bash

BASE_URL="http://localhost:8080/api/v1"

# Login
echo "Testing login..."
TOKEN=$(curl -s -X POST $BASE_URL/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@company.com","password":"admin123"}' \
  | jq -r '.token')

echo "Token: $TOKEN"

# Get employees
echo "Testing get employees..."
curl -s -X GET $BASE_URL/employees \
  -H "Authorization: Bearer $TOKEN" \
  | jq '.data | length'

# Create leave request
echo "Testing create leave request..."
curl -s -X POST $BASE_URL/leave-requests \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "leave_type_id": 1,
    "from_date": "2024-12-01",
    "to_date": "2024-12-03",
    "reason": "Family vacation"
  }' \
  | jq '.data.id'

echo "All tests passed!"
```

---

## FAQ

**Q: How long is the JWT token valid?**  
A: Access tokens are valid for 1 hour. Use refresh token to get new access token.

**Q: Can I use the API from different domains?**  
A: Yes, CORS is enabled for configured origins.

**Q: Is there API versioning?**  
A: Yes, current version is `/api/v1`. Breaking changes will use `/api/v2`.

**Q: How to upload files?**  
A: Use `multipart/form-data` content type. See attendance and document upload examples.

**Q: What's the file size limit for uploads?**  
A: 10MB for attendance files, 50MB for documents.

**Q: How are dates formatted?**  
A: Dates use ISO 8601 format: `YYYY-MM-DD` for dates, `YYYY-MM-DDTHH:mm:ssZ` for timestamps.

---

## Support

- **Documentation:** [View OpenAPI Spec](./api-documentation.yaml)
- **Issues:** Create issue on GitHub
- **Email:** support@company.com

---

## Changelog

### v1.0.0 (2024-10-20)
- Initial API release
- All core features implemented
- Full CRUD operations for all entities
- JWT authentication
- Role-based access control