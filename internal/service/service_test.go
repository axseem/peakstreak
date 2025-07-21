package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/axseem/peakstreak/internal/domain"
	"github.com/axseem/peakstreak/internal/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreateUser(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockRepository) GetUserByEmailOrUsername(ctx context.Context, identifier string) (*domain.User, error) {
	args := m.Called(ctx, identifier)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockRepository) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockRepository) CreateHabit(ctx context.Context, habit *domain.Habit) error {
	habit.CreatedAt = time.Now()
	args := m.Called(ctx, habit)
	return args.Error(0)
}

func (m *MockRepository) GetHabitsByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Habit, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Habit), args.Error(1)
}

func (m *MockRepository) GetHabitByID(ctx context.Context, habitID uuid.UUID) (*domain.Habit, error) {
	args := m.Called(ctx, habitID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Habit), args.Error(1)
}

func (m *MockRepository) UpsertHabitLog(ctx context.Context, log *domain.HabitLog) error {
	log.ID = uuid.New()
	log.CreatedAt = time.Now()
	log.UpdatedAt = time.Now()
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *MockRepository) GetHabitLogs(ctx context.Context, habitID uuid.UUID, start, end time.Time) ([]domain.HabitLog, error) {
	args := m.Called(ctx, habitID, start, end)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.HabitLog), args.Error(1)
}

func (m *MockRepository) GetLogsForHabits(ctx context.Context, habitIDs []uuid.UUID) ([]domain.HabitLog, error) {
	args := m.Called(ctx, habitIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.HabitLog), args.Error(1)
}

func (m *MockRepository) GetUsers(ctx context.Context) ([]domain.User, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.User), args.Error(1)
}

func TestCreateUser_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	s := New(mockRepo)
	ctx := context.Background()

	params := CreateUserParams{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	mockRepo.On("CreateUser", ctx, mock.AnythingOfType("*domain.User")).Return(nil)

	user, err := s.CreateUser(ctx, params)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, params.Username, user.Username)
	assert.Equal(t, params.Email, user.Email)
	assert.Empty(t, user.HashedPassword, "Hashed password should be cleared from response")
	mockRepo.AssertExpectations(t)
}

func TestLoginUser_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	s := New(mockRepo)
	ctx := context.Background()

	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	testUser := &domain.User{
		ID:             uuid.New(),
		Username:       "testuser",
		Email:          "test@example.com",
		HashedPassword: string(hashedPassword),
	}

	mockRepo.On("GetUserByEmailOrUsername", ctx, "testuser").Return(testUser, nil)

	loggedInUser, err := s.LoginUser(ctx, LoginUserParams{Identifier: "testuser", Password: password})

	assert.NoError(t, err)
	assert.NotNil(t, loggedInUser)
	assert.Equal(t, testUser.ID, loggedInUser.ID)
	assert.Empty(t, loggedInUser.HashedPassword)
	mockRepo.AssertExpectations(t)
}

func TestLoginUser_InvalidPassword(t *testing.T) {
	mockRepo := new(MockRepository)
	s := New(mockRepo)
	ctx := context.Background()

	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	testUser := &domain.User{
		ID:             uuid.New(),
		Username:       "testuser",
		HashedPassword: string(hashedPassword),
	}

	mockRepo.On("GetUserByEmailOrUsername", ctx, "testuser").Return(testUser, nil)

	loggedInUser, err := s.LoginUser(ctx, LoginUserParams{Identifier: "testuser", Password: "wrongpassword"})

	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidCredentials))
	assert.Nil(t, loggedInUser)
	mockRepo.AssertExpectations(t)
}

func TestLogHabit_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	s := New(mockRepo)
	ctx := context.Background()

	userID := uuid.New()
	habitID := uuid.New()
	testHabit := &domain.Habit{ID: habitID, UserID: userID, Name: "Read"}

	params := LogHabitParams{
		HabitID: habitID,
		Date:    time.Now().Truncate(24 * time.Hour),
		Status:  true,
	}

	mockRepo.On("GetHabitByID", ctx, habitID).Return(testHabit, nil)
	mockRepo.On("UpsertHabitLog", ctx, mock.AnythingOfType("*domain.HabitLog")).Return(nil)

	log, err := s.LogHabit(ctx, params, userID)

	assert.NoError(t, err)
	assert.NotNil(t, log)
	assert.Equal(t, params.Status, log.Status)
	assert.Equal(t, params.HabitID, log.HabitID)
	mockRepo.AssertExpectations(t)
}

func TestLogHabit_AccessDenied(t *testing.T) {
	mockRepo := new(MockRepository)
	s := New(mockRepo)
	ctx := context.Background()

	ownerUserID := uuid.New()
	attackerUserID := uuid.New()
	habitID := uuid.New()
	testHabit := &domain.Habit{ID: habitID, UserID: ownerUserID, Name: "Read"}

	params := LogHabitParams{
		HabitID: habitID,
		Date:    time.Now(),
		Status:  true,
	}

	mockRepo.On("GetHabitByID", ctx, habitID).Return(testHabit, nil)

	log, err := s.LogHabit(ctx, params, attackerUserID)

	assert.Error(t, err)
	assert.Nil(t, log)
	assert.True(t, errors.Is(err, ErrUserAccessDenied))
	mockRepo.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "UpsertHabitLog", ctx, mock.Anything)
}

func TestGetProfileData_Owner_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	s := New(mockRepo)
	ctx := context.Background()

	testUser := &domain.User{
		ID:             uuid.New(),
		Username:       "testuser",
		HashedPassword: "a_hash",
	}
	habit1ID := uuid.New()
	testHabits := []domain.Habit{{ID: habit1ID, UserID: testUser.ID, Name: "Workout"}}
	testLogs := []domain.HabitLog{{ID: uuid.New(), HabitID: habit1ID, Status: true}}

	mockRepo.On("GetUserByUsername", ctx, "testuser").Return(testUser, nil)
	mockRepo.On("GetHabitsByUserID", ctx, testUser.ID).Return(testHabits, nil)
	mockRepo.On("GetLogsForHabits", ctx, []uuid.UUID{habit1ID}).Return(testLogs, nil)

	profileData, err := s.GetProfileData(ctx, "testuser", testUser.ID)

	assert.NoError(t, err)
	assert.NotNil(t, profileData)
	assert.True(t, profileData.IsOwner)
	assert.Equal(t, "testuser", profileData.User.Username)
	assert.Empty(t, profileData.User.HashedPassword, "Hashed password should be cleared")
	assert.Len(t, profileData.Habits, 1)
	assert.Len(t, profileData.Habits[0].Logs, 1, "Logs should be included for the owner")
	mockRepo.AssertExpectations(t)
}

func TestGetProfileData_NotOwner_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	s := New(mockRepo)
	ctx := context.Background()

	testUser := &domain.User{
		ID:             uuid.New(),
		Username:       "testuser",
		HashedPassword: "a_hash",
	}
	visitorID := uuid.New()
	habit1ID := uuid.New()
	testHabits := []domain.Habit{{ID: habit1ID, UserID: testUser.ID, Name: "Workout"}}
	testLogs := []domain.HabitLog{{ID: uuid.New(), HabitID: habit1ID, Status: true}}

	mockRepo.On("GetUserByUsername", ctx, "testuser").Return(testUser, nil)
	mockRepo.On("GetHabitsByUserID", ctx, testUser.ID).Return(testHabits, nil)
	mockRepo.On("GetLogsForHabits", ctx, []uuid.UUID{habit1ID}).Return(testLogs, nil)

	profileData, err := s.GetProfileData(ctx, "testuser", visitorID)

	assert.NoError(t, err)
	assert.NotNil(t, profileData)
	assert.False(t, profileData.IsOwner)
	assert.Equal(t, "testuser", profileData.User.Username)
	assert.Len(t, profileData.Habits, 1)
	assert.Len(t, profileData.Habits[0].Logs, 1, "Logs should be included for a visitor")
	mockRepo.AssertExpectations(t)
}

func TestGetProfileData_UserNotFound(t *testing.T) {
	mockRepo := new(MockRepository)
	s := New(mockRepo)
	ctx := context.Background()

	mockRepo.On("GetUserByUsername", ctx, "nonexistent").Return(nil, repository.ErrUserNotFound)

	profileData, err := s.GetProfileData(ctx, "nonexistent", uuid.New())

	assert.Error(t, err)
	assert.True(t, errors.Is(err, repository.ErrUserNotFound))
	assert.Nil(t, profileData)
	mockRepo.AssertExpectations(t)
}

func TestGetAllHabitsWithLogs_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	s := New(mockRepo)
	ctx := context.Background()

	userID := uuid.New()
	habit1ID := uuid.New()
	habit2ID := uuid.New()

	testHabits := []domain.Habit{
		{ID: habit1ID, UserID: userID, Name: "Workout"},
		{ID: habit2ID, UserID: userID, Name: "Read"},
	}

	testLogs := []domain.HabitLog{
		{ID: uuid.New(), HabitID: habit1ID, LogDate: time.Now().AddDate(0, 0, -1), Status: true},
		{ID: uuid.New(), HabitID: habit1ID, LogDate: time.Now().AddDate(0, 0, -2), Status: true},
	}

	mockRepo.On("GetHabitsByUserID", ctx, userID).Return(testHabits, nil)
	mockRepo.On("GetLogsForHabits", ctx, []uuid.UUID{habit1ID, habit2ID}).Return(testLogs, nil)

	habitsWithLogs, err := s.GetAllHabitsWithLogs(ctx, userID)

	assert.NoError(t, err)
	assert.NotNil(t, habitsWithLogs)
	assert.Len(t, habitsWithLogs, 2)

	assert.Equal(t, habit1ID, habitsWithLogs[0].ID)
	assert.Equal(t, "Workout", habitsWithLogs[0].Name)
	assert.Len(t, habitsWithLogs[0].Logs, 2)

	assert.Equal(t, habit2ID, habitsWithLogs[1].ID)
	assert.Equal(t, "Read", habitsWithLogs[1].Name)
	assert.Len(t, habitsWithLogs[1].Logs, 0)

	mockRepo.AssertExpectations(t)
}

func TestGetAllHabitsWithLogs_NoHabits(t *testing.T) {
	mockRepo := new(MockRepository)
	s := New(mockRepo)
	ctx := context.Background()
	userID := uuid.New()

	mockRepo.On("GetHabitsByUserID", ctx, userID).Return([]domain.Habit{}, nil)

	habitsWithLogs, err := s.GetAllHabitsWithLogs(ctx, userID)

	assert.NoError(t, err)
	assert.NotNil(t, habitsWithLogs)
	assert.Len(t, habitsWithLogs, 0)

	mockRepo.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "GetLogsForHabits", ctx, mock.Anything)
}
