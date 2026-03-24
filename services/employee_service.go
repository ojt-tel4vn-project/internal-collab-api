package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ojt-tel4vn-project/internal-collab-api/dtos/employee"
	"github.com/ojt-tel4vn-project/internal-collab-api/internal/storage"
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
	GetAllEmployees(requesterRole string) (*employee.ListEmployeesResponse, error)
	GetEmployeesByStatus(status models.Status) (*employee.ListEmployeesResponse, error)
	ApproveEmployee(id uuid.UUID) error
	GetEmployeeByID(id uuid.UUID) (*employee.GetEmployeeResponse, error)
	GetProfile(id uuid.UUID) (*employee.GetEmployeeResponse, error)
	UpdateEmployee(id uuid.UUID, req *employee.UpdateEmployeeRequest) (*employee.UpdateEmployeeResponse, error)
	UpdateProfile(id uuid.UUID, req *employee.UpdateProfileRequest) (*employee.UpdateProfileResponse, error)
	UploadAvatar(ctx context.Context, employeeID uuid.UUID, file io.Reader, filename string) (string, error)
	DeleteEmployee(id uuid.UUID) error
	GetTodayBirthdays() (*employee.ListBirthdayResponse, error)
	GetAllBirthdays() (*employee.ListAllBirthdaysResponse, error)
	GetSubordinates(managerID uuid.UUID) (*employee.ListSubordinatesResponse, error)
	GetBirthdayConfig() (*employee.GetBirthdayConfigResponse, error)
	UpdateBirthdayConfig(req *employee.UpdateBirthdayConfigRequest) (*employee.UpdateBirthdayConfigResponse, error)
	SearchEmployees(query string) (*employee.SearchEmployeeResponse, error)
}

type employeeServiceImpl struct {
	repo         repository.EmployeeRepository
	password     crypto.PasswordService
	emailService email.EmailService
	appConfig    repository.AppConfigRepository
	storage      *storage.SupabaseStorage
}

func NewEmployeeService(repo repository.EmployeeRepository, password crypto.PasswordService, emailService email.EmailService, appConfig repository.AppConfigRepository, stor *storage.SupabaseStorage) EmployeeService {
	return &employeeServiceImpl{
		repo:         repo,
		password:     password,
		emailService: emailService,
		appConfig:    appConfig,
		storage:      stor,
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

	// If no role is specified, assign default "employee" role
	roleID := req.RoleID
	if roleID == nil {
		// Find default "employee" role
		defaultRole, err := s.repo.FindRoleByName("employee")
		if err != nil {
			logger.Warn("CreateEmployee: default 'employee' role not found, proceeding without role", zap.Error(err))
		} else {
			roleID = &defaultRole.ID
			logger.Info("CreateEmployee: assigned default 'employee' role", zap.String("role_id", defaultRole.ID.String()))
		}
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
		RoleID:       roleID,
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

func (s *employeeServiceImpl) GetAllEmployees(requesterRole string) (*employee.ListEmployeesResponse, error) {
	var employees []models.Employee
	var err error

	// Admin sees all, others see only active
	if requesterRole == "admin" {
		employees, err = s.repo.FindAll()
	} else {
		employees, err = s.repo.FindByStatus(models.StatusActive)
	}

	if err != nil {
		logger.Error("GetAllEmployees failed", zap.Error(err))
		return nil, response.InternalServerError("Failed to fetch employees")
	}

	return s.convertToListResponse(employees), nil
}

func (s *employeeServiceImpl) GetEmployeesByStatus(status models.Status) (*employee.ListEmployeesResponse, error) {
	employees, err := s.repo.FindByStatus(status)
	if err != nil {
		logger.Error("GetEmployeesByStatus failed", zap.Error(err), zap.String("status", string(status)))
		return nil, response.InternalServerError("Failed to fetch employees")
	}

	return s.convertToListResponse(employees), nil
}

func (s *employeeServiceImpl) ApproveEmployee(id uuid.UUID) error {
	emp, err := s.repo.FindByID(id)
	if err != nil {
		logger.Warn("ApproveEmployee failed: employee not found", zap.String("id", id.String()))
		return response.NotFound("Employee not found")
	}

	if emp.Status != models.StatusPending {
		return response.BadRequest("Employee is not in pending status")
	}

	emp.Status = models.StatusActive
	if err := s.repo.Update(emp); err != nil {
		logger.Error("Failed to approve employee", zap.Error(err), zap.String("id", id.String()))
		return response.InternalServerError("Failed to approve employee")
	}

	// TODO: Send welcome email
	logger.Info("Employee approved", zap.String("id", id.String()), zap.String("email", emp.Email))
	return nil
}

// Helper function to convert employees to list response
func (s *employeeServiceImpl) convertToListResponse(employees []models.Employee) *employee.ListEmployeesResponse {
	summaries := make([]employee.EmployeeSummary, len(employees))
	for i, emp := range employees {
		var deptBrief *employee.DepartmentBrief
		if emp.Department != nil {
			deptBrief = &employee.DepartmentBrief{
				ID:   emp.Department.ID,
				Name: emp.Department.Name,
			}
		}

		var roleBrief *employee.RoleBrief
		if emp.Role != nil {
			roleBrief = &employee.RoleBrief{
				ID:   emp.Role.ID,
				Name: emp.Role.Name,
			}
		}

		summaries[i] = employee.EmployeeSummary{
			ID:           emp.ID,
			Email:        emp.Email,
			FullName:     emp.FullName,
			EmployeeCode: emp.EmployeeCode,
			Position:     emp.Position,
			Department:   deptBrief,
			Role:         roleBrief,
			AvatarUrl:    emp.AvatarUrl,
			Status:       string(emp.Status),
		}
	}

	return &employee.ListEmployeesResponse{
		Employees: summaries,
		Total:     len(summaries),
	}
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
		AvatarUrl:    emp.AvatarUrl,
		DepartmentID: emp.DepartmentID,
		Position:     emp.Position,
		ManagerID:    emp.ManagerID,
		RoleID:       emp.RoleID,
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

	// Add role info if exists
	if emp.Role != nil {
		resp.Role = &employee.RoleBrief{
			ID:   emp.Role.ID,
			Name: emp.Role.Name,
		}
	}

	return resp, nil
}

func (s *employeeServiceImpl) GetProfile(id uuid.UUID) (*employee.GetEmployeeResponse, error) {
	// Re-using GetEmployeeByID
	return s.GetEmployeeByID(id)
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
	if req.RoleID != nil {
		emp.RoleID = req.RoleID
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

func (s *employeeServiceImpl) UpdateProfile(id uuid.UUID, req *employee.UpdateProfileRequest) (*employee.UpdateProfileResponse, error) {
	emp, err := s.repo.FindByID(id)
	if err != nil {
		logger.Warn("UpdateProfile failed: employee not found", zap.String("id", id.String()))
		return nil, response.NotFound("Employee not found")
	}

	if req.Phone != nil {
		emp.Phone = *req.Phone
	}
	if req.Address != nil {
		emp.Address = *req.Address
	}
	if req.AvatarUrl != nil {
		emp.AvatarUrl = *req.AvatarUrl
	}

	if err := s.repo.Update(emp); err != nil {
		logger.Error("UpdateProfile failed: database update error", zap.Error(err))
		return nil, response.InternalServerError("Failed to update profile")
	}

	logger.Info("Profile updated successfully", zap.String("employee_id", id.String()))

	return &employee.UpdateProfileResponse{
		Message: "Profile updated successfully",
	}, nil
}

// UploadAvatar uploads an avatar image for the employee and stores the URL
func (s *employeeServiceImpl) UploadAvatar(ctx context.Context, employeeID uuid.UUID, file io.Reader, filename string) (string, error) {
	emp, err := s.repo.FindByID(employeeID)
	if err != nil {
		return "", response.NotFound("Employee not found")
	}

	// Validate extension
	ext := strings.ToLower(filepath.Ext(filename))
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}
	if !allowed[ext] {
		return "", response.BadRequest("Only image files are allowed (jpg, jpeg, png, gif, webp)")
	}

	// Upload to Supabase Storage under avatars/
	path := "avatars/" + employeeID.String() + ext
	publicURL, err := s.storage.UploadFile(ctx, path, file)
	if err != nil {
		logger.Error("UploadAvatar: storage upload failed", zap.Error(err))
		return "", response.InternalServerError("Failed to upload avatar")
	}

	// Update avatar_url in DB
	emp.AvatarUrl = publicURL
	if err := s.repo.Update(emp); err != nil {
		logger.Error("UploadAvatar: DB update failed", zap.Error(err))
		return "", response.InternalServerError("Failed to save avatar URL")
	}

	logger.Info("Avatar uploaded", zap.String("employee_id", employeeID.String()), zap.String("url", publicURL))
	return publicURL, nil
}

func (s *employeeServiceImpl) DeleteEmployee(id uuid.UUID) error {
	emp, err := s.repo.FindByID(id)
	if err != nil {
		logger.Warn("DeleteEmployee failed: employee not found", zap.String("id", id.String()))
		return response.NotFound("Employee not found")
	}

	// Soft delete: set status to offboard instead of hard delete
	emp.Status = models.StatusOffBoard
	now := time.Now()
	emp.LeaveDate = &now

	if err := s.repo.Update(emp); err != nil {
		logger.Error("DeleteEmployee failed: database update error", zap.Error(err))
		return response.InternalServerError("Failed to offboard employee")
	}

	logger.Info("Employee offboarded successfully", zap.String("employee_id", id.String()))
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

func (s *employeeServiceImpl) GetAllBirthdays() (*employee.ListAllBirthdaysResponse, error) {
	employees, err := s.repo.FindAllBirthdays()
	if err != nil {
		logger.Error("GetAllBirthdays failed", zap.Error(err))
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

	return &employee.ListAllBirthdaysResponse{
		Employees: summaries,
		Total:     len(summaries),
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

const birthdayConfigKey = "birthday_config"

var defaultBirthdayConfig = employee.BirthdayConfig{
	Enabled:          true,
	NotificationTime: "09:00",
	Channels:         []string{"in_app", "email"},
}

func (s *employeeServiceImpl) GetBirthdayConfig() (*employee.GetBirthdayConfigResponse, error) {
	raw, err := s.appConfig.Get(birthdayConfigKey)
	if err != nil {
		// Key not found yet — return default without error
		return &employee.GetBirthdayConfigResponse{Data: defaultBirthdayConfig}, nil
	}

	var cfg employee.BirthdayConfig
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		logger.Warn("Failed to parse birthday config from DB, using default", zap.Error(err))
		return &employee.GetBirthdayConfigResponse{Data: defaultBirthdayConfig}, nil
	}
	return &employee.GetBirthdayConfigResponse{Data: cfg}, nil
}

func (s *employeeServiceImpl) UpdateBirthdayConfig(req *employee.UpdateBirthdayConfigRequest) (*employee.UpdateBirthdayConfigResponse, error) {
	cfg := employee.BirthdayConfig{
		Enabled:          req.Enabled,
		NotificationTime: req.NotificationTime,
		Channels:         req.Channels,
	}

	raw, err := json.Marshal(cfg)
	if err != nil {
		logger.Error("UpdateBirthdayConfig failed: marshal error", zap.Error(err))
		return nil, response.InternalServerError("Failed to save birthday config")
	}

	if err := s.appConfig.Set(birthdayConfigKey, string(raw)); err != nil {
		logger.Error("UpdateBirthdayConfig failed: DB save error", zap.Error(err))
		return nil, response.InternalServerError("Failed to save birthday config")
	}

	return &employee.UpdateBirthdayConfigResponse{
		Message: "Birthday configuration updated successfully",
		Data:    cfg,
	}, nil
}

func (s *employeeServiceImpl) SearchEmployees(query string) (*employee.SearchEmployeeResponse, error) {
	if query == "" {
		logger.Warn("SearchEmployees failed: empty query")
		return nil, response.BadRequest("Search query cannot be empty")
	}

	employees, err := s.repo.SearchEmployees(query)
	if err != nil {
		logger.Error("SearchEmployees failed: database error", zap.Error(err))
		return nil, response.InternalServerError("Failed to search employees")
	}

	searchResults := make([]employee.SearchEmployeeSummary, len(employees))
	for i, emp := range employees {
		searchResults[i] = employee.SearchEmployeeSummary{
			ID:           emp.ID,
			Email:        emp.Email,
			FullName:     emp.FullName,
			EmployeeCode: emp.EmployeeCode,
			Phone:        emp.Phone,
			Position:     emp.Position,
			AvatarUrl:    emp.AvatarUrl,
			Status:       string(emp.Status),
			JoinDate:     emp.JoinDate,
		}

		if emp.Department != nil {
			searchResults[i].Department = &employee.DepartmentBrief{
				ID:   emp.Department.ID,
				Name: emp.Department.Name,
			}
		}
	}

	return &employee.SearchEmployeeResponse{
		Employees: searchResults,
		Total:     len(searchResults),
	}, nil
}
