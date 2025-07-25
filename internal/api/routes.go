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

	workDir, _ := os.Getwd()

	uploadsDir := http.Dir(filepath.Join(workDir, "uploads"))
	r.Handle("/uploads/*", http.StripPrefix("/uploads/", http.FileServer(uploadsDir)))

	r.Route("/api", func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		})
		r.Get("/profile/{username}", handler.GetProfilePageData)
		r.Get("/profile/{username}/followers", handler.GetFollowers)
		r.Get("/profile/{username}/following", handler.GetFollowing)

		r.Route("/auth", func(r chi.Router) {
			r.Post("/signup", handler.SignUp)
			r.Post("/login", handler.Login)
		})

		r.Group(func(r chi.Router) {
			r.Use(handler.authMiddleware)

			r.Post("/user/avatar", handler.UploadAvatar)

			r.Post("/habit", handler.CreateHabit)
			r.Put("/habit/{habitId}", handler.UpdateHabit)
			r.Post("/habit/{habitId}/log", handler.LogHabit)

			r.Post("/profile/{username}/follow", handler.FollowUser)
			r.Delete("/profile/{username}/follow", handler.UnfollowUser)
		})
	})

	ServeSPA(r, "./frontend/build")

	return r
}

func ServeSPA(r chi.Router, staticPath string) {
	workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, staticPath))

	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api") || strings.HasPrefix(r.URL.Path, "/uploads") {
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
