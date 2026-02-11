package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/ojt-tel4vn-project/internal-collab-api/internal/config"
	"github.com/ojt-tel4vn-project/internal-collab-api/internal/database"
	"github.com/ojt-tel4vn-project/internal-collab-api/pkg/crypto"
	"github.com/ojt-tel4vn-project/internal-collab-api/pkg/email"
	"github.com/ojt-tel4vn-project/internal-collab-api/pkg/logger"
	"github.com/ojt-tel4vn-project/internal-collab-api/repository"
	"github.com/ojt-tel4vn-project/internal-collab-api/routes"
	"github.com/ojt-tel4vn-project/internal-collab-api/services"
)

func main() {
	// Load environment variables
	// Try current directory first, then parent directory
	if err := godotenv.Load(); err != nil {
		if err := godotenv.Load("../.env"); err != nil {
			log.Println("No .env file found")
		}
	}

	// Initialize Logger
	if err := logger.InitDefaultLogger(); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Load configuration
	cfg := config.Load()

	// Initialize JWT
	crypto.InitJWT(cfg.JWT.Secret)

	// Connect to database
	if err := database.Connect(cfg); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Run migrations
	if err := database.Migrate(); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Initialize dependencies
	// Repositories

	employeeRepo := repository.NewEmployeeRepository(database.DB)
	refreshTokenRepo := repository.NewRefreshTokenRepository(database.DB)

	// Utils
	jwtService := crypto.NewJWTService()
	passwordService := crypto.NewPasswordService()

	// Audit Log
	auditLogRepo := repository.NewAuditLogRepository(database.DB)
	auditLogService := services.NewAuditLogService(auditLogRepo)

	// Notifications
	notificationRepo := repository.NewNotificationRepository(database.DB)
	notificationService := services.NewNotificationService(notificationRepo)

	// Email Service
	var emailService email.EmailService
	if cfg.Email.BrevoAPIKey != "" {
		emailService = email.NewBrevoEmailService(
			cfg.Email.BrevoAPIKey,
			cfg.Email.FromEmail,
			cfg.Email.FromName,
		)
		log.Println("Email service initialized (Brevo)")
	} else {
		log.Println("Email service disabled (no BREVO_API_KEY configured)")
	}

	// Services

	authService := services.NewAuthService(employeeRepo, refreshTokenRepo, jwtService, passwordService, emailService)

	authService := services.NewAuthService(employeeRepo, refreshTokenRepo, jwtService, passwordService, emailService)
	employeeService := services.NewEmployeeService(employeeRepo, passwordService, emailService)

	// Cron Service
	cronService := services.NewCronService(employeeRepo, emailService, notificationService)
	cronService.Start()
	defer cronService.Stop()

	// Setup Chi router
	router := chi.NewMux()

	// Setup Huma API
	humaConfig := huma.DefaultConfig("Internal Collab API", "1.0.0")
	api := humachi.New(router, humaConfig)

	// Register routes
	routes.SetupRoutes(api, authService, employeeService, auditLogService, notificationService, jwtService, employeeRepo)

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("Server starting on port %s", cfg.Server.Port)
	log.Printf("API docs available at http://localhost:%s/docs", cfg.Server.Port)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
