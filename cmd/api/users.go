package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/mrityunjaygr8/go-oink/internal/repository"
	"github.com/rs/zerolog/hlog"
)

func (s *Server) UserList() http.HandlerFunc {
	type User struct {
		Email     string    `json:"email"`
		Password  string    `json:"-"`
		Username  string    `json:"username"`
		ID        string    `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at,omitempty"`
	}
	type response struct {
		Users []User `json:"users"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		logger := hlog.FromRequest(r)
		repo := repository.New(s.db, *logger)
		u, err := repo.UserRepository.UsersList(r.Context())

		if err != nil {
			logger.Error().Err(err).Msg("api-UserList-List")
			s.writeJSON(w, http.StatusInternalServerError, envelope{"error": http.StatusText(http.StatusInternalServerError)}, nil)
			return
		}

		users := make([]User, 0)
		for _, user := range *u {
			users = append(users, User{
				Email:     user.Email,
				Username:  user.Username,
				ID:        user.ID,
				CreatedAt: user.CreatedAt,
				UpdatedAt: user.UpdatedAt,
			})
		}

		s.writeJSON(w, http.StatusOK, envelope{"users": users}, nil)
	}
}

func (s *Server) UserCreate() http.HandlerFunc {
	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Username string `json:"username"`
	}

	type response struct {
		Email     string    `json:"email"`
		Username  string    `json:"username"`
		ID        string    `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var req request
		logger := hlog.FromRequest(r)
		err := s.readJSON(w, r, &req)
		if err != nil {
			logger.Error().Err(err).Msg("api-UserCreate-readJson")
			s.writeJSON(w, http.StatusBadRequest, envelope{"error": err.Error()}, nil)
			return
		}

		tx, err := s.db.Begin()
		if err != nil {
			logger.Error().Err(err).Msg("error creating transaction")
			s.writeJSON(w, http.StatusInternalServerError, envelope{"error": http.StatusText(http.StatusInternalServerError)}, nil)
			return
		}

		repo := repository.New(tx, *logger)
		user, err := repo.UserRepository.UserCreate(r.Context(), req.Email, req.Password, req.Username)
		if err != nil {
			if rollback := tx.Rollback(); rollback != nil {
				logger.Error().Err(err).Msg("api-UserCreate-UserCreate-RollbackError")
				s.writeJSON(w, http.StatusInternalServerError, envelope{"error": http.StatusText(http.StatusInternalServerError)}, nil)
				return
			}
			if errors.Is(err, repository.ErrUserExists) {
				s.writeJSON(w, http.StatusBadRequest, envelope{"error": repository.ErrUserExists.Error()}, nil)
				return
			}
			logger.Error().Err(err).Msg("api-UserCreate-UserCreate")
			s.writeJSON(w, http.StatusInternalServerError, envelope{"error": http.StatusText(http.StatusInternalServerError)}, nil)
			return
		}

		if err := tx.Commit(); err != nil {
			logger.Error().Err(err).Msg("api-UserCreate-UserCreate-CommitError")
			if rollback := tx.Rollback(); rollback != nil {
				logger.Error().Err(err).Msg("api-UserCreate-UserCreate-RollbackError")
				s.writeJSON(w, http.StatusInternalServerError, envelope{"error": http.StatusText(http.StatusInternalServerError)}, nil)
				return
			}

			s.writeJSON(w, http.StatusInternalServerError, envelope{"error": http.StatusText(http.StatusInternalServerError)}, nil)
			return
		}

		resp := response{
			Email:     user.Email,
			Username:  user.Username,
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		}

		s.writeJSON(w, http.StatusCreated, envelope{"user": resp}, nil)
	}
}

func (s *Server) UserRetrieve() http.HandlerFunc {
	type response struct {
		Email     string    `json:"email"`
		Username  string    `json:"username"`
		ID        string    `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		userID := chi.URLParam(r, "userID")
		logger := hlog.FromRequest(r)
		repo := repository.New(s.db, *logger)
		user, err := repo.UserRepository.UserRetrieve(r.Context(), userID)
		if err != nil {
			logger.Error().Err(err).Msg("api-UserRetrieve-UserRetrieve")
			s.writeJSON(w, http.StatusNotFound, envelope{"error": err.Error()}, nil)
			return
		}

		res := response{
			Email:     user.Email,
			Username:  user.Username,
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		}

		s.writeJSON(w, http.StatusOK, envelope{"user": res}, nil)
	}
}

func (s *Server) UserUpdatePassword() http.HandlerFunc {
	type request struct {
		Password string `json:"password"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		logger := hlog.FromRequest(r)

		var req request
		err := s.readJSON(w, r, &req)
		if err != nil {
			logger.Error().Err(err).Msg("api-UserUpdatePassword-readJson")
		}
		userID := chi.URLParam(r, "userID")
		repo := repository.New(s.db, *logger)

		err = repo.UserRepository.UserUpdatePassword(r.Context(), userID, req.Password)
		if err != nil {
			if errors.Is(err, repository.ErrUserNotFound) {
				s.writeJSON(w, http.StatusNotFound, envelope{"error": err.Error()}, nil)
				return
			}
			logger.Error().Err(err).Msg("api-UserUpdatePassword-UserUpdatePassword")
			s.writeJSON(w, http.StatusInternalServerError, envelope{"error": http.StatusText(http.StatusInternalServerError)}, nil)
			return
		}

		s.writeJSON(w, http.StatusOK, envelope{"status": "User Password Updated Successfully"}, nil)

	}
}

// TODO!: on delete handler is needed for the user and token relationship
// and maybe other future relations as well
func (s *Server) UserDelete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := hlog.FromRequest(r)
		userID := chi.URLParam(r, "userID")
		repo := repository.New(s.db, *logger)

		err := repo.UserRepository.UserDelete(r.Context(), userID)
		if err != nil {
			if errors.Is(err, repository.ErrUserNotFound) {
				s.writeJSON(w, http.StatusNotFound, envelope{"error": http.StatusText(http.StatusNotFound)}, nil)
				return
			}

			logger.Error().Err(err).Msg("api-UserDelete-UserDelete")
			s.writeJSON(w, http.StatusInternalServerError, envelope{"error": http.StatusText(http.StatusInternalServerError)}, nil)
			return
		}

		s.writeJSON(w, http.StatusOK, envelope{"status": "User deleted successfully"}, nil)

	}
}
