package model

import "time"

type ExchangeOption struct {
	ID          uint64 `gorm:"primaryKey"`
	ActivityID  uint64 `gorm:"column:activity_id"`
	PetalAmount uint64 `gorm:"column:petal_amount"`
	CoinCost    uint64 `gorm:"column:coin_cost"`
	SortNo      int    `gorm:"column:sort_no"`
	Status      uint8
	Remark      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

func (ExchangeOption) TableName() string { return "exchange_options" }

type ExchangeOrder struct {
	ID               uint64 `gorm:"primaryKey"`
	OrderNo          string `gorm:"column:order_no"`
	UserID           uint64 `gorm:"column:user_id"`
	ActivityID       uint64 `gorm:"column:activity_id"`
	ExchangeOptionID uint64 `gorm:"column:exchange_option_id"`
	CoinCost         uint64 `gorm:"column:coin_cost"`
	PetalAmount      uint64 `gorm:"column:petal_amount"`
	Status           uint8
	RequestID        string `gorm:"column:request_id"`
	CreatedAt        time.Time
}

func (ExchangeOrder) TableName() string { return "exchange_orders" }

type AssetTransaction struct {
	ID            uint64  `gorm:"primaryKey"`
	UserID        uint64  `gorm:"column:user_id"`
	ActivityID    *uint64 `gorm:"column:activity_id"`
	AssetType     string  `gorm:"column:asset_type"`
	ChangeAmount  int64   `gorm:"column:change_amount"`
	BalanceBefore int64   `gorm:"column:balance_before"`
	BalanceAfter  int64   `gorm:"column:balance_after"`
	ReasonCode    string  `gorm:"column:reason_code"`
	BizType       string  `gorm:"column:biz_type"`
	BizID         *uint64 `gorm:"column:biz_id"`
	RequestID     string  `gorm:"column:request_id"`
	Remark        string
	CreatedAt     time.Time
}

func (AssetTransaction) TableName() string { return "asset_transactions" }
