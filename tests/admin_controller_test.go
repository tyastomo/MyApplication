package tests

import (
	"encoding/json"
	"net/http"
	"payslip-generator/pkg/models"
	"payslip-generator/pkg/utils"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper to get a valid admin token for testing protected routes
func getAdminToken(t *testing.T, username, password string) string {
	clearTestData() // Usually clear before setting up specific users for a test block
	hashedPassword, _ := utils.HashPassword(password)
	admin := models.Admin{Username: username, Password: hashedPassword}
	err := testDB.Create(&admin).Error
	require.NoError(t, err)

	loginPayload := fiber.Map{"username": username, "password": password}
	resp, err := makeRequest("POST", "/api/v1/admin/login", createJSONBody(loginPayload))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.Code)

	var body map[string]interface{}
	err = json.Unmarshal(resp.Body.Bytes(), &body)
	require.NoError(t, err)
	return body["token"].(string)
}


func TestCreateAttendancePeriod_Success(t *testing.T) {
	adminToken := getAdminToken(t, "periodadmin", "periodpass")

	payload := fiber.Map{
		"start_date": "2024-01-01",
		"end_date":   "2024-01-31",
	}
	resp, err := makeRequest("POST", "/api/v1/admin/attendance-periods", createJSONBody(payload), adminToken)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.Code)

	var body map[string]interface{}
	json.Unmarshal(resp.Body.Bytes(), &body)
	assert.Equal(t, "success", body["status"])

	data := body["data"].(map[string]interface{})
	assert.NotEmpty(t, data["ID"])
	assert.Equal(t, "2024-01-01T00:00:00Z", data["StartDate"]) // GORM stores timestamptz in UTC
	assert.Equal(t, "2024-01-31T00:00:00Z", data["EndDate"])

	// Verify in DB
	var period models.AttendancePeriod
	err = testDB.First(&period, "id = ?", data["ID"]).Error
	require.NoError(t, err)
	assert.Equal(t, "2024-01-01", period.StartDate.Format("2006-01-02"))
}

func TestCreateAttendancePeriod_ValidationErrors(t *testing.T) {
	adminToken := getAdminToken(t, "periodadmin2", "periodpass2")

	testCases := []struct {
		name          string
		payload       fiber.Map
		expectedCode  int
		expectedMsgSubstr string
	}{
		{
			name:          "End date before start date",
			payload:       fiber.Map{"start_date": "2024-02-15", "end_date": "2024-02-10"},
			expectedCode:  http.StatusBadRequest,
			expectedMsgSubstr: "End date must be after start date",
		},
		{
			name:          "Invalid date format",
			payload:       fiber.Map{"start_date": "2024/03/01", "end_date": "2024-03-15"},
			expectedCode:  http.StatusBadRequest,
			expectedMsgSubstr: "Invalid start date format",
		},
		{
			name:          "Missing start_date",
			payload:       fiber.Map{"end_date": "2024-03-15"},
			// Fiber's default binding might allow this if fields are optional,
			// controller logic for parsing time.Parse will fail for empty string.
			// Add `validate:"required"` to payload struct tags for better handling by validator.
			// For now, time.Parse will error.
			expectedCode:  http.StatusBadRequest,
			expectedMsgSubstr: "Invalid start date format",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := makeRequest("POST", "/api/v1/admin/attendance-periods", createJSONBody(tc.payload), adminToken)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedCode, resp.Code)

			var body map[string]interface{}
			json.Unmarshal(resp.Body.Bytes(), &body)
			assert.Contains(t, body["message"], tc.expectedMsgSubstr)
		})
	}
}

func TestCreateAttendancePeriod_Unauthorized(t *testing.T) {
	clearTestData() // No specific user needed, just testing auth

	// 1. No token
	payload := fiber.Map{"start_date": "2024-01-01", "end_date": "2024-01-31"}
	respNoToken, _ := makeRequest("POST", "/api/v1/admin/attendance-periods", createJSONBody(payload))
	assert.Equal(t, http.StatusUnauthorized, respNoToken.Code)

	// 2. Employee token (assuming getEmployeeToken helper exists or created)
	// hashedPassword, _ := utils.HashPassword("emppass")
	// employee := models.Employee{Username: "testemp", Password: hashedPassword, Salary:100}
	// testDB.Create(&employee)
	// empToken := getEmployeeToken(t, "testemp", "emppass") // Placeholder for actual employee token generation
	// respEmpToken, _ := makeRequest("POST", "/api/v1/admin/attendance-periods", createJSONBody(payload), empToken)
	// assert.Equal(t, http.StatusForbidden, respEmpToken.Code)
	// This part is covered in TestProtectedAdminRoute_AccessControl in auth_integration_test.go
}

// TODO: Add tests for RunPayroll and GetPayslipsSummary
// TestRunPayroll will be complex due to its dependencies (employees, attendance, overtime, reimbursements)
// It would require significant setup for each test case.
// Example structure for TestRunPayroll:
/*
func TestRunPayroll_Success(t *testing.T) {
    adminToken := getAdminToken(t, "payrolladmin", "payrollpass")
    clearTestData() // Ensure clean slate

    // 1. Create Admin (done by getAdminToken if it clears)
    var adminUser models.Admin
    testDB.Where("username = ?", "payrolladmin").First(&adminUser)


    // 2. Create Attendance Period
    startDate := time.Date(2024, time.March, 1, 0, 0, 0, 0, time.UTC)
    endDate := time.Date(2024, time.March, 31, 0, 0, 0, 0, time.UTC)
    ipAddr := "127.0.0.1"
    attPeriod := models.AttendancePeriod{
        StartDate: startDate,
        EndDate:   endDate,
        BaseModel: models.BaseModel{CreatedBy: &adminUser.ID, IPAddress: &ipAddr},
    }
    testDB.Create(&attPeriod)

    // 3. Create Employees
    emp1Salary := 60000.0
    emp1 := models.Employee{Username: "emp1payroll", Password: "pw", Salary: emp1Salary, BaseModel: models.BaseModel{CreatedBy: &adminUser.ID}}
    testDB.Create(&emp1)

    // 4. Create Attendance Records for emp1
    // Example: emp1 attended all working days
    totalWorkingDays := utils.CalculateWorkingDays(startDate, endDate)
    currentDate := startDate
    for i := 0; i < int(totalWorkingDays); {
        if currentDate.Weekday() != time.Saturday && currentDate.Weekday() != time.Sunday {
            attRec := models.AttendanceRecord{
                EmployeeID: emp1.ID,
                AttendancePeriodID: attPeriod.ID,
                Date: currentDate,
                CheckInTime: currentDate.Add(9 * time.Hour), // 9 AM check-in
                BaseModel: models.BaseModel{CreatedBy: &emp1.ID},
            }
            testDB.Create(&attRec)
            i++
        }
        currentDate = currentDate.AddDate(0,0,1)
    }

    // 5. (Optional) Create Overtime, Reimbursements

    // 6. Run Payroll
    payrollPayload := fiber.Map{"attendance_period_id": attPeriod.ID.String()}
    resp, err := makeRequest("POST", "/api/v1/admin/payroll", createJSONBody(payrollPayload), adminToken)
    require.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp.Code)

    var body map[string]interface{}
    json.Unmarshal(resp.Body.Bytes(), &body)
    assert.Equal(t, "success", body["status"])

    // 7. Verify Payslip created for emp1
    var payslip models.Payslip
    err = testDB.Where("employee_id = ? AND attendance_period_id = ?", emp1.ID, attPeriod.ID).First(&payslip).Error
    require.NoError(t, err)

    expectedProratedSalary := emp1Salary // Since attended all days
    assert.Equal(t, expectedProratedSalary, payslip.ProratedSalary)
    assert.Equal(t, totalWorkingDays, payslip.AttendanceCount)
    // ... other assertions for overtime, reimbursement, take home pay
}
*/
