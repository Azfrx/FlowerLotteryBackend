package repository

import (
	"flower-lottery-backend/model"
	"gorm.io/gorm"
	"time"
)

type UserRepository interface {
	FindByUserNo(userNo string) (*model.User, error)
	FindByID(id uint64) (*model.User, error)
	CreateUser(user *model.User) error
	UpdateProfile(id uint64, nickname string) error
	UpdateAvatar(id uint64, avatarURL string) error
	UpdatePassword(id uint64, passwordHash string) error
	UpdateLastLogin(id uint64, at time.Time) error
	SaveRefreshToken(token *model.RefreshToken) error
	FindRefreshToken(hash string) (*model.RefreshToken, error)
	RevokeRefreshToken(hash string, at time.Time) error
	RevokeUserRefreshTokens(userID uint64, at time.Time) error
}
type userRepository struct{ db *gorm.DB }

func NewUserRepository(db *gorm.DB) UserRepository { return &userRepository{db: db} }
func (r *userRepository) FindByUserNo(no string) (*model.User, error) {
	var v model.User
	err := r.db.Where("user_no = ? AND deleted_at IS NULL", no).First(&v).Error
	return &v, err
}
func (r *userRepository) FindByID(id uint64) (*model.User, error) {
	var v model.User
	err := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&v).Error
	return &v, err
}
func (r *userRepository) CreateUser(user *model.User) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		return tx.Create(&model.UserWallet{UserID: user.ID}).Error
	})
}
func (r *userRepository) UpdateProfile(id uint64, nickname string) error {
	return r.db.Model(&model.User{}).Where("id=? AND deleted_at IS NULL", id).Update("nickname", nickname).Error
}
func (r *userRepository) UpdateAvatar(id uint64, avatarURL string) error {
	return r.db.Model(&model.User{}).Where("id=? AND deleted_at IS NULL", id).Update("avatar_url", avatarURL).Error
}
func (r *userRepository) UpdatePassword(id uint64, passwordHash string) error {
	return r.db.Model(&model.User{}).Where("id=? AND deleted_at IS NULL", id).Update("password_hash", passwordHash).Error
}
func (r *userRepository) UpdateLastLogin(id uint64, at time.Time) error {
	return r.db.Model(&model.User{}).Where("id = ?", id).Update("last_login_at", at).Error
}
func (r *userRepository) SaveRefreshToken(v *model.RefreshToken) error { return r.db.Create(v).Error }
func (r *userRepository) FindRefreshToken(hash string) (*model.RefreshToken, error) {
	var v model.RefreshToken
	err := r.db.Where("token_hash = ?", hash).First(&v).Error
	return &v, err
}
func (r *userRepository) RevokeRefreshToken(hash string, at time.Time) error {
	return r.db.Model(&model.RefreshToken{}).Where("token_hash = ? AND revoked_at IS NULL", hash).Update("revoked_at", at).Error
}
func (r *userRepository) RevokeUserRefreshTokens(userID uint64, at time.Time) error {
	return r.db.Model(&model.RefreshToken{}).Where("subject_type='user' AND subject_id=? AND revoked_at IS NULL", userID).Update("revoked_at", at).Error
}
