package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	dbmodels "github.com/mrityunjaygr8/go-oink/internal/db/models"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type OinkService struct {
	DB boil.ContextExecutor
	l  zerolog.Logger
}

type OinksServiceInterface interface {
	Exists(context.Context, string) (bool, error)
	Insert(context.Context, *Oink) error
	List(context.Context) (*[]Oink, error)
	Retrieve(context.Context, string) (*Oink, error)
	Delete(context.Context, string) error
}

func dbToServiceOink(dbOink dbmodels.Oink) *Oink {
	return &Oink{
		Creator:     dbOink.R.CreatorUser.ID,
		Description: dbOink.Description.String,
		Name:        dbOink.Name,
		CreatedAt:   dbOink.CreatedAt,
		UpdatedAt:   dbOink.UpdatedAt,
		ID:          dbOink.ID,
	}
}

func dbToServiceOinks(dbOinks dbmodels.OinkSlice) *[]Oink {
	oinks := make([]Oink, 0)

	for _, o := range dbOinks {
		oinks = append(oinks, *dbToServiceOink(*o))
	}

	return &oinks
}

type Oink struct {
	Name        string
	ID          string
	Description string
	Creator     string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (o *OinkService) Exists(ctx context.Context, oinkName string) (bool, error) {
	oinkExists, err := dbmodels.Oinks(dbmodels.OinkWhere.Name.EQ(oinkName)).Exists(ctx, o.DB)
	if err != nil {
		o.l.Error().Err(err).Msg("services-OinksService-Exists")
		return false, err
	}

	return oinkExists, nil
}

func (o *OinkService) Insert(ctx context.Context, oink *Oink) error {
	dbOink := dbmodels.Oink{}
	dbOink.Name = oink.Name
	dbOink.Description = null.StringFrom(oink.Description)
	dbOink.ID = uuid.New().String()
	dbOink.Creator = oink.Creator
	err := dbOink.Insert(ctx, o.DB, boil.Infer())
	if err != nil {
		o.l.Error().Err(err).Msg("services-OinksService-Insert")
		return err
	}
	oink.ID = dbOink.ID
	oink.CreatedAt = dbOink.CreatedAt
	oink.UpdatedAt = dbOink.UpdatedAt
	return nil
}

func (o *OinkService) List(ctx context.Context) (*[]Oink, error) {
	oinkSlice, err := dbmodels.Oinks(qm.Load(dbmodels.OinkRels.CreatorUser)).All(ctx, o.DB)
	if err != nil {
		o.l.Error().Err(err).Msg("service-oinkService-List")
		return nil, err
	}
	return dbToServiceOinks(oinkSlice), nil
}

func (o *OinkService) Retrieve(ctx context.Context, oinkID string) (*Oink, error) {
	oink, err := dbmodels.Oinks(qm.Load(dbmodels.OinkRels.CreatorUser), dbmodels.OinkWhere.ID.GT(oinkID)).One(ctx, o.DB)
	if err != nil {
		o.l.Error().Err(err).Msg("service-OinkRetrieve-bind")
		return nil, err
	}

	return dbToServiceOink(*oink), nil
}

func (u *OinkService) Delete(ctx context.Context, oinkID string) error {
	oink, err := dbmodels.FindOink(ctx, u.DB, oinkID)
	if err != nil {
		u.l.Error().Err(err).Msg("service-oink-delete-findOinks")
		return err
	}

	_, err = oink.Delete(ctx, u.DB)
	if err != nil {
		u.l.Error().Err(err).Msg("service-oink-delete-delete")
		return err
	}

	return nil
}
