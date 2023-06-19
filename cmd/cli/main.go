package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"math/big"
	"os"

	_ "github.com/lib/pq"
	"github.com/mrityunjaygr8/go-oink/internal/config"
	"github.com/mrityunjaygr8/go-oink/internal/repository"
	"github.com/mrityunjaygr8/go-oink/internal/services"
	"github.com/rs/zerolog"
)

func main() {
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()
	path, err := os.Getwd()
	if err != nil {
		logger.Fatal().Err(err)
	}
	c, err := config.GetConfig(path, logger)
	if err != nil {
		logger.Fatal().Err(err)
	}

	if c.Env == config.EnvDevelopment {
		logger = logger.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	logger.Info().Any("config", c).Msg("")
	if c.DbDsn == "" {
		c.DbDsn = fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s", c.DbUser, c.DbPass, c.DbHost, c.DbPort, c.DbName, c.DbSSL)
	}
	logger.Info().Msg(c.DbDsn)
	db, err := sql.Open("postgres", c.DbDsn)
	if err != nil {
		// logger.WithFields(c.toFields()).Fatal(err)
		logger.Fatal().Any("config", c).Err(err)
	}

	err = db.Ping()
	if err != nil {
		// logger.WithFields(c.toFields()).Fatal(err)
		logger.Fatal().Any("config", c).Err(err)
	}

	repo := repository.New(db, logger)
	nRand, err := rand.Int(rand.Reader, big.NewInt(10000))
	if err != nil {
		logger.Error().Err(err).Msg("error creating random suffix")
	}

	user := repository.User{}
	user.Email = fmt.Sprintf("im_%s@parham.in", nRand)
	user.Password = "password"
	user.Username = fmt.Sprintf("im_%s@parham", nRand)

	u, err := repo.UserRepository.UserCreate(context.Background(), user.Email, user.Password, user.Username)
	if err != nil {
		logger.Fatal().Err(err).Msg("user indert error")
	}
	logger.Info().Any("user", u).Msg("user insert")

	service := services.New(db, logger)

	oink := &services.Oink{
		Name:        fmt.Sprintf("chelsea_%s", nRand),
		Description: "the official oink for chelsea FC",
		Creator:     u.ID,
	}

	err = service.OinkService.Insert(context.Background(), oink)
	if err != nil {
		logger.Fatal().Err(err).Msg("user indert error")
	}
	logger.Info().Any("oink", oink).Msg("oink insert")

	// oinks, err := service.OinkService.List(context.Background())
	// if err != nil {
	// 	logger.Fatal().Err(err).Msg("user list error")
	// }
	// for _, o := range *oinks {
	// 	logger.Info().Any("o", o).Msg("oink list")
	// }

	logger.Info().Any("oinkID", oink.ID).Msg("oink ID")
	oo, err := service.OinkService.Retrieve(context.Background(), oink.ID)
	if err != nil {
		logger.Fatal().Err(err).Msg("user retrieve error")
	}
	logger.Info().Any("oo", oo).Msg("oink retrieve")

	err = service.OinkService.Delete(context.Background(), oo.ID)
	if err != nil {
		logger.Fatal().Err(err).Msg("user delete error")
	}
	oo1, err := service.OinkService.Retrieve(context.Background(), oo.ID)
	if err != nil {
		logger.Fatal().Err(err).Msg("oink retrieve after delete error")
	}
	logger.Info().Any("oo1", oo1).Msg("testing oink retrieve after delete")
}
