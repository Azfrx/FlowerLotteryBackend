package jwt

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"flower-lottery-backend/config"
	jwtv5 "github.com/golang-jwt/jwt/v5"
	"time"
)

type Claims struct {
	SubjectID   uint64 `json:"subject_id"`
	SubjectType string `json:"subject_type"`
	jwtv5.RegisteredClaims
}
type Pair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}
type Manager struct {
	cfg config.JWT
	now func() time.Time
}

func New(cfg config.JWT) *Manager { return &Manager{cfg: cfg, now: time.Now} }
func (m *Manager) Issue(subjectID uint64, subjectType string) (Pair, time.Time, error) {
	now := m.now()
	accessExp := now.Add(time.Duration(m.cfg.AccessExpireMinutes) * time.Minute)
	refreshExp := now.Add(time.Duration(m.cfg.RefreshExpireHours) * time.Hour)
	access, err := m.sign(subjectID, subjectType, "access", accessExp)
	if err != nil {
		return Pair{}, time.Time{}, err
	}
	refresh, err := m.sign(subjectID, subjectType, "refresh", refreshExp)
	return Pair{AccessToken: access, RefreshToken: refresh, ExpiresIn: m.cfg.AccessExpireMinutes * 60}, refreshExp, err
}
func (m *Manager) sign(id uint64, typ, use string, exp time.Time) (string, error) {
	c := Claims{SubjectID: id, SubjectType: typ, RegisteredClaims: jwtv5.RegisteredClaims{Issuer: m.cfg.Issuer, Subject: use, IssuedAt: jwtv5.NewNumericDate(m.now()), ExpiresAt: jwtv5.NewNumericDate(exp)}}
	return jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, c).SignedString([]byte(m.cfg.Secret))
}
func (m *Manager) Parse(token, use string) (*Claims, error) {
	parsed, err := jwtv5.ParseWithClaims(token, &Claims{}, func(t *jwtv5.Token) (any, error) {
		if t.Method != jwtv5.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(m.cfg.Secret), nil
	}, jwtv5.WithIssuer(m.cfg.Issuer))
	if err != nil {
		return nil, err
	}
	c, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid || c.Subject != use {
		return nil, errors.New("invalid token")
	}
	return c, nil
}
func Hash(raw string) string { sum := sha256.Sum256([]byte(raw)); return hex.EncodeToString(sum[:]) }
