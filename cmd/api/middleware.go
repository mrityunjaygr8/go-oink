package main

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/mrityunjaygr8/go-oink/internal/repository"
	"github.com/rs/zerolog/hlog"
)

// HTTP middleware setting a value on the request context
func (s *Server) AddUserCtx() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr := r.Header.Get("Authorization")
			tokenSlice := strings.Split(tokenStr, " ")
			if len(tokenSlice) != 2 {
				ctx := context.WithValue(r.Context(), "user", nil)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			tokenID := tokenSlice[1]
			logger := hlog.FromRequest(r)
			repo := repository.New(s.db, *logger)
			token, err := repo.TokenRepository.TokenRetrieve(r.Context(), tokenID)
			if err != nil {
				if errors.Is(err, repository.ErrTokenNotFound) {
					ctx := context.WithValue(r.Context(), "user", nil)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
				logger.Error().Err(err).Msg("middleware-AddUserCtx-TokenRetrieve")
				s.writeJSON(w, http.StatusInternalServerError, envelope{"error": http.StatusText(http.StatusInternalServerError)}, nil)
				return
			}

			user, err := repo.UserRepository.UserRetrieve(r.Context(), token.UserID)
			if err != nil {
				logger.Error().Err(err).Msg("middleware-AddUserCtx-UserRetrieve")
				s.writeJSON(w, http.StatusInternalServerError, envelope{"error": http.StatusText(http.StatusInternalServerError)}, nil)
				return
			}

			tmpCtx := context.WithValue(r.Context(), "user", user)
			ctx := context.WithValue(tmpCtx, "token", tokenID)

			// call the next handler in the chain, passing the response writer and
			// the updated request object with the new context value.
			//
			// note: context.Context values are nested, so any previously set
			// values will be accessible as well, and the new `"user"` key
			// will be accessible from this point forward.
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
func (s *Server) AuthorizedGuard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value("user")
		if user == nil {
			s.writeJSON(w, http.StatusForbidden, envelope{"error": http.StatusText(http.StatusForbidden)}, nil)
			return
		}

		next.ServeHTTP(w, r)
	})
}
func (s *Server) UnauthorizedGuard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value("user")

		_, ok := user.(*repository.User)
		if ok {
			s.writeJSON(w, http.StatusBadRequest, envelope{"error": "You are already logged in"}, nil)
			return
		}

		next.ServeHTTP(w, r)
	})
}
