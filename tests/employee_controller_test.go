package tests

import (
	"encoding/json"
	"net/http"
	"payslip-generator/pkg/models"
	"payslip-generator/pkg/utils"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper to get a valid employee token
func getEmployeeToken(t *testing.T, username, password string, salary float64) string {
	// No need to clearTestData here as it's usually called by the test case itself
	// or a higher-level setup for a group of tests.
	hashedPassword, _ := utils.HashPassword(password)
	employee := models.Employee{Username: username, Password: hashedPassword, Salary: salary}
	err := testDB.Create(&employee).Error
	require.NoError(t, err, "Failed to create employee for token generation")

	loginPayload := fiber.Map{"username": username, "password": password}
	resp, err := makeRequest("POST", "/api/v1/employee/login", createJSONBody(loginPayload))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.Code, "Employee login failed for token generation")

	var body map[string]interface{}
	err = json.Unmarshal(resp.Body.Bytes(), &body)
	require.NoError(t, err)
	token, ok := body["token"].(string)
	require.True(t, ok, "Token not found or not a string in login response")
	return token
}

func TestSubmitAttendance_Success(t *testing.T) {
	clearTestData()
	empToken := getEmployeeToken(t, "attemp", "attpass", 50000)

	// Get admin token to create an attendance period
	// In a real scenario, admin might have a separate test setup or use fixed IDs
	adminToken := getAdminToken(t, "att_admin", "adminpass")

	// Create an active attendance period that covers today
	// Ensure today is not a weekend for this test to pass reliably, or mock time.
	// For simplicity, we assume test isn't run on a weekend or this part needs time mocking.
	today := time.Now()
	if today.Weekday() == time.Saturday || today.Weekday() == time.Sunday {
		t.Skip("Skipping TestSubmitAttendance_Success on weekend to avoid conflict with weekend submission rule, needs time mocking for robustness.")
	}

	periodPayload := fiber.Map{
		"start_date": today.Format("2006-01-02"),
		"end_date":   today.AddDate(0, 0, 5).Format("2006-01-02"), // Period covers today
	}
	respPeriod, err := makeRequest("POST", "/api/v1/admin/attendance-periods", createJSONBody(periodPayload), adminToken)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, respPeriod.Code, "Failed to create attendance period for test")
    var periodBody map[string]interface{}
    json.Unmarshal(respPeriod.Body.Bytes(), &periodBody)
    periodData := periodBody["data"].(map[string]interface{})
    _ = periodData["ID"].(string) // Period ID

	// Employee submits attendance
	resp, err := makeRequest("POST", "/api/v1/employee/attendance", nil, empToken)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.Code)

	var body map[string]interface{}
	json.Unmarshal(resp.Body.Bytes(), &body)
	assert.Equal(t, "success", body["status"])
	attData := body["data"].(map[string]interface{})
	assert.NotEmpty(t, attData["ID"])
	assert.Contains(t, attData["CheckInTime"], today.Format("2006-01-02")) // CheckInTime is a full timestamp

	// Verify in DB
	var attRecord models.AttendanceRecord
	err = testDB.First(&attRecord, "id = ?", attData["ID"]).Error
	require.NoError(t, err)
	assert.Equal(t, today.Format("2006-01-02"), attRecord.Date.Format("2006-01-02"))
}

func TestSubmitAttendance_AlreadySubmitted(t *testing.T) {
	clearTestData()
	empToken := getEmployeeToken(t, "attemp2", "attpass2", 50000)
	adminToken := getAdminToken(t, "att_admin2", "adminpass2")

	today := time.Now()
	if today.Weekday() == time.Saturday || today.Weekday() == time.Sunday {
		t.Skip("Skipping TestSubmitAttendance_AlreadySubmitted on weekend.")
	}
	periodPayload := fiber.Map{
		"start_date": today.Format("2006-01-02"),
		"end_date":   today.AddDate(0, 0, 5).Format("2006-01-02"),
	}
	makeRequest("POST", "/api/v1/admin/attendance-periods", createJSONBody(periodPayload), adminToken)


	// First submission - should succeed
	resp1, err1 := makeRequest("POST", "/api/v1/employee/attendance", nil, empToken)
	require.NoError(t, err1)
	require.Equal(t, http.StatusCreated, resp1.Code)

	// Second submission for the same day - should fail
	resp2, err2 := makeRequest("POST", "/api/v1/employee/attendance", nil, empToken)
	require.NoError(t, err2)
	assert.Equal(t, http.StatusConflict, resp2.Code)

	var body map[string]interface{}
	json.Unmarshal(resp2.Body.Bytes(), &body)
	assert.Equal(t, "fail", body["status"])
	assert.Contains(t, body["message"], "Attendance already submitted for today")
}

func TestSubmitAttendance_NoActivePeriod(t *testing.T) {
	clearTestData()
	empToken := getEmployeeToken(t, "attemp3", "attpass3", 50000)
	// No attendance period created for today

	today := time.Now()
	if today.Weekday() == time.Saturday || today.Weekday() == time.Sunday {
		t.Skip("Skipping TestSubmitAttendance_NoActivePeriod on weekend.")
	}

	resp, err := makeRequest("POST", "/api/v1/employee/attendance", nil, empToken)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.Code)
	var body map[string]interface{}
	json.Unmarshal(resp.Body.Bytes(), &body)
	assert.Contains(t, body["message"], "No active attendance period for today")
}

// TODO: Add tests for SubmitOvertime, SubmitReimbursement, GetMyPayslip
// These will follow similar patterns:
// 1. Setup necessary state (admin, employee, tokens, attendance periods, attendance records for overtime etc.)
// 2. Make the request.
// 3. Assert status code and response body.
// 4. Verify changes in the database.
// 5. Test edge cases and validation errors (e.g., overtime hours > 3, submitting overtime for a day not attended).
