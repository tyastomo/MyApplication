package tests

import (
	"encoding/json"
	"net/http"
	"payslip-generator/pkg/config"
	"payslip-generator/pkg/database"
	"payslip-generator/pkg/models"
	"payslip-generator/pkg/utils"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdminLogin_Success(t *testing.T) {
	clearTestData() // Clear data before test

	// Setup: Create an admin user directly in DB
	hashedPassword, _ := utils.HashPassword("strongpassword")
	adminUser := models.Admin{Username: "testadmin", Password: hashedPassword}
	// adminUser.CreatedBy, adminUser.IPAddress can be nil for this test
	err := testDB.Create(&adminUser).Error
	require.NoError(t, err, "Failed to seed admin user for login test")

	payload := fiber.Map{
		"username": "testadmin",
		"password": "strongpassword",
	}
	respRec, err := makeRequest("POST", "/api/v1/admin/login", createJSONBody(payload))
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, respRec.Code)

	var body map[string]interface{}
	err = json.Unmarshal(respRec.Body.Bytes(), &body)
	require.NoError(t, err)

	assert.Equal(t, "success", body["status"])
	assert.NotEmpty(t, body["token"], "Token should be present in response")

	// Optionally, parse token and verify claims
	_, claims, err := utils.ParseJWT(body["token"].(string), config.AppConfig.JWTSecret)
	require.NoError(t, err)
	assert.Equal(t, adminUser.ID.String(), claims.UserID.String())
	assert.Equal(t, "admin", claims.UserType)
}

func TestAdminLogin_InvalidCredentials(t *testing.T) {
	clearTestData()

	hashedPassword, _ := utils.HashPassword("password123")
	adminUser := models.Admin{Username: "realadmin", Password: hashedPassword}
	testDB.Create(&adminUser)

	// Test with wrong password
	payloadWrongPass := fiber.Map{"username": "realadmin", "password": "wrongpassword"}
	respRecWrongPass, _ := makeRequest("POST", "/api/v1/admin/login", createJSONBody(payloadWrongPass))
	assert.Equal(t, http.StatusUnauthorized, respRecWrongPass.Code)
	var bodyWrongPass map[string]interface{}
	json.Unmarshal(respRecWrongPass.Body.Bytes(), &bodyWrongPass)
	assert.Equal(t, "fail", bodyWrongPass["status"])
	assert.Contains(t, bodyWrongPass["message"], "Invalid credentials")

	// Test with non-existent username
	payloadNonExistentUser := fiber.Map{"username": "fakeadmin", "password": "password"}
	respRecNonExistentUser, _ := makeRequest("POST", "/api/v1/admin/login", createJSONBody(payloadNonExistentUser))
	assert.Equal(t, http.StatusUnauthorized, respRecNonExistentUser.Code)
	var bodyNonExistentUser map[string]interface{}
	json.Unmarshal(respRecNonExistentUser.Body.Bytes(), &bodyNonExistentUser)
	assert.Equal(t, "fail", bodyNonExistentUser["status"])
	assert.Contains(t, bodyNonExistentUser["message"], "Invalid credentials")
}


func TestEmployeeLogin_Success(t *testing.T) {
	clearTestData()

	hashedPassword, _ := utils.HashPassword("securepass")
	employeeUser := models.Employee{
		Username: "testemployee",
		Password: hashedPassword,
		Salary: 50000,
	}
	err := testDB.Create(&employeeUser).Error
	require.NoError(t, err, "Failed to seed employee user for login test")

	payload := fiber.Map{
		"username": "testemployee",
		"password": "securepass",
	}
	respRec, err := makeRequest("POST", "/api/v1/employee/login", createJSONBody(payload))
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, respRec.Code)

	var body map[string]interface{}
	json.Unmarshal(respRec.Body.Bytes(), &body)
	assert.Equal(t, "success", body["status"])
	assert.NotEmpty(t, body["token"])

	_, claims, err := utils.ParseJWT(body["token"].(string), config.AppConfig.JWTSecret)
	require.NoError(t, err)
	assert.Equal(t, employeeUser.ID.String(), claims.UserID.String())
	assert.Equal(t, "employee", claims.UserType)
}

func TestEmployeeLogin_InvalidCredentials(t *testing.T) {
	clearTestData()
	// Seed an employee
	database.SeedData(testDB) // SeedData creates an admin and 100 employees

	// Attempt login with definitely wrong credentials
	payload := fiber.Map{"username": "nonexistentuser", "password": "fakepassword"}
	respRec, _ := makeRequest("POST", "/api/v1/employee/login", createJSONBody(payload))
	assert.Equal(t, http.StatusUnauthorized, respRec.Code)
}


// Test access to a protected admin route
func TestProtectedAdminRoute_AccessControl(t *testing.T) {
	clearTestData()
	// Seed admin & employee
	adminPass := "adminpass"
	empPass := "emppass"
	adminHashedPass, _ := utils.HashPassword(adminPass)
	empHashedPass, _ := utils.HashPassword(empPass)

	admin := models.Admin{Username: "prot_admin", Password: adminHashedPass}
	testDB.Create(&admin)
	employee := models.Employee{Username: "prot_emp", Password: empHashedPass, Salary: 1000}
	testDB.Create(&employee)

	// 1. No token
	respNoToken, _ := makeRequest("POST", "/api/v1/admin/attendance-periods", nil)
	assert.Equal(t, http.StatusUnauthorized, respNoToken.Code, "Should be Unauthorized without token")

	// 2. Employee token trying to access admin route
	empLoginPayload := fiber.Map{"username": "prot_emp", "password": empPass}
	respEmpLogin, _ := makeRequest("POST", "/api/v1/employee/login", createJSONBody(empLoginPayload))
	var empLoginBody map[string]interface{}
	json.Unmarshal(respEmpLogin.Body.Bytes(), &empLoginBody)
	empToken := empLoginBody["token"].(string)

	attendancePayload := fiber.Map{"start_date": "2023-01-01", "end_date": "2023-01-15"}
	respEmpToken, _ := makeRequest("POST", "/api/v1/admin/attendance-periods", createJSONBody(attendancePayload), empToken)
	assert.Equal(t, http.StatusForbidden, respEmpToken.Code, "Should be Forbidden for employee token on admin route")

	// 3. Admin token
	adminLoginPayload := fiber.Map{"username": "prot_admin", "password": adminPass}
	respAdminLogin, _ := makeRequest("POST", "/api/v1/admin/login", createJSONBody(adminLoginPayload))
	var adminLoginBody map[string]interface{}
	json.Unmarshal(respAdminLogin.Body.Bytes(), &adminLoginBody)
	adminToken := adminLoginBody["token"].(string)

	respAdminToken, _ := makeRequest("POST", "/api/v1/admin/attendance-periods", createJSONBody(attendancePayload), adminToken)
	assert.Equal(t, http.StatusCreated, respAdminToken.Code, "Should be Created for admin token on admin route")
}
