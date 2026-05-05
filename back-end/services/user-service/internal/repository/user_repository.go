package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/nusaroute/services/user-service/internal/model"
)

// UserRepository defines the interface for user data access.
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	CreateAddress(ctx context.Context, addr *model.Address) error
	GetAddressesByUserID(ctx context.Context, userID string) ([]model.Address, error)
	DeleteAddress(ctx context.Context, id, userID string) error
}

// userRepo is the PostgreSQL implementation of UserRepository.
type userRepo struct {
	db *sqlx.DB
}

// NewUserRepository creates a new PostgreSQL-backed user repository.
func NewUserRepository(db *sqlx.DB) UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) Create(ctx context.Context, user *model.User) error {
	user.ID = uuid.New().String()
	user.IsActive = true
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	query := `
		INSERT INTO users (id, email, phone, full_name, password_hash, role, avatar_url, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.ExecContext(ctx, query,
		user.ID, user.Email, user.Phone, user.FullName, user.Password,
		user.Role, user.AvatarURL, user.IsActive, user.CreatedAt, user.UpdatedAt,
	)
	return err
}

func (r *userRepo) GetByID(ctx context.Context, id string) (*model.User, error) {
	var user model.User
	err := r.db.GetContext(ctx, &user, "SELECT * FROM users WHERE id = $1 AND is_active = true", id)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := r.db.GetContext(ctx, &user, "SELECT * FROM users WHERE email = $1 AND is_active = true", email)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepo) Update(ctx context.Context, user *model.User) error {
	user.UpdatedAt = time.Now()
	query := `
		UPDATE users SET full_name = $1, phone = $2, avatar_url = $3, updated_at = $4
		WHERE id = $5
	`
	_, err := r.db.ExecContext(ctx, query, user.FullName, user.Phone, user.AvatarURL, user.UpdatedAt, user.ID)
	return err
}

func (r *userRepo) CreateAddress(ctx context.Context, addr *model.Address) error {
	addr.ID = uuid.New().String()
	addr.CreatedAt = time.Now()
	addr.UpdatedAt = time.Now()

	// If this is default, un-default all others
	if addr.IsDefault {
		_, _ = r.db.ExecContext(ctx, "UPDATE addresses SET is_default = false WHERE user_id = $1", addr.UserID)
	}

	query := `
		INSERT INTO addresses (id, user_id, label, full_name, phone, province, city, district, 
		                       postal_code, full_address, lat, lng, is_default, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`
	_, err := r.db.ExecContext(ctx, query,
		addr.ID, addr.UserID, addr.Label, addr.FullName, addr.Phone,
		addr.Province, addr.City, addr.District, addr.PostalCode, addr.FullAddress,
		addr.Lat, addr.Lng, addr.IsDefault, addr.CreatedAt, addr.UpdatedAt,
	)
	return err
}

func (r *userRepo) GetAddressesByUserID(ctx context.Context, userID string) ([]model.Address, error) {
	var addresses []model.Address
	err := r.db.SelectContext(ctx, &addresses,
		"SELECT * FROM addresses WHERE user_id = $1 ORDER BY is_default DESC, created_at DESC", userID)
	if err != nil {
		return nil, err
	}
	return addresses, nil
}

func (r *userRepo) DeleteAddress(ctx context.Context, id, userID string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM addresses WHERE id = $1 AND user_id = $2", id, userID)
	return err
}
