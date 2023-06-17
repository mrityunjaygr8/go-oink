package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"

	"github.com/mrityunjaygr8/go-oink/internal/config"
	"github.com/mrityunjaygr8/go-oink/internal/repository"
	"github.com/rs/zerolog"
)

func main() {
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()
	// logger.Logger = logger.Output(zerolog.ConsoleWriter{Out: os.Stderr})
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
		logger.Fatal().Any("config", c).Err(err)
	}

	err = db.Ping()
	if err != nil {
		logger.Fatal().Any("config", c).Err(err)
	}

	repo := repository.New(db, logger)

	user := repository.User{}
	user.Email = "im@oink.in"
	user.Password = "password"
	user.Username = "im@parham"

	logger.Info().Any("user", user).Msg("Setup")
	u, err := repo.UserRepository.UserCreate(context.Background(), user.Email, user.Password, user.Username)
	if err != nil {
		logger.Info().Err(err).Msg("UserCreate")
	}

	logger.Info().Any("user", u).Msg("user created successfully")

}
