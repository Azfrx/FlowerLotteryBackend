package model

import "time"

type User struct {
	ID           uint64     `gorm:"primaryKey" json:"id"`
	UserNo       string     `gorm:"column:user_no;size:64;uniqueIndex" json:"user_id"`
	Nickname     string     `gorm:"size:64" json:"nickname"`
	AvatarURL    string     `gorm:"column:avatar_url;size:512" json:"avatar_url"`
	PasswordHash string     `gorm:"column:password_hash;size:255" json:"-"`
	Status       uint8      `json:"status"`
	LastLoginAt  *time.Time `gorm:"column:last_login_at" json:"last_login_at"`
	Remark       string     `json:"-"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `gorm:"index" json:"-"`
}

func (User) TableName() string { return "users" }

type UserWallet struct {
	ID           uint64 `gorm:"primaryKey"`
	UserID       uint64 `gorm:"column:user_id;uniqueIndex"`
	CoinBalance  int64  `gorm:"column:coin_balance"`
	PetalBalance int64  `gorm:"column:petal_balance"`
	Version      uint64
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (UserWallet) TableName() string { return "user_wallets" }

type RefreshToken struct {
	ID          uint64     `gorm:"primaryKey"`
	SubjectType string     `gorm:"column:subject_type"`
	SubjectID   uint64     `gorm:"column:subject_id"`
	TokenHash   string     `gorm:"column:token_hash"`
	ExpiresAt   time.Time  `gorm:"column:expires_at"`
	RevokedAt   *time.Time `gorm:"column:revoked_at"`
	CreatedAt   time.Time
}

func (RefreshToken) TableName() string { return "refresh_tokens" }
