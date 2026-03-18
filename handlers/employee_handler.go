package handlers

import (
	"context"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/ojt-tel4vn-project/internal-collab-api/dtos/employee"
	authPkg "github.com/ojt-tel4vn-project/internal-collab-api/pkg/auth"
	"github.com/ojt-tel4vn-project/internal-collab-api/pkg/crypto"
	"github.com/ojt-tel4vn-project/internal-collab-api/repository"
	"github.com/ojt-tel4vn-project/internal-collab-api/services"
)

type EmployeeHandler struct {
	service      services.EmployeeService
	jwtService   crypto.JWTService
	employeeRepo repository.EmployeeRepository
}

func NewEmployeeHandler(service services.EmployeeService, jwtService crypto.JWTService, employeeRepo repository.EmployeeRepository) *EmployeeHandler {
	return &EmployeeHandler{
		service:      service,
		jwtService:   jwtService,
		employeeRepo: employeeRepo,
	}
}

func (h *EmployeeHandler) RegisterRoutes(api huma.API) {
	// HR endpoints (require authentication + HR role)
	huma.Register(api, huma.Operation{
		OperationID: "create-employee",
		Method:      http.MethodPost,
		Path:        "/api/v1/hr/employees",
		Summary:     "Create an employee (HR only)",
		Tags:        []string{"Employees"},
		Security: []map[string][]string{
			{"bearerAuth": {}},
		},
	}, h.CreateEmployee)

	huma.Register(api, huma.Operation{
		OperationID: "get-employees",
		Method:      http.MethodGet,
		Path:        "/api/v1/hr/employees",
		Summary:     "Get all employees (HR only)",
		Tags:        []string{"Employees"},
		Security: []map[string][]string{
			{"bearerAuth": {}},
		},
	}, h.GetAllEmployees)

	huma.Register(api, huma.Operation{
		OperationID: "get-employee-by-id",
		Method:      http.MethodGet,
		Path:        "/api/v1/hr/employees/{id}",
		Summary:     "Get employee by ID (HR only)",
		Tags:        []string{"Employees"},
		Security: []map[string][]string{
			{"bearerAuth": {}},
		},
	}, h.GetEmployeeByID)

	huma.Register(api, huma.Operation{
		OperationID: "update-employee",
		Method:      http.MethodPut,
		Path:        "/api/v1/hr/employees/{id}",
		Summary:     "Update employee (HR only)",
		Tags:        []string{"Employees"},
		Security: []map[string][]string{
			{"bearerAuth": {}},
		},
	}, h.UpdateEmployee)

	huma.Register(api, huma.Operation{
		OperationID: "delete-employee",
		Method:      http.MethodDelete,
		Path:        "/api/v1/hr/employees/{id}",
		Summary:     "Offboard employee (HR only)",
		Description: "Soft delete: Sets employee status to 'offboard' and records leave date. Employee will no longer appear in active lists.",
		Tags:        []string{"Employees"},
		Security: []map[string][]string{
			{"bearerAuth": {}},
		},
	}, h.DeleteEmployee)

	huma.Register(api, huma.Operation{
		OperationID: "get-today-birthdays",
		Method:      http.MethodGet,
		Path:        "/api/v1/hr/employees/birthdays/today",
		Summary:     "Get employees with birthdays today (HR/Manager)",
		Tags:        []string{"Employees"},
		Security: []map[string][]string{
			{"bearerAuth": {}},
		},
	}, h.GetTodayBirthdays)

	huma.Register(api, huma.Operation{
		OperationID: "get-all-birthdays",
		Method:      http.MethodGet,
		Path:        "/api/v1/employees/birthdays",
		Summary:     "Get all employee birthdays (for calendar)",
		Tags:        []string{"Employees"},
		Security: []map[string][]string{
			{"bearerAuth": {}},
		},
	}, h.GetAllBirthdays)

	huma.Register(api, huma.Operation{
		OperationID: "get-birthday-config",
		Method:      http.MethodGet,
		Path:        "/api/v1/employees/birthdays/config",
		Summary:     "Get Birthday Notification Config",
		Tags:        []string{"Employees"},
		Security: []map[string][]string{
			{"bearerAuth": {}},
		},
	}, h.GetBirthdayConfig)

	huma.Register(api, huma.Operation{
		OperationID: "update-birthday-config",
		Method:      http.MethodPut,
		Path:        "/api/v1/employees/birthdays/config",
		Summary:     "Update Birthday Notification Config (HR only)",
		Tags:        []string{"Employees"},
		Security: []map[string][]string{
			{"bearerAuth": {}},
		},
	}, h.UpdateBirthdayConfig)

	huma.Register(api, huma.Operation{
		OperationID: "get-subordinates",
		Method:      http.MethodGet,
		Path:        "/api/v1/employees/subordinates",
		Summary:     "Get subordinates for current manager",
		Tags:        []string{"Employees"},
		Security: []map[string][]string{
			{"bearerAuth": {}},
		},
	}, h.GetSubordinates)

	// Self-service profile endpoints
	huma.Register(api, huma.Operation{
		OperationID: "get-profile",
		Method:      http.MethodGet,
		Path:        "/api/v1/employees/me",
		Summary:     "Get current user profile",
		Tags:        []string{"Employees"},
		Security: []map[string][]string{
			{"bearerAuth": {}},
		},
	}, h.GetProfile)

	huma.Register(api, huma.Operation{
		OperationID: "update-profile",
		Method:      http.MethodPut,
		Path:        "/api/v1/employees/me",
		Summary:     "Update current user profile",
		Tags:        []string{"Employees"},
		Security: []map[string][]string{
			{"bearerAuth": {}},
		},
	}, h.UpdateProfile)

	// Upload avatar
	huma.Register(api, huma.Operation{
		OperationID: "upload-avatar",
		Method:      http.MethodPost,
		Path:        "/api/v1/employees/me/avatar",
		Summary:     "Upload profile avatar (multipart/form-data, field: avatar)",
		Tags:        []string{"Employees"},
		Security: []map[string][]string{
			{"bearerAuth": {}},
		},
	}, h.UploadAvatar)
}

// validateHROrManagerAccess validates JWT and checks for HR or Manager role
func (h *EmployeeHandler) validateHROrManagerAccess(authHeader string) error {
	// Validate JWT
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return huma.Error401Unauthorized("Invalid authorization format")
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := h.jwtService.ValidateToken(token)
	if err != nil {
		return huma.Error401Unauthorized("Invalid or expired token")
	}

	// Check HR or Manager role
	err = authPkg.CheckRole(claims.UserID, h.employeeRepo, "hr", "manager", "admin")
	if err != nil {
		return huma.Error403Forbidden("Insufficient permissions. HR, Manager or Admin role required.")
	}

	return nil
}

// validateHRAccess validates JWT and checks for HR role
func (h *EmployeeHandler) validateHRAccess(authHeader string) error {
	// Validate JWT
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return huma.Error401Unauthorized("Invalid authorization format")
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := h.jwtService.ValidateToken(token)
	if err != nil {
		return huma.Error401Unauthorized("Invalid or expired token")
	}

	// Check HR or Manager role
	err = authPkg.CheckRole(claims.UserID, h.employeeRepo, "hr", "manager", "admin")
	if err != nil {
		return huma.Error403Forbidden("Insufficient permissions. HR, Manager or Admin role required.")
	}

	return nil
}

func (h *EmployeeHandler) CreateEmployee(ctx context.Context, input *struct {
	Authorization string `header:"Authorization" required:"true" doc:"Bearer token"`
	Body          employee.CreateEmployeeRequest
}) (*struct {
	Body employee.CreateEmployeeResponse
}, error) {
	// Validate HR access
	if err := h.validateHRAccess(input.Authorization); err != nil {
		return nil, err
	}

	resp, err := h.service.CreateEmployee(&input.Body)
	if err != nil {
		return nil, huma.Error400BadRequest("Failed to create employee", err)
	}
	return &struct {
		Body employee.CreateEmployeeResponse
	}{Body: *resp}, nil
}

func (h *EmployeeHandler) GetAllEmployees(ctx context.Context, input *struct {
	Authorization string `header:"Authorization" required:"true" doc:"Bearer token"`
}) (*struct {
	Body employee.ListEmployeesResponse
}, error) {
	// Validate HR access
	_, err := authPkg.Authorize(
		input.Authorization,
		h.jwtService,
		h.employeeRepo,
		authPkg.AuthOptions{
			Roles: []string{"hr", "admin"},
		},
	)
	if err != nil {
		return nil, err
	}

	resp, err := h.service.GetAllEmployees()
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to fetch employees", err)
	}
	return &struct {
		Body employee.ListEmployeesResponse
	}{Body: *resp}, nil
}

func (h *EmployeeHandler) GetEmployeeByID(ctx context.Context, input *struct {
	Authorization string `header:"Authorization" required:"true" doc:"Bearer token"`
	ID            string `path:"id"`
}) (*struct {
	Body employee.GetEmployeeResponse
}, error) {
	// Validate HR access
	_, err := authPkg.Authorize(
		input.Authorization,
		h.jwtService,
		h.employeeRepo,
		authPkg.AuthOptions{
			Roles: []string{"hr", "admin"},
		},
	)
	if err != nil {
		return nil, err
	}

	uid, err := uuid.Parse(input.ID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid UUID format", err)
	}

	resp, err := h.service.GetEmployeeByID(uid)
	if err != nil {
		return nil, huma.Error404NotFound("Employee not found", err)
	}

	return &struct{ Body employee.GetEmployeeResponse }{Body: *resp}, nil
}

func (h *EmployeeHandler) UpdateEmployee(ctx context.Context, input *struct {
	Authorization string `header:"Authorization" required:"true" doc:"Bearer token"`
	ID            string `path:"id"`
	Body          employee.UpdateEmployeeRequest
}) (*struct {
	Body employee.UpdateEmployeeResponse
}, error) {
	// Validate HR access
	_, err := authPkg.Authorize(
		input.Authorization,
		h.jwtService,
		h.employeeRepo,
		authPkg.AuthOptions{
			Roles: []string{"hr", "admin"},
		},
	)
	if err != nil {
		return nil, err
	}

	uid, err := uuid.Parse(input.ID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid UUID format", err)
	}

	resp, err := h.service.UpdateEmployee(uid, &input.Body)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to update employee", err)
	}

	return &struct {
		Body employee.UpdateEmployeeResponse
	}{Body: *resp}, nil
}

func (h *EmployeeHandler) DeleteEmployee(ctx context.Context, input *struct {
	Authorization string `header:"Authorization" required:"true" doc:"Bearer token"`
	ID            string `path:"id"`
}) (*struct {
	Body struct {
		Message string `json:"message"`
	}
}, error) {
	// Validate HR access
	_, err := authPkg.Authorize(
		input.Authorization,
		h.jwtService,
		h.employeeRepo,
		authPkg.AuthOptions{
			Roles: []string{"hr", "admin"},
		},
	)
	if err != nil {
		return nil, err
	}

	uid, err := uuid.Parse(input.ID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid UUID format", err)
	}

	if err := h.service.DeleteEmployee(uid); err != nil {
		return nil, huma.Error500InternalServerError("Failed to offboard employee", err)
	}

	return &struct {
		Body struct {
			Message string `json:"message"`
		}
	}{
		Body: struct {
			Message string `json:"message"`
		}{
			Message: "Employee offboarded successfully",
		},
	}, nil
}

func (h *EmployeeHandler) GetTodayBirthdays(ctx context.Context, input *struct {
	Authorization string `header:"Authorization" required:"true" doc:"Bearer token"`
}) (*struct {
	Body employee.ListBirthdayResponse
}, error) {
	// Validate HR/Manager access
	if err := h.validateHROrManagerAccess(input.Authorization); err != nil {
		return nil, err
	}

	resp, err := h.service.GetTodayBirthdays()
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to fetch birthdays", err)
	}

	return &struct {
		Body employee.ListBirthdayResponse
	}{Body: *resp}, nil
}

// GetAllBirthdays GET /api/v1/employees/birthdays — any authenticated employee
func (h *EmployeeHandler) GetAllBirthdays(ctx context.Context, input *struct {
	Authorization string `header:"Authorization" required:"true" doc:"Bearer token"`
}) (*struct {
	Body employee.ListAllBirthdaysResponse
}, error) {
	if !strings.HasPrefix(input.Authorization, "Bearer ") {
		return nil, huma.Error401Unauthorized("Invalid authorization format")
	}
	_, err := h.jwtService.ValidateToken(strings.TrimPrefix(input.Authorization, "Bearer "))
	if err != nil {
		return nil, huma.Error401Unauthorized("Invalid or expired token")
	}

	resp, err := h.service.GetAllBirthdays()
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to fetch birthdays", err)
	}

	return &struct {
		Body employee.ListAllBirthdaysResponse
	}{Body: *resp}, nil
}

func (h *EmployeeHandler) GetSubordinates(ctx context.Context, input *struct {
	Authorization string `header:"Authorization" required:"true" doc:"Bearer token"`
}) (*struct {
	Body employee.ListSubordinatesResponse
}, error) {
	// Validate JWT
	if !strings.HasPrefix(input.Authorization, "Bearer ") {
		return nil, huma.Error401Unauthorized("Invalid authorization format")
	}

	token := strings.TrimPrefix(input.Authorization, "Bearer ")
	claims, err := h.jwtService.ValidateToken(token)
	if err != nil {
		return nil, huma.Error401Unauthorized("Invalid or expired token")
	}

	// Fetch subordinates using the ID from the token (the manager's ID)
	resp, err := h.service.GetSubordinates(claims.UserID)
	if err != nil {
		if strings.Contains(err.Error(), "Manager not found") {
			return nil, huma.Error404NotFound("Manager not found", err)
		}
		return nil, huma.Error500InternalServerError("Failed to fetch subordinates", err)
	}

	return &struct {
		Body employee.ListSubordinatesResponse
	}{Body: *resp}, nil
}

func (h *EmployeeHandler) GetProfile(ctx context.Context, input *struct {
	Authorization string `header:"Authorization" required:"true" doc:"Bearer token"`
}) (*struct {
	Body employee.GetEmployeeResponse
}, error) {
	if !strings.HasPrefix(input.Authorization, "Bearer ") {
		return nil, huma.Error401Unauthorized("Invalid authorization format")
	}

	token := strings.TrimPrefix(input.Authorization, "Bearer ")
	claims, err := h.jwtService.ValidateToken(token)
	if err != nil {
		return nil, huma.Error401Unauthorized("Invalid or expired token")
	}

	resp, err := h.service.GetProfile(claims.UserID)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to fetch profile", err)
	}

	return &struct {
		Body employee.GetEmployeeResponse
	}{Body: *resp}, nil
}

func (h *EmployeeHandler) UpdateProfile(ctx context.Context, input *struct {
	Authorization string `header:"Authorization" required:"true" doc:"Bearer token"`
	Body          employee.UpdateProfileRequest
}) (*struct {
	Body employee.UpdateProfileResponse
}, error) {
	if !strings.HasPrefix(input.Authorization, "Bearer ") {
		return nil, huma.Error401Unauthorized("Invalid authorization format")
	}

	token := strings.TrimPrefix(input.Authorization, "Bearer ")
	claims, err := h.jwtService.ValidateToken(token)
	if err != nil {
		return nil, huma.Error401Unauthorized("Invalid or expired token")
	}

	resp, err := h.service.UpdateProfile(claims.UserID, &input.Body)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to update profile", err)
	}

	return &struct {
		Body employee.UpdateProfileResponse
	}{Body: *resp}, nil
}

func (h *EmployeeHandler) GetBirthdayConfig(ctx context.Context, input *struct {
	Authorization string `header:"Authorization" required:"true" doc:"Bearer token"`
}) (*struct {
	Body employee.GetBirthdayConfigResponse
}, error) {
	if !strings.HasPrefix(input.Authorization, "Bearer ") {
		return nil, huma.Error401Unauthorized("Invalid authorization format")
	}

	_, err := h.jwtService.ValidateToken(strings.TrimPrefix(input.Authorization, "Bearer "))
	if err != nil {
		return nil, huma.Error401Unauthorized("Invalid or expired token")
	}

	resp, err := h.service.GetBirthdayConfig()
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to fetch birthday config", err)
	}

	return &struct {
		Body employee.GetBirthdayConfigResponse
	}{Body: *resp}, nil
}

func (h *EmployeeHandler) UpdateBirthdayConfig(ctx context.Context, input *struct {
	Authorization string `header:"Authorization" required:"true" doc:"Bearer token"`
	Body          employee.UpdateBirthdayConfigRequest
}) (*struct {
	Body employee.UpdateBirthdayConfigResponse
}, error) {
	// Validate HR role since only HR should config this
	if err := h.validateHRAccess(input.Authorization); err != nil {
		return nil, err
	}

	resp, err := h.service.UpdateBirthdayConfig(&input.Body)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to update birthday config", err)
	}

	return &struct {
		Body employee.UpdateBirthdayConfigResponse
	}{Body: *resp}, nil
}

// UploadAvatar handles POST /api/v1/employees/me/avatar.
// The multipart form can use any file field name (avatar, file, filename, image, photo).
func (h *EmployeeHandler) UploadAvatar(ctx context.Context, input *struct {
	Authorization string `header:"Authorization" required:"true" doc:"Bearer token"`
	RawBody       multipart.Form
}) (*struct {
	Body struct {
		Message   string `json:"message"`
		AvatarUrl string `json:"avatar_url"`
	}
}, error) {
	if !strings.HasPrefix(input.Authorization, "Bearer ") {
		return nil, huma.Error401Unauthorized("Invalid authorization format")
	}

	token := strings.TrimPrefix(input.Authorization, "Bearer ")
	claims, err := h.jwtService.ValidateToken(token)
	if err != nil {
		return nil, huma.Error401Unauthorized("Invalid or expired token")
	}

	// Accept any file field name (Huma UI sends as "filename"; curl users may use "avatar", "file", etc.)
	var fh *multipart.FileHeader
	for _, headers := range input.RawBody.File {
		if len(headers) > 0 {
			fh = headers[0]
			break
		}
	}

	if fh == nil {
		return nil, huma.Error400BadRequest(
			"No file found. Send the image as a multipart/form-data field (any field name, e.g. 'avatar').",
		)
	}

	// Max size: 5 MB
	if fh.Size > 5<<20 {
		return nil, huma.Error400BadRequest("Avatar file must be under 5 MB")
	}

	f, err := fh.Open()
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to read avatar file")
	}
	defer f.Close()

	avatarURL, err := h.service.UploadAvatar(ctx, claims.UserID, f, fh.Filename)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to upload avatar", err)
	}

	return &struct {
		Body struct {
			Message   string `json:"message"`
			AvatarUrl string `json:"avatar_url"`
		}
	}{Body: struct {
		Message   string `json:"message"`
		AvatarUrl string `json:"avatar_url"`
	}{
		Message:   "Avatar uploaded successfully",
		AvatarUrl: avatarURL,
	}}, nil
}
