package routes

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/ojt-tel4vn-project/internal-collab-api/handlers"
	"github.com/ojt-tel4vn-project/internal-collab-api/pkg/crypto"
	"github.com/ojt-tel4vn-project/internal-collab-api/repository"
	"github.com/ojt-tel4vn-project/internal-collab-api/services"
)

func SetupRoutes(
	api huma.API,
	todoService services.TodoService,
	authService services.AuthService,
	employeeService services.EmployeeService,
	jwtService crypto.JWTService,
	employeeRepo repository.EmployeeRepository,
) {
	// Todo Routes
	todoHandler := handlers.NewTodoHandler(todoService)
	todoHandler.RegisterRoutes(api)

	// Auth Routes (with JWT service)
	authHandler := handlers.NewAuthHandler(authService, jwtService)
	authHandler.RegisterRoutes(api)

	// Employee Routes (with JWT service and employee repo for role checking)
	employeeHandler := handlers.NewEmployeeHandler(employeeService, jwtService, employeeRepo)
	employeeHandler.RegisterRoutes(api)
}
