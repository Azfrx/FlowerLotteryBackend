package service

import (
	"errors"
	"flower-lottery-backend/common"
	"flower-lottery-backend/model"
	"flower-lottery-backend/repository"
	"fmt"
	"gorm.io/gorm"
	"time"
)

type WalletService struct{ repo repository.WalletRepository }

const PetalGiftPackPetalAmount uint64 = 100

type PetalGiftPackResult struct {
	PetalAmount uint64            `json:"petal_amount"`
	Wallet      *model.UserWallet `json:"wallet"`
}

func NewWalletService(repo repository.WalletRepository) *WalletService {
	return &WalletService{repo: repo}
}
func (s *WalletService) Get(userID uint64) (*model.UserWallet, error) { return s.repo.Get(userID) }
func (s *WalletService) Options() ([]model.ExchangeOption, error)     { return s.repo.ListOptions() }
func (s *WalletService) Exchange(userID, optionID, expectedPetalAmount, expectedCoinCost uint64, requestID string) (*model.ExchangeOrder, *model.UserWallet, error) {
	order, wallet, err := s.repo.Exchange(userID, optionID, expectedPetalAmount, expectedCoinCost, requestID, newOrderNo("EX"))
	if errors.Is(err, repository.ErrInsufficientBalance) {
		return nil, nil, common.ErrCoinInsufficient
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, common.ErrExchangeOption
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return nil, nil, common.ErrDuplicateRequest
	}
	if errors.Is(err, repository.ErrExchangeOptionChanged) {
		return nil, nil, common.ErrExchangeOptionChanged
	}
	return order, wallet, err
}
func (s *WalletService) PetalGiftPackStatus(userID uint64) (bool, error) {
	return s.repo.PetalGiftPackPurchased(userID)
}
func (s *WalletService) PurchasePetalGiftPack(userID uint64, requestID string) (*PetalGiftPackResult, error) {
	wallet, err := s.repo.PurchasePetalGiftPack(userID, PetalGiftPackPetalAmount, requestID)
	if errors.Is(err, repository.ErrPetalGiftPackPurchased) {
		return nil, common.ErrPetalGiftPackPurchased
	}
	if err != nil {
		return nil, err
	}
	return &PetalGiftPackResult{PetalAmount: PetalGiftPackPetalAmount, Wallet: wallet}, nil
}
func (s *WalletService) Transactions(userID uint64, page, pageSize int) ([]model.AssetTransaction, int64, error) {
	return s.repo.ListTransactions(userID, page, pageSize)
}
func newOrderNo(prefix string) string { return fmt.Sprintf("%s%d", prefix, time.Now().UnixNano()) }
