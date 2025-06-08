package models

import (
	"github.com/google/uuid"
)

// Payslip represents an employee's payslip for a specific period
type Payslip struct {
	BaseModel
	EmployeeID         uuid.UUID `gorm:"type:uuid;not null"`
	AttendancePeriodID uuid.UUID `gorm:"type:uuid;not null"`
	BaseSalary         float64   `gorm:"type:decimal(10,2);not null"`
	ProratedSalary     float64   `gorm:"type:decimal(10,2);not null"`
	AttendanceCount    int       `gorm:"type:integer;not null"`
	TotalWorkingDays   int       `gorm:"type:integer;not null"`
	OvertimeHours      float64   `gorm:"type:decimal(4,2);default:0"`
	OvertimePay        float64   `gorm:"type:decimal(10,2);default:0"`
	ReimbursementsTotal float64   `gorm:"type:decimal(10,2);default:0"`
	TakeHomePay        float64   `gorm:"type:decimal(10,2);not null"`

	Employee         Employee         `gorm:"foreignKey:EmployeeID"`
	AttendancePeriod AttendancePeriod `gorm:"foreignKey:AttendancePeriodID"`
}

// TableName specifies the table name for Payslip
func (Payslip) TableName() string {
	return "payslips"
}

// We will add the unique constraint `uix_employee_period` for (`EmployeeID`, `AttendancePeriodID`)
// during the auto-migration process in `database.go`.
