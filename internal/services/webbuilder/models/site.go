package models

import (
	"time"

	"gorm.io/gorm"
)

type Site struct {
	gorm.Model
	UserID      uint   `gorm:"column:user_id;index;not null"`
	Slug        string `gorm:"column:slug;uniqueIndex;not null"`
	Title       string
	ConfigJSON  string     `gorm:"column:config;type:jsonb;not null"`
	Status      string     `gorm:"column:status;not null"`
	OutputPath  string     `gorm:"column:output_path;not null"`
	PublishedAt *time.Time `gorm:"column:published_at"`
}

func (Site) TableName() string {
	return "web_builder.web_builder_sites"
}
