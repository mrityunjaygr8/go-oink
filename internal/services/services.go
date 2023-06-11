package services

import (
	"github.com/rs/zerolog"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type Services struct {
	UserService UserServiceInterface
}

func New(db boil.ContextExecutor, logger zerolog.Logger) *Services {
	return &Services{
		UserService: &UserService{l: logger, DB: db},
	}
}
