package auth

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

type postgresRepo struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) RepositoryInterface {
	return &postgresRepo{db}
}

func (pr *postgresRepo) CreateUser(ctx context.Context, user *User, passwordHash string) (int, error) {
	var id int
	err := pr.db.QueryRowContext(ctx, "INSERT INTO users (name, email, password_hash, role) VALUES ($1, $2, $3, $4) RETURNING id", user.Name, user.Email, passwordHash, user.Role).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create user: %w", err)
	}
	return id, nil
}

func (pr *postgresRepo) DeleteRefreshToken(ctx context.Context, refreshToken string) error {
	_, err := pr.db.ExecContext(ctx, "DELETE from refresh_tokens WHERE token = $1", refreshToken)
	if err != nil {
		return fmt.Errorf("failed to delete refresh tokens: %w", err)
	}
	return nil
}

func (pr *postgresRepo) GetUserIDByRefreshToken(ctx context.Context, refreshToken string) (int, error) {
	var userID int
	err := pr.db.QueryRowContext(ctx, "SELECT user_id FROM refresh_tokens WHERE token = $1", refreshToken).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("refresh token not found")
		}
		return 0, err
	}
	return userID, nil
}

func (pr *postgresRepo) GetUserRoleByID(ctx context.Context, userID int) (string, error) {
	var role string
	err := pr.db.QueryRowContext(ctx, "SELECT role FROM users WHERE id = $1", userID).Scan(&role)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("user with this id not found")
		}
		return "", err
	}
	return role, nil
}

func (pr *postgresRepo) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	user := &User{}

	err := pr.db.QueryRowContext(ctx, "SELECT id, name, password_hash, role FROM users WHERE email = $1", email).Scan(&user.ID, &user.Name, &user.Password, &user.Role)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	return user, nil
}

func (pr *postgresRepo) SaveRefreshToken(ctx context.Context, userID int, refreshToken string, refreshTokenExpire time.Time) error {
	_, err := pr.db.ExecContext(ctx, "INSERT INTO refresh_tokens (user_id, token, expires_at) VALUES ($1, $2, $3)", userID, refreshToken, refreshTokenExpire)
	if err != nil {
		return fmt.Errorf("failed to create new refresh token: %w", err)
	}
	return nil
}
