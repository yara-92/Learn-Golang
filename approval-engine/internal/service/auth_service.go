package service

import (
	"context"
	"database/sql"
	"time"

	"github.com/yorkyu/approval-engine/internal/auth"
	"github.com/yorkyu/approval-engine/internal/model"
	"github.com/yorkyu/approval-engine/internal/repository"
)

const tokenTTL = 24 * time.Hour

type AuthService struct {
	db     *sql.DB
	signer *auth.Signer
}

func NewAuthService(db *sql.DB, signer *auth.Signer) *AuthService {
	return &AuthService{db: db, signer: signer}
}

type LoginResult struct {
	Token string     `json:"token"`
	User  model.User `json:"user"`
}

func (s *AuthService) Login(ctx context.Context, username, password string) (*LoginResult, error) {
	u, err := repository.GetUserByUsername(ctx, s.db, username)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, model.ErrInvalidCredential
		}
		return nil, err
	}
	if !auth.CheckPassword(u.PasswordHash, password) {
		return nil, model.ErrInvalidCredential
	}
	token, err := s.signer.Generate(u.ID, u.Username, u.Role, tokenTTL)
	if err != nil {
		return nil, err
	}
	u.PasswordHash = ""
	return &LoginResult{Token: token, User: *u}, nil
}

// Register 提供一个简单的自助注册入口，方便你在演示数据之外自行加人。
func (s *AuthService) Register(ctx context.Context, username, password, displayName, role string) (*model.User, error) {
	existing, err := repository.GetUserByUsername(ctx, s.db, username)
	if err != nil && err != model.ErrNotFound {
		return nil, err
	}
	if existing != nil {
		return nil, model.ErrDuplicateUser
	}
	hash, err := auth.HashPassword(password)
	if err != nil {
		return nil, err
	}
	u := &model.User{Username: username, PasswordHash: hash, DisplayName: displayName, Role: role}
	id, err := repository.CreateUser(ctx, s.db, u)
	if err != nil {
		return nil, err
	}
	u.ID = id
	u.PasswordHash = ""
	return u, nil
}
