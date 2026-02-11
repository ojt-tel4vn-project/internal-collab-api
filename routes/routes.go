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
	authService services.AuthService,
	employeeService services.EmployeeService,
	auditLogService services.AuditLogService,
	notificationService services.NotificationService,
	jwtService crypto.JWTService,
	employeeRepo repository.EmployeeRepository,
	documentService services.DocumentService,
	categoryService services.DocumentCategoryService,
) {
	// Auth Routes (with JWT service)
	authHandler := handlers.NewAuthHandler(authService, jwtService)
	authHandler.RegisterRoutes(api)

	// Employee Routes (with JWT service and employee repo for role checking)
	employeeHandler := handlers.NewEmployeeHandler(employeeService, jwtService, employeeRepo)

	employeeHandler.RegisterRoutes(api)

	// Document Routes (with JWT service and employee repo for role checking)
	documentHandler := handlers.NewDocumentHandler(documentService, jwtService, employeeRepo, categoryService)
	documentHandler.RegisterRoutes(api)
	// Audit Log Routes (Admin only)
	auditLogHandler := handlers.NewAuditLogHandler(auditLogService, jwtService, employeeRepo)
	auditLogHandler.RegisterRoutes(api)

	// Notification Routes
	notificationHandler := handlers.NewNotificationHandler(notificationService, jwtService)
	notificationHandler.RegisterRoutes(api)
}
