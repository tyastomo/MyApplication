package database

import (
	"fmt"
	"log"
	// "os" // No longer needed for direct os.Getenv here
	"payslip-generator/pkg/config" // Added
	"payslip-generator/pkg/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// ConnectDB initializes the database connection and performs auto-migration
// This function is used by the main application.
func ConnectDB() {
	var err error
	dbLoggerMode := logger.Info
	if os.Getenv("APP_ENV") == "test" || os.Getenv("LOG_LEVEL") == "silent" { // Less verbose for tests
		dbLoggerMode = logger.Silent
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		config.AppConfig.DBHost,
		config.AppConfig.DBUser,
		config.AppConfig.DBPassword,
		config.AppConfig.DBName,
		config.AppConfig.DBPort,
	)

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(dbLoggerMode),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database (%s): %v", config.AppConfig.DBName, err)
	}

	log.Printf("Database connection established to %s.", config.AppConfig.DBName)
	migrateDB(DB)
	log.Println("Database migration completed.")
}

// InitTestDB connects to the test database and returns the *gorm.DB instance
// It ensures that the database is migrated.
func InitTestDB() *gorm.DB {
	// Ensure test config is loaded (LoadConfig should handle APP_ENV=test)
	// config.LoadConfig() // Already called by test main or test setup

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		config.AppConfig.DBHost,
		config.AppConfig.DBUser,
		config.AppConfig.DBPassword,
		config.AppConfig.DBName,
		config.AppConfig.DBPort,
	)

	testDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Keep tests quiet
	})
	if err != nil {
		log.Fatalf("Failed to connect to TEST database: %v. DSN: %s", err, dsn)
	}

	log.Println("Test database connection established.")
	migrateDB(testDB) // Ensure schema is up-to-date
	log.Println("Test database migration completed.")
	return testDB
}

func migrateDB(db *gorm.DB) {
	// Auto-migrate models
	err := db.AutoMigrate(
		&models.Employee{},
		&models.Admin{},
		&models.AttendancePeriod{},
		&models.AttendanceRecord{},
		&models.OvertimeRecord{},
		&models.ReimbursementRequest{},
		&models.Payslip{},
		&models.AuditLog{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Add unique constraint for AttendanceRecord (EmployeeID, Date)
	if !db.Migrator().HasConstraint(&models.AttendanceRecord{}, "uix_employee_date") {
		err = db.Migrator().CreateConstraint(&models.AttendanceRecord{}, "uix_employee_date")
		if err != nil {
			log.Printf("Warning: Failed to create constraint uix_employee_date: %v", err)
		}
	}

	// Add unique constraint for Payslip (EmployeeID, AttendancePeriodID)
	if !db.Migrator().HasConstraint(&models.Payslip{}, "uix_employee_period") {
		err = db.Migrator().CreateConstraint(&models.Payslip{}, "uix_employee_period")
		if err != nil {
			log.Printf("Warning: Failed to create constraint uix_employee_period: %v", err)
		}
	}

	// Add check constraint for OvertimeRecord Hours <= 3
	var constraintCount int64
	db.Raw("SELECT count(*) FROM information_schema.check_constraints WHERE constraint_name = ? AND table_name = ?", "ck_overtime_hours", "overtime_records").Scan(&constraintCount)
	if constraintCount == 0 {
		err = db.Exec("ALTER TABLE overtime_records ADD CONSTRAINT ck_overtime_hours CHECK (hours <= 3)").Error
		if err != nil {
			log.Printf("Warning: Failed to create CHECK constraint ck_overtime_hours: %v", err)
		}
	}
}

// ClearAllData empties all known tables in the test database
// Be very careful with this function. Ensure it's only used on a test database.
func ClearAllData(db *gorm.DB) error {
	if os.Getenv("APP_ENV") != "test" {
		return fmt.Errorf("ClearAllData can only be run in test environment")
	}

	// Order matters due to foreign key constraints. Start with tables that are referenced by others.
	// Or, temporarily disable foreign key checks if your DB supports it, but that's riskier.
	tables := []string{
		"audit_logs",
		"payslips",
		"reimbursement_requests",
		"overtime_records",
		"attendance_records",
		"attendance_periods",
		"employees",
		"admins",
	}

	for _, table := range tables {
		// Using Exec for TRUNCATE. CASCADE might be needed if FKs are strict and not handled by order.
		// RESTART IDENTITY resets auto-increment counters.
		if err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table)).Error; err != nil {
			// Log error but try to continue, as some tables might not exist or be empty.
			log.Printf("Warning: Failed to truncate table %s: %v", table, err)
		}
	}
	log.Println("All data cleared from test database.")
	return nil
}
