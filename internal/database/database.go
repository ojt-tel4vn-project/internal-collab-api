package database

import (
	"fmt"
	"log"

	"github.com/ojt-tel4vn-project/internal-collab-api/internal/config"
	"github.com/ojt-tel4vn-project/internal-collab-api/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		// Prevent AutoMigrate from issuing ALTER TABLE ... DROP/ADD CONSTRAINT
		// statements, which can fail on Supabase/pgBouncer.
		DisableForeignKeyConstraintWhenMigrating: true,
		// Raise slow-query threshold to 2s to suppress noise from Supabase
		// network latency during migration schema queries (~200-400ms each).
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	DB = db
	log.Println("Database connected successfully")
	return nil
}

// dropStaleIndex safely drops a GORM-managed unique index using the correct
// DROP INDEX SQL (not ALTER TABLE DROP CONSTRAINT) to avoid SQLSTATE 42704.
func dropStaleIndex(model interface{}, indexName string) {
	if DB.Migrator().HasIndex(model, indexName) {
		if err := DB.Migrator().DropIndex(model, indexName); err != nil {
			log.Printf("[warn] could not drop stale index %s: %v", indexName, err)
		}
	}
}

func Migrate() error {
	// Drop stale unique indexes that GORM tries to remove via ALTER TABLE DROP
	// CONSTRAINT (wrong SQL). Must be listed here whenever a uniqueIndex tag is
	// added/removed from a model after the table already exists in the DB.
	dropStaleIndex(&models.LeaveType{}, "uni_leave_types_name")
	dropStaleIndex(&models.LeaveRequest{}, "uni_leave_requests_action_token")

	return DB.AutoMigrate(
		&models.Department{},
		&models.Role{},
		&models.Employee{},
		&models.RefreshToken{},
		&models.AuditLog{},
		&models.Notification{},
		&models.AppConfig{},
		&models.LeaveType{},
		&models.LeaveQuota{},
		&models.LeaveRequest{},
		&models.Attendance{},
		&models.AttendanceComment{},
	)
}
