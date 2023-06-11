package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/mrityunjaygr8/go-oink/internal/services"
	"github.com/rs/zerolog"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"golang.org/x/crypto/bcrypt"
)

var ErrUserExists = errors.New("User with this email or username already exists")
var ErrUserNotFound = errors.New("User does not exists")
var ErrUserCredsInvalid = errors.New("Invalid email/password")

type UserRepositoryInterface interface {
	UserCreate(ctx context.Context, email, password, username string) (*User, error)
	UserUpdatePassword(ctx context.Context, userID string, password string) error
	UserAuthenticate(ctx context.Context, email, password string) error
}
type User struct {
	Email    string
	ID       string
	Password string
	Username string
}

type UserRepository struct {
	DB boil.ContextExecutor
	l  zerolog.Logger
}

func (u *UserRepository) UserCreate(ctx context.Context, email, password, username string) (*User, error) {
	service := services.New(u.DB, u.l)
	emailExists, err := service.UserService.ExistsByEmail(ctx, email)
	if err != nil {
		u.l.Error().Err(err).Msg("create-user-email-exists")
	}

	if emailExists {
		return nil, ErrUserExists
	}
	usernameExists, err := service.UserService.ExistsByUsername(ctx, username)
	if err != nil {
		u.l.Error().Err(err).Msg("create-user-username-exists")
	}

	if usernameExists {
		return nil, ErrUserExists
	}
	user := services.User{
		Email:    email,
		Username: username,
	}

	user.Password, err = getPasswordHash(password)
	u.l.Info().Any("user", user).Msg("create user")

	err = service.UserService.Insert(ctx, &user)
	if err != nil {
		u.l.Error().Err(err).Msg("create-user-service-insert")
		return nil, err
	}
	return &User{
		Email:    user.Email,
		Password: user.Password,
		ID:       user.ID,
		Username: user.Username,
	}, nil
}

func (u *UserRepository) UserUpdatePassword(ctx context.Context, userID string, password string) error {
	service := services.New(u.DB, u.l)
	exists, err := service.UserService.ExistsByID(ctx, userID)
	if err != nil {
		return err
	}

	if !exists {
		return ErrUserNotFound
	}

	newPassword, err := getPasswordHash(password)
	if err != nil {
		return err
	}

	err = service.UserService.UpdatePassword(ctx, userID, newPassword)
	if err != nil {
		return err
	}

	return nil
}

func getPasswordHash(raw string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(raw), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashed), nil
}
func (u *UserRepository) UserAuthenticate(ctx context.Context, email, password string) error {
	service := services.New(u.DB, u.l)
	user, err := service.UserService.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrUserCredsInvalid
		}
		return err
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrUserCredsInvalid
		}

		return err
	}

	return nil
}
