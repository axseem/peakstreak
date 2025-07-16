package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(handler *APIHandler) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	r.Get("/users/{username}", handler.GetPublicUserProfile)

	r.Route("/auth", func(r chi.Router) {
		r.Post("/signup", handler.SignUp)
		r.Post("/login", handler.Login)
	})

	r.Group(func(r chi.Router) {
		r.Use(handler.authMiddleware)

		r.Post("/habits", handler.CreateHabit)
		r.Post("/habits/{habitId}/logs", handler.LogHabit)
	})

	return r
}
