package models

import (
	"time"

	"github.com/google/uuid"
)

// OvertimeRecord represents an employee's overtime request
type OvertimeRecord struct {
	BaseModel
	EmployeeID     uuid.UUID `gorm:"type:uuid;not null"`
	Date           time.Time `gorm:"type:date;not null"`
	Hours          int       `gorm:"type:integer;not null"` // Consider adding validation for max 3
	SubmittedAt    time.Time `gorm:"type:timestamptz;not null;default:now()"`
	RateMultiplier float64   `gorm:"type:decimal(3,2);default:2.0"`

	Employee Employee `gorm:"foreignKey:EmployeeID"`
}

// TableName specifies the table name for OvertimeRecord
func (OvertimeRecord) TableName() string {
	return "overtime_records"
}

// You might want to add a check constraint for Hours <= 3 directly in PostgreSQL.
// GORM allows creating CHECK constraints during migration:
// db.Migrator().CreateConstraint(&OvertimeRecord{}, "ck_overtime_hours")
// And the constraint would be defined in the struct tag like:
// Hours int `gorm:"type:integer;not null;check:hours <= 3"`
// However, direct support for CHECK constraints in struct tags can be database-dependent.
// For now, we will rely on application-level validation or manual DB constraint.
