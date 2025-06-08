package tests

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http/httptest"
	"os"
	"payslip-generator/pkg/config"
	"payslip-generator/pkg/database"
	"payslip-generator/pkg/middleware"
	"payslip-generator/pkg/routes"
	"testing"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

var (
	testApp *fiber.App
	testDB  *gorm.DB
)

// TestMain is the entry point for tests in this package.
func TestMain(m *testing.M) {
	// Set environment to test
	os.Setenv("APP_ENV", "test")

	// Load test configuration
	// This will load .env.test because APP_ENV=test
	config.LoadConfig()

	// Initialize test database
	testDB = database.InitTestDB()

	// Setup test Fiber app
	testApp = setupTestApp()

	// Run tests
	exitCode := m.Run()

	// Teardown (if any specific needed beyond what individual tests do)
	log.Println("Tests finished. Cleaning up.")
	// You might close DB connection here if InitTestDB opened a unique one for the suite.
	// For now, InitTestDB uses config, so it's like the main app's DB but for test DB.

	os.Exit(exitCode)
}

// setupTestApp initializes a Fiber app instance for testing.
func setupTestApp() *fiber.App {
	app := fiber.New()

	// Use the global testDB instance initialized in TestMain
	database.DB = testDB // Crucial: Point the app's DB to the testDB instance

	// Register middleware similar to main.go
	app.Use(middleware.RequestID)
	app.Use(middleware.Logger) // You might want to disable verbose logging for tests or use a test-specific logger config
	app.Use(middleware.DeserializeUser)

	// Group API routes
	api := app.Group("/api/v1")

	// Setup Admin Routes
	adminAPI := api.Group("/admin")
	routes.SetupAdminRoutes(adminAPI)

	// Setup Employee Routes
	employeeAPI := api.Group("/employee")
	routes.SetupEmployeeRoutes(employeeAPI)

	return app
}

// Helper function to make requests and return response (primarily for integration tests)
// This simplifies making requests to the testApp.
func makeRequest(method, url string, body io.Reader, token ...string) (*httptest.ResponseRecorder, error) {
	req := httptest.NewRequest(method, url, body)
	req.Header.Set("Content-Type", "application/json")
	if len(token) > 0 && token[0] != "" {
		req.Header.Set("Authorization", "Bearer "+token[0])
	}

	respRec := httptest.NewRecorder()
	testApp.Test(req, -1) // -1 for no timeout, or set a reasonable timeout
	return respRec, nil
}

// Helper to create JSON body for requests
func createJSONBody(data interface{}) *bytes.Buffer {
	bodyBytes, _ := json.Marshal(data)
	return bytes.NewBuffer(bodyBytes)
}

// clearTestData is a helper to clear data before each relevant test or test group
func clearTestData() {
	err := database.ClearAllData(testDB)
	if err != nil {
		log.Fatalf("Failed to clear test data: %v", err)
	}
}
