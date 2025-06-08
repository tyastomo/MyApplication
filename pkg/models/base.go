package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BaseModel defines the common fields for all models
type BaseModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CreatedAt time.Time  `gorm:"type:timestamptz;not null;default:now()"`
	UpdatedAt time.Time  `gorm:"type:timestamptz;not null;default:now()"`
	CreatedBy *uuid.UUID `gorm:"type:uuid"` // Pointer to allow nil
	UpdatedBy *uuid.UUID `gorm:"type:uuid"` // Pointer to allow nil
	IPAddress *string    `gorm:"type:varchar(255)"` // Pointer to allow nil
}

// BeforeCreate will set a UUID if it's not already set.
func (base *BaseModel) BeforeCreate(tx *gorm.DB) (err error) {
	if base.ID == uuid.Nil { // Only generate a new UUID if ID is not already set
		base.ID = uuid.New()
	}
	// CreatedAt and UpdatedAt are handled by default:now() or GORM's BeforeSave/BeforeUpdate hooks if needed for application-time setting
	return
}

// It's also common to use BeforeSave to update UpdatedAt timestamp
// func (base *BaseModel) BeforeSave(tx *gorm.DB) (err error) {
//   now := time.Now()
//   base.UpdatedAt = now
//   if base.CreatedAt.IsZero() { // Set CreatedAt only if it's not set
//    base.CreatedAt = now
//   }
//   return
// }
// However, `default:now()` in struct tags for CreatedAt and UpdatedAt usually suffices for DB level defaults.
