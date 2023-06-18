package repository

import (
	"context"
	"time"

	"github.com/mrityunjaygr8/go-oink/internal/services"
	"github.com/rs/zerolog"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type OinkRepositoryInterface interface {
	OinkList(context.Context) (*[]Oink, error)
}

type OinkRepository struct {
	DB boil.ContextExecutor
	l  zerolog.Logger
}
type Oink struct {
	Name        string
	ID          string
	CreatorID   string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func serviceToRepositoryOink(oink services.Oink) *Oink {
	return &Oink{
		Name:        oink.Name,
		CreatorID:   oink.Creator,
		ID:          oink.ID,
		Description: oink.Description,
		CreatedAt:   oink.CreatedAt,
		UpdatedAt:   oink.UpdatedAt,
	}
}

func serviceToRepositoryOinks(o []services.Oink) *[]Oink {
	oinks := make([]Oink, 0)
	for _, oink := range o {
		oinks = append(oinks, *serviceToRepositoryOink(oink))
	}

	return &oinks
}

func (o *OinkRepository) OinkList(ctx context.Context) (*[]Oink, error) {
	service := services.New(o.DB, o.l)

	oinks, err := service.OinkService.List(ctx)
	if err != nil {
		o.l.Error().Err(err).Msg("repository-OinkList-OinkList")
		return nil, err
	}

	return serviceToRepositoryOinks(*oinks), nil
}
