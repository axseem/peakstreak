package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/axseem/peakstreak/internal/auth"
	"github.com/axseem/peakstreak/internal/domain"
	"github.com/axseem/peakstreak/internal/repository"
	"github.com/axseem/peakstreak/internal/storage"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserAccessDenied   = errors.New("user does not have permission to access this resource")
	ErrCannotFollowSelf   = errors.New("cannot follow yourself")
)

const (
	KB = 1024
	MB = 1024 * KB
)

const (
	MAX_AVATAR_SIZE = 2 * MB
)

var allowedMimeTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
}

type Service struct {
	repo    repository.IRepository
	storage storage.FileStorage
}

func New(repo repository.IRepository, storage storage.FileStorage) *Service {
	return &Service{
		repo:    repo,
		storage: storage,
	}
}

type CreateUserParams struct {
	Username string
	Email    string
	Password string
}

func (s *Service) CreateUser(ctx context.Context, params CreateUserParams) (*domain.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(params.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &domain.User{
		ID:             uuid.New(),
		Username:       params.Username,
		Email:          params.Email,
		HashedPassword: string(hashedPassword),
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	user.HashedPassword = ""
	return user, nil
}

type LoginUserParams struct {
	Identifier string
	Password   string
}

func (s *Service) LoginUser(ctx context.Context, params LoginUserParams, jwtSecret string, jwtExpiresIn time.Duration) (*domain.User, string, error) {
	user, err := s.repo.GetUserByEmailOrUsername(ctx, params.Identifier)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, "", ErrInvalidCredentials
		}
		return nil, "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(params.Password))
	if err != nil {
		return nil, "", ErrInvalidCredentials
	}

	token, err := auth.GenerateToken(user.ID, jwtSecret, jwtExpiresIn)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	user.HashedPassword = ""
	return user, token, nil
}

func (s *Service) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	// First, get the avatar URL so we can delete the file after the DB entry is gone.
	avatarURL, err := s.repo.GetUserAvatar(ctx, userID)
	if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
		slog.Warn("could not get user avatar before deletion", "userID", userID, "error", err)
		// Continue with deletion even if we can't get the avatar URL.
	}

	// Now, delete the user from the database.
	if err := s.repo.DeleteUser(ctx, userID); err != nil {
		return err
	}

	// If user deletion was successful and an avatar existed, delete the file.
	if avatarURL != nil && *avatarURL != "" {
		if err := s.storage.Delete(ctx, *avatarURL); err != nil {
			// Log a warning, but don't return an error since the user is already deleted.
			slog.Warn("failed to delete user's avatar file after account deletion", "url", *avatarURL, "error", err)
		}
	}

	return nil
}

func (s *Service) UpdateUserAvatar(ctx context.Context, userID uuid.UUID, file multipart.File, header *multipart.FileHeader) (string, error) {
	if header.Size > MAX_AVATAR_SIZE {
		return "", errors.New("file size exceeds the 2MB limit")
	}

	buff := make([]byte, 512)
	if _, err := file.Read(buff); err != nil {
		return "", fmt.Errorf("could not read file for mime type detection: %w", err)
	}
	if _, err := file.Seek(0, 0); err != nil {
		return "", fmt.Errorf("could not reset file read pointer: %w", err)
	}
	mimeType := http.DetectContentType(buff)
	if !allowedMimeTypes[mimeType] {
		return "", fmt.Errorf("invalid file type: %s. Only jpeg and png are allowed", mimeType)
	}

	oldAvatarURL, err := s.repo.GetUserAvatar(ctx, userID)
	if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
		return "", fmt.Errorf("could not get user's current avatar: %w", err)
	}

	newFilename := uuid.New().String() + ".png" // Standardize on a single extension
	publicURL, err := s.storage.Save(ctx, file, newFilename)
	if err != nil {
		return "", fmt.Errorf("could not save file: %w", err)
	}

	if err := s.repo.UpdateUserAvatar(ctx, userID, &publicURL); err != nil {
		// Attempt to clean up the newly saved file on DB error
		if cleanupErr := s.storage.Delete(ctx, publicURL); cleanupErr != nil {
			slog.Warn("failed to cleanup orphaned avatar file", "url", publicURL, "error", cleanupErr)
		}
		return "", fmt.Errorf("could not update user avatar in database: %w", err)
	}

	if oldAvatarURL != nil && *oldAvatarURL != "" {
		if err := s.storage.Delete(ctx, *oldAvatarURL); err != nil {
			slog.Warn("failed to delete old avatar file", "url", *oldAvatarURL, "error", err)
		}
	}

	return publicURL, nil
}

func (s *Service) SearchUsers(ctx context.Context, query string) ([]domain.PublicUser, error) {
	if strings.TrimSpace(query) == "" {
		return []domain.PublicUser{}, nil
	}
	return s.repo.SearchUsersByUsername(ctx, query)
}

func (s *Service) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	user, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	user.HashedPassword = ""
	return user, nil
}

type ProfileData struct {
	User           *domain.User           `json:"user"`
	Habits         []domain.HabitWithLogs `json:"habits"`
	IsOwner        bool                   `json:"isOwner"`
	FollowersCount int                    `json:"followersCount"`
	FollowingCount int                    `json:"followingCount"`
	IsFollowing    bool                   `json:"isFollowing"`
}

func (s *Service) GetProfileData(ctx context.Context, username string, authenticatedUserID uuid.UUID) (*ProfileData, error) {
	user, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	user.HashedPassword = ""

	habits, err := s.GetAllHabitsWithLogs(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get habits: %w", err)
	}

	followersCount, err := s.repo.GetFollowerCount(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get follower count: %w", err)
	}

	followingCount, err := s.repo.GetFollowingCount(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get following count: %w", err)
	}

	isOwner := user.ID == authenticatedUserID
	var isFollowing bool
	if !isOwner && authenticatedUserID != uuid.Nil {
		isFollowing, err = s.repo.IsFollowing(ctx, authenticatedUserID, user.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to check following status: %w", err)
		}
	}

	return &ProfileData{
		User:           user,
		Habits:         habits,
		IsOwner:        isOwner,
		FollowersCount: followersCount,
		FollowingCount: followingCount,
		IsFollowing:    isFollowing,
	}, nil
}

type CreateHabitParams struct {
	Name      string
	ColorHue  int
	IsBoolean bool
}

func (s *Service) CreateHabit(ctx context.Context, params CreateHabitParams, userID uuid.UUID) (*domain.Habit, error) {
	habit := &domain.Habit{
		ID:        uuid.New(),
		UserID:    userID,
		Name:      params.Name,
		ColorHue:  params.ColorHue,
		IsBoolean: params.IsBoolean,
	}

	if err := s.repo.CreateHabit(ctx, habit); err != nil {
		return nil, err
	}

	return habit, nil
}

type UpdateHabitParams struct {
	Name     string
	ColorHue int
}

func (s *Service) UpdateHabit(ctx context.Context, params UpdateHabitParams, habitID, userID uuid.UUID) (*domain.Habit, error) {
	habit, err := s.repo.GetHabitByID(ctx, habitID)
	if err != nil {
		return nil, err
	}
	if habit.UserID != userID {
		return nil, ErrUserAccessDenied
	}

	habit.Name = params.Name
	habit.ColorHue = params.ColorHue

	if err := s.repo.UpdateHabit(ctx, habit); err != nil {
		return nil, err
	}

	return habit, nil
}

func (s *Service) DeleteHabit(ctx context.Context, habitID, userID uuid.UUID) error {
	err := s.repo.DeleteHabit(ctx, habitID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrHabitNotFound) {
			return ErrUserAccessDenied
		}
		return err
	}
	return nil
}

func (s *Service) GetAllHabitsWithLogs(ctx context.Context, userID uuid.UUID) ([]domain.HabitWithLogs, error) {
	habits, err := s.repo.GetHabitsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if len(habits) == 0 {
		return []domain.HabitWithLogs{}, nil
	}

	habitIDs := make([]uuid.UUID, len(habits))
	for i, habit := range habits {
		habitIDs[i] = habit.ID
	}

	logs, err := s.repo.GetLogsForHabits(ctx, habitIDs)
	if err != nil {
		return nil, err
	}

	logsByHabitID := make(map[uuid.UUID][]domain.HabitLog)
	for _, log := range logs {
		logsByHabitID[log.HabitID] = append(logsByHabitID[log.HabitID], log)
	}

	habitsWithLogs := make([]domain.HabitWithLogs, len(habits))
	for i, habit := range habits {
		logsForHabit, ok := logsByHabitID[habit.ID]
		if !ok {
			logsForHabit = make([]domain.HabitLog, 0)
		}
		habitsWithLogs[i] = domain.HabitWithLogs{
			Habit: habit,
			Logs:  logsForHabit,
		}
	}

	return habitsWithLogs, nil
}

type LogHabitParams struct {
	HabitID uuid.UUID
	Date    time.Time
	Value   int
}

func (s *Service) LogHabit(ctx context.Context, params LogHabitParams, userID uuid.UUID) (*domain.HabitLog, error) {
	habit, err := s.repo.GetHabitByID(ctx, params.HabitID)
	if err != nil {
		return nil, err
	}
	if habit.UserID != userID {
		return nil, ErrUserAccessDenied
	}

	log := &domain.HabitLog{
		ID:      uuid.New(),
		HabitID: params.HabitID,
		LogDate: params.Date,
		Value:   params.Value,
	}

	if err := s.repo.UpsertHabitLog(ctx, log); err != nil {
		return nil, err
	}

	return log, nil
}

func (s *Service) FollowUserByUsername(ctx context.Context, followerID uuid.UUID, usernameToFollow string) error {
	userToFollow, err := s.repo.GetUserByUsername(ctx, usernameToFollow)
	if err != nil {
		return err
	}
	if followerID == userToFollow.ID {
		return ErrCannotFollowSelf
	}
	return s.repo.FollowUser(ctx, followerID, userToFollow.ID)
}

func (s *Service) UnfollowUserByUsername(ctx context.Context, followerID uuid.UUID, usernameToUnfollow string) error {
	userToUnfollow, err := s.repo.GetUserByUsername(ctx, usernameToUnfollow)
	if err != nil {
		return err // Propagates ErrUserNotFound
	}
	return s.repo.UnfollowUser(ctx, followerID, userToUnfollow.ID)
}

func (s *Service) GetFollowers(ctx context.Context, username string) ([]domain.PublicUser, error) {
	user, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	return s.repo.GetFollowers(ctx, user.ID)
}

func (s *Service) GetFollowing(ctx context.Context, username string) ([]domain.PublicUser, error) {
	user, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	return s.repo.GetFollowing(ctx, user.ID)
}

func (s *Service) GetLeaderboard(ctx context.Context) ([]domain.LeaderboardEntry, error) {
	return s.repo.GetLeaderboard(ctx, 50)
}

func (s *Service) GetExplorePage(ctx context.Context) ([]domain.ExploreEntry, error) {
	return s.repo.GetExplorePage(ctx, 20)
}
