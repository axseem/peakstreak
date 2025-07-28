package service

import (
	"context"
	"errors"
	"io"
	"mime/multipart"
	"strings"
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
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	// Handle the case where a nil pointer is returned intentionally
	if val, ok := args.Get(0).(*string); ok {
		return val, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockRepository) UpdateUserAvatar(ctx context.Context, userID uuid.UUID, avatarURL *string) error {
	args := m.Called(ctx, userID, avatarURL)
	return args.Error(0)
}

func (m *MockRepository) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
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

func (m *MockRepository) DeleteHabit(ctx context.Context, habitID, userID uuid.UUID) error {
	args := m.Called(ctx, habitID, userID)
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

func (m *MockRepository) GetLogsForHabits(ctx context.Context, habitIDs []uuid.UUID) ([]domain.HabitLog, error) {
	args := m.Called(ctx, habitIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.HabitLog), args.Error(1)
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

func (m *MockRepository) GetLeaderboard(ctx context.Context, limit int) ([]domain.LeaderboardEntry, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.LeaderboardEntry), args.Error(1)
}

func (m *MockRepository) GetExplorePage(ctx context.Context, limit int) ([]domain.ExploreEntry, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.ExploreEntry), args.Error(1)
}

type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) Save(ctx context.Context, file io.Reader, filename string) (string, error) {
	args := m.Called(ctx, file, filename)
	return args.String(0), args.Error(1)
}

func (m *MockStorage) Delete(ctx context.Context, url string) error {
	args := m.Called(ctx, url)
	return args.Error(0)
}

type mockMultipartFile struct {
	*strings.Reader
}

func (m *mockMultipartFile) Close() error {
	return nil
}

func TestCreateUser_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	mockStorage := new(MockStorage)
	s := New(mockRepo, mockStorage)
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
	mockStorage := new(MockStorage)
	s := New(mockRepo, mockStorage)
	ctx := context.Background()

	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	testUser := &domain.User{
		ID:             uuid.New(),
		Username:       "testuser",
		HashedPassword: string(hashedPassword),
	}

	mockRepo.On("GetUserByEmailOrUsername", ctx, "testuser").Return(testUser, nil)

	loggedInUser, token, err := s.LoginUser(ctx, LoginUserParams{Identifier: "testuser", Password: password}, "secret", time.Hour)

	assert.NoError(t, err)
	assert.NotNil(t, loggedInUser)
	assert.NotEmpty(t, token)
	assert.Equal(t, testUser.ID, loggedInUser.ID)
	assert.Empty(t, loggedInUser.HashedPassword)
	mockRepo.AssertExpectations(t)
}

func TestLoginUser_InvalidPassword(t *testing.T) {
	mockRepo := new(MockRepository)
	mockStorage := new(MockStorage)
	s := New(mockRepo, mockStorage)
	ctx := context.Background()

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	testUser := &domain.User{ID: uuid.New(), Username: "testuser", HashedPassword: string(hashedPassword)}

	mockRepo.On("GetUserByEmailOrUsername", ctx, "testuser").Return(testUser, nil)

	_, _, err := s.LoginUser(ctx, LoginUserParams{Identifier: "testuser", Password: "wrongpassword"}, "secret", time.Hour)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidCredentials))
	mockRepo.AssertExpectations(t)
}

func TestCreateHabit_Integer(t *testing.T) {
	mockRepo := new(MockRepository)
	mockStorage := new(MockStorage)
	s := New(mockRepo, mockStorage)
	ctx := context.Background()

	userID := uuid.New()
	params := CreateHabitParams{
		Name:      "Drink Water",
		ColorHue:  210,
		IsBoolean: false,
	}

	mockRepo.On("CreateHabit", ctx, mock.MatchedBy(func(h *domain.Habit) bool {
		return h.UserID == userID &&
			h.Name == params.Name &&
			h.ColorHue == params.ColorHue &&
			!h.IsBoolean
	})).Return(nil)

	habit, err := s.CreateHabit(ctx, params, userID)

	assert.NoError(t, err)
	assert.NotNil(t, habit)
	assert.Equal(t, params.Name, habit.Name)
	assert.False(t, habit.IsBoolean)
	mockRepo.AssertExpectations(t)
}

func TestLogHabit_Integer(t *testing.T) {
	mockRepo := new(MockRepository)
	mockStorage := new(MockStorage)
	s := New(mockRepo, mockStorage)
	ctx := context.Background()

	userID := uuid.New()
	habitID := uuid.New()
	logDate := time.Now().Truncate(24 * time.Hour)

	testHabit := &domain.Habit{ID: habitID, UserID: userID, IsBoolean: false}
	params := LogHabitParams{
		HabitID: habitID,
		Date:    logDate,
		Value:   5,
	}

	mockRepo.On("GetHabitByID", ctx, habitID).Return(testHabit, nil)
	mockRepo.On("UpsertHabitLog", ctx, mock.MatchedBy(func(l *domain.HabitLog) bool {
		return l.HabitID == habitID && l.Value == 5
	})).Return(nil)

	log, err := s.LogHabit(ctx, params, userID)

	assert.NoError(t, err)
	assert.NotNil(t, log)
	assert.Equal(t, 5, log.Value)
	mockRepo.AssertExpectations(t)
}

func TestLogHabit_Boolean(t *testing.T) {
	mockRepo := new(MockRepository)
	mockStorage := new(MockStorage)
	s := New(mockRepo, mockStorage)
	ctx := context.Background()

	userID := uuid.New()
	habitID := uuid.New()
	logDate := time.Now().Truncate(24 * time.Hour)

	testHabit := &domain.Habit{ID: habitID, UserID: userID, IsBoolean: true}
	params := LogHabitParams{
		HabitID: habitID,
		Date:    logDate,
		Value:   1,
	}

	mockRepo.On("GetHabitByID", ctx, habitID).Return(testHabit, nil)
	mockRepo.On("UpsertHabitLog", ctx, mock.MatchedBy(func(l *domain.HabitLog) bool {
		return l.HabitID == habitID && l.Value == 1
	})).Return(nil)

	log, err := s.LogHabit(ctx, params, userID)

	assert.NoError(t, err)
	assert.NotNil(t, log)
	assert.Equal(t, 1, log.Value)
	mockRepo.AssertExpectations(t)
}

func TestLogHabit_AccessDenied(t *testing.T) {
	mockRepo := new(MockRepository)
	mockStorage := new(MockStorage)
	s := New(mockRepo, mockStorage)
	ctx := context.Background()

	ownerID := uuid.New()
	attackerID := uuid.New()
	habitID := uuid.New()
	logDate := time.Now().Truncate(24 * time.Hour)

	testHabit := &domain.Habit{ID: habitID, UserID: ownerID, IsBoolean: true}
	params := LogHabitParams{
		HabitID: habitID,
		Date:    logDate,
		Value:   1,
	}

	mockRepo.On("GetHabitByID", ctx, habitID).Return(testHabit, nil)

	_, err := s.LogHabit(ctx, params, attackerID)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrUserAccessDenied))
	mockRepo.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "UpsertHabitLog", ctx, mock.Anything)
}

func TestUpdateHabit_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	mockStorage := new(MockStorage)
	s := New(mockRepo, mockStorage)
	ctx := context.Background()

	userID := uuid.New()
	habitID := uuid.New()
	testHabit := &domain.Habit{ID: habitID, UserID: userID, Name: "Old Name", ColorHue: 100, IsBoolean: true}
	params := UpdateHabitParams{Name: "New Name", ColorHue: 200}

	mockRepo.On("GetHabitByID", ctx, habitID).Return(testHabit, nil)
	mockRepo.On("UpdateHabit", ctx, mock.MatchedBy(func(h *domain.Habit) bool {
		return h.ID == habitID && h.Name == "New Name" && h.ColorHue == 200
	})).Return(nil)

	updatedHabit, err := s.UpdateHabit(ctx, params, habitID, userID)

	assert.NoError(t, err)
	assert.NotNil(t, updatedHabit)
	assert.Equal(t, "New Name", updatedHabit.Name)
	assert.Equal(t, 200, updatedHabit.ColorHue)
	mockRepo.AssertExpectations(t)
}

func TestUpdateHabit_AccessDenied(t *testing.T) {
	mockRepo := new(MockRepository)
	mockStorage := new(MockStorage)
	s := New(mockRepo, mockStorage)
	ctx := context.Background()

	ownerID := uuid.New()
	attackerID := uuid.New()
	habitID := uuid.New()
	testHabit := &domain.Habit{ID: habitID, UserID: ownerID, Name: "Secret Habit", IsBoolean: true}
	params := UpdateHabitParams{Name: "Pwned Habit"}

	mockRepo.On("GetHabitByID", ctx, habitID).Return(testHabit, nil)

	_, err := s.UpdateHabit(ctx, params, habitID, attackerID)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrUserAccessDenied))
	mockRepo.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "UpdateHabit", ctx, mock.Anything)
}

func TestGetProfileData_Visitor(t *testing.T) {
	mockRepo := new(MockRepository)
	mockStorage := new(MockStorage)
	s := New(mockRepo, mockStorage)
	ctx := context.Background()

	profileUserID := uuid.New()
	visitorID := uuid.New()
	testUser := &domain.User{ID: profileUserID, Username: "testuser"}

	mockRepo.On("GetUserByUsername", ctx, "testuser").Return(testUser, nil)
	mockRepo.On("GetHabitsByUserID", ctx, profileUserID).Return([]domain.Habit{}, nil)
	// The call to GetLogsForHabits is removed as it's not expected when there are no habits.
	mockRepo.On("GetFollowerCount", ctx, profileUserID).Return(10, nil)
	mockRepo.On("GetFollowingCount", ctx, profileUserID).Return(5, nil)
	mockRepo.On("IsFollowing", ctx, visitorID, profileUserID).Return(true, nil)

	profileData, err := s.GetProfileData(ctx, "testuser", visitorID)

	assert.NoError(t, err)
	assert.NotNil(t, profileData)
	assert.False(t, profileData.IsOwner)
	assert.True(t, profileData.IsFollowing)
	assert.Equal(t, 10, profileData.FollowersCount)
	mockRepo.AssertExpectations(t)
}

func TestFollowUserByUsername_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	mockStorage := new(MockStorage)
	s := New(mockRepo, mockStorage)
	ctx := context.Background()

	followerID := uuid.New()
	userToFollow := &domain.User{ID: uuid.New(), Username: "followedUser"}

	mockRepo.On("GetUserByUsername", ctx, "followedUser").Return(userToFollow, nil)
	mockRepo.On("FollowUser", ctx, followerID, userToFollow.ID).Return(nil)

	err := s.FollowUserByUsername(ctx, followerID, "followedUser")

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestFollowUserByUsername_CannotFollowSelf(t *testing.T) {
	mockRepo := new(MockRepository)
	mockStorage := new(MockStorage)
	s := New(mockRepo, mockStorage)
	ctx := context.Background()

	followerID := uuid.New()
	userToFollow := &domain.User{ID: followerID, Username: "selfFollower"}

	mockRepo.On("GetUserByUsername", ctx, "selfFollower").Return(userToFollow, nil)

	err := s.FollowUserByUsername(ctx, followerID, "selfFollower")

	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrCannotFollowSelf))
	mockRepo.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "FollowUser", ctx, mock.Anything, mock.Anything)
}

func TestUpdateUserAvatar_Success_DeletesOld(t *testing.T) {
	mockRepo := new(MockRepository)
	mockStorage := new(MockStorage)
	s := New(mockRepo, mockStorage)
	ctx := context.Background()

	userID := uuid.New()
	oldAvatarURL := "/uploads/avatars/old.png"
	newAvatarURL := "/uploads/avatars/new.png"
	header := &multipart.FileHeader{Filename: "new.png", Size: 100}

	mockFile := &mockMultipartFile{strings.NewReader("\x89PNG\r\n\x1a\n")}

	mockRepo.On("GetUserAvatar", ctx, userID).Return(&oldAvatarURL, nil)
	mockStorage.On("Save", ctx, mock.Anything, mock.AnythingOfType("string")).Return(newAvatarURL, nil)
	mockRepo.On("UpdateUserAvatar", ctx, userID, &newAvatarURL).Return(nil)
	mockStorage.On("Delete", ctx, oldAvatarURL).Return(nil)

	url, err := s.UpdateUserAvatar(ctx, userID, mockFile, header)

	assert.NoError(t, err)
	assert.Equal(t, newAvatarURL, url)
	mockRepo.AssertExpectations(t)
	mockStorage.AssertExpectations(t)
}

func TestUpdateUserAvatar_CleanupOnDBError(t *testing.T) {
	mockRepo := new(MockRepository)
	mockStorage := new(MockStorage)
	s := New(mockRepo, mockStorage)
	ctx := context.Background()

	userID := uuid.New()
	newAvatarURL := "/uploads/avatars/orphan.png"
	header := &multipart.FileHeader{Filename: "orphan.png", Size: 100}
	dbError := errors.New("database is down")

	mockFile := &mockMultipartFile{strings.NewReader("\x89PNG\r\n\x1a\n")}

	mockRepo.On("GetUserAvatar", ctx, userID).Return(nil, nil)
	mockStorage.On("Save", ctx, mock.Anything, mock.AnythingOfType("string")).Return(newAvatarURL, nil)
	mockRepo.On("UpdateUserAvatar", ctx, userID, &newAvatarURL).Return(dbError)
	mockStorage.On("Delete", ctx, newAvatarURL).Return(nil)

	_, err := s.UpdateUserAvatar(ctx, userID, mockFile, header)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), dbError.Error())
	mockRepo.AssertExpectations(t)
	mockStorage.AssertExpectations(t)
}

func TestGetLeaderboard_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	mockStorage := new(MockStorage)
	s := New(mockRepo, mockStorage)
	ctx := context.Background()

	expectedLeaderboard := []domain.LeaderboardEntry{
		{User: domain.PublicUser{Username: "user1"}, TotalLoggedDays: 100},
	}
	mockRepo.On("GetLeaderboard", ctx, 50).Return(expectedLeaderboard, nil)

	leaderboard, err := s.GetLeaderboard(ctx)

	assert.NoError(t, err)
	assert.Equal(t, expectedLeaderboard, leaderboard)
	mockRepo.AssertExpectations(t)
}

func TestDeleteHabit_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	mockStorage := new(MockStorage)
	s := New(mockRepo, mockStorage)
	ctx := context.Background()

	userID := uuid.New()
	habitID := uuid.New()

	mockRepo.On("DeleteHabit", ctx, habitID, userID).Return(nil)

	err := s.DeleteHabit(ctx, habitID, userID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestDeleteHabit_AccessDenied(t *testing.T) {
	mockRepo := new(MockRepository)
	mockStorage := new(MockStorage)
	s := New(mockRepo, mockStorage)
	ctx := context.Background()

	attackerID := uuid.New()
	habitID := uuid.New()

	mockRepo.On("DeleteHabit", ctx, habitID, attackerID).Return(repository.ErrHabitNotFound)

	err := s.DeleteHabit(ctx, habitID, attackerID)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrUserAccessDenied), "Expected user access denied error")
	mockRepo.AssertExpectations(t)
}

func TestDeleteHabit_GenericDBError(t *testing.T) {
	mockRepo := new(MockRepository)
	mockStorage := new(MockStorage)
	s := New(mockRepo, mockStorage)
	ctx := context.Background()

	userID := uuid.New()
	habitID := uuid.New()
	dbError := errors.New("something went wrong with the db")

	mockRepo.On("DeleteHabit", ctx, habitID, userID).Return(dbError)

	err := s.DeleteHabit(ctx, habitID, userID)

	assert.Error(t, err)
	assert.Equal(t, dbError, err)
	assert.False(t, errors.Is(err, ErrUserAccessDenied))
	mockRepo.AssertExpectations(t)
}

func TestDeleteUser_Success_WithAvatar(t *testing.T) {
	mockRepo := new(MockRepository)
	mockStorage := new(MockStorage)
	s := New(mockRepo, mockStorage)
	ctx := context.Background()

	userID := uuid.New()
	avatarURL := "/uploads/avatars/user-avatar.png"

	mockRepo.On("GetUserAvatar", ctx, userID).Return(&avatarURL, nil)
	mockRepo.On("DeleteUser", ctx, userID).Return(nil)
	mockStorage.On("Delete", ctx, avatarURL).Return(nil)

	err := s.DeleteUser(ctx, userID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockStorage.AssertExpectations(t)
}

func TestDeleteUser_Success_NoAvatar(t *testing.T) {
	mockRepo := new(MockRepository)
	mockStorage := new(MockStorage)
	s := New(mockRepo, mockStorage)
	ctx := context.Background()

	userID := uuid.New()

	mockRepo.On("GetUserAvatar", ctx, userID).Return(nil, nil)
	mockRepo.On("DeleteUser", ctx, userID).Return(nil)

	err := s.DeleteUser(ctx, userID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockStorage.AssertNotCalled(t, "Delete", ctx, mock.Anything)
}

func TestDeleteUser_RepoError(t *testing.T) {
	mockRepo := new(MockRepository)
	mockStorage := new(MockStorage)
	s := New(mockRepo, mockStorage)
	ctx := context.Background()

	userID := uuid.New()
	dbError := errors.New("database is down")
	avatarURL := "/uploads/avatars/user-avatar.png"

	mockRepo.On("GetUserAvatar", ctx, userID).Return(&avatarURL, nil)
	mockRepo.On("DeleteUser", ctx, userID).Return(dbError)

	err := s.DeleteUser(ctx, userID)

	assert.Error(t, err)
	assert.Equal(t, dbError, err)
	mockRepo.AssertExpectations(t)
	mockStorage.AssertNotCalled(t, "Delete", ctx, mock.Anything)
}

func TestDeleteUser_UserNotFound(t *testing.T) {
	mockRepo := new(MockRepository)
	mockStorage := new(MockStorage)
	s := New(mockRepo, mockStorage)
	ctx := context.Background()

	userID := uuid.New()

	mockRepo.On("GetUserAvatar", ctx, userID).Return(nil, repository.ErrUserNotFound)
	mockRepo.On("DeleteUser", ctx, userID).Return(repository.ErrUserNotFound)

	err := s.DeleteUser(ctx, userID)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, repository.ErrUserNotFound))
	mockRepo.AssertExpectations(t)
	mockStorage.AssertNotCalled(t, "Delete", ctx, mock.Anything)
}
