package main

import (
	"log"
	"os"
	"payslip-generator/pkg/config"
	"payslip-generator/pkg/database"
	"payslip-generator/pkg/middleware"
	"payslip-generator/pkg/routes"

	"github.com/gofiber/fiber/v2"

	_ "payslip-generator/docs" // Import generated docs
	swagger "github.com/swaggo/fiber-swagger"
)

// @title Payslip Generation API
// @version 1.0
// @description This is a Fiber-based API for a payslip generation system.
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.url https://github.com/your-repo/payslip-generator
// @contact.email support@example.com
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
// @host localhost:8080
// @BasePath /api/v1
// @schemes http https
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	// Load configuration
	config.LoadConfig()

	// Connect to the database
	database.ConnectDB()

	// Seed data - In a real app, you might control this with a flag
	database.SeedData(database.DB)

	app := fiber.New()

	// Register middleware
	app.Use(middleware.RequestID)       // Add RequestID middleware first
	app.Use(middleware.Logger)          // Add Logger middleware next
	app.Use(middleware.DeserializeUser) // DeserializeUser for auth

	// Group API routes
	api := app.Group("/api/v1")

	// Setup Admin Routes
	adminAPI := api.Group("/admin")
	routes.SetupAdminRoutes(adminAPI)

	// Setup Employee Routes
	employeeAPI := api.Group("/employee")
	routes.SetupEmployeeRoutes(employeeAPI)


	// Default route
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World from Payslip Generator! API docs at /swagger/index.html")
	})

	// Swagger route
	app.Get("/swagger/*", swagger.WrapHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port if not specified
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(app.Listen(":" + port))
}
