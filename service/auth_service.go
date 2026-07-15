package service

import (
	"errors"
	"flower-lottery-backend/common"
	"flower-lottery-backend/model"
	tokenjwt "flower-lottery-backend/pkg/jwt"
	"flower-lottery-backend/repository"
	"flower-lottery-backend/utils"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"strings"
	"time"
)

const newUserInitialCoinBalance int64 = 10_000_000

type AuthService struct {
	users  repository.UserRepository
	tokens *tokenjwt.Manager
	now    func() time.Time
}

func NewAuthService(users repository.UserRepository, tokens *tokenjwt.Manager) *AuthService {
	return &AuthService{users: users, tokens: tokens, now: time.Now}
}

func (s *AuthService) Register(userNo, nickname, avatarURL, password string) (*model.User, tokenjwt.Pair, error) {
	userNo = strings.TrimSpace(userNo)
	nickname = strings.TrimSpace(nickname)
	if !utils.ValidAccount(userNo) {
		return nil, tokenjwt.Pair{}, common.ErrAccountInvalid
	}
	if nickname == "" {
		return nil, tokenjwt.Pair{}, common.ErrNicknameRequired
	}
	if _, err := s.users.FindByUserNo(userNo); err == nil {
		return nil, tokenjwt.Pair{}, common.ErrAccountExists
	} else if err != gorm.ErrRecordNotFound {
		return nil, tokenjwt.Pair{}, err
	}
	hash, err := utils.HashPassword(password)
	if err != nil {
		return nil, tokenjwt.Pair{}, err
	}
	user := &model.User{
		UserNo:       userNo,
		Nickname:     nickname,
		AvatarURL:    avatarURL,
		PasswordHash: hash,
		Status:       1,
	}
	if err = s.users.CreateUser(user, newUserInitialCoinBalance); err != nil {
		if isDuplicateUserError(err) {
			return nil, tokenjwt.Pair{}, common.ErrAccountExists
		}
		return nil, tokenjwt.Pair{}, err
	}
	return s.Login(userNo, password)
}
func (s *AuthService) Login(userNo, password string) (*model.User, tokenjwt.Pair, error) {
	user, err := s.users.FindByUserNo(userNo)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, tokenjwt.Pair{}, common.ErrInvalidCredentials
		}
		return nil, tokenjwt.Pair{}, err
	}
	if !utils.CheckPassword(user.PasswordHash, password) {
		return nil, tokenjwt.Pair{}, common.ErrInvalidCredentials
	}
	if user.Status != 1 {
		return nil, tokenjwt.Pair{}, common.ErrAccountDisabled
	}
	pair, refreshExp, err := s.tokens.Issue(user.ID, "user")
	if err != nil {
		return nil, tokenjwt.Pair{}, err
	}
	if err = s.users.SaveRefreshToken(&model.RefreshToken{SubjectType: "user", SubjectID: user.ID, TokenHash: tokenjwt.Hash(pair.RefreshToken), ExpiresAt: refreshExp}); err != nil {
		return nil, tokenjwt.Pair{}, err
	}
	now := s.now()
	_ = s.users.UpdateLastLogin(user.ID, now)
	user.LastLoginAt = &now
	return user, pair, nil
}
func (s *AuthService) Refresh(raw string) (tokenjwt.Pair, error) {
	claims, err := s.tokens.Parse(raw, "refresh")
	if err != nil {
		return tokenjwt.Pair{}, common.ErrUnauthorized
	}
	record, err := s.users.FindRefreshToken(tokenjwt.Hash(raw))
	if err != nil || record.RevokedAt != nil || record.ExpiresAt.Before(s.now()) {
		return tokenjwt.Pair{}, common.ErrUnauthorized
	}
	_ = s.users.RevokeRefreshToken(record.TokenHash, s.now())
	pair, exp, err := s.tokens.Issue(claims.SubjectID, claims.SubjectType)
	if err != nil {
		return pair, err
	}
	err = s.users.SaveRefreshToken(&model.RefreshToken{SubjectType: claims.SubjectType, SubjectID: claims.SubjectID, TokenHash: tokenjwt.Hash(pair.RefreshToken), ExpiresAt: exp})
	return pair, err
}
func (s *AuthService) Logout(raw string) error {
	if raw == "" {
		return nil
	}
	return s.users.RevokeRefreshToken(tokenjwt.Hash(raw), s.now())
}
func (s *AuthService) Me(id uint64) (*model.User, error) { return s.users.FindByID(id) }
func (s *AuthService) UpdateProfile(id uint64, nickname string) (*model.User, error) {
	nickname = strings.TrimSpace(nickname)
	if nickname == "" {
		return nil, common.ErrNicknameRequired
	}
	if err := s.users.UpdateProfile(id, nickname); err != nil {
		return nil, err
	}
	return s.users.FindByID(id)
}
func (s *AuthService) UpdateAvatar(id uint64, avatarURL string) (*model.User, error) {
	if err := s.users.UpdateAvatar(id, avatarURL); err != nil {
		return nil, err
	}
	return s.users.FindByID(id)
}
func (s *AuthService) ChangePassword(id uint64, currentPassword, newPassword string) error {
	user, err := s.users.FindByID(id)
	if err != nil {
		return err
	}
	if !utils.CheckPassword(user.PasswordHash, currentPassword) {
		return common.ErrCurrentPassword
	}
	if utils.CheckPassword(user.PasswordHash, newPassword) {
		return common.ErrPasswordUnchanged
	}
	hash, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}
	if err = s.users.UpdatePassword(id, hash); err != nil {
		return err
	}
	return s.users.RevokeUserRefreshTokens(id, s.now())
}

func isDuplicateUserError(err error) bool {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}
