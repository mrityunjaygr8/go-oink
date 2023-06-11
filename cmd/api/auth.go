package main

import (
	"errors"
	"net/http"

	"github.com/mrityunjaygr8/go-oink/internal/repository"
	"github.com/rs/zerolog/hlog"
)

func (s *Server) AuthLogin() http.HandlerFunc {
	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type response struct{}

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
		err = repo.UserRepository.UserAuthenticate(r.Context(), req.Email, req.Password)
		if err != nil {
			if errors.Is(err, repository.ErrUserCredsInvalid) {
				s.writeJSON(w, http.StatusForbidden, envelope{"error": repository.ErrUserCredsInvalid.Error()}, nil)
				return
			}

			logger.Error().Err(err).Msg("api-AuthLogin-UserAuthenticate")
			s.writeJSON(w, http.StatusInternalServerError, envelope{"error": http.StatusText(http.StatusInternalServerError)}, nil)
			return
		}

		s.writeJSON(w, http.StatusOK, envelope{"status": "you are logged in "}, nil)

	}
}
