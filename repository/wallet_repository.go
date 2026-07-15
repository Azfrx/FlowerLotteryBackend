package repository

import (
	"errors"
	"flower-lottery-backend/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type WalletRepository interface {
	Get(userID uint64) (*model.UserWallet, error)
	ListOptions() ([]model.ExchangeOption, error)
	Exchange(userID, optionID, expectedPetalAmount, expectedCoinCost uint64, requestID, orderNo string) (*model.ExchangeOrder, *model.UserWallet, error)
	PetalGiftPackPurchased(userID uint64) (bool, error)
	PurchasePetalGiftPack(userID, petalAmount uint64, requestID string) (*model.UserWallet, error)
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
	activityID, err := currentActivityID(r.db)
	if err != nil {
		return nil, err
	}
	err = r.db.Where("activity_id = ? AND status = 1 AND deleted_at IS NULL", activityID).
		Order("sort_no,id").Find(&v).Error
	return v, err
}
func (r *walletRepository) Exchange(userID, optionID, expectedPetalAmount, expectedCoinCost uint64, requestID, orderNo string) (*model.ExchangeOrder, *model.UserWallet, error) {
	var order model.ExchangeOrder
	var wallet model.UserWallet
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ? AND request_id = ?", userID, requestID).First(&order).Error; err == nil {
			return gorm.ErrDuplicatedKey
		} else if err != gorm.ErrRecordNotFound {
			return err
		}
		activityID, err := currentActivityID(tx)
		if err != nil {
			return err
		}
		var option model.ExchangeOption
		if err := tx.Where("id = ? AND activity_id = ? AND status = 1 AND deleted_at IS NULL", optionID, activityID).First(&option).Error; err != nil {
			return err
		}
		if option.PetalAmount != expectedPetalAmount || option.CoinCost != expectedCoinCost {
			return ErrExchangeOptionChanged
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

func currentActivityID(db *gorm.DB) (uint64, error) {
	var activity model.Activity
	now := time.Now()
	if err := db.Where(
		"status = 2 AND starts_at <= ? AND ends_at > ? AND deleted_at IS NULL",
		now,
		now,
	).First(&activity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, ErrActivityNotPlayable
		}
		return 0, err
	}
	return activity.ID, nil
}

func (r *walletRepository) PetalGiftPackPurchased(userID uint64) (bool, error) {
	activityID, err := currentActivityID(r.db)
	if err != nil {
		return false, err
	}
	var count int64
	err = r.db.Model(&model.AssetTransaction{}).
		Where("user_id=? AND activity_id=? AND reason_code=?", userID, activityID, "petal_gift_pack").
		Count(&count).Error
	return count > 0, err
}

func (r *walletRepository) PurchasePetalGiftPack(userID, petalAmount uint64, requestID string) (*model.UserWallet, error) {
	var wallet model.UserWallet
	err := r.db.Transaction(func(tx *gorm.DB) error {
		activityID, err := currentActivityID(tx)
		if err != nil {
			return err
		}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id=?", userID).First(&wallet).Error; err != nil {
			return err
		}
		var existingPurchase model.AssetTransaction
		err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id=? AND activity_id=? AND reason_code=?", userID, activityID, "petal_gift_pack").
			Order("id DESC").
			First(&existingPurchase).Error
		if err == nil {
			return ErrPetalGiftPackPurchased
		}
		if err != gorm.ErrRecordNotFound {
			return err
		}

		petalBefore := wallet.PetalBalance
		wallet.PetalBalance += int64(petalAmount)
		wallet.Version++
		if err := tx.Save(&wallet).Error; err != nil {
			return err
		}

		activity := activityID
		transaction := model.AssetTransaction{
			UserID:        userID,
			ActivityID:    &activity,
			AssetType:     "petal",
			ChangeAmount:  int64(petalAmount),
			BalanceBefore: petalBefore,
			BalanceAfter:  wallet.PetalBalance,
			ReasonCode:    "petal_gift_pack",
			BizType:       "gift_pack",
			RequestID:     requestID,
			Remark:        "30元花瓣特惠礼包",
		}
		return tx.Create(&transaction).Error
	})
	return &wallet, err
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
var ErrPetalGiftPackPurchased = errors.New("petal gift pack already purchased")
var ErrExchangeOptionChanged = errors.New("exchange option changed")
