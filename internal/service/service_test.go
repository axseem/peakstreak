package service

import (
	"context"
	"errors"
	"mime/multipart"
	"os"
	"path/filepath"
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

func (m *MockRepository) GetUserAvatar(ctx context.Context, userID uuid.UUID) (*string, error) {
	args := m.Called(ctx, userID)
	// Handle the case where Get(0) is nil (for a nil *string)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*string), args.Error(1)
}

func (m *MockRepository) UpdateUserAvatar(ctx context.Context, userID uuid.UUID, avatarURL *string) error {
	args := m.Called(ctx, userID, avatarURL)
	return args.Error(0)
}

func (m *MockRepository) CreateHabit(ctx context.Context, habit *domain.Habit) error {
	habit.CreatedAt = time.Now()
	args := m.Called(ctx, habit)
	return args.Error(0)
}

func (m *MockRepository) UpdateHabit(ctx context.Context, habit *domain.Habit) error {
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

func (m *MockRepository) SearchUsersByUsername(ctx context.Context, query string) ([]domain.PublicUser, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.PublicUser), args.Error(1)
}

func (m *MockRepository) FollowUser(ctx context.Context, followerID, followingID uuid.UUID) error {
	args := m.Called(ctx, followerID, followingID)
	return args.Error(0)
}

func (m *MockRepository) UnfollowUser(ctx context.Context, followerID, followingID uuid.UUID) error {
	args := m.Called(ctx, followerID, followingID)
	return args.Error(0)
}

func (m *MockRepository) IsFollowing(ctx context.Context, followerID, followingID uuid.UUID) (bool, error) {
	args := m.Called(ctx, followerID, followingID)
	return args.Bool(0), args.Error(1)
}

func (m *MockRepository) GetFollowerCount(ctx context.Context, userID uuid.UUID) (int, error) {
	args := m.Called(ctx, userID)
	return args.Int(0), args.Error(1)
}

func (m *MockRepository) GetFollowingCount(ctx context.Context, userID uuid.UUID) (int, error) {
	args := m.Called(ctx, userID)
	return args.Int(0), args.Error(1)
}

func (m *MockRepository) GetFollowers(ctx context.Context, userID uuid.UUID) ([]domain.PublicUser, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.PublicUser), args.Error(1)
}

func (m *MockRepository) GetFollowing(ctx context.Context, userID uuid.UUID) ([]domain.PublicUser, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.PublicUser), args.Error(1)
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
	avatarURL := "/uploads/avatars/test.jpg"
	testUser := &domain.User{
		ID:             uuid.New(),
		Username:       "testuser",
		Email:          "test@example.com",
		HashedPassword: string(hashedPassword),
		AvatarURL:      &avatarURL,
	}

	mockRepo.On("GetUserByEmailOrUsername", ctx, "testuser").Return(testUser, nil)

	loggedInUser, err := s.LoginUser(ctx, LoginUserParams{Identifier: "testuser", Password: password})

	assert.NoError(t, err)
	assert.NotNil(t, loggedInUser)
	assert.Equal(t, testUser.ID, loggedInUser.ID)
	assert.NotNil(t, loggedInUser.AvatarURL)
	assert.Equal(t, avatarURL, *loggedInUser.AvatarURL)
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

func TestUpdateHabit_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	s := New(mockRepo)
	ctx := context.Background()

	userID := uuid.New()
	habitID := uuid.New()
	testHabit := &domain.Habit{ID: habitID, UserID: userID, Name: "Old Name"}
	params := UpdateHabitParams{Name: "New Name"}

	mockRepo.On("GetHabitByID", ctx, habitID).Return(testHabit, nil)
	mockRepo.On("UpdateHabit", ctx, mock.MatchedBy(func(h *domain.Habit) bool {
		return h.ID == habitID && h.Name == "New Name"
	})).Return(nil)

	updatedHabit, err := s.UpdateHabit(ctx, params, habitID, userID)

	assert.NoError(t, err)
	assert.NotNil(t, updatedHabit)
	assert.Equal(t, "New Name", updatedHabit.Name)
	mockRepo.AssertExpectations(t)
}

func TestUpdateHabit_AccessDenied(t *testing.T) {
	mockRepo := new(MockRepository)
	s := New(mockRepo)
	ctx := context.Background()

	ownerID := uuid.New()
	attackerID := uuid.New()
	habitID := uuid.New()
	testHabit := &domain.Habit{ID: habitID, UserID: ownerID, Name: "Old Name"}
	params := UpdateHabitParams{Name: "New Name"}

	mockRepo.On("GetHabitByID", ctx, habitID).Return(testHabit, nil)

	_, err := s.UpdateHabit(ctx, params, habitID, attackerID)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrUserAccessDenied))
	mockRepo.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "UpdateHabit", ctx, mock.Anything)
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

	avatarURL := "/uploads/avatars/profile.png"
	testUser := &domain.User{
		ID:             uuid.New(),
		Username:       "testuser",
		HashedPassword: "a_hash",
		AvatarURL:      &avatarURL,
	}
	habit1ID := uuid.New()
	testHabits := []domain.Habit{{ID: habit1ID, UserID: testUser.ID, Name: "Workout"}}
	testLogs := []domain.HabitLog{{ID: uuid.New(), HabitID: habit1ID, Status: true}}

	mockRepo.On("GetUserByUsername", ctx, "testuser").Return(testUser, nil)
	mockRepo.On("GetHabitsByUserID", ctx, testUser.ID).Return(testHabits, nil)
	mockRepo.On("GetLogsForHabits", ctx, []uuid.UUID{habit1ID}).Return(testLogs, nil)
	mockRepo.On("GetFollowerCount", ctx, testUser.ID).Return(10, nil)
	mockRepo.On("GetFollowingCount", ctx, testUser.ID).Return(5, nil)

	profileData, err := s.GetProfileData(ctx, "testuser", testUser.ID)

	assert.NoError(t, err)
	assert.NotNil(t, profileData)
	assert.True(t, profileData.IsOwner)
	assert.Equal(t, "testuser", profileData.User.Username)
	assert.NotNil(t, profileData.User.AvatarURL)
	assert.Equal(t, avatarURL, *profileData.User.AvatarURL)
	assert.Empty(t, profileData.User.HashedPassword, "Hashed password should be cleared")
	assert.Len(t, profileData.Habits, 1)
	assert.Len(t, profileData.Habits[0].Logs, 1, "Logs should be included for the owner")
	assert.Equal(t, 10, profileData.FollowersCount)
	assert.Equal(t, 5, profileData.FollowingCount)
	assert.False(t, profileData.IsFollowing, "Owner should not be 'following' themselves")

	mockRepo.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "IsFollowing", ctx, mock.Anything, mock.Anything)
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
	mockRepo.On("GetFollowerCount", ctx, testUser.ID).Return(25, nil)
	mockRepo.On("GetFollowingCount", ctx, testUser.ID).Return(15, nil)
	mockRepo.On("IsFollowing", ctx, visitorID, testUser.ID).Return(true, nil)

	profileData, err := s.GetProfileData(ctx, "testuser", visitorID)

	assert.NoError(t, err)
	assert.NotNil(t, profileData)
	assert.False(t, profileData.IsOwner)
	assert.Equal(t, "testuser", profileData.User.Username)
	assert.Len(t, profileData.Habits, 1)
	assert.Len(t, profileData.Habits[0].Logs, 1, "Logs should be included for a visitor")
	assert.Equal(t, 25, profileData.FollowersCount)
	assert.Equal(t, 15, profileData.FollowingCount)
	assert.True(t, profileData.IsFollowing)

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

func TestFollowUser_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	s := New(mockRepo)
	ctx := context.Background()

	followerID := uuid.New()
	userToFollowID := uuid.New()

	mockRepo.On("FollowUser", ctx, followerID, userToFollowID).Return(nil)

	err := s.FollowUser(ctx, followerID, userToFollowID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestFollowUser_CannotFollowSelf(t *testing.T) {
	mockRepo := new(MockRepository)
	s := New(mockRepo)
	ctx := context.Background()

	userID := uuid.New()

	err := s.FollowUser(ctx, userID, userID)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrCannotFollowSelf))
	mockRepo.AssertNotCalled(t, "FollowUser", ctx, mock.Anything, mock.Anything)
}

func TestUnfollowUser_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	s := New(mockRepo)
	ctx := context.Background()

	followerID := uuid.New()
	userToUnfollowID := uuid.New()

	mockRepo.On("UnfollowUser", ctx, followerID, userToUnfollowID).Return(nil)

	err := s.UnfollowUser(ctx, followerID, userToUnfollowID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestGetUserByUsername_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	s := New(mockRepo)
	ctx := context.Background()

	testUser := &domain.User{
		ID:             uuid.New(),
		Username:       "testuser",
		Email:          "test@example.com",
		HashedPassword: "a_very_secret_hash",
	}

	mockRepo.On("GetUserByUsername", ctx, "testuser").Return(testUser, nil)

	user, err := s.GetUserByUsername(ctx, "testuser")

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "testuser", user.Username)
	assert.Empty(t, user.HashedPassword, "Hashed password should be cleared from response")
	mockRepo.AssertExpectations(t)
}

func TestGetFollowers_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	s := New(mockRepo)
	ctx := context.Background()

	testUser := &domain.User{
		ID:       uuid.New(),
		Username: "testuser",
	}
	avatarURL1 := "/uploads/avatars/follower1.png"
	followers := []domain.PublicUser{
		{ID: uuid.New(), Username: "follower1", AvatarURL: &avatarURL1},
		{ID: uuid.New(), Username: "follower2"},
	}

	mockRepo.On("GetUserByUsername", ctx, "testuser").Return(testUser, nil)
	mockRepo.On("GetFollowers", ctx, testUser.ID).Return(followers, nil)

	result, err := s.GetFollowers(ctx, "testuser")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)
	assert.Equal(t, "follower1", result[0].Username)
	assert.NotNil(t, result[0].AvatarURL)
	assert.Equal(t, avatarURL1, *result[0].AvatarURL)
	assert.Nil(t, result[1].AvatarURL)

	mockRepo.AssertExpectations(t)
}

func TestGetFollowers_UserNotFound(t *testing.T) {
	mockRepo := new(MockRepository)
	s := New(mockRepo)
	ctx := context.Background()

	mockRepo.On("GetUserByUsername", ctx, "nonexistent").Return(nil, repository.ErrUserNotFound)

	result, err := s.GetFollowers(ctx, "nonexistent")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, repository.ErrUserNotFound))

	mockRepo.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "GetFollowers", ctx, mock.Anything)
}

func TestGetFollowing_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	s := New(mockRepo)
	ctx := context.Background()

	testUser := &domain.User{
		ID:       uuid.New(),
		Username: "testuser",
	}
	avatarURL := "/uploads/avatars/following1.png"
	following := []domain.PublicUser{
		{ID: uuid.New(), Username: "following1", AvatarURL: &avatarURL},
	}

	mockRepo.On("GetUserByUsername", ctx, "testuser").Return(testUser, nil)
	mockRepo.On("GetFollowing", ctx, testUser.ID).Return(following, nil)

	result, err := s.GetFollowing(ctx, "testuser")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 1)
	assert.Equal(t, "following1", result[0].Username)
	assert.NotNil(t, result[0].AvatarURL)
	assert.Equal(t, avatarURL, *result[0].AvatarURL)

	mockRepo.AssertExpectations(t)
}

func TestGetFollowing_UserNotFound(t *testing.T) {
	mockRepo := new(MockRepository)
	s := New(mockRepo)
	ctx := context.Background()

	mockRepo.On("GetUserByUsername", ctx, "nonexistent").Return(nil, repository.ErrUserNotFound)

	result, err := s.GetFollowing(ctx, "nonexistent")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, repository.ErrUserNotFound))

	mockRepo.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "GetFollowing", ctx, mock.Anything)
}

// --- Avatar Tests ---

func TestUpdateUserAvatar_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	s := New(mockRepo)
	ctx := context.Background()
	userID := uuid.New()

	if err := os.MkdirAll(AVATAR_PATH, os.ModePerm); err != nil {
		t.Fatalf("could not create test avatar dir: %v", err)
	}
	defer func() {
		files, _ := filepath.Glob(filepath.Join(AVATAR_PATH, "*"))
		for _, f := range files {
			os.Remove(f)
		}
	}()

	pngHeaderBytes := []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}
	tmpFile, err := os.CreateTemp("", "test-*.png")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	_, err = tmpFile.Write(pngHeaderBytes)
	assert.NoError(t, err)
	_, err = tmpFile.Seek(0, 0)
	assert.NoError(t, err)

	header := &multipart.FileHeader{
		Filename: "test.png",
		Size:     int64(len(pngHeaderBytes)),
	}

	mockRepo.On("GetUserAvatar", ctx, userID).Return(nil, nil)
	mockRepo.On("UpdateUserAvatar", ctx, userID, mock.AnythingOfType("*string")).Return(nil)

	avatarURL, err := s.UpdateUserAvatar(ctx, userID, tmpFile, header)

	assert.NoError(t, err)
	assert.NotEmpty(t, avatarURL)
	assert.Contains(t, avatarURL, "/uploads/avatars/")
	assert.Contains(t, avatarURL, ".png")
	mockRepo.AssertExpectations(t)
}

func TestUpdateUserAvatar_DeletesOld(t *testing.T) {
	mockRepo := new(MockRepository)
	s := New(mockRepo)
	ctx := context.Background()
	userID := uuid.New()

	if err := os.MkdirAll(AVATAR_PATH, os.ModePerm); err != nil {
		t.Fatalf("could not create test avatar dir: %v", err)
	}
	defer func() {
		files, _ := filepath.Glob(filepath.Join(AVATAR_PATH, "*"))
		for _, f := range files {
			os.Remove(f)
		}
	}()

	oldFileName := "old_avatar.png"
	oldFilePath := filepath.Join(AVATAR_PATH, oldFileName)
	oldFile, err := os.Create(oldFilePath)
	assert.NoError(t, err)
	oldFile.Close()

	_, err = os.Stat(oldFilePath)
	assert.NoError(t, err, "Old file should exist before test")

	oldAvatarURL := "/uploads/avatars/" + oldFileName

	mockRepo.On("GetUserAvatar", ctx, userID).Return(&oldAvatarURL, nil)
	mockRepo.On("UpdateUserAvatar", ctx, userID, mock.AnythingOfType("*string")).Return(nil)

	pngHeaderBytes := []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}
	newTmpFile, err := os.CreateTemp("", "new-*.png")
	assert.NoError(t, err)
	defer os.Remove(newTmpFile.Name())
	defer newTmpFile.Close()

	_, err = newTmpFile.Write(pngHeaderBytes)
	assert.NoError(t, err)
	_, err = newTmpFile.Seek(0, 0)
	assert.NoError(t, err)

	newHeader := &multipart.FileHeader{
		Filename: "new_avatar.png",
		Size:     int64(len(pngHeaderBytes)),
	}

	newAvatarURL, err := s.UpdateUserAvatar(ctx, userID, newTmpFile, newHeader)

	assert.NoError(t, err)
	assert.NotEmpty(t, newAvatarURL)
	assert.NotEqual(t, oldAvatarURL, newAvatarURL)
	mockRepo.AssertExpectations(t)

	_, err = os.Stat(oldFilePath)
	assert.True(t, os.IsNotExist(err), "Old file should have been deleted")
}

func TestUpdateUserAvatar_FileTooLarge(t *testing.T) {
	mockRepo := new(MockRepository)
	s := New(mockRepo)
	ctx := context.Background()
	userID := uuid.New()

	tmpFile, err := os.CreateTemp("", "large-*.png")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	header := &multipart.FileHeader{
		Filename: "large.png",
		Size:     MAX_AVATAR_SIZE + 1,
	}

	_, err = s.UpdateUserAvatar(ctx, userID, tmpFile, header)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file size exceeds")
	mockRepo.AssertNotCalled(t, "GetUserAvatar")
}

func TestUpdateUserAvatar_InvalidMimeType(t *testing.T) {
	mockRepo := new(MockRepository)
	s := New(mockRepo)
	ctx := context.Background()
	userID := uuid.New()

	fileContent := "this is a text file"
	tmpFile, err := os.CreateTemp("", "test-*.txt")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	_, err = tmpFile.Write([]byte(fileContent))
	assert.NoError(t, err)
	_, err = tmpFile.Seek(0, 0)
	assert.NoError(t, err)

	header := &multipart.FileHeader{
		Filename: "document.txt",
		Size:     int64(len(fileContent)),
	}

	_, err = s.UpdateUserAvatar(ctx, userID, tmpFile, header)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid file type")
	mockRepo.AssertNotCalled(t, "GetUserAvatar")
}

// --- Search Tests ---

func TestSearchUsers_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	s := New(mockRepo)
	ctx := context.Background()

	query := "test"
	expectedUsers := []domain.PublicUser{
		{ID: uuid.New(), Username: "testuser1"},
		{ID: uuid.New(), Username: "another_test"},
	}

	mockRepo.On("SearchUsersByUsername", ctx, query).Return(expectedUsers, nil)

	users, err := s.SearchUsers(ctx, query)

	assert.NoError(t, err)
	assert.Len(t, users, 2)
	assert.Equal(t, "testuser1", users[0].Username)
	mockRepo.AssertExpectations(t)
}

func TestSearchUsers_EmptyQuery(t *testing.T) {
	mockRepo := new(MockRepository)
	s := New(mockRepo)
	ctx := context.Background()

	users, err := s.SearchUsers(ctx, "  ")

	assert.NoError(t, err)
	assert.Len(t, users, 0)
	mockRepo.AssertNotCalled(t, "SearchUsersByUsername", ctx, mock.Anything)
}

func TestSearchUsers_NoResults(t *testing.T) {
	mockRepo := new(MockRepository)
	s := New(mockRepo)
	ctx := context.Background()

	query := "nonexistent"
	mockRepo.On("SearchUsersByUsername", ctx, query).Return([]domain.PublicUser{}, nil)

	users, err := s.SearchUsers(ctx, query)

	assert.NoError(t, err)
	assert.Len(t, users, 0)
	mockRepo.AssertExpectations(t)
}
