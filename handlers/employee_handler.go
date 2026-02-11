package handlers

import (
	"context"
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
		Summary:     "Delete employee (HR only)",
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
		OperationID: "get-subordinates",
		Method:      http.MethodGet,
		Path:        "/api/v1/employees/subordinates",
		Summary:     "Get subordinates for current manager",
		Tags:        []string{"Employees"},
		Security: []map[string][]string{
			{"bearerAuth": {}},
		},
	}, h.GetSubordinates)
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

	// Check HR role
	err = authPkg.CheckRole(claims.UserID, h.employeeRepo, "hr", "admin")
	if err != nil {
		return huma.Error403Forbidden("Insufficient permissions. HR or Admin role required.")
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
	if err := h.validateHRAccess(input.Authorization); err != nil {
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
	if err := h.validateHRAccess(input.Authorization); err != nil {
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
	if err := h.validateHRAccess(input.Authorization); err != nil {
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
	if err := h.validateHRAccess(input.Authorization); err != nil {
		return nil, err
	}

	uid, err := uuid.Parse(input.ID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid UUID format", err)
	}

	if err := h.service.DeleteEmployee(uid); err != nil {
		return nil, huma.Error500InternalServerError("Failed to delete employee", err)
	}

	return &struct {
		Body struct {
			Message string `json:"message"`
		}
	}{
		Body: struct {
			Message string `json:"message"`
		}{
			Message: "Employee deleted successfully",
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
