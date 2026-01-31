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
- 🔐 JWT Authentication
- 👥 Employee Management
- 📅 Attendance Tracking
- 🏖️ Leave Management
- 🎁 Reward System (Stickers)
- 📚 Document Management
- 🔔 Real-time Notifications

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
  "status": "confirmed",
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

#### List Departments
```http
GET /departments
Authorization: Bearer <token>
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