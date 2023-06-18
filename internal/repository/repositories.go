package repository

import (
	"github.com/rs/zerolog"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type Repository struct {
	UserRepository  UserRepositoryInterface
	TokenRepository TokenRepositoryInterface
	OinkRepository  OinkRepositoryInterface
}

func New(db boil.ContextExecutor, l zerolog.Logger) *Repository {
	return &Repository{
		UserRepository:  &UserRepository{DB: db, l: l},
		TokenRepository: &TokenRepository{DB: db, l: l},
		OinkRepository:  &OinkRepository{DB: db, l: l},
	}
}
