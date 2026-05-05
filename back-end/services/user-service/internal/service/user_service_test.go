//go:build unit

package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/nusaroute/services/user-service/internal/model"
	"github.com/nusaroute/services/user-service/internal/service"
)

// MockUserRepository implements repository.UserRepository for unit testing.
type MockUserRepository struct {
	users     map[string]*model.User
	addresses map[string][]model.Address
	createErr error
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users:     make(map[string]*model.User),
		addresses: make(map[string][]model.Address),
	}
}

func (m *MockUserRepository) Create(ctx context.Context, user *model.User) error {
	if m.createErr != nil { return m.createErr }
	user.ID = "test-user-id"
	m.users[user.Email] = user
	return nil
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
	for _, u := range m.users {
		if u.ID == id { return u, nil }
	}
	return nil, errors.New("not found")
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	u, ok := m.users[email]
	if !ok { return nil, errors.New("not found") }
	return u, nil
}

func (m *MockUserRepository) Update(ctx context.Context, user *model.User) error {
	m.users[user.Email] = user
	return nil
}

func (m *MockUserRepository) CreateAddress(ctx context.Context, addr *model.Address) error {
	addr.ID = "test-addr-id"
	m.addresses[addr.UserID] = append(m.addresses[addr.UserID], *addr)
	return nil
}

func (m *MockUserRepository) GetAddressesByUserID(ctx context.Context, userID string) ([]model.Address, error) {
	return m.addresses[userID], nil
}

func (m *MockUserRepository) DeleteAddress(ctx context.Context, id, userID string) error {
	return nil
}

// ============================================================
// Unit Tests
// ============================================================

func TestRegister_Success(t *testing.T) {
	repo := NewMockUserRepository()
	svc := service.NewUserService(repo)

	user, err := svc.Register(context.Background(), model.RegisterRequest{
		Email: "test@nusaroute.id", Password: "SecurePass123!", FullName: "Budi Setiawan",
	})

	if err != nil { t.Fatalf("expected no error, got %v", err) }
	if user.Email != "test@nusaroute.id" { t.Errorf("expected email test@nusaroute.id, got %s", user.Email) }
	if user.FullName != "Budi Setiawan" { t.Errorf("expected name Budi Setiawan, got %s", user.FullName) }
	if user.Role != "USER" { t.Errorf("expected role USER, got %s", user.Role) }
}

func TestRegister_DuplicateEmail(t *testing.T) {
	repo := NewMockUserRepository()
	svc := service.NewUserService(repo)

	// First registration
	svc.Register(context.Background(), model.RegisterRequest{
		Email: "dup@nusaroute.id", Password: "Pass123!", FullName: "User One",
	})

	// Duplicate
	_, err := svc.Register(context.Background(), model.RegisterRequest{
		Email: "dup@nusaroute.id", Password: "Pass456!", FullName: "User Two",
	})

	if err == nil { t.Fatal("expected error for duplicate email") }
	if err.Error() != "email already registered" { t.Errorf("unexpected error: %v", err) }
}

func TestRegister_MissingFields(t *testing.T) {
	repo := NewMockUserRepository()
	svc := service.NewUserService(repo)

	_, err := svc.Register(context.Background(), model.RegisterRequest{})
	if err == nil { t.Fatal("expected error for missing fields") }
}

func TestLogin_Success(t *testing.T) {
	repo := NewMockUserRepository()
	svc := service.NewUserService(repo)

	// Register first
	svc.Register(context.Background(), model.RegisterRequest{
		Email: "login@nusaroute.id", Password: "MyPassword1!", FullName: "Login User",
	})

	result, err := svc.Login(context.Background(), model.LoginRequest{
		Email: "login@nusaroute.id", Password: "MyPassword1!",
	}, "test-jwt-secret")

	if err != nil { t.Fatalf("expected no error, got %v", err) }
	if result.Token == "" { t.Error("expected non-empty token") }
	if result.User.Email != "login@nusaroute.id" { t.Errorf("wrong email: %s", result.User.Email) }
}

func TestLogin_WrongPassword(t *testing.T) {
	repo := NewMockUserRepository()
	svc := service.NewUserService(repo)

	svc.Register(context.Background(), model.RegisterRequest{
		Email: "wrong@nusaroute.id", Password: "CorrectPass!", FullName: "Test",
	})

	_, err := svc.Login(context.Background(), model.LoginRequest{
		Email: "wrong@nusaroute.id", Password: "WrongPass!",
	}, "test-jwt-secret")

	if err == nil { t.Fatal("expected error for wrong password") }
}

func TestAddAddress_Success(t *testing.T) {
	repo := NewMockUserRepository()
	svc := service.NewUserService(repo)

	addr, err := svc.AddAddress(context.Background(), "user-1", model.CreateAddressRequest{
		FullName: "Budi", FullAddress: "Jl. Merdeka No. 1, Bandung",
		Label: "Rumah", Lat: -6.917, Lng: 107.619,
	})

	if err != nil { t.Fatalf("expected no error, got %v", err) }
	if addr.Label != "Rumah" { t.Errorf("wrong label: %s", addr.Label) }
	if addr.Lat != -6.917 { t.Errorf("wrong lat: %f", addr.Lat) }
}

func TestAddAddress_MissingRequired(t *testing.T) {
	repo := NewMockUserRepository()
	svc := service.NewUserService(repo)

	_, err := svc.AddAddress(context.Background(), "user-1", model.CreateAddressRequest{})
	if err == nil { t.Fatal("expected error for missing fields") }
}

func TestGetProfile_NotFound(t *testing.T) {
	repo := NewMockUserRepository()
	svc := service.NewUserService(repo)

	_, err := svc.GetProfile(context.Background(), "nonexistent")
	if err == nil { t.Fatal("expected error for non-existent user") }
}
