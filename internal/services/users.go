package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	dbmodels "github.com/mrityunjaygr8/go-oink/internal/db/models"
	"github.com/rs/zerolog"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type UserService struct {
	DB boil.ContextExecutor
	l  zerolog.Logger
}

var ErrUserNotFound = errors.New("User Not Found")

type UserServiceInterface interface {
	List(context.Context) (*[]User, error)
	Insert(context.Context, *User) error
	Exists(ctx context.Context, query string) (bool, error)
	ExistsByID(ctx context.Context, query string) (bool, error)
	ExistsByEmail(ctx context.Context, query string) (bool, error)
	ExistsByUsername(ctx context.Context, query string) (bool, error)
	UpdatePassword(ctx context.Context, userID string, password string) error
	GetByID(ctx context.Context, userID string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	Delete(ctx context.Context, userID string) error
}

type User struct {
	Email     string
	ID        string
	Password  string
	Username  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (u *UserService) Insert(ctx context.Context, user *User) error {
	dbUser := dbmodels.User{}
	dbUser.Username = user.Username
	dbUser.Email = user.Email
	dbUser.ID = uuid.New().String()
	dbUser.Password = user.Password

	err := dbUser.Insert(ctx, u.DB, boil.Infer())
	if err != nil {
		u.l.Error().Err(err).Str("err-type", fmt.Sprintf("%T", err)).Msg("")
		return err
	}

	user.Password = dbUser.Password
	user.ID = dbUser.ID
	user.CreatedAt = dbUser.CreatedAt
	user.UpdatedAt = dbUser.UpdatedAt

	return nil
}

func (u *UserService) List(ctx context.Context) (*[]User, error) {
	var users []User

	err := dbmodels.Users().Bind(ctx, u.DB, &users)
	if err != nil {
		u.l.Error().Err(err).Msg("in-list-erro")
		return nil, err
	}
	return &users, nil
}

func (u *UserService) Exists(ctx context.Context, query string) (bool, error) {
	exists, err := dbmodels.Users(qm.Expr(dbmodels.UserWhere.Email.EQ(query), qm.Or2(dbmodels.UserWhere.Username.EQ(query)))).Exists(ctx, u.DB)
	if err != nil {
		u.l.Error().Err(err).Msg("service-user-exists")
		return false, err
	}

	return exists, err
}
func (u *UserService) ExistsByID(ctx context.Context, query string) (bool, error) {
	exists, err := dbmodels.Users(dbmodels.UserWhere.ID.EQ(query)).Exists(ctx, u.DB)
	if err != nil {
		u.l.Error().Err(err).Msg("service-user-existsByID")
		return false, err
	}

	return exists, err
}
func (u *UserService) ExistsByEmail(ctx context.Context, query string) (bool, error) {
	exists, err := dbmodels.Users(dbmodels.UserWhere.Email.EQ(query)).Exists(ctx, u.DB)
	if err != nil {
		u.l.Error().Err(err).Msg("service-user-existsByEmail")
		return false, err
	}

	return exists, err
}
func (u *UserService) ExistsByUsername(ctx context.Context, query string) (bool, error) {
	exists, err := dbmodels.Users(dbmodels.UserWhere.Username.EQ(query)).Exists(ctx, u.DB)
	if err != nil {
		u.l.Error().Err(err).Msg("service-user-existsByUsername")
		return false, err
	}

	return exists, err
}

func (u *UserService) GetByID(ctx context.Context, userID string) (*User, error) {
	var user User
	err := dbmodels.Users(dbmodels.UserWhere.ID.EQ(userID)).Bind(ctx, u.DB, &user)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		u.l.Error().Err(err).Msg("service-user-getByID")
		return nil, err
	}

	return &user, nil
}
func (u *UserService) GetByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	err := dbmodels.Users(dbmodels.UserWhere.Email.EQ(email)).Bind(ctx, u.DB, &user)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		u.l.Error().Err(err).Msg("service-user-getByEmail")
		return nil, err
	}

	return &user, nil
}
func (u *UserService) GetByUsername(ctx context.Context, username string) (*User, error) {
	var user User
	err := dbmodels.Users(dbmodels.UserWhere.Username.EQ(username)).Bind(ctx, u.DB, &user)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		u.l.Error().Err(err).Msg("service-user-getByUsername")
		return nil, err
	}

	return &user, nil
}

func (u *UserService) UpdatePassword(ctx context.Context, userID string, password string) error {
	user, err := dbmodels.FindUser(ctx, u.DB, userID)
	if err != nil {
		u.l.Error().Err(err).Msg("service-user-UpdatePassword-findUser")
		return err
	}

	user.Password = password

	_, err = user.Update(ctx, u.DB, boil.Whitelist("password", "updated_at"))
	if err != nil {
		u.l.Error().Err(err).Msg("service-user-UpdatePassword-update")
		return err
	}

	return nil
}

func (u *UserService) Delete(ctx context.Context, userID string) error {
	user, err := dbmodels.FindUser(ctx, u.DB, userID)
	if err != nil {
		u.l.Error().Err(err).Msg("service-user-delete-findUser")
		return err
	}

	_, err = user.Delete(ctx, u.DB)
	if err != nil {
		u.l.Error().Err(err).Msg("service-user-delete-delete")
		return err
	}

	return nil
}
