package models

import "gorm.io/gorm"

type Role struct {
	gorm.Model
	Name        string `gorm:"uniqueIndex;not null"`
	Description string
	Permissions []Permission `gorm:"many2many:iam.role_permissions;"`
}

func (Role) TableName() string {
	return "iam.roles"
}
