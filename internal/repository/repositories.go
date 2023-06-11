package repository

import (
	"github.com/rs/zerolog"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type Repository struct {
	UserRepository UserRepositoryInterface
}

func New(db boil.ContextExecutor, l zerolog.Logger) *Repository {
	return &Repository{
		UserRepository: &UserRepository{DB: db, l: l},
	}
}