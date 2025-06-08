package routes

import (
	"payslip-generator/pkg/controllers"
	"payslip-generator/pkg/middleware"

	"github.com/gofiber/fiber/v2"
)

// SetupAdminRoutes sets up the admin routes
func SetupAdminRoutes(api fiber.Router) {
	// Public admin routes (e.g., login)
	api.Post("/login", controllers.AdminLogin)

	// Group for protected admin routes
	// This group applies RequireLoggedIn and then RequireUserType("admin")
	adminProtectedGroup := api.Group("", middleware.RequireLoggedIn(), middleware.RequireUserType("admin"))

	adminProtectedGroup.Post("/attendance-periods", controllers.CreateAttendancePeriod)
	adminProtectedGroup.Post("/payroll", controllers.RunPayroll)
	adminProtectedGroup.Get("/payslips-summary", controllers.GetPayslipsSummary)

	// Example of another protected route:
	// adminProtectedGroup.Get("/dashboard", func(c *fiber.Ctx) error {
	// 	userID, _ := utils.GetUserIDFromContext(c) // Assuming utils has this helper
	// 	userType, _ := utils.GetUserTypeFromContext(c)
	// 	return c.JSON(fiber.Map{
	// 		"message":  "Welcome to Admin Dashboard!",
	// 		"userID":   userID,
	// 		"userType": userType,
	// 	})
	// })
}
