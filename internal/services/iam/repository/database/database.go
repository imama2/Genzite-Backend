package repository

import (
	"github.com/imama2/Genzite-Backend/internal/services/iam/repository/database/users"
	"gorm.io/gorm"
)

func New(db *gorm.DB) Repository {
	return users.New(db)
}
