package service

import (
	"errors"
	"flower-lottery-backend/common"
	"flower-lottery-backend/model"
	"flower-lottery-backend/repository"
	"testing"
)

type walletRepositoryStub struct {
	purchased      bool
	purchaseAmount uint64
	purchaseErr    error
	exchangeErr    error
	wallet         *model.UserWallet
}

func (r *walletRepositoryStub) Get(uint64) (*model.UserWallet, error) {
	return r.wallet, nil
}
func (r *walletRepositoryStub) ListOptions() ([]model.ExchangeOption, error) {
	return nil, nil
}
func (r *walletRepositoryStub) Exchange(uint64, uint64, uint64, uint64, string, string) (*model.ExchangeOrder, *model.UserWallet, error) {
	return nil, r.wallet, r.exchangeErr
}

func TestExchangeMapsChangedOptionError(t *testing.T) {
	repo := &walletRepositoryStub{exchangeErr: repository.ErrExchangeOptionChanged}
	_, _, err := NewWalletService(repo).Exchange(8, 3, 50, 3000, "exchange-request")
	if !errors.Is(err, common.ErrExchangeOptionChanged) {
		t.Fatalf("expected changed option error, got %v", err)
	}
}
func (r *walletRepositoryStub) PetalGiftPackPurchased(uint64) (bool, error) {
	return r.purchased, nil
}
func (r *walletRepositoryStub) PurchasePetalGiftPack(_ uint64, amount uint64, _ string) (*model.UserWallet, error) {
	r.purchaseAmount = amount
	return r.wallet, r.purchaseErr
}
func (r *walletRepositoryStub) ListTransactions(uint64, int, int) ([]model.AssetTransaction, int64, error) {
	return nil, 0, nil
}

func TestPetalGiftPackStatus(t *testing.T) {
	repo := &walletRepositoryStub{purchased: true}
	purchased, err := NewWalletService(repo).PetalGiftPackStatus(8)
	if err != nil {
		t.Fatalf("PetalGiftPackStatus returned error: %v", err)
	}
	if !purchased {
		t.Fatal("expected purchased status")
	}
}

func TestPurchasePetalGiftPackGrantsOneHundredPetals(t *testing.T) {
	wallet := &model.UserWallet{UserID: 8, PetalBalance: 160}
	repo := &walletRepositoryStub{wallet: wallet}
	result, err := NewWalletService(repo).PurchasePetalGiftPack(8, "gift-pack-request")
	if err != nil {
		t.Fatalf("PurchasePetalGiftPack returned error: %v", err)
	}
	if repo.purchaseAmount != 100 || result.PetalAmount != 100 {
		t.Fatalf("expected 100 petals, repository=%d result=%d", repo.purchaseAmount, result.PetalAmount)
	}
	if result.Wallet != wallet {
		t.Fatal("expected updated wallet in result")
	}
}

func TestPurchasePetalGiftPackMapsPurchaseLimitError(t *testing.T) {
	repo := &walletRepositoryStub{purchaseErr: repository.ErrPetalGiftPackPurchased}
	_, err := NewWalletService(repo).PurchasePetalGiftPack(8, "gift-pack-request")
	if !errors.Is(err, common.ErrPetalGiftPackPurchased) {
		t.Fatalf("expected purchase limit error, got %v", err)
	}
}
