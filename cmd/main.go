package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/ojt-tel4vn-project/internal-collab-api/config"
	"github.com/ojt-tel4vn-project/internal-collab-api/database"
	"github.com/ojt-tel4vn-project/internal-collab-api/repository"
	"github.com/ojt-tel4vn-project/internal-collab-api/routes"
	"github.com/ojt-tel4vn-project/internal-collab-api/services"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Load configuration
	cfg := config.Load()

	// Connect to database
	if err := database.Connect(cfg); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Run migrations
	if err := database.Migrate(); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Initialize dependencies
	todoRepo := repository.NewTodoRepository(database.DB)
	todoService := services.NewTodoService(todoRepo)

	// Setup Chi router
	router := chi.NewMux()

	// Setup Huma API
	api := humachi.New(router, huma.DefaultConfig("Todo List API", "1.0.0"))

	// Register routes
	routes.SetupRoutes(api, todoService)

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("Server starting on port %s", cfg.Server.Port)
	log.Printf("API docs available at http://localhost:%s/docs", cfg.Server.Port)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
