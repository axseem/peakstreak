package service

import (
	"context"
	"errors"
	"io"
	"mime/multipart"
	"testing"
	"time"

	"github.com/axseem/peakstreak/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// --- Mock Implementations ---

// MockRepository provides a mock implementation of the IRepository interface for testing.
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

// MockStorage provides a mock implementation of the FileStorage interface.
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

// --- Test Cases ---

func TestCreateUser_Success(t *testing.T) {
	// Why this test exists:
	// To verify the primary success path of user creation. It ensures that valid input results
	// in a call to the repository and, critically, that the returned user object has its
	// sensitive HashedPassword field cleared.
	// Principle: Confidence. This test gives confidence in the most critical function of the
	// auth system: account creation, including the security measure of not leaking password hashes.
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
	// Why this test exists:
	// To confirm the entire successful login flow: fetching a user, correctly comparing the
	// password, generating a JWT, and returning a sanitized user object.
	// Principle: Predictability. This ensures the function's output (user, token, nil error)
	// is predictable and correct for valid inputs.
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
	// Why this test exists:
	// To test a critical failure path: the user exists, but the password is wrong. It verifies
	// that the service correctly identifies the mismatch and returns a specific, user-friendly
	// error (ErrInvalidCredentials) rather than a generic or database-level error.
	// Principle: Confidence. This guarantees we don't leak information about whether a user
	// exists when a bad password is provided and that we don't accidentally log in users with
	// wrong passwords.
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

func TestUpdateHabit_Success(t *testing.T) {
	// Why this test exists:
	// To verify the authorization logic within the service. It ensures that a user can only
	// update a habit they own.
	// Principle: Predictability. Confirms the service correctly enforces ownership rules before
	// persisting changes.
	mockRepo := new(MockRepository)
	mockStorage := new(MockStorage)
	s := New(mockRepo, mockStorage)
	ctx := context.Background()

	userID := uuid.New()
	habitID := uuid.New()
	testHabit := &domain.Habit{ID: habitID, UserID: userID, Name: "Old Name", ColorHue: 100}
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
	// Why this test exists:
	// This is the negative case for the above test. It verifies that if a user tries to
	// update a habit they *don't* own, the service returns ErrUserAccessDenied and, crucially,
	// **does not** call the repository's UpdateHabit method.
	// Principle: Confidence. Provides strong confidence in the application's security by
	// preventing unauthorized modifications.
	mockRepo := new(MockRepository)
	mockStorage := new(MockStorage)
	s := New(mockRepo, mockStorage)
	ctx := context.Background()

	ownerID := uuid.New()
	attackerID := uuid.New()
	habitID := uuid.New()
	testHabit := &domain.Habit{ID: habitID, UserID: ownerID, Name: "Secret Habit"}
	params := UpdateHabitParams{Name: "Pwned Habit"}

	mockRepo.On("GetHabitByID", ctx, habitID).Return(testHabit, nil)

	_, err := s.UpdateHabit(ctx, params, habitID, attackerID)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrUserAccessDenied))
	mockRepo.AssertExpectations(t)
	// We assert that UpdateHabit was *not* called, which is the key part of this test.
	mockRepo.AssertNotCalled(t, "UpdateHabit", ctx, mock.Anything)
}

func TestGetProfileData_Visitor(t *testing.T) {
	// Why this test exists:
	// To verify the data aggregation for the profile page from a visitor's perspective. It
	// confirms that all necessary data is fetched sequentially and that the boolean flags
	// `IsOwner` and `IsFollowing` are set correctly, as they drive UI behavior.
	// Principle: Confidence. This high-value test ensures a key feature works correctly
	// by verifying the assembled business outcome.
	mockRepo := new(MockRepository)
	mockStorage := new(MockStorage)
	s := New(mockRepo, mockStorage)
	ctx := context.Background()

	profileUserID := uuid.New()
	visitorID := uuid.New()
	testUser := &domain.User{ID: profileUserID, Username: "testuser"}

	mockRepo.On("GetUserByUsername", ctx, "testuser").Return(testUser, nil)
	// Mocking GetAllHabitsWithLogs which is an internal composition in the service
	mockRepo.On("GetHabitsByUserID", ctx, profileUserID).Return([]domain.Habit{}, nil)
	mockRepo.On("GetLogsForHabits", ctx, mock.Anything).Return([]domain.HabitLog{}, nil)
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
	// Why this test exists:
	// To test the user-facing follow mechanism. It verifies the service correctly orchestrates
	// the two-step process: looking up the target user by name, then creating the relationship.
	// Principle: Predictability. Ensures the sequence of operations is correct.
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
	// Why this test exists:
	// To test a specific business rule: a user cannot follow themselves. It verifies the
	// service checks this condition *before* hitting the database to create the relationship.
	// Principle: Clarity. Enforces a clear business rule in the service layer where it belongs.
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
	// Why this test exists:
	// To test the full orchestration of an avatar update, including resource cleanup. It verifies:
	// 1. Old avatar URL is fetched.
	// 2. New file is saved via the storage service.
	// 3. Database is updated with the new URL.
	// 4. Old file is deleted via the storage service.
	// Principle: Confidence. Guarantees the service correctly manages its resources and
	// cleans up after itself, preventing orphaned files.
	mockRepo := new(MockRepository)
	mockStorage := new(MockStorage)
	s := New(mockRepo, mockStorage)
	ctx := context.Background()

	userID := uuid.New()
	oldAvatarURL := "/uploads/avatars/old.png"
	newAvatarURL := "/uploads/avatars/new.png"
	header := &multipart.FileHeader{Filename: "new.png", Size: 100}

	mockRepo.On("GetUserAvatar", ctx, userID).Return(&oldAvatarURL, nil)
	mockStorage.On("Save", ctx, mock.Anything, mock.AnythingOfType("string")).Return(newAvatarURL, nil)
	mockRepo.On("UpdateUserAvatar", ctx, userID, &newAvatarURL).Return(nil)
	mockStorage.On("Delete", ctx, oldAvatarURL).Return(nil)

	url, err := s.UpdateUserAvatar(ctx, userID, nil, header) // file reader is nil as it's not used by the mock

	assert.NoError(t, err)
	assert.Equal(t, newAvatarURL, url)
	mockRepo.AssertExpectations(t)
	mockStorage.AssertExpectations(t)
}

func TestUpdateUserAvatar_CleanupOnDBError(t *testing.T) {
	// Why this test exists:
	// To test a critical rollback scenario. If the new avatar is saved to storage but the
	// database update fails, the service MUST clean up the newly created (but now orphaned) file.
	// Principle: Confidence. Ensures the system remains in a consistent state even when
	// things go wrong.
	mockRepo := new(MockRepository)
	mockStorage := new(MockStorage)
	s := New(mockRepo, mockStorage)
	ctx := context.Background()

	userID := uuid.New()
	newAvatarURL := "/uploads/avatars/orphan.png"
	header := &multipart.FileHeader{Filename: "orphan.png", Size: 100}
	dbError := errors.New("database is down")

	mockRepo.On("GetUserAvatar", ctx, userID).Return(nil, nil)
	mockStorage.On("Save", ctx, mock.Anything, mock.AnythingOfType("string")).Return(newAvatarURL, nil)
	mockRepo.On("UpdateUserAvatar", ctx, userID, &newAvatarURL).Return(dbError)
	// This is the crucial assertion: the service must attempt to delete the file it just saved.
	mockStorage.On("Delete", ctx, newAvatarURL).Return(nil)

	_, err := s.UpdateUserAvatar(ctx, userID, nil, header)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), dbError.Error())
	mockRepo.AssertExpectations(t)
	mockStorage.AssertExpectations(t)
}

func TestGetLeaderboard_Success(t *testing.T) {
	// Why this test exists:
	// After refactoring, the service's role is simply to delegate to the repository. This
	// test verifies that delegation. It mocks the repository's `GetLeaderboard` method and
	// ensures the service returns its result directly, without modification.
	// Principle: Clarity. The simplicity of this test reflects the clarity of the refactored
	// code. It confirms the service is not doing any complex data manipulation itself.
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
