package api

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/axseem/peakstreak/internal/config"
	"github.com/axseem/peakstreak/internal/domain"
	"github.com/axseem/peakstreak/internal/repository"
	"github.com/axseem/peakstreak/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type APIHandler struct {
	service  *service.Service
	validate *validator.Validate
	cfg      *config.Config
}

func NewAPIHandler(s *service.Service, cfg *config.Config) *APIHandler {
	return &APIHandler{
		service:  s,
		validate: validator.New(),
		cfg:      cfg,
	}
}

type SignUpRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50,alphanum"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,printascii"`
}

func (h *APIHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	var req SignUpRequest
	if err := readJSON(r, &req); err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		validationErrorResponse(w, err)
		return
	}

	params := service.CreateUserParams{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	}

	user, err := h.service.CreateUser(r.Context(), params)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateUsername) || errors.Is(err, repository.ErrDuplicateEmail) {
			errorResponse(w, http.StatusConflict, err.Error())
			return
		}
		errorResponse(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	writeJSON(w, http.StatusCreated, user)
}

type LoginRequest struct {
	Identifier string `json:"identifier" validate:"required"`
	Password   string `json:"password" validate:"required"`
}

type LoginResponse struct {
	Token string       `json:"token"`
	User  *domain.User `json:"user"`
}

func (h *APIHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := readJSON(r, &req); err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		validationErrorResponse(w, err)
		return
	}

	params := service.LoginUserParams{
		Identifier: req.Identifier,
		Password:   req.Password,
	}

	user, token, err := h.service.LoginUser(r.Context(), params, h.cfg.JWTSecret, h.cfg.JWTExpiresIn)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			errorResponse(w, http.StatusUnauthorized, "Invalid credentials")
			return
		}
		slog.Error("failed to login", "error", err)
		errorResponse(w, http.StatusInternalServerError, "Failed to login")
		return
	}

	resp := LoginResponse{
		Token: token,
		User:  user,
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *APIHandler) GetProfilePageData(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	authenticatedUserID, _ := getUserIDFromContext(r.Context())

	profileData, err := h.service.GetProfileData(r.Context(), username, authenticatedUserID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			errorResponse(w, http.StatusNotFound, "User not found")
			return
		}
		slog.Error("could not retrieve user profile", "error", err)
		errorResponse(w, http.StatusInternalServerError, "Could not retrieve user profile")
		return
	}

	writeJSON(w, http.StatusOK, profileData)
}

type CreateHabitRequest struct {
	Name      string `json:"name" validate:"required,min=1,max=100"`
	ColorHue  int    `json:"colorHue" validate:"min=0,max=360"`
	IsBoolean bool   `json:"isBoolean"`
}

func (h *APIHandler) CreateHabit(w http.ResponseWriter, r *http.Request) {
	var req CreateHabitRequest
	if err := readJSON(r, &req); err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		validationErrorResponse(w, err)
		return
	}

	userID, ok := getUserIDFromContext(r.Context())
	if !ok {
		errorResponse(w, http.StatusUnauthorized, "Authentication error")
		panic("unreachable: middleware is not correctly applied")
	}

	params := service.CreateHabitParams{
		Name:      req.Name,
		ColorHue:  req.ColorHue,
		IsBoolean: req.IsBoolean,
	}

	habit, err := h.service.CreateHabit(r.Context(), params, userID)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "Failed to create habit")
		return
	}

	writeJSON(w, http.StatusCreated, habit)
}

type UpdateHabitRequest struct {
	Name     string `json:"name" validate:"required,min=1,max=100"`
	ColorHue int    `json:"colorHue" validate:"required,min=0,max=360"`
}

func (h *APIHandler) UpdateHabit(w http.ResponseWriter, r *http.Request) {
	habitIDStr := chi.URLParam(r, "habitId")
	habitID, err := uuid.Parse(habitIDStr)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid habit ID format")
		return
	}

	var req UpdateHabitRequest
	if err := readJSON(r, &req); err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		validationErrorResponse(w, err)
		return
	}

	userID, ok := getUserIDFromContext(r.Context())
	if !ok {
		errorResponse(w, http.StatusUnauthorized, "Authentication error")
		return
	}

	params := service.UpdateHabitParams{
		Name:     req.Name,
		ColorHue: req.ColorHue,
	}

	_, err = h.service.UpdateHabit(r.Context(), params, habitID, userID)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrHabitNotFound):
			errorResponse(w, http.StatusNotFound, "Habit not found")
		case errors.Is(err, service.ErrUserAccessDenied):
			errorResponse(w, http.StatusForbidden, "You do not have permission to update this habit")
		default:
			errorResponse(w, http.StatusInternalServerError, "Failed to update habit")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *APIHandler) DeleteHabit(w http.ResponseWriter, r *http.Request) {
	habitIDStr := chi.URLParam(r, "habitId")
	habitID, err := uuid.Parse(habitIDStr)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid habit ID format")
		return
	}

	userID, ok := getUserIDFromContext(r.Context())
	if !ok {
		errorResponse(w, http.StatusUnauthorized, "Authentication error")
		return
	}

	err = h.service.DeleteHabit(r.Context(), habitID, userID)
	if err != nil {
		if errors.Is(err, service.ErrUserAccessDenied) {
			errorResponse(w, http.StatusForbidden, "You do not have permission to delete this habit")
		} else {
			errorResponse(w, http.StatusInternalServerError, "Failed to delete habit")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

const DATE_FORMAT = "2006-01-02"

type LogHabitRequest struct {
	Date  string `json:"date" validate:"required"`
	Value int    `json:"value" validate:"min=0"`
}

func (h *APIHandler) LogHabit(w http.ResponseWriter, r *http.Request) {
	habitIDStr := chi.URLParam(r, "habitId")
	habitID, err := uuid.Parse(habitIDStr)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid habit ID format")
		return
	}

	var req LogHabitRequest
	if err := readJSON(r, &req); err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		validationErrorResponse(w, err)
		return
	}

	userID, ok := getUserIDFromContext(r.Context())
	if !ok {
		errorResponse(w, http.StatusUnauthorized, "Authentication error")
		return
	}

	logDate, err := time.Parse(DATE_FORMAT, req.Date)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid date format, please use YYYY-MM-DD")
		return
	}

	params := service.LogHabitParams{
		HabitID: habitID,
		Date:    logDate,
		Value:   req.Value,
	}

	log, err := h.service.LogHabit(r.Context(), params, userID)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrHabitNotFound):
			errorResponse(w, http.StatusNotFound, "Habit not found")
		case errors.Is(err, service.ErrUserAccessDenied):
			errorResponse(w, http.StatusForbidden, "You do not have permission to log this habit")
		default:
			errorResponse(w, http.StatusInternalServerError, "Failed to log habit")
		}
		return
	}

	writeJSON(w, http.StatusOK, log)
}

func (h *APIHandler) FollowUser(w http.ResponseWriter, r *http.Request) {
	usernameToFollow := chi.URLParam(r, "username")

	followerID, ok := getUserIDFromContext(r.Context())
	if !ok {
		errorResponse(w, http.StatusUnauthorized, "Authentication error")
		return
	}

	if err := h.service.FollowUserByUsername(r.Context(), followerID, usernameToFollow); err != nil {
		switch {
		case errors.Is(err, service.ErrCannotFollowSelf):
			errorResponse(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, repository.ErrUserNotFound):
			errorResponse(w, http.StatusNotFound, "User to follow not found")
		default:
			slog.Error("failed to follow user", "error", err)
			errorResponse(w, http.StatusInternalServerError, "Failed to follow user")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *APIHandler) UnfollowUser(w http.ResponseWriter, r *http.Request) {
	usernameToUnfollow := chi.URLParam(r, "username")

	followerID, ok := getUserIDFromContext(r.Context())
	if !ok {
		errorResponse(w, http.StatusUnauthorized, "Authentication error")
		return
	}

	if err := h.service.UnfollowUserByUsername(r.Context(), followerID, usernameToUnfollow); err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			errorResponse(w, http.StatusNotFound, "User to unfollow not found")
			return
		}
		slog.Error("failed to unfollow user", "error", err)
		errorResponse(w, http.StatusInternalServerError, "Failed to unfollow user")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *APIHandler) GetFollowers(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	users, err := h.service.GetFollowers(r.Context(), username)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			errorResponse(w, http.StatusNotFound, "User not found")
			return
		}
		slog.Error("could not retrieve followers", "error", err)
		errorResponse(w, http.StatusInternalServerError, "Unexpected internal server error")
		return
	}
	writeJSON(w, http.StatusOK, users)
}

func (h *APIHandler) GetFollowing(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	users, err := h.service.GetFollowing(r.Context(), username)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			errorResponse(w, http.StatusNotFound, "User not found")
			return
		}
		slog.Error("could not retrieve following list", "error", err)
		errorResponse(w, http.StatusInternalServerError, "Unexpected internal server error")
		return
	}
	writeJSON(w, http.StatusOK, users)
}

func (h *APIHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromContext(r.Context())
	if !ok {
		errorResponse(w, http.StatusUnauthorized, "Authentication error")
		return
	}

	if err := r.ParseMultipartForm(service.MAX_AVATAR_SIZE); err != nil {
		errorResponse(w, http.StatusBadRequest, "File too large, max 2MB")
		return
	}

	file, header, err := r.FormFile("avatar")
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "Could not get avatar from form")
		return
	}
	defer file.Close()

	avatarURL, err := h.service.UpdateUserAvatar(r.Context(), userID, file, header)
	if err != nil {
		slog.Error("failed to update avatar", "error", err, "userID", userID)
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"avatarUrl": avatarURL})
}

func (h *APIHandler) SearchUsers(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")

	users, err := h.service.SearchUsers(r.Context(), query)
	if err != nil {
		slog.Error("failed to search users", "error", err, "query", query)
		errorResponse(w, http.StatusInternalServerError, "Failed to search for users")
		return
	}

	writeJSON(w, http.StatusOK, users)
}

func (h *APIHandler) GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	leaderboardData, err := h.service.GetLeaderboard(r.Context())
	if err != nil {
		slog.Error("could not retrieve leaderboard data", "error", err)
		errorResponse(w, http.StatusInternalServerError, "Could not retrieve leaderboard")
		return
	}

	writeJSON(w, http.StatusOK, leaderboardData)
}

func (h *APIHandler) GetExplorePage(w http.ResponseWriter, r *http.Request) {
	exploreData, err := h.service.GetExplorePage(r.Context())
	if err != nil {
		slog.Error("could not retrieve explore page data", "error", err)
		errorResponse(w, http.StatusInternalServerError, "Could not retrieve explore page data")
		return
	}

	writeJSON(w, http.StatusOK, exploreData)
}

func (h *APIHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromContext(r.Context())
	if !ok {
		errorResponse(w, http.StatusUnauthorized, "Authentication error")
		return
	}

	if err := h.service.DeleteUser(r.Context(), userID); err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			// This shouldn't happen if the token is valid, but handle it defensively.
			errorResponse(w, http.StatusNotFound, "User not found")
			return
		}
		slog.Error("failed to delete user", "userID", userID, "error", err)
		errorResponse(w, http.StatusInternalServerError, "Failed to delete account")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
