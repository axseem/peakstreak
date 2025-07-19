package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/axseem/peakstreak/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

const UNIQUE_VIOLATION_CODE = "23505"

func (r *PostgresRepository) CreateUser(ctx context.Context, user *domain.User) error {
	query := `INSERT INTO users (id, username, email, hashed_password) VALUES ($1, $2, $3, $4)`
	_, err := r.db.Exec(ctx, query, user.ID, user.Username, user.Email, user.HashedPassword)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == UNIQUE_VIOLATION_CODE {
			if strings.Contains(pgErr.ConstraintName, "username") {
				return ErrDuplicateUsername
			}
			if strings.Contains(pgErr.ConstraintName, "email") {
				return ErrDuplicateEmail
			}
		}
		return err
	}
	return nil
}

func (r *PostgresRepository) GetUserByEmailOrUsername(ctx context.Context, identifier string) (*domain.User, error) {
	query := `SELECT id, username, email, hashed_password, created_at FROM users WHERE username = $1 OR email = $1`
	var user domain.User
	err := r.db.QueryRow(ctx, query, identifier).Scan(&user.ID, &user.Username, &user.Email, &user.HashedPassword, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *PostgresRepository) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	query := `SELECT id, username, email, hashed_password, created_at FROM users WHERE username = $1`
	var user domain.User
	err := r.db.QueryRow(ctx, query, username).Scan(&user.ID, &user.Username, &user.Email, &user.HashedPassword, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *PostgresRepository) GetUsers(ctx context.Context) ([]domain.User, error) {
	query := `SELECT id, username, email, created_at FROM users`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.User])
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (r *PostgresRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `SELECT id, username, email, created_at FROM users WHERE id = $1`
	var user domain.User
	err := r.db.QueryRow(ctx, query, id).Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *PostgresRepository) CreateHabit(ctx context.Context, habit *domain.Habit) error {
	query := `INSERT INTO habits (id, user_id, name) VALUES ($1, $2, $3) RETURNING created_at`
	return r.db.QueryRow(ctx, query, habit.ID, habit.UserID, habit.Name).Scan(&habit.CreatedAt)
}

func (r *PostgresRepository) GetHabitsByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Habit, error) {
	query := `SELECT id, user_id, name, created_at FROM habits WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	habits, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.Habit])
	if err != nil {
		return nil, err
	}
	return habits, nil
}

func (r *PostgresRepository) GetHabitByID(ctx context.Context, habitID uuid.UUID) (*domain.Habit, error) {
	query := `SELECT id, user_id, name, created_at FROM habits WHERE id = $1`
	var habit domain.Habit
	err := r.db.QueryRow(ctx, query, habitID).Scan(&habit.ID, &habit.UserID, &habit.Name, &habit.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrHabitNotFound
		}
		return nil, err
	}
	return &habit, nil
}

func (r *PostgresRepository) UpsertHabitLog(ctx context.Context, log *domain.HabitLog) error {
	query := `
        INSERT INTO habit_logs (id, habit_id, log_date, status)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (habit_id, log_date) DO UPDATE SET status = EXCLUDED.status, updated_at = NOW()
        RETURNING id, created_at, updated_at`

	return r.db.QueryRow(ctx, query, log.ID, log.HabitID, log.LogDate, log.Status).Scan(&log.ID, &log.CreatedAt, &log.UpdatedAt)
}

func (r *PostgresRepository) GetHabitLogs(ctx context.Context, habitID uuid.UUID, startDate, endDate time.Time) ([]domain.HabitLog, error) {
	query := `
        SELECT id, habit_id, log_date, status, created_at, updated_at
        FROM habit_logs
        WHERE habit_id = $1 AND log_date >= $2 AND log_date <= $3
        ORDER BY log_date ASC`

	rows, err := r.db.Query(ctx, query, habitID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.HabitLog])
	if err != nil {
		return nil, err
	}
	return logs, nil
}

func (r *PostgresRepository) GetLogsForHabits(ctx context.Context, habitIDs []uuid.UUID, startDate, endDate time.Time) ([]domain.HabitLog, error) {
	if len(habitIDs) == 0 {
		return []domain.HabitLog{}, nil
	}
	query := `
        SELECT id, habit_id, log_date, status, created_at, updated_at
        FROM habit_logs
        WHERE habit_id = ANY($1) AND log_date >= $2 AND log_date <= $3
        ORDER BY habit_id, log_date ASC`

	rows, err := r.db.Query(ctx, query, habitIDs, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.HabitLog])
	if err != nil {
		return nil, err
	}
	return logs, nil
}
