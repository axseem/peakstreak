package repository

import (
	"context"
	"time"

	"github.com/axseem/peakstreak/internal/domain"
	"github.com/google/uuid"
)

var (
	ErrUserNotFound      = NewRepositoryError("user not found")
	ErrHabitNotFound     = NewRepositoryError("habit not found")
	ErrDuplicateUsername = NewRepositoryError("username already exists")
	ErrDuplicateEmail    = NewRepositoryError("email already exists")
	ErrDuplicateHabitLog = NewRepositoryError("habit log for this date already exists")
)

type RepositoryError struct {
	s string
}

func (e *RepositoryError) Error() string {
	return e.s
}

func NewRepositoryError(s string) error {
	return &RepositoryError{s}
}

type UserRepository interface {
	CreateUser(ctx context.Context, user *domain.User) error
	GetUserByUsername(ctx context.Context, username string) (*domain.User, error)
	GetUserByEmailOrUsername(ctx context.Context, identifier string) (*domain.User, error)
	GetUsers(ctx context.Context) ([]domain.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
}

type HabitRepository interface {
	CreateHabit(ctx context.Context, habit *domain.Habit) error
	GetHabitsByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Habit, error)
	GetHabitByID(ctx context.Context, habitID uuid.UUID) (*domain.Habit, error)
	UpdateHabit(ctx context.Context, habit *domain.Habit) error
	UpsertHabitLog(ctx context.Context, log *domain.HabitLog) error
	GetHabitLogs(ctx context.Context, habitID uuid.UUID, startDate, endDate time.Time) ([]domain.HabitLog, error)
	GetLogsForHabits(ctx context.Context, habitIDs []uuid.UUID) ([]domain.HabitLog, error)
}

type AllInOneRepository interface {
	UserRepository
	HabitRepository
}
