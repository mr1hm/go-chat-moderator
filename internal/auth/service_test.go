package auth

import (
	"errors"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

// Mock repository for testing
type mockUserRepo struct {
	users       map[string]*User
	createErr   error
	findByEmail func(email string) (*User, error)
}

func newMockRepo() *mockUserRepo {
	return &mockUserRepo{
		users: make(map[string]*User),
	}
}

func (m *mockUserRepo) Create(user *User) error {
	if m.createErr != nil {
		return m.createErr
	}
	// Check for duplicates
	for _, u := range m.users {
		if u.Email == user.Email {
			return ErrEmailExists
		}
		if u.Username == user.Username {
			return ErrUsernameExists
		}
	}
	user.ID = "generated-id"
	m.users[user.ID] = user
	return nil
}

func (m *mockUserRepo) FindByEmail(email string) (*User, error) {
	if m.findByEmail != nil {
		return m.findByEmail(email)
	}
	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, ErrUserNotFound
}

func (m *mockUserRepo) FindByID(id string) (*User, error) {
	if u, ok := m.users[id]; ok {
		return u, nil
	}
	return nil, ErrUserNotFound
}

func TestAuthService_Register_Success(t *testing.T) {
	repo := newMockRepo()
	svc := NewAuthService(repo)

	req := &RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Username: "testuser",
	}

	user, err := svc.Register(req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if user.Email != req.Email {
		t.Errorf("expected email %s, got %s", req.Email, user.Email)
	}

	if user.Username != req.Username {
		t.Errorf("expected username %s, got %s", req.Username, user.Username)
	}

	// Password should be hashed
	if user.PasswordHash == req.Password {
		t.Error("password should be hashed, not stored in plain text")
	}

	// Verify hash is valid
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		t.Error("password hash should be valid bcrypt hash")
	}
}

func TestAuthService_Register_EmailExists(t *testing.T) {
	repo := newMockRepo()
	svc := NewAuthService(repo)

	// Create first user
	req := &RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Username: "testuser1",
	}
	svc.Register(req)

	// Try to create another user with same email
	req2 := &RegisterRequest{
		Email:    "test@example.com",
		Password: "password456",
		Username: "testuser2",
	}
	_, err := svc.Register(req2)

	if !errors.Is(err, ErrEmailExists) {
		t.Errorf("expected ErrEmailExists, got %v", err)
	}
}

func TestAuthService_Register_UsernameExists(t *testing.T) {
	repo := newMockRepo()
	svc := NewAuthService(repo)

	// Create first user
	req := &RegisterRequest{
		Email:    "test1@example.com",
		Password: "password123",
		Username: "testuser",
	}
	svc.Register(req)

	// Try to create another user with same username
	req2 := &RegisterRequest{
		Email:    "test2@example.com",
		Password: "password456",
		Username: "testuser",
	}
	_, err := svc.Register(req2)

	if !errors.Is(err, ErrUsernameExists) {
		t.Errorf("expected ErrUsernameExists, got %v", err)
	}
}

func TestAuthService_Login_Success(t *testing.T) {
	repo := newMockRepo()
	svc := NewAuthService(repo)

	// Register user first
	password := "password123"
	req := &RegisterRequest{
		Email:    "test@example.com",
		Password: password,
		Username: "testuser",
	}
	svc.Register(req)

	// Login
	loginReq := &LoginRequest{
		Email:    "test@example.com",
		Password: password,
	}

	user, err := svc.Login(loginReq)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if user.Email != req.Email {
		t.Errorf("expected email %s, got %s", req.Email, user.Email)
	}
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	repo := newMockRepo()
	svc := NewAuthService(repo)

	loginReq := &LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "password123",
	}

	_, err := svc.Login(loginReq)
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	repo := newMockRepo()
	svc := NewAuthService(repo)

	// Register user first
	req := &RegisterRequest{
		Email:    "test@example.com",
		Password: "correctpassword",
		Username: "testuser",
	}
	svc.Register(req)

	// Login with wrong password
	loginReq := &LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}

	_, err := svc.Login(loginReq)
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthService_GetUser_Success(t *testing.T) {
	repo := newMockRepo()
	svc := NewAuthService(repo)

	// Register user first
	req := &RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Username: "testuser",
	}
	registeredUser, _ := svc.Register(req)

	// Get user
	user, err := svc.GetUser(registeredUser.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if user.ID != registeredUser.ID {
		t.Errorf("expected user ID %s, got %s", registeredUser.ID, user.ID)
	}
}

func TestAuthService_GetUser_NotFound(t *testing.T) {
	repo := newMockRepo()
	svc := NewAuthService(repo)

	_, err := svc.GetUser("nonexistent-id")
	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}
