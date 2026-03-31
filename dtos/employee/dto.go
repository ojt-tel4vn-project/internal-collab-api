package employee

import (
	"time"

	"github.com/google/uuid"
)

// Create Employee DTOs (HR only)
type CreateEmployeeRequest struct {
	Email        string     `json:"email" binding:"required,email" format:"email" doc:"Valid employee email address" example:"john.doe@example.com"`
	FirstName    string     `json:"first_name" binding:"required"`
	LastName     string     `json:"last_name" binding:"required"`
	DateOfBirth  string     `json:"date_of_birth" binding:"required"` // Format: YYYY-MM-DD
	Phone        string     `json:"phone" pattern:"^0[0-9]{9}$" doc:"10-digit phone number starting with 0" example:"0912345678"`
	Address      string     `json:"address"`
	DepartmentID *uuid.UUID `json:"department_id"`
	Position     string     `json:"position" binding:"required"`
	ManagerID    *uuid.UUID `json:"manager_id,omitempty"`
	RoleID       *uuid.UUID `json:"role_id,omitempty"`
	JoinDate     string     `json:"join_date"` // Format: YYYY-MM-DD, defaults to today
}

type CreateEmployeeResponse struct {
	Message           string `json:"message"`
	TemporaryPassword string `json:"temporary_password"`
	Employee          struct {
		ID           uuid.UUID  `json:"id"`
		Email        string     `json:"email"`
		FullName     string     `json:"full_name"`
		EmployeeCode string     `json:"employee_code"`
		Position     string     `json:"position"`
		DepartmentID *uuid.UUID `json:"department_id"`
		Status       string     `json:"status"`
		JoinDate     time.Time  `json:"join_date"`
	} `json:"employee"`
}

// Update Employee DTOs
type UpdateEmployeeRequest struct {
	FirstName    *string    `json:"first_name"`
	LastName     *string    `json:"last_name"`
	DateOfBirth  *string    `json:"date_of_birth"` // Format: YYYY-MM-DD
	Phone        *string    `json:"phone" pattern:"^0[0-9]{9}$" doc:"10-digit phone number starting with 0" example:"0912345678"`
	Address      *string    `json:"address"`
	DepartmentID *uuid.UUID `json:"department_id"`
	Position     *string    `json:"position"`
	ManagerID    *uuid.UUID `json:"manager_id"`
	RoleID       *uuid.UUID `json:"role_id"`
	Status       *string    `json:"status"` // 'active', 'offboard'
}

type UpdateEmployeeResponse struct {
	Message  string `json:"message"`
	Employee struct {
		ID           uuid.UUID  `json:"id"`
		Email        string     `json:"email"`
		FullName     string     `json:"full_name"`
		EmployeeCode string     `json:"employee_code"`
		Position     string     `json:"position"`
		DepartmentID *uuid.UUID `json:"department_id"`
		Status       string     `json:"status"`
	} `json:"employee"`
}

// Update Profile DTOs (Self-Service)
type UpdateProfileRequest struct {
	Phone     *string `json:"phone" pattern:"^0[0-9]{9}$" doc:"10-digit phone number starting with 0" example:"0912345678"`
	Address   *string `json:"address"`
	AvatarUrl *string `json:"avatar_url"`
}

type UpdateProfileResponse struct {
	Message string `json:"message"`
}

// List Employees DTOs
type ListEmployeesResponse struct {
	Employees []EmployeeSummary `json:"employees"`
	Total     int               `json:"total"`
}

type EmployeeSummary struct {
	ID           uuid.UUID        `json:"id"`
	Email        string           `json:"email"`
	FullName     string           `json:"full_name"`
	EmployeeCode string           `json:"employee_code"`
	Position     string           `json:"position"`
	Department   *DepartmentBrief `json:"department"`
	Role         *RoleBrief       `json:"role"`
	AvatarUrl    string           `json:"avatar_url"`
	Status       string           `json:"status"`
}

// DepartmentBrief is a slim department object
type DepartmentBrief struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

// RoleBrief is a slim role object
type RoleBrief struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

// Get Employee Detail DTOs
type GetEmployeeResponse struct {
	ID           uuid.UUID  `json:"id"`
	Email        string     `json:"email"`
	EmployeeCode string     `json:"employee_code"`
	FirstName    string     `json:"first_name"`
	LastName     string     `json:"last_name"`
	FullName     string     `json:"full_name"`
	DateOfBirth  time.Time  `json:"date_of_birth"`
	Phone        string     `json:"phone"`
	Address      string     `json:"address"`
	AvatarUrl    string     `json:"avatar_url"`
	DepartmentID *uuid.UUID `json:"department_id"`
	Department   *struct {
		ID   uuid.UUID `json:"id"`
		Name string    `json:"name"`
	} `json:"department,omitempty"`
	Position  string     `json:"position"`
	ManagerID *uuid.UUID `json:"manager_id"`
	Manager   *struct {
		ID       uuid.UUID `json:"id"`
		FullName string    `json:"full_name"`
	} `json:"manager,omitempty"`
	RoleID      *uuid.UUID `json:"role_id"`
	Role        *RoleBrief `json:"role,omitempty"`
	JoinDate    time.Time  `json:"join_date"`
	LeaveDate   *time.Time `json:"leave_date,omitempty"`
	Status      string     `json:"status"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// Check Birthday DTOs
type ListBirthdayResponse struct {
	Employees []BirthdaySummary `json:"employees"`
	Total     int               `json:"total"`
	Message   string            `json:"message"`
}

// ListAllBirthdaysResponse for the calendar endpoint (all employees, all dates)
type ListAllBirthdaysResponse struct {
	Employees []BirthdaySummary `json:"employees"`
	Total     int               `json:"total"`
}

type BirthdaySummary struct {
	ID         uuid.UUID `json:"id"`
	FullName   string    `json:"full_name"`
	Email      string    `json:"email"`
	Department string    `json:"department"`
	Position   string    `json:"position"`
	BirthDate  string    `json:"birth_date"`
}

// Birthday Config DTOs
type BirthdayConfig struct {
	Enabled          bool     `json:"enabled"`
	NotificationTime string   `json:"notification_time"`
	Channels         []string `json:"channels"`
}

type GetBirthdayConfigResponse struct {
	Data BirthdayConfig `json:"data"`
}

type UpdateBirthdayConfigRequest struct {
	Enabled          bool     `json:"enabled"`
	NotificationTime string   `json:"notification_time"`
	Channels         []string `json:"channels"`
}

type UpdateBirthdayConfigResponse struct {
	Message string         `json:"message"`
	Data    BirthdayConfig `json:"data"`
}

type SearchEmployeeRequest struct {
	Query string `json:"query" query:"query" doc:"Search by name, email"`
}

type SearchEmployeeResponse struct {
	Employees []SearchEmployeeSummary `json:"employees"`
	Total     int                     `json:"total"`
}

type SearchEmployeeSummary struct {
	ID           uuid.UUID        `json:"id"`
	Email        string           `json:"email"`
	FullName     string           `json:"full_name"`
	EmployeeCode string           `json:"employee_code"`
	Phone        string           `json:"phone"`
	Position     string           `json:"position"`
	Department   *DepartmentBrief `json:"department"`
	AvatarUrl    string           `json:"avatar_url"`
	Status       string           `json:"status"`
	JoinDate     time.Time        `json:"join_date"`
}
