package services

import (
	"context"
	"fmt"

	"github.com/betpro/server/internal/db"
	"github.com/betpro/server/internal/models"
	"github.com/betpro/server/pkg/money"
)

type UserService struct {
	db *db.DB
}

func NewUserService(database *db.DB) *UserService {
	return &UserService{db: database}
}

func (s *UserService) CreateUser(ctx context.Context, username, email, passwordHash string) (*models.User, error) {
	user := &models.User{
		Username: username,
		Email:    email,
		Password: passwordHash,
		Role:     models.RoleUser,
		IsActive: true,
		Balance:  money.FromInt(0),
	}

	query := `
		INSERT INTO users (username, email, password, role, is_active, balance)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`

	err := s.db.QueryRow(ctx, query,
		user.Username, user.Email, user.Password, user.Role, user.IsActive, user.Balance,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}

func (s *UserService) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, username, email, password, role, is_active, balance, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	err := s.db.QueryRow(ctx, query, userID).Scan(
		&user.ID, &user.Username, &user.Email, &user.Password,
		&user.Role, &user.IsActive, &user.Balance, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	return user, nil
}

func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, username, email, password, role, is_active, balance, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	err := s.db.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Username, &user.Email, &user.Password,
		&user.Role, &user.IsActive, &user.Balance, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	return user, nil
}

func (s *UserService) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, username, email, password, role, is_active, balance, created_at, updated_at
		FROM users
		WHERE username = $1
	`

	err := s.db.QueryRow(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.Password,
		&user.Role, &user.IsActive, &user.Balance, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("get user by username: %w", err)
	}

	return user, nil
}

func (s *UserService) GetProfile(ctx context.Context, userID string) (*models.Profile, error) {
	profile := &models.Profile{}
	query := `
		SELECT id, balance, is_active, role
		FROM users
		WHERE id = $1
	`

	err := s.db.QueryRow(ctx, query, userID).Scan(
		&profile.ID, &profile.Balance, &profile.IsActive, &profile.Role,
	)

	if err != nil {
		return nil, fmt.Errorf("get profile: %w", err)
	}

	return profile, nil
}

func (s *UserService) UpdateBalance(ctx context.Context, userID string, amount money.Money) error {
	query := `
		UPDATE users
		SET balance = balance + $1, updated_at = NOW()
		WHERE id = $2
	`

	_, err := s.db.Exec(ctx, query, amount, userID)
	return err
}

func (s *UserService) SetBalance(ctx context.Context, userID string, balance money.Money) error {
	query := `
		UPDATE users
		SET balance = $1, updated_at = NOW()
		WHERE id = $2
	`

	_, err := s.db.Exec(ctx, query, balance, userID)
	return err
}

func (s *UserService) UpdateRole(ctx context.Context, userID, role string) error {
	if err := models.ValidateRole(role); err != nil {
		return err
	}

	query := `
		UPDATE users
		SET role = $1, updated_at = NOW()
		WHERE id = $2
	`

	_, err := s.db.Exec(ctx, query, role, userID)
	return err
}

func (s *UserService) SetActive(ctx context.Context, userID string, active bool) error {
	query := `
		UPDATE users
		SET is_active = $1, updated_at = NOW()
		WHERE id = $2
	`

	_, err := s.db.Exec(ctx, query, active, userID)
	return err
}
