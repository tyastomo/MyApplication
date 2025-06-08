package models

import (
	"github.com/google/uuid"
)

// ReimbursementRequest represents an employee's request for reimbursement
type ReimbursementRequest struct {
	BaseModel
	EmployeeID         uuid.UUID `gorm:"type:uuid;not null"`
	AttendancePeriodID uuid.UUID `gorm:"type:uuid"` // Nullable
	Description        string    `gorm:"type:text;not null"`
	Amount             float64   `gorm:"type:decimal(10,2);not null"`
	Status             string    `gorm:"type:varchar(50);default:'pending'"` // e.g., pending, approved, rejected, paid

	Employee         Employee         `gorm:"foreignKey:EmployeeID"`
	AttendancePeriod AttendancePeriod `gorm:"foreignKey:AttendancePeriodID"`
}

// TableName specifies the table name for ReimbursementRequest
func (ReimbursementRequest) TableName() string {
	return "reimbursement_requests"
}
