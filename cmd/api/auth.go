package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/mrityunjaygr8/go-oink/internal/repository"
	"github.com/rs/zerolog/hlog"
)

func (s *Server) AuthLogin() http.HandlerFunc {
	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type response struct {
		Token    string    `json:"token"`
		Type     string    `json:"type"`
		CreateAt time.Time `json:"create_at"`
		UserID   string    `json:"userID"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		logger := hlog.FromRequest(r)
		var req request

		err := s.readJSON(w, r, &req)
		if err != nil {
			logger.Error().Err(err).Msg("api-AuthLogin-readJson")
			s.writeJSON(w, http.StatusBadRequest, envelope{"error": err}, nil)
			return
		}

		repo := repository.New(s.db, *logger)
		user, err := repo.UserRepository.UserAuthenticate(r.Context(), req.Email, req.Password)
		if err != nil {
			if errors.Is(err, repository.ErrUserCredsInvalid) {
				s.writeJSON(w, http.StatusForbidden, envelope{"error": repository.ErrUserCredsInvalid.Error()}, nil)
				return
			}

			if errors.Is(err, repository.ErrUserNotFound) {
				s.writeJSON(w, http.StatusForbidden, envelope{"error": repository.ErrUserCredsInvalid.Error()}, nil)
				return
			}

			logger.Error().Err(err).Msg("api-AuthLogin-UserAuthenticate")
			s.writeJSON(w, http.StatusInternalServerError, envelope{"error": http.StatusText(http.StatusInternalServerError)}, nil)
			return
		}

		tokens, err := repo.TokenRepository.TokenListUserType(r.Context(), user.ID, repository.TokenTypeLogin)
		if err != nil {
			logger.Error().Err(err).Msg("api-AuthLogin-TokenListUserType")
			s.writeJSON(w, http.StatusInternalServerError, envelope{"error": http.StatusText(http.StatusInternalServerError)}, nil)
			return
		}
		var token *repository.Token
		if len(*tokens) == 0 {
			token, err = repo.TokenRepository.TokenLoginCreate(r.Context(), user.ID)
			if err != nil {
				logger.Error().Err(err).Msg("api-AuthLogin-TokenLoginCreate")
				s.writeJSON(w, http.StatusInternalServerError, envelope{"error": http.StatusText(http.StatusInternalServerError)}, nil)
				return
			}
		} else {
			token = &(*tokens)[0]
		}

		resp := response{Token: token.Token, Type: string(token.Type), CreateAt: token.CreatedAt, UserID: token.UserID}
		s.writeJSON(w, http.StatusOK, envelope{"token": resp}, nil)

	}
}

func (s *Server) AuthMe() http.HandlerFunc {
	type response struct {
		Email     string    `json:"email"`
		Password  string    `json:"-"`
		Username  string    `json:"username"`
		ID        string    `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at,omitempty"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		logger := hlog.FromRequest(r)
		user := r.Context().Value("user")
		if user == nil {
			logger.Error().Any("User", user).Msg("api-AuthMe-userNil")
			s.writeJSON(w, http.StatusInternalServerError, envelope{"error": http.StatusText(http.StatusInternalServerError)}, nil)
			return
		}

		u, ok := user.(*repository.User)
		if !ok {
			logger.Error().Any("User", user).Msg("api-AuthMe-userTypeAssertion")
			s.writeJSON(w, http.StatusInternalServerError, envelope{"error": http.StatusText(http.StatusInternalServerError)}, nil)
			return
		}

		res := response{
			Email:     u.Email,
			Username:  u.Username,
			ID:        u.ID,
			CreatedAt: u.CreatedAt,
			UpdatedAt: u.UpdatedAt,
		}
		s.writeJSON(w, http.StatusOK, envelope{"user": res}, nil)
	}
}

// func (s *Server) AuthLogout() http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
//
// 	}
// }
