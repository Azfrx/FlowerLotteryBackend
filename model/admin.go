package model

import "time"

type AdminUser struct {
	ID           uint64     `gorm:"primaryKey" json:"id"`
	Username     string     `json:"username"`
	DisplayName  string     `json:"display_name"`
	PasswordHash string     `json:"-"`
	Status       uint8      `json:"status"`
	LastLoginAt  *time.Time `json:"last_login_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"-"`
}

func (AdminUser) TableName() string { return "admin_users" }

type AdminOperationLog struct {
	ID           uint64    `gorm:"primaryKey" json:"id"`
	AdminUserID  uint64    `gorm:"column:admin_user_id" json:"admin_user_id"`
	RequestID    string    `gorm:"column:request_id" json:"request_id"`
	Method       string    `json:"method"`
	Path         string    `json:"path"`
	Action       string    `json:"action"`
	TargetType   string    `gorm:"column:target_type" json:"target_type"`
	TargetID     string    `gorm:"column:target_id" json:"target_id"`
	RequestBody  []byte    `gorm:"column:request_body" json:"-"`
	ResponseCode int       `gorm:"column:response_code" json:"response_code"`
	IP           string    `gorm:"column:ip" json:"ip"`
	UserAgent    string    `gorm:"column:user_agent" json:"user_agent"`
	DurationMS   uint      `gorm:"column:duration_ms" json:"duration_ms"`
	CreatedAt    time.Time `json:"created_at"`
}

func (AdminOperationLog) TableName() string { return "admin_operation_logs" }
