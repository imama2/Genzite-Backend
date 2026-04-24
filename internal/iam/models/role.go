package models

import "gorm.io/gorm"

type Role struct {
	gorm.Model
	Name        string       `gorm:"uniqueIndex;not null"`
	Description string
	Permissions []Permission `gorm:"many2many:role_permissions;"`
}
