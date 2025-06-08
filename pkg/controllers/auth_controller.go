package controllers

import (
	// "payslip-generator/pkg/config"
	// "payslip-generator/pkg/database"
	// "payslip-generator/pkg/models"
	// "payslip-generator/pkg/utils"
	"payslip-generator/pkg/config"
	"payslip-generator/pkg/database"
	"payslip-generator/pkg/models"
	"payslip-generator/pkg/utils"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	// "github.com/google/uuid"
)

// LoginPayload struct for parsing login request
type LoginPayload struct {
	Username string `json:"username" xml:"username" form:"username" validate:"required"`
	Password string `json:"password" xml:"password" form:"password" validate:"required"`
}

// AdminLogin godoc
// @Summary Admin Login
// @Description Authenticates an admin and returns a JWT token.
// @Tags Auth
// @Accept json
// @Produce json
// @Param credentials body LoginPayload true "Admin Credentials"
// @Success 200 {object} map[string]interface{} `json:"{"status":"success", "token":"jwt_token_here"}"`
// @Failure 400 {object} map[string]string `json:"{"status":"fail", "message":"error_message"}"`
// @Failure 401 {object} map[string]string `json:"{"status":"fail", "message":"Invalid credentials."}"`
// @Failure 500 {object} map[string]string `json:"{"status":"error", "message":"Database error / Could not generate token."}"`
// @Router /admin/login [post]
func AdminLogin(c *fiber.Ctx) error {
	var payload LoginPayload
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status": "fail", "message": err.Error(),
		})
	}

	// Basic validation
	if payload.Username == "" || payload.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status": "fail", "message": "Username and password are required.",
		})
	}

	var admin models.Admin
	result := database.DB.Where("username = ?", payload.Username).First(&admin)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": "fail", "message": "Invalid credentials."})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Database error."})
	}

	if !utils.CheckPasswordHash(payload.Password, admin.Password) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": "fail", "message": "Invalid credentials."})
	}

	token, err := utils.GenerateJWT(admin.ID, "admin", config.AppConfig.JWTSecret)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Could not generate token."})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "token": token})
}

// EmployeeLogin godoc
// @Summary Employee Login
// @Description Authenticates an employee and returns a JWT token.
// @Tags Auth
// @Accept json
// @Produce json
// @Param credentials body LoginPayload true "Employee Credentials"
// @Success 200 {object} map[string]interface{} `json:"{"status":"success", "token":"jwt_token_here"}"`
// @Failure 400 {object} map[string]string `json:"{"status":"fail", "message":"error_message"}"`
// @Failure 401 {object} map[string]string `json:"{"status":"fail", "message":"Invalid credentials."}"`
// @Failure 500 {object} map[string]string `json:"{"status":"error", "message":"Database error / Could not generate token."}"`
// @Router /employee/login [post]
func EmployeeLogin(c *fiber.Ctx) error {
	var payload LoginPayload
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status": "fail", "message": err.Error(),
		})
	}

	if payload.Username == "" || payload.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status": "fail", "message": "Username and password are required.",
		})
	}

	var employee models.Employee
	result := database.DB.Where("username = ?", payload.Username).First(&employee)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": "fail", "message": "Invalid credentials."})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Database error."})
	}

	if !utils.CheckPasswordHash(payload.Password, employee.Password) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": "fail", "message": "Invalid credentials."})
	}

	token, err := utils.GenerateJWT(employee.ID, "employee", config.AppConfig.JWTSecret)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Could not generate token."})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "token": token})
}


// Login (Conceptual - to be implemented fully later)
// This is a placeholder to illustrate where JWT generation would occur.
// Actual implementation will involve database checks, password verification, etc.
/*
func LoginEmployee(c *fiber.Ctx) error {
	var payload LoginPayload
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status": "fail", "message": err.Error(),
		})
	}

	// 1. Validate input (e.g., using a validator library)

	// 2. Fetch employee from database by username
	var employee models.Employee
	// result := database.DB.Where("username = ?", payload.Username).First(&employee)
	// if result.Error != nil {
	//    return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": "fail", "message": "Invalid credentials"})
	// }

	// 3. Check password
	// if !utils.CheckPasswordHash(payload.Password, employee.Password) {
	//    return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": "fail", "message": "Invalid credentials"})
	// }

	// 4. Generate JWT
	// token, err := utils.GenerateJWT(employee.ID, "employee", config.AppConfig.JWTSecret)
	// if err != nil {
	//    return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Could not generate token"})
	// }

	// return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "token": token})
	return c.SendStatus(fiber.StatusNotImplemented) // Placeholder
}

func LoginAdmin(c *fiber.Ctx) error {
	// Similar logic for admin login
	return c.SendStatus(fiber.StatusNotImplemented) // Placeholder
}
*/
