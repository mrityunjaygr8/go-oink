package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/mrityunjaygr8/go-oink/internal/repository"
	"github.com/rs/zerolog/hlog"
)

func (s *Server) OinkList() http.HandlerFunc {
	type Oink struct {
		Name        string    `json:"name"`
		Description string    `json:"description"`
		ID          string    `json:"id"`
		CreatorID   string    `json:"creator_id"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at,omitempty"`
	}
	type response struct {
		Oinks []Oink `json:"oinks"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		logger := hlog.FromRequest(r)
		repo := repository.New(s.db, *logger)
		o, err := repo.OinkRepository.OinkList(r.Context())
		if err != nil {
			logger.Error().Err(err).Msg("api-OinkList-List")
			s.writeJSON(w, http.StatusInternalServerError, envelope{"error": http.StatusText(http.StatusInternalServerError)}, nil)
			return
		}

		oinks := make([]Oink, 0)
		for _, oink := range *o {
			oinks = append(oinks, Oink{
				Name:        oink.Name,
				Description: oink.Description,
				ID:          oink.ID,
				CreatorID:   oink.CreatorID,
				CreatedAt:   oink.CreatedAt,
				UpdatedAt:   oink.UpdatedAt,
			})
		}

		s.writeJSON(w, http.StatusOK, envelope{"oinks": oinks}, nil)
	}
}

func (s *Server) OinkRetrieve() http.HandlerFunc {
	type response struct {
		Name        string    `json:"name"`
		Description string    `json:"description"`
		ID          string    `json:"id"`
		CreatorID   string    `json:"creator_id"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at,omitempty"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		logger := hlog.FromRequest(r)
		repo := repository.New(s.db, *logger)
		oinkName := chi.URLParam(r, "oinkName")
		oink, err := repo.OinkRepository.OinkRetrieve(r.Context(), oinkName)
		if err != nil {
			if errors.Is(err, repository.ErrOinkNotFound) {
				s.writeJSON(w, http.StatusNotFound, envelope{"error": err.Error()}, nil)
				return
			}
			logger.Error().Err(err).Msg("api-OinkRetrieve-Retrieve")
			s.writeJSON(w, http.StatusInternalServerError, envelope{"error": http.StatusText(http.StatusInternalServerError)}, nil)
			return
		}

		s.writeJSON(w, http.StatusOK, envelope{"oink": oink}, nil)
	}
}

func (s *Server) OinkDelete() http.HandlerFunc {
	type response struct {
		Name        string    `json:"name"`
		Description string    `json:"description"`
		ID          string    `json:"id"`
		CreatorID   string    `json:"creator_id"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at,omitempty"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		logger := hlog.FromRequest(r)
		repo := repository.New(s.db, *logger)
		oinkName := chi.URLParam(r, "oinkName")
		err := repo.OinkRepository.OinkDelete(r.Context(), oinkName)
		if err != nil {
			if errors.Is(err, repository.ErrOinkNotFound) {
				s.writeJSON(w, http.StatusNotFound, envelope{"error": err.Error()}, nil)
				return
			}
			logger.Error().Err(err).Msg("api-OinkDelete-Delete")
			s.writeJSON(w, http.StatusInternalServerError, envelope{"error": http.StatusText(http.StatusInternalServerError)}, nil)
			return
		}

		s.writeJSON(w, http.StatusNoContent, nil, nil)
	}
}

func (s *Server) OinkInsert() http.HandlerFunc {
	type request struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	type response struct {
		Name        string    `json:"name"`
		Description string    `json:"description"`
		ID          string    `json:"id"`
		CreatorID   string    `json:"creator_id"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at,omitempty"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		logger := hlog.FromRequest(r)
		var req request
		err := s.readJSON(w, r, &req)
		if err != nil {
			logger.Error().Err(err).Msg("api-OinkInsert-readJson")
			s.writeJSON(w, http.StatusBadRequest, envelope{"error": err.Error()}, nil)
			return
		}
		repo := repository.New(s.db, *logger)

		user := r.Context().Value("user")

		u, ok := user.(*repository.User)
		if !ok {
			logger.Error().Any("User", user).Msg("api-OinkInsert-userTypeAssertion")
			s.writeJSON(w, http.StatusInternalServerError, envelope{"error": http.StatusText(http.StatusInternalServerError)}, nil)
			return
		}
		tx, err := s.db.Begin()
		if err != nil {
			logger.Error().Err(err).Msg("error creating transaction")
			s.writeJSON(w, http.StatusInternalServerError, envelope{"error": http.StatusText(http.StatusInternalServerError)}, nil)
			return
		}

		oink, err := repo.OinkRepository.OinkInsert(r.Context(), req.Name, req.Description, u.ID)
		if err != nil {
			if rollback := tx.Rollback(); rollback != nil {
				logger.Error().Err(err).Msg("api-OinkInsert-OinkInsert-RollbackError")
				s.writeJSON(w, http.StatusInternalServerError, envelope{"error": http.StatusText(http.StatusInternalServerError)}, nil)
				return
			}
			if errors.Is(err, repository.ErrOinkExists) {
				s.writeJSON(w, http.StatusBadRequest, envelope{"error": repository.ErrOinkExists.Error()}, nil)
				return
			}
			logger.Error().Err(err).Msg("api-OinkInsert-OinkInsert")
			s.writeJSON(w, http.StatusInternalServerError, envelope{"error": http.StatusText(http.StatusInternalServerError)}, nil)
			return
		}

		if err := tx.Commit(); err != nil {
			logger.Error().Err(err).Msg("api-OinkInsert-OinkInsert-CommitError")
			if rollback := tx.Rollback(); rollback != nil {
				logger.Error().Err(err).Msg("api-OinkInsert-OinkInsert-RollbackError")
				s.writeJSON(w, http.StatusInternalServerError, envelope{"error": http.StatusText(http.StatusInternalServerError)}, nil)
				return
			}

			s.writeJSON(w, http.StatusInternalServerError, envelope{"error": http.StatusText(http.StatusInternalServerError)}, nil)
			return
		}
		resp := response{
			Name:        oink.Name,
			Description: oink.Description,
			ID:          oink.ID,
			CreatedAt:   oink.CreatedAt,
			UpdatedAt:   oink.UpdatedAt,
		}

		s.writeJSON(w, http.StatusCreated, envelope{"oink": resp}, nil)
	}
}
