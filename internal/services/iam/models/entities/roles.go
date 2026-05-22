package entities

import "gorm.io/gorm"

type Roles struct {
	gorm.Model
	Name        string `gorm:"uniqueIndex;not null"`
	Description string
	Permissions []Permissions `gorm:"many2many:iam.role_permissions;"`
}

func (Roles) TableName() string {
	return "iam.roles"
}
