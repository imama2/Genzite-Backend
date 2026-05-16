package models

import "gorm.io/gorm"

type Permission struct {
	gorm.Model
	Name        string `gorm:"uniqueIndex;not null"`
	Description string
}

func (Permission) TableName() string {
	return "iam.permissions"
}
