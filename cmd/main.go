package main

import (
	"fmt"
	"log"
	"net/http"

	"time"

	ratelimit "github.com/JGLTechnologies/gin-rate-limit"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/ojt-tel4vn-project/internal-collab-api/internal/config"
	"github.com/ojt-tel4vn-project/internal-collab-api/internal/database"
	"github.com/ojt-tel4vn-project/internal-collab-api/internal/storage"
	"github.com/ojt-tel4vn-project/internal-collab-api/pkg/crypto"
	"github.com/ojt-tel4vn-project/internal-collab-api/pkg/email"
	"github.com/ojt-tel4vn-project/internal-collab-api/pkg/logger"
	"github.com/ojt-tel4vn-project/internal-collab-api/pkg/sse"
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

	// Initialize Storage
	storageService := storage.NewSupabaseStorage(
		cfg.Supabase.URL,
		cfg.Supabase.Bucket,
		cfg.Supabase.APIKey,
	)

	// Run migrations
	if err := database.Migrate(); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Initialize dependencies
	// Repositories
	employeeRepo := repository.NewEmployeeRepository(database.DB)
	refreshTokenRepo := repository.NewRefreshTokenRepository(database.DB)
	categoryRepo := repository.NewDocumentCategoryRepository(database.DB)
	documentRepo := repository.NewDocumentRepository(database.DB)
	appConfigRepo := repository.NewAppConfigRepository(database.DB)
	departmentRepo := repository.NewDepartmentRepository(database.DB)

	// Utils
	jwtService := crypto.NewJWTService()
	passwordService := crypto.NewPasswordService()

	// Audit Log
	auditLogRepo := repository.NewAuditLogRepository(database.DB)
	auditLogService := services.NewAuditLogService(auditLogRepo)

	// Department
	departmentService := services.NewDepartmentService(departmentRepo)

	// Notifications & SSE
	sseBroker := sse.NewSSEBroker(jwtService)
	sseBroker.Start()
	defer sseBroker.Stop()

	notificationRepo := repository.NewNotificationRepository(database.DB)
	notificationService := services.NewNotificationService(notificationRepo, sseBroker)

	// Email Service
	var emailService email.EmailService
	if cfg.Email.BrevoAPIKey != "" {
		baseEmailService := email.NewBrevoEmailService(
			cfg.Email.BrevoAPIKey,
			cfg.Email.FromEmail,
			cfg.Email.FromName,
		)
		// Wrap with Async Email Service (3 concurrent workers)
		emailService = email.NewAsyncEmailService(baseEmailService, 3)
		log.Println("Async Email service initialized (Brevo)")
	} else {
		log.Println("Email service disabled (no BREVO_API_KEY configured)")
	}

	leaveRepo := repository.NewLeaveRepository(database.DB)
	
	// Services
	authService := services.NewAuthService(employeeRepo, refreshTokenRepo, jwtService, passwordService, emailService, cfg.Server.FrontendURL)
	employeeService := services.NewEmployeeService(employeeRepo, passwordService, emailService, appConfigRepo, storageService, leaveRepo)

	categoryService := services.NewDocumentCategoryService(categoryRepo)
	documentService := services.NewDocumentService(documentRepo, storageService)

	stickerRepo := repository.NewStickerRepository(database.DB)
	stickerService := services.NewStickerService(stickerRepo, repository.NewPointConfigRepository(database.DB), database.DB)

	leaveService := services.NewLeaveService(leaveRepo, employeeRepo, jwtService, notificationService)

	attendanceRepo := repository.NewAttendanceRepository(database.DB)
	attendanceService := services.NewAttendanceService(attendanceRepo, employeeRepo, appConfigRepo)

	commentRepo := repository.NewCommentRepository(database.DB)
	commentService := services.NewCommentService(commentRepo)

	// Cron Service
	cronService := services.NewCronService(employeeRepo, emailService, notificationService)
	cronService.Start()
	defer cronService.Stop()

	// Setup Gin router
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// Health check (no rate limit — used by Docker/k8s)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "internal-collab-api",
		})
	})

	// Setup Rate Limiter (In-Memory)
	store := ratelimit.InMemoryStore(&ratelimit.InMemoryOptions{
		Rate:  time.Minute,
		Limit: 300,
	})
	mw := ratelimit.RateLimiter(store, &ratelimit.Options{
		ErrorHandler: func(c *gin.Context, info ratelimit.Info) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many requests. Please try again later.",
			})
		},
		KeyFunc: func(c *gin.Context) string {
			return c.ClientIP()
		},
	})
	router.Use(mw)

	// SSE Route (Bypassing Huma because it strictly expects JSON standard response)
	router.GET("/api/v1/notifications/stream", sseBroker.ServeHTTP())

	// Setup Huma API
	humaConfig := huma.DefaultConfig("Internal Collab API", "1.0.0")
	api := humagin.New(router, humaConfig)

	// Register routes
	routes.SetupRoutes(api,
		authService,
		employeeService,
		auditLogService,
		notificationService,
		jwtService,
		employeeRepo,
		documentService,
		categoryService,
		leaveService,
		attendanceService,
		stickerService,
		commentService,
		departmentService)

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("Server starting on port %s", cfg.Server.Port)
	log.Printf("API docs available at http://localhost:%s/docs", cfg.Server.Port)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
