package entities

import "gorm.io/gorm"

type Permissions struct {
	gorm.Model
	Name        string `gorm:"uniqueIndex;not null"`
	Description string
}

func (Permissions) TableName() string {
	return "iam.permissions"
}
