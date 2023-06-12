package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	dbmodels "github.com/mrityunjaygr8/go-oink/internal/db/models"
	"github.com/rs/zerolog"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type TokenServiceInterface interface {
	TokenListUser(ctx context.Context, userID string) (*[]Token, error)
	TokenListUserType(ctx context.Context, userID string, tokenType string) (*[]Token, error)
	TokenCreate(ctx context.Context, token *Token) error
	TokenRetrieve(ctx context.Context, tokenID string) (*Token, error)
	TokenDelete(ctx context.Context, tokenID string) error
	TokenExistsUserType(ctx context.Context, tokenID string, userID string, tokenType string) (bool, error)
}

type TokenService struct {
	DB boil.ContextExecutor
	l  zerolog.Logger
}

type Token struct {
	Token     string
	UserID    string `boil:"user"`
	Type      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func dbToServiceToken(dbToken dbmodels.Token) *Token {
	return &Token{
		Token:     dbToken.Token,
		UserID:    dbToken.R.TokenUser.ID,
		Type:      dbToken.Type,
		CreatedAt: dbToken.CreatedAt,
		UpdatedAt: dbToken.UpdatedAt,
	}
}
func dbToServiceTokens(dbTokens dbmodels.TokenSlice) *[]Token {
	tokens := make([]Token, 0)

	for _, t := range dbTokens {
		tokens = append(tokens, *dbToServiceToken(*t))
	}

	return &tokens
}

func (t *TokenService) TokenCreate(ctx context.Context, token *Token) error {
	dbToken := dbmodels.Token{}
	dbToken.User = token.UserID
	dbToken.Type = token.Type

	dbToken.Token = uuid.New().String()
	err := dbToken.Insert(ctx, t.DB, boil.Infer())
	if err != nil {
		t.l.Error().Err(err).Msg("service-TokenCreate-Insert")
		return err
	}

	token.Token = dbToken.Token
	token.CreatedAt = dbToken.CreatedAt
	token.UpdatedAt = dbToken.UpdatedAt
	return nil
}

func (t *TokenService) TokenListUser(ctx context.Context, userID string) (*[]Token, error) {
	tokenSlice, err := dbmodels.Tokens(qm.Load(dbmodels.TokenRels.TokenUser), dbmodels.TokenWhere.User.EQ(userID)).All(ctx, t.DB)
	if err != nil {
		t.l.Error().Err(err).Msg("service-TokenListUser-List")
		return nil, err
	}
	return dbToServiceTokens(tokenSlice), nil
}

func (t *TokenService) TokenListUserType(ctx context.Context, userID string, tokenType string) (*[]Token, error) {
	tokenSlice, err := dbmodels.Tokens(qm.Load(dbmodels.TokenRels.TokenUser), dbmodels.TokenWhere.User.EQ(userID), dbmodels.TokenWhere.Type.EQ(tokenType), qm.OrderBy(dbmodels.TokenColumns.CreatedAt+" desc")).All(ctx, t.DB)
	if err != nil {
		t.l.Error().Err(err).Msg("service-TokenListUserType-List")
		return nil, err
	}

	return dbToServiceTokens(tokenSlice), nil

}

func (t *TokenService) TokenRetrieve(ctx context.Context, tokenID string) (*Token, error) {
	token, err := dbmodels.Tokens(qm.Load(dbmodels.TokenRels.TokenUser), dbmodels.TokenWhere.Token.EQ(tokenID)).One(ctx, t.DB)
	if err != nil {
		t.l.Error().Err(err).Msg("service-TokenRetrieve-bind")
		return nil, err
	}

	return dbToServiceToken(*token), nil
}

func (t *TokenService) TokenDelete(ctx context.Context, tokenID string) error {
	token, err := dbmodels.FindToken(ctx, t.DB, tokenID)
	if err != nil {
		t.l.Error().Err(err).Msg("service-TokenDelete-findToken")
		return err
	}

	_, err = token.Delete(ctx, t.DB)
	if err != nil {
		t.l.Error().Err(err).Msg("service-TokenDelete-delete")
		return err
	}

	return nil
}
func (t *TokenService) TokenExistsUserType(ctx context.Context, tokenID string, userID string, tokenType string) (bool, error) {
	return dbmodels.Tokens(dbmodels.TokenWhere.Token.EQ(tokenID), dbmodels.TokenWhere.Type.EQ(tokenType), dbmodels.TokenWhere.User.EQ(userID)).Exists(ctx, t.DB)
}
