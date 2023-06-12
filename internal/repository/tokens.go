package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/mrityunjaygr8/go-oink/internal/services"
	"github.com/rs/zerolog"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type TokenType string

const (
	TokenTypeLogin TokenType = "login"
)

var ErrTokenNotFound = errors.New("Token not Found")

type TokenRepositoryInterface interface {
	TokenListUser(ctx context.Context, userID string) (*[]Token, error)
	TokenListUserType(ctx context.Context, userID string, tokenType TokenType) (*[]Token, error)
	TokenLoginDelete(ctx context.Context, tokenID string, userID string) error
	TokenLoginCreate(ctx context.Context, userID string) (*Token, error)
	TokenRetrieve(ctx context.Context, tokenID string) (*Token, error)
}

type TokenRepository struct {
	DB boil.ContextExecutor
	l  zerolog.Logger
}

type Token struct {
	Token     string
	UserID    string
	Type      TokenType
	CreatedAt time.Time
	UpdatedAt time.Time
}

func serviceToRepositoryToken(token services.Token) *Token {
	return &Token{
		Token:     token.Token,
		UserID:    token.UserID,
		Type:      TokenType(token.Type),
		CreatedAt: token.CreatedAt,
		UpdatedAt: token.UpdatedAt,
	}
}

func serviceToRepositoryTokens(t []services.Token) *[]Token {
	tokens := make([]Token, 0)
	for _, token := range t {
		tokens = append(tokens, *serviceToRepositoryToken(token))
	}

	return &tokens
}

func (t *TokenRepository) TokenLoginCreate(ctx context.Context, userID string) (*Token, error) {
	service := services.New(t.DB, t.l)

	token := services.Token{
		UserID: userID,
		Type:   string(TokenTypeLogin),
	}

	err := service.TokenService.TokenCreate(ctx, &token)
	if err != nil {
		t.l.Error().Err(err).Msg("repository-TokenLoginCreate-TokenCreate")
		return nil, err
	}

	return serviceToRepositoryToken(token), nil
}

func (t *TokenRepository) TokenLoginDelete(ctx context.Context, tokenID string, userID string) error {
	service := services.New(t.DB, t.l)

	exists, err := service.TokenService.TokenExistsUserType(ctx, tokenID, userID, string(TokenTypeLogin))
	if err != nil {
		t.l.Error().Err(err).Msg("repository-TokenLoginDelete-TokenExistsUserType")
		return err
	}

	if !exists {
		return ErrTokenNotFound
	}

	err = service.TokenService.TokenDelete(ctx, tokenID)
	if err != nil {
		t.l.Error().Err(err).Msg("repository-TokenLoginDelete-TokenDelete")
		return err
	}

	return nil
}

func (t *TokenRepository) TokenListUser(ctx context.Context, userID string) (*[]Token, error) {
	service := services.New(t.DB, t.l)

	tokens, err := service.TokenService.TokenListUser(ctx, userID)
	if err != nil {
		t.l.Error().Err(err).Msg("repository-TokenList-TokenListUser")
		return nil, err
	}

	return serviceToRepositoryTokens(*tokens), nil
}
func (t *TokenRepository) TokenListUserType(ctx context.Context, userID string, tokenType TokenType) (*[]Token, error) {
	service := services.New(t.DB, t.l)

	tokens, err := service.TokenService.TokenListUserType(ctx, userID, string(tokenType))
	if err != nil {
		t.l.Error().Err(err).Msg("repository-TokenList-TokenListUserType")
		return nil, err
	}

	return serviceToRepositoryTokens(*tokens), nil
}

func (t *TokenRepository) TokenRetrieve(ctx context.Context, tokenID string) (*Token, error) {
	service := services.New(t.DB, t.l)

	token, err := service.TokenService.TokenRetrieve(ctx, tokenID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTokenNotFound
		}
		t.l.Error().Err(err).Msg("repository-TokenRetrieve-TokenRetrieve")
		return nil, err
	}

	return serviceToRepositoryToken(*token), nil
}
