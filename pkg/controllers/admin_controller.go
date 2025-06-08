package controllers

import (
	"fmt"
	"payslip-generator/pkg/config"
	"encoding/json"
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

// CreateAttendancePeriodPayload struct for creating attendance period
type CreateAttendancePeriodPayload struct {
	StartDate string `json:"start_date" validate:"required,datetime=2006-01-02"`
	EndDate   string `json:"end_date" validate:"required,datetime=2006-01-02"`
}

// CreateAttendancePeriod godoc
// @Summary Create Attendance Period
// @Description Allows an admin to create a new attendance period.
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param period body CreateAttendancePeriodPayload true "Attendance Period Details"
// @Success 201 {object} object{status=string,data=models.AttendancePeriod} "Successful response with created attendance period"
// @Failure 400 {object} object{status=string,message=string} "Validation error or invalid input"
// @Failure 401 {object} object{status=string,message=string} "Unauthorized - Admin ID not found or invalid token"
// @Failure 500 {object} object{status=string,message=string} "Internal server error"
// @Router /admin/attendance-periods [post]
func CreateAttendancePeriod(c *fiber.Ctx) error {
	var payload CreateAttendancePeriodPayload
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": err.Error()})
	}

	// TODO: Add proper validation using a library like go-playground/validator

	startDate, err := time.Parse("2006-01-02", payload.StartDate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": "Invalid start date format. Use YYYY-MM-DD."})
	}
	endDate, err := time.Parse("2006-01-02", payload.EndDate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": "Invalid end date format. Use YYYY-MM-DD."})
	}

	if endDate.Before(startDate) || endDate.Equal(startDate) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": "End date must be after start date."})
	}

	adminID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": "fail", "message": "Admin ID not found in token or invalid."})
	}
	ipAddress := c.IP()
	requestIDVal := c.Locals(constants.RequestIDKey.String())
	requestID, _ := requestIDVal.(string)

	period := models.AttendancePeriod{
		StartDate: startDate,
		EndDate:   endDate,
	}
	period.CreatedBy = &adminID // Pointer
	period.UpdatedBy = &adminID // Pointer
	period.IPAddress = &ipAddress // Pointer

	if err := database.DB.Create(&period).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": fmt.Sprintf("Could not create attendance period: %v", err)})
	}

	// Audit Log
	auditService := services.NewAuditService(database.DB)
	auditService.CreateAuditLog(services.AuditLogEntryParams{
		UserID:           adminID, // Admin performing the action
		UserType:         "admin",
		Action:           "create_attendance_period",
		TargetResource:   "attendance_period",
		TargetResourceID: period.ID,
		Changes:          period, // Log the created object
		IPAddress:        ipAddress,
		RequestID:        requestID,
		PerformedBy:      adminID,
	})

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"status": "success", "data": period})
}

// RunPayrollPayload struct for running payroll
type RunPayrollPayload struct {
	AttendancePeriodID string `json:"attendance_period_id" validate:"required,uuid"`
}

// RunPayroll godoc
// @Summary Run Payroll
// @Description Allows an admin to run payroll for a specified attendance period.
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payroll_run body RunPayrollPayload true "Payroll Run Details"
// @Success 200 {object} map[string]string `json:"{"status":"success", "message":"Payroll run successfully for period XYZ"}"`
// @Failure 400 {object} map[string]string `json:"{"status":"fail", "message":"error_message (e.g., invalid ID, payroll already run, zero working days)"}"`
// @Failure 401 {object} map[string]string `json:"{"status":"fail", "message":"Admin ID not found in token or invalid."}"`
// @Failure 404 {object} map[string]string `json:"{"status":"fail", "message":"Attendance period not found."}"`
// @Failure 500 {object} map[string]string `json:"{"status":"error", "message":"An internal error occurred during payroll processing."}"`
// @Router /admin/payroll [post]
func RunPayroll(c *fiber.Ctx) error {
	var payload RunPayrollPayload
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": err.Error()})
	}

	periodID, err := uuid.Parse(payload.AttendancePeriodID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": "Invalid AttendancePeriodID format."})
	}

	adminID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": "fail", "message": "Admin ID not found in token or invalid."})
	}
	ipAddress := c.IP()
	requestIDVal := c.Locals(constants.RequestIDKey.String())
	requestID, _ := requestIDVal.(string)
	auditService := services.NewAuditService(database.DB) // auditService for transaction

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		// Re-initialize auditService with transactional DB
		txAuditService := services.NewAuditService(tx)

		var attendancePeriod models.AttendancePeriod
		if err := tx.First(&attendancePeriod, "id = ?", periodID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return fiber.NewError(fiber.StatusNotFound, "Attendance period not found.")
			}
			return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("Database error: %v", err))
		}

		if attendancePeriod.PayrollRunAt != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Payroll already run for this period.")
		}

		var employees []models.Employee
		if err := tx.Find(&employees).Error; err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("Failed to fetch employees: %v", err))
		}

		totalWorkingDays := utils.CalculateWorkingDays(attendancePeriod.StartDate, attendancePeriod.EndDate)
		if totalWorkingDays == 0 { // Avoid division by zero later, implies no possible workdays
			// Update period and log, then return. No payslips generated.
			now := time.Now()
			attendancePeriod.PayrollRunAt = &now
			attendancePeriod.UpdatedBy = &adminID
			attendancePeriod.IPAddress = &ipAddress
			if err := tx.Save(&attendancePeriod).Error; err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("Failed to update attendance period: %v", err))
			}
			// Log audit for payroll run attempt on zero working day period
			txAuditService.CreateAuditLog(services.AuditLogEntryParams{
				UserID:           adminID,
				UserType:         "admin",
				Action:           "run_payroll_zero_working_days",
				TargetResource:   "attendance_period",
				TargetResourceID: attendancePeriod.ID,
				Changes:          map[string]interface{}{"message": "Attempted payroll run for period with zero working days."},
				IPAddress:        ipAddress,
				RequestID:        requestID,
				PerformedBy:      adminID,
			})
			return fiber.NewError(fiber.StatusBadRequest, "Payroll cannot be run for a period with zero total working days.")
		}

		var payslipsGenerated int
		for _, emp := range employees {
			var attendanceRecords []models.AttendanceRecord
			tx.Where("employee_id = ? AND date BETWEEN ? AND ?", emp.ID, attendancePeriod.StartDate, attendancePeriod.EndDate).Find(&attendanceRecords)

			uniqueAttendanceDates := make(map[string]struct{})
			for _, ar := range attendanceRecords {
				uniqueAttendanceDates[ar.Date.Format("2006-01-02")] = struct{}{}
			}
			attendanceCount := len(uniqueAttendanceDates)

			proratedSalary := decimal.NewFromFloat(0)
			if totalWorkingDays > 0 {
				proratedSalary = decimal.NewFromFloat(emp.Salary).Div(decimal.NewFromInt(int64(totalWorkingDays))).Mul(decimal.NewFromInt(int64(attendanceCount)))
			}

			var overtimeRecords []models.OvertimeRecord
			tx.Where("employee_id = ? AND date BETWEEN ? AND ?", emp.ID, attendancePeriod.StartDate, attendancePeriod.EndDate).Find(&overtimeRecords)

			overtimeHours := decimal.NewFromFloat(0)
			overtimePay := decimal.NewFromFloat(0)
			if totalWorkingDays > 0 { // Avoid division by zero if no working days
				dailySalary := decimal.NewFromFloat(emp.Salary).Div(decimal.NewFromInt(int64(totalWorkingDays)))
				hourlySalary := dailySalary.Div(decimal.NewFromInt(8)) // Assuming 8 hour work day

				for _, ot := range overtimeRecords {
					hoursDecimal := decimal.NewFromInt(int64(ot.Hours))
					rateMultiplierDecimal := decimal.NewFromFloat(ot.RateMultiplier)
					currentOvertimePay := hourlySalary.Mul(hoursDecimal).Mul(rateMultiplierDecimal)
					overtimePay = overtimePay.Add(currentOvertimePay)
					overtimeHours = overtimeHours.Add(hoursDecimal)
				}
			}

			var reimbursementRequests []models.ReimbursementRequest
			tx.Where("employee_id = ? AND status = ? AND (attendance_period_id IS NULL OR attendance_period_id = ?)", emp.ID, "approved", periodID).Find(&reimbursementRequests)

			reimbursementsTotal := decimal.NewFromFloat(0)
			for i := range reimbursementRequests { // Use index to modify slice elements
				rr := &reimbursementRequests[i]
				reimbursementsTotal = reimbursementsTotal.Add(decimal.NewFromFloat(rr.Amount))
				rr.AttendancePeriodID = periodID
				rr.Status = "paid"
				rr.UpdatedBy = &adminID
				rr.IPAddress = &ipAddress
				if err := tx.Save(rr).Error; err != nil {
					return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("Failed to update reimbursement request %s: %v", rr.ID, err))
				}
			}

			takeHomePay := proratedSalary.Add(overtimePay).Add(reimbursementsTotal)

			payslip := models.Payslip{
				EmployeeID:          emp.ID,
				AttendancePeriodID:  periodID,
				BaseSalary:          emp.Salary,
				ProratedSalary:      proratedSalary.InexactFloat64(),
				AttendanceCount:     attendanceCount,
				TotalWorkingDays:    totalWorkingDays,
				OvertimeHours:       overtimeHours.InexactFloat64(),
				OvertimePay:         overtimePay.InexactFloat64(),
				ReimbursementsTotal: reimbursementsTotal.InexactFloat64(),
				TakeHomePay:         takeHomePay.InexactFloat64(),
			}
			payslip.CreatedBy = &adminID
			payslip.UpdatedBy = &adminID
			payslip.IPAddress = &ipAddress

			if err := tx.Create(&payslip).Error; err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("Failed to create payslip for employee %s: %v", emp.ID, err))
			}
			payslipsGenerated++
		}

		now := time.Now()
		attendancePeriod.PayrollRunAt = &now
		attendancePeriod.UpdatedBy = &adminID
		attendancePeriod.IPAddress = &ipAddress
		if err := tx.Save(&attendancePeriod).Error; err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("Failed to update attendance period: %v", err))
		}

		// Audit Log for successful payroll run
		txAuditService.CreateAuditLog(services.AuditLogEntryParams{
			UserID:           adminID,
			UserType:         "admin",
			Action:           "run_payroll",
			TargetResource:   "attendance_period",
			TargetResourceID: attendancePeriod.ID,
			Changes:          map[string]interface{}{"payslips_generated": payslipsGenerated, "period_id": attendancePeriod.ID},
			IPAddress:        ipAddress,
			RequestID:        requestID,
			PerformedBy:      adminID,
		})

		return nil // Commit transaction
	})

	if err != nil {
		// Log the original error before returning a generic one to client
		utils.Logger.Error("Payroll transaction failed", zap.Error(err), zap.String("request_id", requestID))
		if fe, ok := err.(*fiber.Error); ok {
			return c.Status(fe.Code).JSON(fiber.Map{"status": "fail", "message": fe.Message})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "An internal error occurred during payroll processing."})
	}
	// Explicitly log success after transaction for clarity
	utils.Logger.Info("Payroll run completed successfully", zap.String("period_id", periodID.String()), zap.String("request_id", requestID))
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "message": "Payroll run successfully for period " + periodID.String()})
}


// GetPayslipsSummaryResponse defines the structure for payslip summary
type GetPayslipsSummaryResponse struct {
	PeriodID                      uuid.UUID        `json:"period_id"`
	Summary                       []PayslipSummary `json:"summary"`
	TotalTakeHomePayAllEmployees float64          `json:"total_take_home_pay_all_employees"`
}

// PayslipSummary holds individual employee payslip info
type PayslipSummary struct {
	EmployeeID   uuid.UUID `json:"employee_id"`
	Username     string    `json:"username"` // Assuming Employee has Username
	TakeHomePay float64   `json:"take_home_pay"`
}

// GetPayslipsSummary godoc
// @Summary Get Payslips Summary
// @Description Allows an admin to retrieve a summary of all payslips for a given attendance period.
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param period_id query string true "Attendance Period ID (UUID)" format(uuid)
// @Success 200 {object} map[string]interface{} `json:"{"status":"success", "data": GetPayslipsSummaryResponse}"`
// @Failure 400 {object} map[string]string `json:"{"status":"fail", "message":"period_id query parameter is required / Invalid period_id format."}"`
// @Failure 401 {object} map[string]string `json:"{"status":"fail", "message":"Unauthorized"}"` // Implicit via middleware
// @Failure 500 {object} map[string]string `json:"{"status":"error", "message":"Failed to fetch payslips: error_message"}"`
// @Router /admin/payslips-summary [get]
func GetPayslipsSummary(c *fiber.Ctx) error {
	periodIDStr := c.Query("period_id")
	if periodIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": "period_id query parameter is required."})
	}
	periodID, err := uuid.Parse(periodIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": "Invalid period_id format."})
	}

	var payslips []models.Payslip
	// Preload Employee to get Username
	if err := database.DB.Preload("Employee").Where("attendance_period_id = ?", periodID).Find(&payslips).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": fmt.Sprintf("Failed to fetch payslips: %v", err)})
	}

	var summaryList []PayslipSummary
	totalTakeHomePay := decimal.NewFromFloat(0)

	for _, p := range payslips {
		summaryList = append(summaryList, PayslipSummary{
			EmployeeID:  p.EmployeeID,
			Username:    p.Employee.Username, // Accessing preloaded employee's username
			TakeHomePay: p.TakeHomePay,
		})
		totalTakeHomePay = totalTakeHomePay.Add(decimal.NewFromFloat(p.TakeHomePay))
	}

	response := GetPayslipsSummaryResponse{
		PeriodID:                      periodID,
		Summary:                       summaryList,
		TotalTakeHomePayAllEmployees: totalTakeHomePay.InexactFloat64(),
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "data": response})
}
