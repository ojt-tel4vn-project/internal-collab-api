package employee

import (
	"time"

	"github.com/google/uuid"
)

// Create Employee DTOs (HR only)
type CreateEmployeeRequest struct {
	Email        string     `json:"email" binding:"required,email"`
	FirstName    string     `json:"first_name" binding:"required"`
	LastName     string     `json:"last_name" binding:"required"`
	DateOfBirth  string     `json:"date_of_birth" binding:"required"` // Format: YYYY-MM-DD
	Phone        string     `json:"phone"`
	Address      string     `json:"address"`
	DepartmentID *uuid.UUID `json:"department_id,omitempty"`
	Position     string     `json:"position" binding:"required"`
	ManagerID    *uuid.UUID `json:"manager_id,omitempty"`
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
	Phone        *string    `json:"phone"`
	Address      *string    `json:"address"`
	DepartmentID *uuid.UUID `json:"department_id"`
	Position     *string    `json:"position"`
	ManagerID    *uuid.UUID `json:"manager_id"`
	Status       *string    `json:"status"` // 'active', 'inactive'
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

// List Employees DTOs
type ListEmployeesResponse struct {
	Employees []EmployeeSummary `json:"employees"`
	Total     int               `json:"total"`
}

type EmployeeSummary struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	FullName     string    `json:"full_name"`
	EmployeeCode string    `json:"employee_code"`
	Position     string    `json:"position"`
	Department   string    `json:"department"`
	Status       string    `json:"status"`
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
	JoinDate    time.Time  `json:"join_date"`
	LeaveDate   *time.Time `json:"leave_date,omitempty"`
	Status      string     `json:"status"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
