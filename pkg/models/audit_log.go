package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// AuditLog records actions performed in the system for auditing purposes
type AuditLog struct {
	ID               uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Timestamp        time.Time      `gorm:"type:timestamptz;not null;default:now()"`
	UserID           uuid.UUID      `gorm:"type:uuid"` // Nullable, as some actions might be system-initiated
	UserType         string         `gorm:"type:varchar(50)"` // 'employee', 'admin', or 'system'
	Action           string         `gorm:"type:varchar(255);not null"`
	TargetResource   string         `gorm:"type:varchar(255)"`
	TargetResourceID uuid.UUID      `gorm:"type:uuid"`
	Changes          datatypes.JSON `gorm:"type:jsonb"` // Or json depending on PostgreSQL version and preference
	IPAddress        string         `gorm:"type:varchar(255)"`
	RequestID        string         `gorm:"type:varchar(255)"`
}

// TableName specifies the table name for AuditLog
func (AuditLog) TableName() string {
	return "audit_logs"
}
