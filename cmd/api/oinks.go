package main

import (
	"net/http"
	"time"

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
