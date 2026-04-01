package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/ojt-tel4vn-project/internal-collab-api/internal/config"
	"github.com/ojt-tel4vn-project/internal-collab-api/internal/database"
	"github.com/ojt-tel4vn-project/internal-collab-api/models"
)

func main() {
	if err := godotenv.Load(); err != nil {
		if err := godotenv.Load("../.env"); err != nil {
			log.Println("No .env file found")
		}
	}
	cfg := config.Load()
	if err := database.Connect(cfg); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	var types []models.LeaveType
	database.DB.Find(&types)

	for _, t := range types {
		fmt.Printf("ID: %s | Name: %s\n", t.ID, t.Name)
	}
}
