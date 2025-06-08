package models

import "time"

// AttendancePeriod represents a payroll period
type AttendancePeriod struct {
	BaseModel
	StartDate    time.Time  `gorm:"type:date;not null"`
	EndDate      time.Time  `gorm:"type:date;not null"`
	PayrollRunAt *time.Time `gorm:"type:timestamptz"`
}
