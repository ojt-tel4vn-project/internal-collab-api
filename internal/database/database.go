package database

import (
	"fmt"
	"log"

	"github.com/ojt-tel4vn-project/internal-collab-api/internal/config"
	"github.com/ojt-tel4vn-project/internal-collab-api/models"
	docModels "github.com/ojt-tel4vn-project/internal-collab-api/models/document"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect(cfg *config.Config) error {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	DB = db
	log.Println("Database connected successfully")
	return nil
}

func Migrate() error {
	return DB.AutoMigrate(
		&models.Employee{},
		&models.Role{},
		&models.Department{},
		&docModels.Document{},
		&docModels.DocumentRead{},
	)
}
