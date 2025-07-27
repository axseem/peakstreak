package repository

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

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

func (r *PostgresRepository) SearchUsersByUsername(ctx context.Context, query string) ([]domain.PublicUser, error) {
	sqlQuery := `
		SELECT id, username, avatar_url
		FROM users
		WHERE username ILIKE $1
		ORDER BY username
		LIMIT 40`

	rows, err := r.db.Query(ctx, sqlQuery, "%"+query+"%")
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
	query := `INSERT INTO habits (id, user_id, name, color_hue) VALUES ($1, $2, $3, $4) RETURNING created_at`
	return r.db.QueryRow(ctx, query, habit.ID, habit.UserID, habit.Name, habit.ColorHue).Scan(&habit.CreatedAt)
}

func (r *PostgresRepository) GetHabitsByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Habit, error) {
	query := `SELECT id, user_id, name, color_hue, created_at FROM habits WHERE user_id = $1 ORDER BY created_at DESC`
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
	query := `SELECT id, user_id, name, color_hue, created_at FROM habits WHERE id = $1`
	var habit domain.Habit
	err := r.db.QueryRow(ctx, query, habitID).Scan(&habit.ID, &habit.UserID, &habit.Name, &habit.ColorHue, &habit.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrHabitNotFound
		}
		return nil, err
	}
	return &habit, nil
}

func (r *PostgresRepository) UpdateHabit(ctx context.Context, habit *domain.Habit) error {
	query := `UPDATE habits SET name = $1, color_hue = $2 WHERE id = $3`
	tag, err := r.db.Exec(ctx, query, habit.Name, habit.ColorHue, habit.ID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrHabitNotFound
	}
	return nil
}

func (r *PostgresRepository) DeleteHabit(ctx context.Context, habitID, userID uuid.UUID) error {
	query := `DELETE FROM habits WHERE id = $1 AND user_id = $2`
	tag, err := r.db.Exec(ctx, query, habitID, userID)
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

	err := r.db.QueryRow(ctx, query, log.ID, log.HabitID, log.LogDate, log.Status).Scan(&log.ID, &log.CreatedAt, &log.UpdatedAt)
	if err != nil {
		return err
	}
	return nil
}

func (r *PostgresRepository) GetLogsForHabits(ctx context.Context, habitIDs []uuid.UUID) ([]domain.HabitLog, error) {
	if len(habitIDs) == 0 {
		return []domain.HabitLog{}, nil
	}
	query := `
        SELECT id, habit_id, log_date, status, created_at, updated_at
        FROM habit_logs
        WHERE habit_id = ANY($1) AND status = TRUE
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

func (r *PostgresRepository) GetLeaderboard(ctx context.Context, limit int) ([]domain.LeaderboardEntry, error) {
	query := `
WITH RankedUsers AS (
    SELECT
        u.id,
        u.username,
        u.avatar_url,
        COUNT(hl.id) as total_logged_days
    FROM
        users u
    JOIN
        habits h ON u.id = h.user_id
    JOIN
        habit_logs hl ON h.id = hl.habit_id
    WHERE
        hl.status = TRUE
    GROUP BY
        u.id
    ORDER BY
        total_logged_days DESC
    LIMIT $1
),
HabitsWithLogs AS (
    SELECT
        h.user_id,
        json_agg(
            json_build_object(
                'id', h.id,
                'userId', h.user_id,
                'name', h.name,
                'colorHue', h.color_hue,
                'createdAt', to_jsonb(h.created_at),
                'logs', COALESCE(
                    (SELECT json_agg(
                        json_build_object(
                            'id', hl.id,
                            'habitId', hl.habit_id,
                            'date', to_jsonb(hl.log_date::timestamp AT TIME ZONE 'UTC'),
                            'status', hl.status,
                            'createdAt', to_jsonb(hl.created_at),
                            'updatedAt', to_jsonb(hl.updated_at)
                        ) ORDER BY hl.log_date ASC
                    ) FROM habit_logs hl WHERE hl.habit_id = h.id AND hl.status = TRUE),
                    '[]'::json
                )
            )
        ) AS habits
    FROM
        habits h
    WHERE
        h.user_id IN (SELECT id FROM RankedUsers)
    GROUP BY
        h.user_id
)
SELECT json_build_object(
    'user', json_build_object('id', ru.id, 'username', ru.username, 'avatarUrl', ru.avatar_url),
    'totalLoggedDays', ru.total_logged_days,
    'habits', COALESCE(hwl.habits, '[]'::json)
)
FROM RankedUsers ru
LEFT JOIN HabitsWithLogs hwl ON ru.id = hwl.user_id
ORDER BY ru.total_logged_days DESC;
`
	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []domain.LeaderboardEntry
	for rows.Next() {
		var jsonData []byte
		if err := rows.Scan(&jsonData); err != nil {
			return nil, err
		}

		var entry domain.LeaderboardEntry
		if err := json.Unmarshal(jsonData, &entry); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	return entries, rows.Err()
}

func (r *PostgresRepository) GetExplorePage(ctx context.Context, limit int) ([]domain.ExploreEntry, error) {
	query := `
WITH LatestUserLogs AS (
    SELECT DISTINCT ON (h.user_id)
        h.user_id,
        h.id AS habit_id,
        hl.updated_at
    FROM habit_logs hl
    JOIN habits h ON hl.habit_id = h.id
    WHERE hl.status = TRUE
    ORDER BY h.user_id, hl.updated_at DESC
),
ExploreHabits AS (
    SELECT
        lul.user_id,
        h.id,
        h.name,
        h.color_hue,
        h.created_at,
        COALESCE(
            (SELECT json_agg(
                json_build_object(
                    'id', hl.id,
                    'habitId', hl.habit_id,
                    'date', to_jsonb(hl.log_date::timestamp AT TIME ZONE 'UTC'),
                    'status', hl.status,
                    'createdAt', to_jsonb(hl.created_at),
                    'updatedAt', to_jsonb(hl.updated_at)
                ) ORDER BY hl.log_date ASC
            ) FROM habit_logs hl WHERE hl.habit_id = h.id AND hl.status = TRUE),
            '[]'::json
        ) AS logs
    FROM LatestUserLogs lul
    JOIN habits h ON lul.habit_id = h.id
    ORDER BY lul.updated_at DESC
    LIMIT $1
)
SELECT json_build_object(
    'user', json_build_object('id', u.id, 'username', u.username, 'avatarUrl', u.avatar_url),
    'habit', json_build_object(
        'id', eh.id,
        'userId', eh.user_id,
        'name', eh.name,
        'colorHue', eh.color_hue,
        'createdAt', to_jsonb(eh.created_at),
        'logs', eh.logs
    )
)
FROM ExploreHabits eh
JOIN users u ON eh.user_id = u.id;
`

	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []domain.ExploreEntry
	for rows.Next() {
		var jsonData []byte
		if err := rows.Scan(&jsonData); err != nil {
			return nil, err
		}

		var entry domain.ExploreEntry
		if err := json.Unmarshal(jsonData, &entry); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	return entries, rows.Err()
}
