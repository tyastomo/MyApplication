package controllers

import (
	"fmt"
	"payslip-generator/pkg/constants"
	"payslip-generator/pkg/database"
	"payslip-generator/pkg/models"
	"payslip-generator/pkg/services" // Added
	"payslip-generator/pkg/utils"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// SubmitAttendance godoc
// @Summary Submit Employee Attendance
// @Description Allows an authenticated employee to submit their attendance for the current day.
// @Tags Employee
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 201 {object} map[string]interface{} `json:"{"status":"success", "data": models.AttendanceRecord}"`
// @Failure 400 {object} map[string]string `json:"{"status":"fail", "message":"error_message (e.g., weekend, no active period, payroll run)"}"`
// @Failure 401 {object} map[string]string `json:"{"status":"fail", "message":"User not authenticated."}"`
// @Failure 409 {object} map[string]string `json:"{"status":"fail", "message":"Attendance already submitted for today."}"`
// @Failure 500 {object} map[string]string `json:"{"status":"error", "message":"Database error / Could not submit attendance"}"`
// @Router /employee/attendance [post]
func SubmitAttendance(c *fiber.Ctx) error {
	employeeID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": "fail", "message": "User not authenticated."})
	}
	ipAddress := c.IP()
	requestIDVal := c.Locals(constants.RequestIDKey.String())
	requestID, _ := requestIDVal.(string)
	auditService := services.NewAuditService(database.DB)
	today := time.Now() // Consider standardizing timezone, e.g., today = time.Now().UTC()

	// Check if today is a weekend
	weekday := today.Weekday()
	if weekday == time.Saturday || weekday == time.Sunday {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": "Attendance submission not allowed on weekends."})
	}

	// Find active attendance period
	var activePeriod models.AttendancePeriod
	dateOnlyToday := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())

	err = database.DB.Where("start_date <= ? AND end_date >= ? AND payroll_run_at IS NULL", dateOnlyToday, dateOnlyToday).First(&activePeriod).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": "No active attendance period for today, or payroll has been run."})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Database error finding attendance period."})
	}
    // Redundant check, already in query: if activePeriod.PayrollRunAt != nil {
	// 	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": "Payroll has already been run for the current period."})
	// }


	// Check if attendance already submitted for today
	var existingRecord models.AttendanceRecord
	err = database.DB.Where("employee_id = ? AND date = ?", employeeID, dateOnlyToday).First(&existingRecord).Error
	if err == nil { // Record found
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"status": "fail", "message": "Attendance already submitted for today."})
	}
	if err != gorm.ErrRecordNotFound { // Some other database error
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Database error checking existing attendance."})
	}

	attendanceRecord := models.AttendanceRecord{
		EmployeeID:         employeeID,
		AttendancePeriodID: activePeriod.ID,
		Date:               dateOnlyToday,
		CheckInTime:        today, // Full timestamp for check-in
	}
	attendanceRecord.CreatedBy = &employeeID // Pointer
	attendanceRecord.UpdatedBy = &employeeID // Pointer
	attendanceRecord.IPAddress = &ipAddress  // Pointer

	if err := database.DB.Create(&attendanceRecord).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": fmt.Sprintf("Could not submit attendance: %v", err)})
	}

	// Audit Log
	auditService.CreateAuditLog(services.AuditLogEntryParams{
		UserID:           employeeID,
		UserType:         "employee",
		Action:           "submit_attendance",
		TargetResource:   "attendance_record",
		TargetResourceID: attendanceRecord.ID,
		Changes:          attendanceRecord,
		IPAddress:        ipAddress,
		RequestID:        requestID,
		PerformedBy:      employeeID,
	})

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"status": "success", "data": attendanceRecord})
}

// SubmitOvertimePayload struct for submitting overtime
type SubmitOvertimePayload struct {
	Date  string `json:"date" validate:"required,datetime=2006-01-02"`
	Hours int    `json:"hours" validate:"required,min=1,max=3"` // Example: Using struct tags for validation info in docs
}

// SubmitOvertime godoc
// @Summary Submit Employee Overtime
// @Description Allows an authenticated employee to submit an overtime record.
// @Tags Employee
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param overtime_details body SubmitOvertimePayload true "Overtime Details"
// @Success 201 {object} map[string]interface{} `json:"{"status":"success", "data": models.OvertimeRecord}"`
// @Failure 400 {object} map[string]string `json:"{"status":"fail", "message":"error_message (e.g., invalid hours, invalid date, submission timing, no attendance, payroll run)"}"`
// @Failure 401 {object} map[string]string `json:"{"status":"fail", "message":"User not authenticated."}"`
// @Failure 500 {object} map[string]string `json:"{"status":"error", "message":"Database error / Could not submit overtime"}"`
// @Router /employee/overtime [post]
func SubmitOvertime(c *fiber.Ctx) error {
	var payload SubmitOvertimePayload
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": err.Error()})
	}
	// TODO: Add proper validation using a library like go-playground/validator

	if payload.Hours <= 0 || payload.Hours > 3 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": "Overtime hours must be between 1 and 3."})
	}

	overtimeDate, err := time.Parse("2006-01-02", payload.Date)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": "Invalid date format. Use YYYY-MM-DD."})
	}

	employeeID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": "fail", "message": "User not authenticated."})
	}
	ipAddress := c.IP()
	requestIDVal := c.Locals(constants.RequestIDKey.String())
	requestID, _ := requestIDVal.(string)
	auditService := services.NewAuditService(database.DB)
	now := time.Now()

	// Validation: If Date is today, current time must be after 5 PM.
	if overtimeDate.Year() == now.Year() && overtimeDate.Month() == now.Month() && overtimeDate.Day() == now.Day() {
		if now.Hour() < 17 { // 17:00 is 5 PM
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": "Overtime for today can only be submitted after 5 PM."})
		}
	}
    if overtimeDate.After(now) {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": "Overtime date cannot be in the future."})
    }


	// Check if an AttendanceRecord exists for this employee on the given Date
	var attendanceRecord models.AttendanceRecord
	err = database.DB.Where("employee_id = ? AND date = ?", employeeID, overtimeDate).First(&attendanceRecord).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": "Cannot submit overtime for a day you did not attend."})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Database error checking attendance for overtime."})
	}

	// Find the AttendancePeriod for the overtime Date. If payroll for that period is already run, disallow submission.
	var attendancePeriod models.AttendancePeriod
	err = database.DB.Where("start_date <= ? AND end_date >= ? AND payroll_run_at IS NULL", overtimeDate, overtimeDate).First(&attendancePeriod).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": "No active attendance period for the overtime date, or payroll has been run."})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Database error finding attendance period for overtime."})
	}


	overtimeRecord := models.OvertimeRecord{
		EmployeeID:     employeeID,
		Date:           overtimeDate,
		Hours:          payload.Hours,
		SubmittedAt:    now,
		RateMultiplier: 2.0, // Default, can be made configurable if needed
	}
	overtimeRecord.CreatedBy = &employeeID // Pointer
	overtimeRecord.UpdatedBy = &employeeID // Pointer
	overtimeRecord.IPAddress = &ipAddress  // Pointer

	if err := database.DB.Create(&overtimeRecord).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": fmt.Sprintf("Could not submit overtime: %v", err)})
	}

	// Audit Log
	auditService.CreateAuditLog(services.AuditLogEntryParams{
		UserID:           employeeID,
		UserType:         "employee",
		Action:           "submit_overtime",
		TargetResource:   "overtime_record",
		TargetResourceID: overtimeRecord.ID,
		Changes:          overtimeRecord,
		IPAddress:        ipAddress,
		RequestID:        requestID,
		PerformedBy:      employeeID,
	})

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"status": "success", "data": overtimeRecord})
}


// SubmitReimbursementPayload for reimbursement requests
type SubmitReimbursementPayload struct {
	Amount      float64 `json:"amount" validate:"required,gt=0"`
	Description string  `json:"description" validate:"required"`
}

// SubmitReimbursement godoc
// @Summary Submit Employee Reimbursement Request
// @Description Allows an authenticated employee to submit a reimbursement request.
// @Tags Employee
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param reimbursement_details body SubmitReimbursementPayload true "Reimbursement Details"
// @Success 201 {object} map[string]interface{} `json:"{"status":"success", "data": models.ReimbursementRequest}"`
// @Failure 400 {object} map[string]string `json:"{"status":"fail", "message":"error_message (e.g., invalid amount, missing description)"}"`
// @Failure 401 {object} map[string]string `json:"{"status":"fail", "message":"User not authenticated."}"`
// @Failure 500 {object} map[string]string `json:"{"status":"error", "message":"Could not submit reimbursement request"}"`
// @Router /employee/reimbursements [post]
func SubmitReimbursement(c *fiber.Ctx) error {
	var payload SubmitReimbursementPayload
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": err.Error()})
	}
	// TODO: Add proper validation using a library

	if payload.Amount <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": "Amount must be greater than zero."})
	}
	if payload.Description == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": "Description is required."})
	}

	employeeID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": "fail", "message": "User not authenticated."})
	}
	ipAddress := c.IP()
	requestIDVal := c.Locals(constants.RequestIDKey.String())
	requestID, _ := requestIDVal.(string)
	auditService := services.NewAuditService(database.DB)

	reimbursementRequest := models.ReimbursementRequest{
		EmployeeID:  employeeID, // This is the FK, not part of BaseModel's CreatedBy
		Amount:      payload.Amount,
		Description: payload.Description,
		Status:      "pending", // Default status
	}
	reimbursementRequest.CreatedBy = &employeeID // Pointer for BaseModel field
	reimbursementRequest.UpdatedBy = &employeeID // Pointer for BaseModel field
	reimbursementRequest.IPAddress = &ipAddress  // Pointer for BaseModel field

	// Optionally, try to link to an active, non-payroll-run period
	var activePeriod models.AttendancePeriod
	today := time.Now()
	dateOnlyToday := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	if err := database.DB.Where("start_date <= ? AND end_date >= ? AND payroll_run_at IS NULL", dateOnlyToday, dateOnlyToday).First(&activePeriod).Error; err == nil {
		// Note: AttendancePeriodID in ReimbursementRequest is uuid.UUID, not *uuid.UUID
		// So, if activePeriod.ID is uuid.Nil (from a zero-value struct if not found), it will be stored as such.
		// This is fine as it's nullable in the DB or will be the actual ID.
		reimbursementRequest.AttendancePeriodID = activePeriod.ID
	}


	if err := database.DB.Create(&reimbursementRequest).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": fmt.Sprintf("Could not submit reimbursement request: %v", err)})
	}

	// Audit Log
	auditService.CreateAuditLog(services.AuditLogEntryParams{
		UserID:           employeeID,
		UserType:         "employee",
		Action:           "submit_reimbursement",
		TargetResource:   "reimbursement_request",
		TargetResourceID: reimbursementRequest.ID,
		Changes:          reimbursementRequest,
		IPAddress:        ipAddress,
		RequestID:        requestID,
		PerformedBy:      employeeID,
	})

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"status": "success", "data": reimbursementRequest})
}

// PayslipDetailResponse structure for employee's view of their payslip
type PayslipDetailResponse struct {
	PayslipID                    uuid.UUID `json:"payslip_id"`
	PeriodStartDate              string    `json:"period_start_date"`
	PeriodEndDate                string    `json:"period_end_date"`
	BaseSalary                   float64   `json:"base_salary"`
	ProratedSalary               float64   `json:"prorated_salary"`
	AttendanceCount              int       `json:"attendance_count"`
	TotalWorkingDaysInPeriod     int       `json:"total_working_days_in_period"`
	OvertimeHours                float64   `json:"overtime_hours"`
	OvertimePay                  float64   `json:"overtime_pay"`
	Reimbursements               []models.ReimbursementRequest `json:"reimbursements"` // List of actual RRs
	TotalReimbursements          float64   `json:"total_reimbursements"` // This should be sum of Reimbursements array amounts
	TakeHomePay                  float64   `json:"take_home_pay"`
}

// GetMyPayslip godoc
// @Summary Get Employee Payslip
// @Description Allows an authenticated employee to retrieve their own payslip for a specified period.
// @Tags Employee
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param period_id query string true "Attendance Period ID (UUID) for the payslip" format(uuid)
// @Success 200 {object} map[string]interface{} `json:"{"status":"success", "data": PayslipDetailResponse}"`
// @Failure 400 {object} map[string]string `json:"{"status":"fail", "message":"period_id query parameter is required / Invalid period_id format."}"`
// @Failure 401 {object} map[string]string `json:"{"status":"fail", "message":"User not authenticated."}"`
// @Failure 404 {object} map[string]string `json:"{"status":"fail", "message":"Payslip not found for this period."}"`
// @Failure 500 {object} map[string]string `json:"{"status":"error", "message":"Database error"}"`
// @Router /employee/payslip [get]
func GetMyPayslip(c *fiber.Ctx) error {
	employeeID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": "fail", "message": "User not authenticated."})
	}

	periodIDStr := c.Query("period_id")
	if periodIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": "period_id query parameter is required."})
	}
	periodID, err := uuid.Parse(periodIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": "Invalid period_id format."})
	}

	var payslip models.Payslip
	err = database.DB.Preload("AttendancePeriod").
		Where("employee_id = ? AND attendance_period_id = ?", employeeID, periodID).
		First(&payslip).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"status": "fail", "message": "Payslip not found for this period."})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": fmt.Sprintf("Database error: %v", err)})
	}

	var paidReimbursements []models.ReimbursementRequest
	err = database.DB.Where("employee_id = ? AND attendance_period_id = ? AND status = ?", employeeID, periodID, "paid").Find(&paidReimbursements).Error
	if err != nil {
		// Log this error but don't fail the request, as payslip itself was found
		fmt.Printf("Error fetching paid reimbursements for payslip %s: %v\n", payslip.ID, err)
	}

    // Recalculate total reimbursements from the fetched list for accuracy, though payslip.ReimbursementsTotal should be correct.
    actualReimbursementsTotal := decimal.NewFromFloat(0)
    for _, rr := range paidReimbursements {
        actualReimbursementsTotal = actualReimbursementsTotal.Add(decimal.NewFromFloat(rr.Amount))
    }


	response := PayslipDetailResponse{
		PayslipID:                    payslip.ID,
		PeriodStartDate:              payslip.AttendancePeriod.StartDate.Format("2006-01-02"),
		PeriodEndDate:                payslip.AttendancePeriod.EndDate.Format("2006-01-02"),
		BaseSalary:                   payslip.BaseSalary, // This is the employee's salary at the time of payroll run
		ProratedSalary:               payslip.ProratedSalary,
		AttendanceCount:              payslip.AttendanceCount,
		TotalWorkingDaysInPeriod:     payslip.TotalWorkingDays,
		OvertimeHours:                payslip.OvertimeHours,
		OvertimePay:                  payslip.OvertimePay,
		Reimbursements:               paidReimbursements,
		TotalReimbursements:          actualReimbursementsTotal.InexactFloat64(), // Use sum from actual RRs
		TakeHomePay:                  payslip.TakeHomePay,
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "data": response})
}
