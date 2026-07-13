package repository

import (
	"errors"
	"flower-lottery-backend/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type WalletRepository interface {
	Get(userID uint64) (*model.UserWallet, error)
	ListOptions() ([]model.ExchangeOption, error)
	Exchange(userID, optionID uint64, requestID, orderNo string) (*model.ExchangeOrder, *model.UserWallet, error)
	ListTransactions(userID uint64, page, pageSize int) ([]model.AssetTransaction, int64, error)
}
type walletRepository struct{ db *gorm.DB }

func NewWalletRepository(db *gorm.DB) WalletRepository { return &walletRepository{db: db} }
func (r *walletRepository) Get(userID uint64) (*model.UserWallet, error) {
	var v model.UserWallet
	err := r.db.Where("user_id = ?", userID).First(&v).Error
	return &v, err
}
func (r *walletRepository) ListOptions() ([]model.ExchangeOption, error) {
	var v []model.ExchangeOption
	err := r.db.Where("status = 1 AND deleted_at IS NULL").Order("sort_no,id").Find(&v).Error
	return v, err
}
func (r *walletRepository) Exchange(userID, optionID uint64, requestID, orderNo string) (*model.ExchangeOrder, *model.UserWallet, error) {
	var order model.ExchangeOrder
	var wallet model.UserWallet
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ? AND request_id = ?", userID, requestID).First(&order).Error; err == nil {
			return gorm.ErrDuplicatedKey
		} else if err != gorm.ErrRecordNotFound {
			return err
		}
		var option model.ExchangeOption
		if err := tx.Where("id = ? AND status = 1 AND deleted_at IS NULL", optionID).First(&option).Error; err != nil {
			return err
		}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ?", userID).First(&wallet).Error; err != nil {
			return err
		}
		if wallet.CoinBalance < int64(option.CoinCost) {
			return ErrInsufficientBalance
		}
		coinBefore := wallet.CoinBalance
		petalBefore := wallet.PetalBalance
		wallet.CoinBalance -= int64(option.CoinCost)
		wallet.PetalBalance += int64(option.PetalAmount)
		wallet.Version++
		if err := tx.Save(&wallet).Error; err != nil {
			return err
		}
		order = model.ExchangeOrder{OrderNo: orderNo, UserID: userID, ActivityID: option.ActivityID, ExchangeOptionID: option.ID, CoinCost: option.CoinCost, PetalAmount: option.PetalAmount, Status: 1, RequestID: requestID}
		if err := tx.Create(&order).Error; err != nil {
			return err
		}
		biz := order.ID
		activity := option.ActivityID
		rows := []model.AssetTransaction{{UserID: userID, ActivityID: &activity, AssetType: "coin", ChangeAmount: -int64(option.CoinCost), BalanceBefore: coinBefore, BalanceAfter: wallet.CoinBalance, ReasonCode: "petal_exchange", BizType: "exchange", BizID: &biz, RequestID: requestID}, {UserID: userID, ActivityID: &activity, AssetType: "petal", ChangeAmount: int64(option.PetalAmount), BalanceBefore: petalBefore, BalanceAfter: wallet.PetalBalance, ReasonCode: "petal_exchange", BizType: "exchange", BizID: &biz, RequestID: requestID}}
		return tx.Create(&rows).Error
	})
	return &order, &wallet, err
}
func (r *walletRepository) ListTransactions(userID uint64, page, pageSize int) ([]model.AssetTransaction, int64, error) {
	var list []model.AssetTransaction
	var total int64
	q := r.db.Model(&model.AssetTransaction{}).Where("user_id = ?", userID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

var ErrInsufficientBalance = errors.New("insufficient balance")
