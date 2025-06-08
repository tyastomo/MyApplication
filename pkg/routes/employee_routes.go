package routes

import (
	"payslip-generator/pkg/controllers"
	"payslip-generator/pkg/middleware"

	"github.com/gofiber/fiber/v2"
)

// SetupEmployeeRoutes sets up the employee routes
func SetupEmployeeRoutes(api fiber.Router) {
	// Public employee routes (e.g., login)
	api.Post("/login", controllers.EmployeeLogin)

	// Group for protected employee routes
	// This group applies RequireLoggedIn and then RequireUserType("employee")
	employeeProtectedGroup := api.Group("", middleware.RequireLoggedIn(), middleware.RequireUserType("employee"))

	employeeProtectedGroup.Post("/attendance", controllers.SubmitAttendance)
	employeeProtectedGroup.Post("/overtime", controllers.SubmitOvertime)
	employeeProtectedGroup.Post("/reimbursements", controllers.SubmitReimbursement)
	employeeProtectedGroup.Get("/payslip", controllers.GetMyPayslip)

	// Example of another protected route:
	// employeeProtectedGroup.Get("/profile", func(c *fiber.Ctx) error {
	// 	userID, _ := utils.GetUserIDFromContext(c) // Assuming utils has this helper
	// 	userType, _ := utils.GetUserTypeFromContext(c)
	// 	return c.JSON(fiber.Map{
	// 		"message":  "Welcome to your Employee Profile!",
	// 		"userID":   userID,
	// 		"userType": userType,
	// 	})
	// })
}
