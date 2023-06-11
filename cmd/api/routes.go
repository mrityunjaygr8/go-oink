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
	router := chi.NewRouter()

	// middleware.DefaultLogger = middleware.RequestLogger(customLogFormatter{logger: a.l})

	// router.Use(middleware.RequestID)
	// router.Use(middleware.RealIP)
	// // router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Use(hlog.NewHandler(s.l))
	router.Use(hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
		hlog.FromRequest(r).Info().Str("method", r.Method).Stringer("url", r.URL).Int("status", status).
			Int("size", size).
			Dur("duration", duration).
			Msg("")

	}))
	router.Use(hlog.RemoteAddrHandler("ip"))
	router.Use(hlog.UserAgentHandler("user_agent"))
	router.Use(hlog.RefererHandler("refer"))
	router.Use(hlog.RequestIDHandler("req_id", "Request-Id"))
	// router.Use()

	router.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		hlog.FromRequest(r).Info().Str("user", "current user").Str("status", "ok").Msg("woo")
		w.Write([]byte("pong"))
	})

	router.Post("/ping", func(w http.ResponseWriter, r *http.Request) {
		var req request

		s.readJSON(w, r, &req)

		resp := response{
			NameResp: req.Name,
			GameResp: req.Game,
		}

		s.writeJSON(w, http.StatusOK, envelope{"resp": resp}, nil)

	})

	router.Get("/api/users", s.UserList())
	router.Post("/api/users", s.UserCreate())

	router.Get("/api/users/{userID}", s.UserRetrieve())
	router.Delete("/api/users/{userID}", s.UserDelete())
	router.Post("/api/users/{userID}/password", s.UserUpdatePassword())

	router.Post("/api/auth/login", s.AuthLogin())

	router.Get("/health", health.NewHandler(s.health))

	return router
}
