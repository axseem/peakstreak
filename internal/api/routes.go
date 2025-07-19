package api

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
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

	r.Route("/api", func(r chi.Router) {
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

			// r.Get("/me", handler.GetCurrentUserProfile)
			r.Post("/habits", handler.CreateHabit)
			r.Post("/habits/{habitId}/logs", handler.LogHabit)
		})
	})

	// Serve the Single Page Application (SPA)
	ServeSPA(r, "./frontend/build")

	return r
}

func ServeSPA(r chi.Router, staticPath string) {
	workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, staticPath))

	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api") {
			http.NotFound(w, r)
			return
		}

		path := r.URL.Path
		if _, err := os.Stat(filepath.Join(staticPath, path)); !os.IsNotExist(err) {
			http.FileServer(filesDir).ServeHTTP(w, r)
		} else {
			http.ServeFile(w, r, filepath.Join(staticPath, "index.html"))
		}
	})
}
