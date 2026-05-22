package entities

import "gorm.io/gorm"

type Users struct {
	gorm.Model
	UUID         string `gorm:"uniqueIndex;not null"`
	Email        string `gorm:"uniqueIndex;not null"`
	Name         string
	PasswordHash *string `gorm:"column:password_hash"`
	GoogleID     *string `gorm:"column:google_id;uniqueIndex"`
	Roles        []Roles `gorm:"many2many:iam.user_roles;"`
}

func (Users) TableName() string {
	return "iam.users"
}
