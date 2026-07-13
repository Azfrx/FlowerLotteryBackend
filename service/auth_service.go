package service

import (
	"flower-lottery-backend/common"
	"flower-lottery-backend/model"
	tokenjwt "flower-lottery-backend/pkg/jwt"
	"flower-lottery-backend/repository"
	"flower-lottery-backend/utils"
	"gorm.io/gorm"
	"time"
)

type AuthService struct {
	users  repository.UserRepository
	tokens *tokenjwt.Manager
	now    func() time.Time
}

func NewAuthService(users repository.UserRepository, tokens *tokenjwt.Manager) *AuthService {
	return &AuthService{users: users, tokens: tokens, now: time.Now}
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
