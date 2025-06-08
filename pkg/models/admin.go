package models

import "gorm.io/gorm"

// Admin represents an admin user in the system
type Admin struct {
	BaseModel
	Username string `gorm:"type:varchar(255);unique;not null"`
	Password string `gorm:"type:varchar(255);not null"`
}

// BeforeSave hashes the admin's password before saving
func (a *Admin) BeforeSave(tx *gorm.DB) (err error) {
	// TODO: Implement password hashing
	return
}
