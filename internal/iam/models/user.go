package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Email        string `gorm:"uniqueIndex;not null"`
	Name         string
	PasswordHash *string `gorm:"column:password_hash"`
	GoogleID     *string `gorm:"column:google_id;uniqueIndex"`
	Roles        []Role  `gorm:"many2many:user_roles;"`
}
