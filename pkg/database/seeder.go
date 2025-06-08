package database

import (
	"fmt"
	"log"
	"math/rand"
	"payslip-generator/pkg/models"
	"payslip-generator/pkg/utils"

	"github.com/go-faker/faker/v4"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// SeedData populates the database with initial data
func SeedData(db *gorm.DB) {
	log.Println("Starting to seed data...")

	// Seed Admin
	hashedAdminPassword, err := utils.HashPassword("adminpassword")
	if err != nil {
		log.Fatalf("Failed to hash admin password: %v", err)
	}

	admin := models.Admin{
		Username: "admin",
		Password: hashedAdminPassword,
	}
	// admin.CreatedBy is already a pointer, so it's nil by default if not set.
	// If you had a specific system UUID:
	// var systemUserID = uuid.Nil // or uuid.New() for a specific system user
	// admin.CreatedBy = &systemUserID

	var existingAdmin models.Admin
	if err := db.Where("username = ?", admin.Username).First(&existingAdmin).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			if err := db.Create(&admin).Error; err != nil {
				log.Printf("Failed to seed admin user: %v", err)
			} else {
				log.Println("Admin user seeded.")
			}
		} else {
			log.Printf("Error checking for existing admin: %v", err)
		}
	} else {
		log.Println("Admin user already exists.")
	}

	// Seed Employees
	seededEmployees := 0
	for i := 0; i < 100; i++ {
		username := faker.Username()
		var existingEmployee models.Employee
		// Check if employee with this username already exists
		if err := db.Where("username = ?", username).First(&existingEmployee).Error; err == nil {
			// log.Printf("Employee with username %s already exists, skipping or generating new.", username)
			// Option: generate a new username or simply skip
			i-- // Decrement i to ensure 100 unique employees are attempted
			continue
		} else if err != gorm.ErrRecordNotFound {
			log.Printf("Error checking for existing employee %s: %v", username, err)
			continue
		}

		plainPassword := faker.Password()
		hashedPassword, err := utils.HashPassword(plainPassword)
		if err != nil {
			log.Printf("Failed to hash password for employee %s: %v", username, err)
			continue
		}

		// Generate salary between 30000 and 150000
		// faker.Number().Float(2, 30000, 150000) can be used if faker is configured for it
		// For now, using math/rand and shopspring/decimal for precision
		salaryVal := 30000 + rand.Float64()*(150000-30000)
		salary := decimal.NewFromFloat(salaryVal).Round(2).InexactFloat64() // Store as float64 after rounding

		employee := models.Employee{
			Username: username,
			Password: hashedPassword,
			Salary:   salary,
		}
		// employee.CreatedBy is already a pointer, nil by default.
		// employee.IPAddress can also be nil by default.

		if err := db.Create(&employee).Error; err != nil {
			log.Printf("Failed to seed employee %s: %v", username, err)
		} else {
			seededEmployees++
		}
	}
	log.Printf("%d fake employees seeded.", seededEmployees)
	log.Println("Data seeding completed.")
}
