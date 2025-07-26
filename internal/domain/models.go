package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID             uuid.UUID `json:"id"`
	Username       string    `json:"username"`
	Email          string    `json:"email"`
	HashedPassword string    `json:"-"`
	AvatarURL      *string   `json:"avatarUrl,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
}

type PublicUser struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	AvatarURL *string   `json:"avatarUrl,omitempty"`
}

type Habit struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"userId"`
	Name      string    `json:"name"`
	ColorHue  int       `json:"colorHue"`
	CreatedAt time.Time `json:"createdAt"`
}

type HabitLog struct {
	ID        uuid.UUID `json:"id"`
	HabitID   uuid.UUID `json:"habitId"`
	LogDate   time.Time `json:"date"`
	Status    bool      `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type HabitWithLogs struct {
	Habit
	Logs []HabitLog `json:"logs"`
}
