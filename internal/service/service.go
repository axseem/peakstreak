package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/axseem/peakstreak/internal/domain"
	"github.com/axseem/peakstreak/internal/repository"
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
	AVATAR_PATH     = "./uploads/avatars"
)

var allowedMimeTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
}

type Service struct {
	repo repository.AllInOneRepository
}

func New(repo repository.AllInOneRepository) *Service {
	return &Service{repo: repo}
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

func (s *Service) LoginUser(ctx context.Context, params LoginUserParams) (*domain.User, error) {
	user, err := s.repo.GetUserByEmailOrUsername(ctx, params.Identifier)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(params.Password))
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	user.HashedPassword = ""
	return user, nil
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

	ext := filepath.Ext(header.Filename)
	newFilename := uuid.New().String() + ext
	filePath := filepath.Join(AVATAR_PATH, newFilename)

	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("could not create file on server: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return "", fmt.Errorf("could not save file: %w", err)
	}

	publicURL := "/uploads/avatars/" + newFilename
	if err := s.repo.UpdateUserAvatar(ctx, userID, &publicURL); err != nil {
		os.Remove(filePath) // Clean up the newly saved file on DB error
		return "", fmt.Errorf("could not update user avatar in database: %w", err)
	}

	if oldAvatarURL != nil && *oldAvatarURL != "" {
		oldFilePath := filepath.Join(".", *oldAvatarURL)
		if err := os.Remove(oldFilePath); err != nil {
			slog.Warn("failed to delete old avatar file", "path", oldFilePath, "error", err)
		}
	}

	return publicURL, nil
}

func (s *Service) GetUsers(ctx context.Context) ([]domain.User, error) {
	users, err := s.repo.GetUsers(ctx)
	if err != nil {
		return nil, err
	}
	return users, nil
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

	isOwner := user.ID == authenticatedUserID

	var habitsWithLogs []domain.HabitWithLogs
	var followersCount, followingCount int
	var isFollowing bool
	var wg sync.WaitGroup
	var errHabits, errFollowers, errFollowing, errIsFollowing error

	wg.Add(4)

	go func() {
		defer wg.Done()
		habitsWithLogs, errHabits = s.GetAllHabitsWithLogs(ctx, user.ID)
	}()

	go func() {
		defer wg.Done()
		followersCount, errFollowers = s.repo.GetFollowerCount(ctx, user.ID)
	}()

	go func() {
		defer wg.Done()
		followingCount, errFollowing = s.repo.GetFollowingCount(ctx, user.ID)
	}()

	go func() {
		defer wg.Done()
		if !isOwner {
			isFollowing, errIsFollowing = s.repo.IsFollowing(ctx, authenticatedUserID, user.ID)
		}
	}()

	wg.Wait()

	if errHabits != nil {
		return nil, fmt.Errorf("failed to get habits: %w", errHabits)
	}
	if errFollowers != nil {
		return nil, fmt.Errorf("failed to get followers count: %w", errFollowers)
	}
	if errFollowing != nil {
		return nil, fmt.Errorf("failed to get following count: %w", errFollowing)
	}
	if errIsFollowing != nil {
		return nil, fmt.Errorf("failed to check following status: %w", errIsFollowing)
	}

	return &ProfileData{
		User:           user,
		Habits:         habitsWithLogs,
		IsOwner:        isOwner,
		FollowersCount: followersCount,
		FollowingCount: followingCount,
		IsFollowing:    isFollowing,
	}, nil
}

type CreateHabitParams struct {
	Name     string
	ColorHue int
}

func (s *Service) CreateHabit(ctx context.Context, params CreateHabitParams, userID uuid.UUID) (*domain.Habit, error) {
	habit := &domain.Habit{
		ID:       uuid.New(),
		UserID:   userID,
		Name:     params.Name,
		ColorHue: params.ColorHue,
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

func (s *Service) GetHabitDetailsForUser(ctx context.Context, habitID, requestingUserID uuid.UUID) (*domain.HabitWithLogs, error) {
	habit, err := s.repo.GetHabitByID(ctx, habitID)
	if err != nil {
		return nil, err
	}

	if habit.UserID != requestingUserID {
		return nil, ErrUserAccessDenied
	}

	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)

	logs, err := s.repo.GetHabitLogs(ctx, habitID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	details := &domain.HabitWithLogs{
		Habit: *habit,
		Logs:  logs,
	}

	return details, nil
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
		logsForHabit := logsByHabitID[habit.ID]
		if logsForHabit == nil {
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
	Status  bool
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
		Status:  params.Status,
	}

	if err := s.repo.UpsertHabitLog(ctx, log); err != nil {
		return nil, err
	}

	return log, nil
}

func (s *Service) FollowUser(ctx context.Context, followerID, userToFollowID uuid.UUID) error {
	if followerID == userToFollowID {
		return ErrCannotFollowSelf
	}
	return s.repo.FollowUser(ctx, followerID, userToFollowID)
}

func (s *Service) UnfollowUser(ctx context.Context, followerID, userToUnfollowID uuid.UUID) error {
	return s.repo.UnfollowUser(ctx, followerID, userToUnfollowID)
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
