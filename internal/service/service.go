package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/axseem/peakstreak/internal/domain"
	"github.com/axseem/peakstreak/internal/repository"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserAccessDenied   = errors.New("user does not have permission to access this resource")
)

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

func (s *Service) GetUsers(ctx context.Context) ([]domain.User, error) {
	users, err := s.repo.GetUsers(ctx)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (s *Service) GetPublicUserProfile(ctx context.Context, username string) (*domain.User, []domain.Habit, error) {
	user, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, nil, err
	}

	habits, err := s.repo.GetHabitsByUserID(ctx, user.ID)
	if err != nil {
		return nil, nil, err
	}

	user.HashedPassword = ""
	return user, habits, nil
}

type CreateHabitParams struct {
	Name string
}

func (s *Service) CreateHabit(ctx context.Context, params CreateHabitParams, userID uuid.UUID) (*domain.Habit, error) {
	habit := &domain.Habit{
		ID:     uuid.New(),
		UserID: userID,
		Name:   params.Name,
	}

	if err := s.repo.CreateHabit(ctx, habit); err != nil {
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

	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -90)

	logs, err := s.repo.GetLogsForHabits(ctx, habitIDs, startDate, endDate)
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
