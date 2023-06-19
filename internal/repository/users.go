package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/mrityunjaygr8/go-oink/internal/services"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserExists       = errors.New("User with this email or username already exists")
	ErrUserNotFound     = errors.New("User does not exists")
	ErrUserCredsInvalid = errors.New("Invalid email/password")
)

type UserRepositoryInterface interface {
	UserCreate(ctx context.Context, email, password, username string) (*User, error)
	UserUpdatePassword(ctx context.Context, userID string, password string) error
	UserAuthenticate(ctx context.Context, email, password string) (*User, error)
	UserRetrieve(ctx context.Context, userID string) (*User, error)
	UsersList(ctx context.Context) (*[]User, error)
	UserDelete(ctx context.Context, userID string) error
}
type User struct {
	Email     string
	ID        string
	Password  string
	Username  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func serviceToRepositoryUser(user services.User) *User {
	return &User{
		Email:     user.Email,
		ID:        user.ID,
		Password:  user.Password,
		Username:  user.Username,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

func serviceToRepositoryUsers(u []services.User) *[]User {
	users := make([]User, 0)
	for _, user := range u {
		users = append(users, *serviceToRepositoryUser(user))
	}

	return &users
}

type UserRepository struct {
	DB boil.ContextExecutor
	l  zerolog.Logger
}

func (u *UserRepository) UsersList(ctx context.Context) (*[]User, error) {
	_, _ = hlog.IDFromCtx(ctx)
	service := services.New(u.DB, u.l)

	serviceUsers, err := service.UserService.List(ctx)
	if err != nil {
		return nil, err
	}

	users := serviceToRepositoryUsers(*serviceUsers)

	return users, nil
}

func (u *UserRepository) UserCreate(ctx context.Context, email, password, username string) (*User, error) {
	service := services.New(u.DB, u.l)
	emailExists, err := service.UserService.ExistsByEmail(ctx, email)
	if err != nil {
		u.l.Error().Err(err).Msg("repository-user-UserCreate-existsByEmail")
		return nil, err
	}

	if emailExists {
		return nil, ErrUserExists
	}
	usernameExists, err := service.UserService.ExistsByUsername(ctx, username)
	if err != nil {
		u.l.Error().Err(err).Msg("repository-user-UserCreate-existsByUser")
		return nil, err
	}

	if usernameExists {
		return nil, ErrUserExists
	}
	user := services.User{
		Email:    email,
		Username: username,
	}

	user.Password, err = getPasswordHash(password)
	if err != nil {
		u.l.Error().Err(err).Msg("repository-user-UserCreate-getPasswordHash")
		return nil, err
	}

	err = service.UserService.Insert(ctx, &user)
	if err != nil {
		u.l.Error().Err(err).Msg("repository-user-UserCreate-insert")
		return nil, err
	}

	return serviceToRepositoryUser(user), nil
}

func (u *UserRepository) UserUpdatePassword(ctx context.Context, userID string, password string) error {
	service := services.New(u.DB, u.l)
	exists, err := service.UserService.ExistsByID(ctx, userID)
	if err != nil {
		u.l.Error().Err(err).Msg("repository-user-UserUpdatePassword-ExistsByID")
		return err
	}

	if !exists {
		return ErrUserNotFound
	}

	newPassword, err := getPasswordHash(password)
	if err != nil {
		u.l.Error().Err(err).Msg("repository-user-UserUpdatePassword-getPasswordHash")
		return err
	}

	err = service.UserService.UpdatePassword(ctx, userID, newPassword)
	if err != nil {
		u.l.Error().Err(err).Msg("repository-user-UserUpdatePassword-UpdatePassword")
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

func (u *UserRepository) UserAuthenticate(ctx context.Context, email, password string) (*User, error) {
	service := services.New(u.DB, u.l)
	user, err := service.UserService.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			return nil, ErrUserCredsInvalid
		}
		u.l.Error().Err(err).Msg("repository-user-UserAuthenticate-GetByEmail")
		return nil, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, ErrUserCredsInvalid
		}

		u.l.Error().Err(err).Msg("repository-user-UserAuthenticate-compare")
		return nil, err
	}

	return serviceToRepositoryUser(*user), nil
}

func (u *UserRepository) UserRetrieve(ctx context.Context, userID string) (*User, error) {
	service := services.New(u.DB, u.l)
	user, err := service.UserService.GetByID(ctx, userID)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		u.l.Error().Err(err).Msg("repository-user-UserRetrieve-GetByID")
		return nil, err
	}

	return serviceToRepositoryUser(*user), nil
}

func (u *UserRepository) UserDelete(ctx context.Context, userID string) error {
	service := services.New(u.DB, u.l)
	_, err := service.UserService.GetByID(ctx, userID)
	if err != nil {
		if err != sql.ErrNoRows {
			return ErrUserNotFound
		}
		u.l.Error().Err(err).Msg("repository-user-UserRetrieve-GetByID")
		return err
	}

	err = service.UserService.Delete(ctx, userID)
	if err != nil {
		u.l.Error().Err(err).Msg("repository-user-UserDelete-Delete")
		return err
	}

	return nil
}
