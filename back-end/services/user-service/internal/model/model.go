package model

import "time"

// User represents a registered user in the system.
type User struct {
	ID        string    `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	Phone     string    `json:"phone" db:"phone"`
	FullName  string    `json:"full_name" db:"full_name"`
	Password  string    `json:"-" db:"password_hash"` // never expose in JSON
	Role      string    `json:"role" db:"role"`       // USER, COURIER, ADMIN
	AvatarURL string    `json:"avatar_url,omitempty" db:"avatar_url"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Address represents a saved address in the user's address book.
type Address struct {
	ID        string    `json:"id" db:"id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Label     string    `json:"label" db:"label"` // Rumah, Kantor, etc.
	FullName  string    `json:"full_name" db:"full_name"`
	Phone     string    `json:"phone" db:"phone"`
	Province  string    `json:"province" db:"province"`
	City      string    `json:"city" db:"city"`
	District  string    `json:"district" db:"district"`
	PostalCode string   `json:"postal_code" db:"postal_code"`
	FullAddress string  `json:"full_address" db:"full_address"`
	Lat       float64   `json:"lat" db:"lat"`
	Lng       float64   `json:"lng" db:"lng"`
	IsDefault bool      `json:"is_default" db:"is_default"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// RegisterRequest is the input for user registration.
type RegisterRequest struct {
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	FullName string `json:"full_name"`
	Password string `json:"password"`
	Role     string `json:"role"` // defaults to USER
}

// LoginRequest is the input for user login.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse contains the JWT token and user info.
type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// UpdateProfileRequest is the input for updating user profile.
type UpdateProfileRequest struct {
	FullName  string `json:"full_name,omitempty"`
	Phone     string `json:"phone,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

// CreateAddressRequest is the input for adding an address.
type CreateAddressRequest struct {
	Label       string  `json:"label"`
	FullName    string  `json:"full_name"`
	Phone       string  `json:"phone"`
	Province    string  `json:"province"`
	City        string  `json:"city"`
	District    string  `json:"district"`
	PostalCode  string  `json:"postal_code"`
	FullAddress string  `json:"full_address"`
	Lat         float64 `json:"lat"`
	Lng         float64 `json:"lng"`
	IsDefault   bool    `json:"is_default"`
}
