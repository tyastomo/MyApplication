package models

import "gorm.io/gorm"

// Employee represents an employee in the system
type Employee struct {
	BaseModel
	Username string  `gorm:"type:varchar(255);unique;not null"`
	Password string  `gorm:"type:varchar(255);not null"`
	Salary   float64 `gorm:"type:decimal(10,2);not null"`
}

// BeforeSave hashes the employee's password before saving
func (e *Employee) BeforeSave(tx *gorm.DB) (err error) {
	// TODO: Implement password hashing
	return
}
