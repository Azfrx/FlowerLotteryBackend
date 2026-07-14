package service

import (
	"flower-lottery-backend/common"
	"flower-lottery-backend/config"
	"flower-lottery-backend/model"
	tokenjwt "flower-lottery-backend/pkg/jwt"
	"flower-lottery-backend/utils"
	"gorm.io/gorm"
	"testing"
	"time"
)

type authServiceTestRepository struct {
	users         map[uint64]*model.User
	userIDs       map[string]uint64
	refreshTokens map[string]*model.RefreshToken
	walletCreated bool
	walletCoins   int64
	tokensRevoked bool
	nextUserID    uint64
}

func newAuthServiceTestRepository() *authServiceTestRepository {
	return &authServiceTestRepository{
		users:         make(map[uint64]*model.User),
		userIDs:       make(map[string]uint64),
		refreshTokens: make(map[string]*model.RefreshToken),
		nextUserID:    1,
	}
}

func (r *authServiceTestRepository) FindByUserNo(userNo string) (*model.User, error) {
	id, exists := r.userIDs[userNo]
	if !exists {
		return nil, gorm.ErrRecordNotFound
	}
	return r.FindByID(id)
}

func (r *authServiceTestRepository) FindByID(id uint64) (*model.User, error) {
	user, exists := r.users[id]
	if !exists {
		return nil, gorm.ErrRecordNotFound
	}
	copy := *user
	return &copy, nil
}

func (r *authServiceTestRepository) CreateUser(user *model.User, initialCoinBalance int64) error {
	if _, exists := r.userIDs[user.UserNo]; exists {
		return gorm.ErrDuplicatedKey
	}
	user.ID = r.nextUserID
	r.nextUserID++
	copy := *user
	r.users[user.ID] = &copy
	r.userIDs[user.UserNo] = user.ID
	r.walletCreated = true
	r.walletCoins = initialCoinBalance
	return nil
}

func (r *authServiceTestRepository) UpdateProfile(id uint64, nickname string) error {
	r.users[id].Nickname = nickname
	return nil
}

func (r *authServiceTestRepository) UpdateAvatar(id uint64, avatarURL string) error {
	r.users[id].AvatarURL = avatarURL
	return nil
}

func (r *authServiceTestRepository) UpdatePassword(id uint64, passwordHash string) error {
	r.users[id].PasswordHash = passwordHash
	return nil
}

func (r *authServiceTestRepository) UpdateLastLogin(id uint64, at time.Time) error {
	r.users[id].LastLoginAt = &at
	return nil
}

func (r *authServiceTestRepository) SaveRefreshToken(token *model.RefreshToken) error {
	r.refreshTokens[token.TokenHash] = token
	return nil
}

func (r *authServiceTestRepository) FindRefreshToken(hash string) (*model.RefreshToken, error) {
	token, exists := r.refreshTokens[hash]
	if !exists {
		return nil, gorm.ErrRecordNotFound
	}
	return token, nil
}

func (r *authServiceTestRepository) RevokeRefreshToken(hash string, at time.Time) error {
	if token := r.refreshTokens[hash]; token != nil {
		token.RevokedAt = &at
	}
	return nil
}

func (r *authServiceTestRepository) RevokeUserRefreshTokens(userID uint64, at time.Time) error {
	r.tokensRevoked = true
	for _, token := range r.refreshTokens {
		if token.SubjectID == userID && token.SubjectType == "user" {
			token.RevokedAt = &at
		}
	}
	return nil
}

func newAuthServiceForTest(repository *authServiceTestRepository) *AuthService {
	manager := tokenjwt.New(config.JWT{
		Issuer:              "flower-lottery-test",
		Secret:              "test-secret",
		AccessExpireMinutes: 15,
		RefreshExpireHours:  24,
	})
	return NewAuthService(repository, manager)
}

func TestAuthServiceRegisterCreatesLoginReadyUser(t *testing.T) {
	repository := newAuthServiceTestRepository()
	service := newAuthServiceForTest(repository)

	user, pair, err := service.Register("newuser01", "新用户", "", "123456")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if user.UserNo != "newuser01" || user.Nickname != "新用户" {
		t.Fatalf("unexpected registered user: %+v", user)
	}
	if !repository.walletCreated {
		t.Fatal("registration did not create a wallet")
	}
	if repository.walletCoins != newUserInitialCoinBalance {
		t.Fatalf("wallet initial coins = %d, want %d", repository.walletCoins, newUserInitialCoinBalance)
	}
	if pair.AccessToken == "" || pair.RefreshToken == "" {
		t.Fatal("registration did not return a login token pair")
	}
	if !utils.CheckPassword(repository.users[user.ID].PasswordHash, "123456") {
		t.Fatal("registration did not hash the password")
	}
}

func TestAuthServiceRegisterRejectsNonAlphanumericAccount(t *testing.T) {
	repository := newAuthServiceTestRepository()
	service := newAuthServiceForTest(repository)

	for _, account := range []string{"ab", "flower_user", "花愿账号", "account-01"} {
		if _, _, err := service.Register(account, "新用户", "", "123456"); err != common.ErrAccountInvalid {
			t.Fatalf("Register(%q) error = %v, want %v", account, err, common.ErrAccountInvalid)
		}
	}
}

func TestAuthServiceProfileAndPasswordUpdates(t *testing.T) {
	repository := newAuthServiceTestRepository()
	service := newAuthServiceForTest(repository)
	user, _, err := service.Register("profileuser", "旧昵称", "", "123456")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	updated, err := service.UpdateProfile(user.ID, "  新昵称  ")
	if err != nil {
		t.Fatalf("UpdateProfile() error = %v", err)
	}
	if updated.Nickname != "新昵称" {
		t.Fatalf("unexpected updated profile: %+v", updated)
	}

	if err = service.ChangePassword(user.ID, "wrong-password", "654321"); err != common.ErrCurrentPassword {
		t.Fatalf("wrong current password error = %v", err)
	}
	if err = service.ChangePassword(user.ID, "123456", "123456"); err != common.ErrPasswordUnchanged {
		t.Fatalf("unchanged password error = %v", err)
	}
	if err = service.ChangePassword(user.ID, "123456", "654321"); err != nil {
		t.Fatalf("ChangePassword() error = %v", err)
	}
	if !utils.CheckPassword(repository.users[user.ID].PasswordHash, "654321") {
		t.Fatal("password was not replaced with the new hash")
	}
	if !repository.tokensRevoked {
		t.Fatal("password update did not revoke refresh tokens")
	}
}
