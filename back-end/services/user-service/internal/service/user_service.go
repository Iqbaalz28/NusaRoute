package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/nusaroute/pkg/middleware"
	"github.com/nusaroute/services/user-service/internal/model"
	"github.com/nusaroute/services/user-service/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

// UserService defines the interface for user business logic.
type UserService interface {
	Register(ctx context.Context, req model.RegisterRequest) (*model.User, error)
	Login(ctx context.Context, req model.LoginRequest, jwtSecret string) (*model.LoginResponse, error)
	GetProfile(ctx context.Context, userID string) (*model.User, error)
	UpdateProfile(ctx context.Context, userID string, req model.UpdateProfileRequest) (*model.User, error)
	AddAddress(ctx context.Context, userID string, req model.CreateAddressRequest) (*model.Address, error)
	ListAddresses(ctx context.Context, userID string) ([]model.Address, error)
	DeleteAddress(ctx context.Context, addressID, userID string) error
}

type userService struct {
	repo repository.UserRepository
}

// NewUserService creates a new UserService instance.
func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) Register(ctx context.Context, req model.RegisterRequest) (*model.User, error) {
	if req.Email == "" || req.Password == "" || req.FullName == "" {
		return nil, errors.New("email, password, and full_name are required")
	}

	existing, _ := s.repo.GetByEmail(ctx, req.Email)
	if existing != nil {
		return nil, errors.New("email already registered")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	role := req.Role
	if role == "" { role = "USER" }

	user := &model.User{
		Email:    req.Email,
		Phone:    req.Phone,
		FullName: req.FullName,
		Password: string(hashedPassword),
		Role:     role,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (s *userService) Login(ctx context.Context, req model.LoginRequest, jwtSecret string) (*model.LoginResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, errors.New("email and password are required")
	}

	user, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	token, err := middleware.GenerateToken(user.ID, user.Email, user.Role, jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &model.LoginResponse{
		Token: token,
		User:  *user,
	}, nil
}

func (s *userService) GetProfile(ctx context.Context, userID string) (*model.User, error) {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (s *userService) UpdateProfile(ctx context.Context, userID string, req model.UpdateProfileRequest) (*model.User, error) {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil { return nil, errors.New("user not found") }

	if req.FullName != "" { user.FullName = req.FullName }
	if req.Phone != "" { user.Phone = req.Phone }
	if req.AvatarURL != "" { user.AvatarURL = req.AvatarURL }

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}
	return user, nil
}

func (s *userService) AddAddress(ctx context.Context, userID string, req model.CreateAddressRequest) (*model.Address, error) {
	if req.FullAddress == "" || req.FullName == "" {
		return nil, errors.New("full_address and full_name are required")
	}

	addr := &model.Address{
		UserID: userID, Label: req.Label, FullName: req.FullName,
		Phone: req.Phone, Province: req.Province, City: req.City,
		District: req.District, PostalCode: req.PostalCode,
		FullAddress: req.FullAddress, Lat: req.Lat, Lng: req.Lng, IsDefault: req.IsDefault,
	}

	if err := s.repo.CreateAddress(ctx, addr); err != nil {
		return nil, fmt.Errorf("failed to add address: %w", err)
	}
	return addr, nil
}

func (s *userService) ListAddresses(ctx context.Context, userID string) ([]model.Address, error) {
	return s.repo.GetAddressesByUserID(ctx, userID)
}

func (s *userService) DeleteAddress(ctx context.Context, addressID, userID string) error {
	return s.repo.DeleteAddress(ctx, addressID, userID)
}
