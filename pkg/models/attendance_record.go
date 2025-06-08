package models

import (
	"time"

	"github.com/google/uuid"
)

// AttendanceRecord represents an employee's attendance for a specific day
type AttendanceRecord struct {
	BaseModel
	EmployeeID         uuid.UUID `gorm:"type:uuid;not null"`
	AttendancePeriodID uuid.UUID `gorm:"type:uuid;not null"`
	Date               time.Time `gorm:"type:date;not null"`
	CheckInTime        time.Time `gorm:"type:timestamptz;not null"`

	Employee         Employee         `gorm:"foreignKey:EmployeeID"`
	AttendancePeriod AttendancePeriod `gorm:"foreignKey:AttendancePeriodID"`
}

// TableName specifies the table name for AttendanceRecord and adds a unique constraint
func (AttendanceRecord) TableName() string {
	return "attendance_records"
}

// GORM V2 uses a different way to specify composite unique constraints.
// This is typically done during migration using db.Migrator().CreateConstraint(&AttendanceRecord{}, "uix_employee_date")
// or directly in the struct tag for simpler cases, but composite foreign keys need explicit handling.
// For now, we'll rely on manual constraint creation or a migration tool.
// GORM's default behavior for struct tags like `gorm:"uniqueIndex:uix_employee_date"` might work for some databases.
// However, for cross-database compatibility and clarity, explicit constraint creation during migration is often preferred.

// We will add the unique constraint `uix_employee_date` for (`EmployeeID`, `Date`)
// during the auto-migration process in `database.go`.
