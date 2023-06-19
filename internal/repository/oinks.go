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

type OinkRepositoryInterface interface {
	OinkList(context.Context) (*[]Oink, error)
	OinkRetrieve(context.Context, string) (*Oink, error)
	OinkDelete(context.Context, string) error
	OinkInsert(context.Context, string, string, string) (*Oink, error)
}

var (
	ErrOinkNotFound = errors.New("Oink does not exist")
	ErrOinkExists   = errors.New("Oink with this name already exists")
)

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

func (o *OinkRepository) OinkRetrieve(ctx context.Context, oinkName string) (*Oink, error) {
	service := services.New(o.DB, o.l)

	oink, err := service.OinkService.RetrieveByName(ctx, oinkName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrOinkNotFound
		}
		o.l.Error().Err(err).Msg("repository-OinkRetrieve-OinkRetrieve")
		return nil, err
	}
	return serviceToRepositoryOink(*oink), nil
}

func (o *OinkRepository) OinkDelete(ctx context.Context, oinkName string) error {
	service := services.New(o.DB, o.l)

	err := service.OinkService.Delete(ctx, oinkName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrOinkNotFound
		}
		o.l.Error().Err(err).Msg("repository-OinkDelete-OinkDelete")
		return err
	}
	return nil
}

func (o *OinkRepository) OinkInsert(ctx context.Context, name string, description string, creatorID string) (*Oink, error) {
	service := services.New(o.DB, o.l)

	exists, err := service.OinkService.Exists(ctx, name)
	if err != nil {
		o.l.Error().Err(err).Msg("repository-OinkInsert-Exists")
		return nil, err
	}

	if exists {
		return nil, ErrOinkExists
	}

	oink := services.Oink{
		Name:        name,
		Description: description,
		Creator:     creatorID,
	}
	err = service.OinkService.Insert(ctx, &oink)
	if err != nil {
		o.l.Error().Err(err).Msg("repository-oink-OinkInsert-insert")
		return nil, err
	}

	return serviceToRepositoryOink(oink), nil
}
