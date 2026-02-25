package services

import (
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/google/uuid"
	"github.com/ojt-tel4vn-project/internal-collab-api/dtos/employee"
	"github.com/ojt-tel4vn-project/internal-collab-api/models"
	"github.com/ojt-tel4vn-project/internal-collab-api/pkg/crypto"
	"github.com/ojt-tel4vn-project/internal-collab-api/pkg/email"
	"github.com/ojt-tel4vn-project/internal-collab-api/pkg/logger"
	"github.com/ojt-tel4vn-project/internal-collab-api/pkg/response"
	"github.com/ojt-tel4vn-project/internal-collab-api/pkg/utils"
	"github.com/ojt-tel4vn-project/internal-collab-api/repository"
	"go.uber.org/zap"
)

type EmployeeService interface {
	CreateEmployee(req *employee.CreateEmployeeRequest) (*employee.CreateEmployeeResponse, error)
	GetAllEmployees() (*employee.ListEmployeesResponse, error)
	GetEmployeeByID(id uuid.UUID) (*employee.GetEmployeeResponse, error)
	UpdateEmployee(id uuid.UUID, req *employee.UpdateEmployeeRequest) (*employee.UpdateEmployeeResponse, error)
	DeleteEmployee(id uuid.UUID) error
	GetTodayBirthdays() (*employee.ListBirthdayResponse, error)
	GetSubordinates(managerID uuid.UUID) (*employee.ListSubordinatesResponse, error)
}

type employeeServiceImpl struct {
	repo         repository.EmployeeRepository
	password     crypto.PasswordService
	emailService email.EmailService
}

func NewEmployeeService(repo repository.EmployeeRepository, password crypto.PasswordService, emailService email.EmailService) EmployeeService {
	return &employeeServiceImpl{
		repo:         repo,
		password:     password,
		emailService: emailService,
	}
}

// Generate random temporary password
func generateTemporaryPassword() string {
	b := make([]byte, 12)
	rand.Read(b)
	return "TEMP_" + base64.URLEncoding.EncodeToString(b)[:12]
}

func (s *employeeServiceImpl) CreateEmployee(req *employee.CreateEmployeeRequest) (*employee.CreateEmployeeResponse, error) {
	// Check if email already exists
	_, err := s.repo.FindByEmail(req.Email)
	if err == nil {
		logger.Warn("CreateEmployee failed: email already exists", zap.String("email", req.Email))
		return nil, response.Conflict("Email already exists")
	}

	// Generate temporary password
	tempPassword := generateTemporaryPassword()
	hashedPassword, err := s.password.HashPassword(tempPassword)
	if err != nil {
		logger.Error("CreateEmployee failed: password hashing error", zap.Error(err))
		return nil, response.InternalServerError("Failed to create employee")
	}

	// Parse dates
	var dob time.Time
	if req.DateOfBirth != "" {
		dob, err = time.Parse("2006-01-02", req.DateOfBirth)
		if err != nil {
			logger.Warn("CreateEmployee failed: invalid date of birth", zap.Error(err))
			return nil, response.BadRequest("Invalid date of birth format. Use YYYY-MM-DD")
		}
	}

	var joinDate time.Time
	if req.JoinDate != "" {
		joinDate, err = time.Parse("2006-01-02", req.JoinDate)
		if err != nil {
			logger.Warn("CreateEmployee failed: invalid join date", zap.Error(err))
			return nil, response.BadRequest("Invalid join date format. Use YYYY-MM-DD")
		}
	} else {
		joinDate = time.Now()
	}

	// Generate employee code
	// Default prefix EMP, or could be derived from department/role
	employeeCode, err := utils.GenerateEmployeeCode("EMP")
	if err != nil {
		logger.Error("CreateEmployee failed: code generation error", zap.Error(err))
		return nil, response.InternalServerError("Failed to generate employee code")
	}

	// Create employee model
	newEmployee := models.Employee{
		Email:        req.Email,
		PasswordHash: hashedPassword,
		EmployeeCode: employeeCode,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		DateOfBirth:  dob,
		Phone:        req.Phone,
		Address:      req.Address,
		DepartmentID: req.DepartmentID,
		Position:     req.Position,
		ManagerID:    req.ManagerID,
		JoinDate:     joinDate,
		Status:       models.StatusPending, // Pending until first-time setup
	}

	// Save to database
	if err := s.repo.Create(&newEmployee); err != nil {
		logger.Error("CreateEmployee failed: database create error", zap.Error(err))
		return nil, response.InternalServerError("Failed to create employee")
	}

	logger.Info("Employee created successfully", zap.String("employee_id", newEmployee.ID.String()))

	// Send welcome email with temporary password
	if s.emailService != nil {
		err = s.emailService.SendWelcomeEmail(newEmployee.Email, newEmployee.FullName, tempPassword)
		if err != nil {
			logger.Warn("Failed to send welcome email",
				zap.Error(err),
				zap.String("employee_id", newEmployee.ID.String()),
				zap.String("email", newEmployee.Email),
			)
			// Don't fail the whole operation if email fails
			// Employee is already created, just log the error
		} else {
			logger.Info("Welcome email sent successfully",
				zap.String("employee_id", newEmployee.ID.String()),
				zap.String("email", newEmployee.Email),
			)
		}
	}

	// Return response
	return &employee.CreateEmployeeResponse{
		Message:           "Employee created successfully. Email sent with login credentials.",
		TemporaryPassword: tempPassword,
		Employee: struct {
			ID           uuid.UUID  `json:"id"`
			Email        string     `json:"email"`
			FullName     string     `json:"full_name"`
			EmployeeCode string     `json:"employee_code"`
			Position     string     `json:"position"`
			DepartmentID *uuid.UUID `json:"department_id"`
			Status       string     `json:"status"`
			JoinDate     time.Time  `json:"join_date"`
		}{
			ID:           newEmployee.ID,
			Email:        newEmployee.Email,
			FullName:     newEmployee.FullName,
			EmployeeCode: newEmployee.EmployeeCode,
			Position:     newEmployee.Position,
			DepartmentID: newEmployee.DepartmentID,
			Status:       string(newEmployee.Status),
			JoinDate:     newEmployee.JoinDate,
		},
	}, nil
}

func (s *employeeServiceImpl) GetAllEmployees() (*employee.ListEmployeesResponse, error) {
	employees, err := s.repo.FindAll()
	if err != nil {
		logger.Error("GetAllEmployees failed", zap.Error(err))
		return nil, response.InternalServerError("Failed to fetch employees")
	}

	// Convert to summary
	summaries := make([]employee.EmployeeSummary, len(employees))
	for i, emp := range employees {
		departmentName := ""
		if emp.Department != nil {
			departmentName = emp.Department.Name
		}

		summaries[i] = employee.EmployeeSummary{
			ID:           emp.ID,
			Email:        emp.Email,
			FullName:     emp.FullName,
			EmployeeCode: emp.EmployeeCode,
			Position:     emp.Position,
			Department:   departmentName,
			Status:       string(emp.Status),
		}
	}

	return &employee.ListEmployeesResponse{
		Employees: summaries,
		Total:     len(summaries),
	}, nil
}

func (s *employeeServiceImpl) GetEmployeeByID(id uuid.UUID) (*employee.GetEmployeeResponse, error) {
	emp, err := s.repo.FindByID(id)
	if err != nil {
		logger.Warn("GetEmployeeByID failed: employee not found", zap.String("id", id.String()))
		return nil, response.NotFound("Employee not found")
	}

	// Convert to response
	resp := &employee.GetEmployeeResponse{
		ID:           emp.ID,
		Email:        emp.Email,
		EmployeeCode: emp.EmployeeCode,
		FirstName:    emp.FirstName,
		LastName:     emp.LastName,
		FullName:     emp.FullName,
		DateOfBirth:  emp.DateOfBirth,
		Phone:        emp.Phone,
		Address:      emp.Address,
		DepartmentID: emp.DepartmentID,
		Position:     emp.Position,
		ManagerID:    emp.ManagerID,
		JoinDate:     emp.JoinDate,
		LeaveDate:    emp.LeaveDate,
		Status:       string(emp.Status),
		LastLoginAt:  emp.LastLoginAt,
		CreatedAt:    emp.CreatedAt,
		UpdatedAt:    emp.UpdatedAt,
	}

	// Add department info if exists
	if emp.Department != nil {
		resp.Department = &struct {
			ID   uuid.UUID `json:"id"`
			Name string    `json:"name"`
		}{
			ID:   emp.Department.ID,
			Name: emp.Department.Name,
		}
	}

	// Add manager info if exists
	if emp.Manager != nil {
		resp.Manager = &struct {
			ID       uuid.UUID `json:"id"`
			FullName string    `json:"full_name"`
		}{
			ID:       emp.Manager.ID,
			FullName: emp.Manager.FullName,
		}
	}

	return resp, nil
}

func (s *employeeServiceImpl) UpdateEmployee(id uuid.UUID, req *employee.UpdateEmployeeRequest) (*employee.UpdateEmployeeResponse, error) {
	emp, err := s.repo.FindByID(id)
	if err != nil {
		logger.Warn("UpdateEmployee failed: employee not found", zap.String("id", id.String()))
		return nil, response.NotFound("Employee not found")
	}

	// Update fields if provided
	if req.FirstName != nil {
		emp.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		emp.LastName = *req.LastName
	}
	if req.DateOfBirth != nil {
		dob, err := time.Parse("2006-01-02", *req.DateOfBirth)
		if err != nil {
			return nil, response.BadRequest("Invalid date of birth format. Use YYYY-MM-DD")
		}
		emp.DateOfBirth = dob
	}
	if req.Phone != nil {
		emp.Phone = *req.Phone
	}
	if req.Address != nil {
		emp.Address = *req.Address
	}
	if req.DepartmentID != nil {
		emp.DepartmentID = req.DepartmentID
	}
	if req.Position != nil {
		emp.Position = *req.Position
	}
	if req.ManagerID != nil {
		emp.ManagerID = req.ManagerID
	}
	if req.Status != nil {
		emp.Status = models.Status(*req.Status)
	}

	// Save updates
	if err := s.repo.Update(emp); err != nil {
		logger.Error("UpdateEmployee failed: database update error", zap.Error(err))
		return nil, response.InternalServerError("Failed to update employee")
	}

	logger.Info("Employee updated successfully", zap.String("employee_id", id.String()))

	return &employee.UpdateEmployeeResponse{
		Message: "Employee updated successfully",
		Employee: struct {
			ID           uuid.UUID  `json:"id"`
			Email        string     `json:"email"`
			FullName     string     `json:"full_name"`
			EmployeeCode string     `json:"employee_code"`
			Position     string     `json:"position"`
			DepartmentID *uuid.UUID `json:"department_id"`
			Status       string     `json:"status"`
		}{
			ID:           emp.ID,
			Email:        emp.Email,
			FullName:     emp.FullName,
			EmployeeCode: emp.EmployeeCode,
			Position:     emp.Position,
			DepartmentID: emp.DepartmentID,
			Status:       string(emp.Status),
		},
	}, nil
}

func (s *employeeServiceImpl) DeleteEmployee(id uuid.UUID) error {
	_, err := s.repo.FindByID(id)
	if err != nil {
		logger.Warn("DeleteEmployee failed: employee not found", zap.String("id", id.String()))
		return response.NotFound("Employee not found")
	}

	if err := s.repo.Delete(id); err != nil {
		logger.Error("DeleteEmployee failed: database delete error", zap.Error(err))
		return response.InternalServerError("Failed to delete employee")
	}

	logger.Info("Employee deleted successfully", zap.String("employee_id", id.String()))
	return nil
}

func (s *employeeServiceImpl) GetTodayBirthdays() (*employee.ListBirthdayResponse, error) {
	today := time.Now()
	month := int(today.Month())
	day := today.Day()

	employees, err := s.repo.FindEmployeesByBirthday(month, day)
	if err != nil {
		logger.Error("GetTodayBirthdays failed", zap.Error(err))
		return nil, response.InternalServerError("Failed to fetch birthdays")
	}

	summaries := make([]employee.BirthdaySummary, len(employees))
	for i, emp := range employees {
		departmentName := ""
		if emp.Department != nil {
			departmentName = emp.Department.Name
		}

		summaries[i] = employee.BirthdaySummary{
			ID:         emp.ID,
			FullName:   emp.FullName,
			Email:      emp.Email,
			Department: departmentName,
			Position:   emp.Position,
			BirthDate:  emp.DateOfBirth.Format("2006-01-02"),
		}
	}

	msg := "No birthdays today"
	if len(summaries) > 0 {
		msg = "Found birthdays today"
	}

	return &employee.ListBirthdayResponse{
		Employees: summaries,
		Total:     len(summaries),
		Message:   msg,
	}, nil
}

func (s *employeeServiceImpl) GetSubordinates(managerID uuid.UUID) (*employee.ListSubordinatesResponse, error) {
	// Find manager info first
	manager, err := s.repo.FindByID(managerID)
	if err != nil {
		logger.Warn("GetSubordinates failed: manager not found", zap.String("id", managerID.String()))
		return nil, response.NotFound("Manager not found")
	}

	subordinates, err := s.repo.FindSubordinates(managerID)
	if err != nil {
		logger.Error("GetSubordinates failed", zap.Error(err))
		return nil, response.InternalServerError("Failed to fetch subordinates")
	}

	summaryList := make([]employee.SubordinateSummary, len(subordinates))
	for i, sub := range subordinates {
		deptName := "N/A"
		if sub.Department != nil {
			deptName = sub.Department.Name
		}
		summaryList[i] = employee.SubordinateSummary{
			ID:           sub.ID,
			EmployeeCode: sub.EmployeeCode,
			FullName:     sub.FullName,
			Email:        sub.Email,
			Position:     sub.Position,
			Department:   deptName,
			Status:       string(sub.Status),
			AvatarUrl:    sub.AvatarUrl,
		}
	}

	return &employee.ListSubordinatesResponse{
		Manager: employee.SubordinateManagerRaw{
			ID:       manager.ID,
			FullName: manager.FullName,
			Position: manager.Position,
		},
		Subordinates: summaryList,
		Total:        len(summaryList),
	}, nil
}
