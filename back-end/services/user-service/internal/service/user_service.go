package service

import (
	"context"
	"errors"

	"github.com/nusaroute/services/user-service/internal/model"
	"github.com/nusaroute/services/user-service/internal/repository"
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
	// TODO: Implement user registration (Tubes Tahap Dua)
	return nil, errors.New("method Register not implemented")
}

func (s *userService) Login(ctx context.Context, req model.LoginRequest, jwtSecret string) (*model.LoginResponse, error) {
	// TODO: Implement user login (Tubes Tahap Dua)
	return nil, errors.New("method Login not implemented")
}

func (s *userService) GetProfile(ctx context.Context, userID string) (*model.User, error) {
	// TODO: Implement get profile (Tubes Tahap Dua)
	return nil, errors.New("method GetProfile not implemented")
}

func (s *userService) UpdateProfile(ctx context.Context, userID string, req model.UpdateProfileRequest) (*model.User, error) {
	return nil, errors.New("method UpdateProfile not implemented")
}

func (s *userService) AddAddress(ctx context.Context, userID string, req model.CreateAddressRequest) (*model.Address, error) {
	// TODO: Implement add address (Tubes Tahap Dua)
	return nil, errors.New("method AddAddress not implemented")
}

func (s *userService) ListAddresses(ctx context.Context, userID string) ([]model.Address, error) {
	return nil, errors.New("method ListAddresses not implemented")
}

func (s *userService) DeleteAddress(ctx context.Context, addressID, userID string) error {
	return errors.New("method DeleteAddress not implemented")
}
