package handlers

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/ojt-tel4vn-project/internal-collab-api/dtos/department"
	"github.com/ojt-tel4vn-project/internal-collab-api/middleware"
	"github.com/ojt-tel4vn-project/internal-collab-api/pkg/crypto"
	"github.com/ojt-tel4vn-project/internal-collab-api/repository"
	"github.com/ojt-tel4vn-project/internal-collab-api/services"
)

type DepartmentHandler struct {
	service      services.DepartmentService
	jwtService   crypto.JWTService
	employeeRepo repository.EmployeeRepository
}

func NewDepartmentHandler(service services.DepartmentService, jwtService crypto.JWTService, employeeRepo repository.EmployeeRepository) *DepartmentHandler {
	return &DepartmentHandler{
		service:      service,
		jwtService:   jwtService,
		employeeRepo: employeeRepo,
	}
}

func (h *DepartmentHandler) RegisterRoutes(api huma.API) {
	// List Departments (Public/All authenticated users)
	huma.Register(api, huma.Operation{
		OperationID:   "get-departments",
		Method:        http.MethodGet,
		Path:          "/api/v1/departments",
		Summary:       "List All Departments",
		Tags:          []string{"Departments"},
		Security:      []map[string][]string{{"bearerAuth": {}}},
	}, h.handleGetDepartments)

	// Create Department (HR/Admin only)
	huma.Register(api, huma.Operation{
		OperationID:   "create-department",
		Method:        http.MethodPost,
		Path:          "/api/v1/hr/departments",
		Summary:       "Create a Department (HR only)",
		Tags:          []string{"Departments"},
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: 201,
	}, h.handleCreateDepartment)
}

func (h *DepartmentHandler) handleGetDepartments(ctx context.Context, input *struct {
	Authorization string `header:"Authorization" required:"true"`
}) (*struct {
	Body department.ListDepartmentResponse
}, error) {
	_, err := middleware.ValidateJWTFromHeader(input.Authorization, h.jwtService)
	if err != nil {
		return nil, huma.Error401Unauthorized("Missing or invalid token")
	}

	depts, err := h.service.GetAllDepartments()
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to fetch departments", err)
	}

	res := department.ListDepartmentResponse{Data: make([]department.DepartmentResponse, len(depts))}
	for i, d := range depts {
		res.Data[i] = department.DepartmentResponse{
			ID:          d.ID,
			Name:        d.Name,
			Description: d.Description,
			CreatedAt:   d.CreatedAt,
			UpdatedAt:   d.UpdatedAt,
		}
	}

	return &struct{ Body department.ListDepartmentResponse }{Body: res}, nil
}

func (h *DepartmentHandler) handleCreateDepartment(ctx context.Context, input *struct {
	Authorization string                             `header:"Authorization" required:"true"`
	Body          department.CreateDepartmentRequest `json:"body"`
}) (*struct {
	Body department.DepartmentResponse
}, error) {
	claims, err := middleware.ValidateJWTFromHeader(input.Authorization, h.jwtService)
	if err != nil {
		return nil, huma.Error401Unauthorized("Missing or invalid token")
	}

	// Only HR or Admin can create departments
	if roleErr := middleware.CheckUserRole(claims.UserID, h.employeeRepo, "hr", "admin"); roleErr != nil {
		return nil, huma.Error403Forbidden("Only HR or Admin can create departments")
	}

	dept, err := h.service.CreateDepartment(input.Body.Name, input.Body.Description)
	if err != nil {
		if err == services.ErrDepartmentExists {
			return nil, huma.Error409Conflict(err.Error())
		}
		return nil, huma.Error500InternalServerError("Failed to create department", err)
	}

	res := department.DepartmentResponse{
		ID:          dept.ID,
		Name:        dept.Name,
		Description: dept.Description,
		CreatedAt:   dept.CreatedAt,
		UpdatedAt:   dept.UpdatedAt,
	}

	return &struct{ Body department.DepartmentResponse }{Body: res}, nil
}
