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
	query := `SELECT id, username, email, hashed_password, avatar_url, created_at FROM users WHERE username = $1 OR email = $1`
	var user domain.User
	err := r.db.QueryRow(ctx, query, identifier).Scan(&user.ID, &user.Username, &user.Email, &user.HashedPassword, &user.AvatarURL, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *PostgresRepository) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	query := `SELECT id, username, email, hashed_password, avatar_url, created_at FROM users WHERE username = $1`
	var user domain.User
	err := r.db.QueryRow(ctx, query, username).Scan(&user.ID, &user.Username, &user.Email, &user.HashedPassword, &user.AvatarURL, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *PostgresRepository) GetUsers(ctx context.Context) ([]domain.User, error) {
	query := `SELECT id, username, email, avatar_url, created_at FROM users`

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
	query := `SELECT id, username, email, avatar_url, created_at FROM users WHERE id = $1`
	var user domain.User
	err := r.db.QueryRow(ctx, query, id).Scan(&user.ID, &user.Username, &user.Email, &user.AvatarURL, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *PostgresRepository) GetUserAvatar(ctx context.Context, userID uuid.UUID) (*string, error) {
	var avatarURL *string
	query := `SELECT avatar_url FROM users WHERE id = $1`
	err := r.db.QueryRow(ctx, query, userID).Scan(&avatarURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return avatarURL, nil
}

func (r *PostgresRepository) UpdateUserAvatar(ctx context.Context, userID uuid.UUID, avatarURL *string) error {
	query := `UPDATE users SET avatar_url = $1 WHERE id = $2`
	_, err := r.db.Exec(ctx, query, avatarURL, userID)
	return err
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

func (r *PostgresRepository) UpdateHabit(ctx context.Context, habit *domain.Habit) error {
	query := `UPDATE habits SET name = $1 WHERE id = $2`
	tag, err := r.db.Exec(ctx, query, habit.Name, habit.ID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrHabitNotFound
	}
	return nil
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

func (r *PostgresRepository) GetLogsForHabits(ctx context.Context, habitIDs []uuid.UUID) ([]domain.HabitLog, error) {
	if len(habitIDs) == 0 {
		return []domain.HabitLog{}, nil
	}
	query := `
        SELECT id, habit_id, log_date, status, created_at, updated_at
        FROM habit_logs
        WHERE habit_id = ANY($1)
        ORDER BY habit_id, log_date ASC`

	rows, err := r.db.Query(ctx, query, habitIDs)
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

func (r *PostgresRepository) FollowUser(ctx context.Context, followerID, followingID uuid.UUID) error {
	query := `INSERT INTO followers (follower_id, following_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := r.db.Exec(ctx, query, followerID, followingID)
	return err
}

func (r *PostgresRepository) UnfollowUser(ctx context.Context, followerID, followingID uuid.UUID) error {
	query := `DELETE FROM followers WHERE follower_id = $1 AND following_id = $2`
	_, err := r.db.Exec(ctx, query, followerID, followingID)
	return err
}

func (r *PostgresRepository) IsFollowing(ctx context.Context, followerID, followingID uuid.UUID) (bool, error) {
	if followerID == uuid.Nil {
		return false, nil
	}
	query := `SELECT EXISTS(SELECT 1 FROM followers WHERE follower_id = $1 AND following_id = $2)`
	var isFollowing bool
	err := r.db.QueryRow(ctx, query, followerID, followingID).Scan(&isFollowing)
	return isFollowing, err
}

func (r *PostgresRepository) GetFollowerCount(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM followers WHERE following_id = $1`
	var count int
	err := r.db.QueryRow(ctx, query, userID).Scan(&count)
	return count, err
}

func (r *PostgresRepository) GetFollowingCount(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM followers WHERE follower_id = $1`
	var count int
	err := r.db.QueryRow(ctx, query, userID).Scan(&count)
	return count, err
}

func (r *PostgresRepository) GetFollowers(ctx context.Context, userID uuid.UUID) ([]domain.PublicUser, error) {
	query := `
		SELECT u.id, u.username, u.avatar_url
		FROM users u
		JOIN followers f ON u.id = f.follower_id
		WHERE f.following_id = $1
		ORDER BY f.created_at DESC`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	users, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.PublicUser])
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (r *PostgresRepository) GetFollowing(ctx context.Context, userID uuid.UUID) ([]domain.PublicUser, error) {
	query := `
		SELECT u.id, u.username, u.avatar_url
		FROM users u
		JOIN followers f ON u.id = f.following_id
		WHERE f.follower_id = $1
		ORDER BY f.created_at DESC`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	users, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.PublicUser])
	if err != nil {
		return nil, err
	}
	return users, nil
}
