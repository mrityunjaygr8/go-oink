package main

import (
	"net/http"
	"time"

	"github.com/alexliesenfeld/health"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/hlog"
)

type request struct {
	Name string `json:"name"`
	Game string `json:"game"`
}

type response struct {
	NameResp string `json:"name"`
	GameResp string `json:"game"`
}

func (s *Server) routes() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)

	r.Use(hlog.NewHandler(s.l))
	r.Use(hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
		hlog.FromRequest(r).Info().Str("method", r.Method).Stringer("url", r.URL).Int("status", status).
			Int("size", size).
			Dur("duration", duration).
			Msg("")
	}))
	r.Use(hlog.RemoteAddrHandler("ip"))
	r.Use(hlog.UserAgentHandler("user_agent"))
	r.Use(hlog.RefererHandler("refer"))
	r.Use(hlog.RequestIDHandler("req_id", "Request-Id"))

	r.Use(s.AddUserCtx())

	r.Route("/api/v1", func(r chi.Router) {
		r.Group(func(unauthorizedOnlyRouter chi.Router) {
			unauthorizedOnlyRouter.Use(s.UnauthorizedGuard)
			unauthorizedOnlyRouter.Post("/auth/login", s.AuthLogin())
		})
		r.Group(func(authorizedOnlyRouter chi.Router) {
			authorizedOnlyRouter.Use(s.AuthorizedGuard)

			authorizedOnlyRouter.Get("/users", s.UserList())
			authorizedOnlyRouter.Post("/users", s.UserCreate())

			authorizedOnlyRouter.Get("/users/{userID}", s.UserRetrieve())
			authorizedOnlyRouter.Delete("/users/{userID}", s.UserDelete())
			authorizedOnlyRouter.Post("/users/{userID}/password", s.UserUpdatePassword())

			authorizedOnlyRouter.Get("/oinks", s.OinkList())
			authorizedOnlyRouter.Get("/auth/me", s.AuthMe())
		})
	})
	r.Get("/health", health.NewHandler(s.health))

	for _, route := range r.Routes() {
		s.l.Info().Any("route", route.Pattern).Msg("asdf")
	}

	return r
}
