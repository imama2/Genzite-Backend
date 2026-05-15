package models

import "gorm.io/gorm"

type Payment struct {
	gorm.Model
	UserID      uint   `gorm:"column:user_id;index;not null"`
	SiteID      uint   `gorm:"column:site_id;index;not null"`
	OrderID     string `gorm:"column:order_id;uniqueIndex;not null"`
	Amount      int64  `gorm:"column:amount;not null"`
	Status      string `gorm:"column:status;not null"`
	Provider    string `gorm:"column:provider;not null"`
	RedirectURL string `gorm:"column:redirect_url"`
	SnapToken   string `gorm:"column:snap_token"`
}

func (Payment) TableName() string {
	return "payments"
}
