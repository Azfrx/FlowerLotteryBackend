package model

import "time"

type AdminUser struct {
	ID           uint64 `gorm:"primaryKey"`
	Username     string
	DisplayName  string
	PasswordHash string
	Status       uint8
	LastLoginAt  *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
}

func (AdminUser) TableName() string { return "admin_users" }
