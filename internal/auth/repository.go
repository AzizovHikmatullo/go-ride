package auth

import (
	"context"
	"database/sql"
	"encoding/hex"
	"fmt"
	"math/rand"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type postgresRepo struct {
	db              *sqlx.DB
	jwtSecret       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

func NewRepository(db *sqlx.DB, jwtSecret string, accessTokenTTL, refreshTokenTTL time.Duration) RepositoryInterface {
	return &postgresRepo{db, jwtSecret, accessTokenTTL, refreshTokenTTL}
}

func (pr *postgresRepo) CreateUser(ctx context.Context, user *User) (int, error) {
	passwordHash, err := getPasswordHash(user.Password)
	if err != nil {
		return 0, fmt.Errorf("failed to generate hash for password: %w", err)
	}

	var id int
	err = pr.db.QueryRowContext(ctx, "INSERT INTO users (name, email, password_hash, role) VALUES ($1, $2, $3, $4) RETURNING id", user.Name, user.Email, passwordHash, user.Role).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create user: %w", err)
	}
	return id, nil
}

func (pr *postgresRepo) LogoutUser(ctx context.Context, refreshToken string) error {
	_, err := pr.db.ExecContext(ctx, "DELETE from refresh_tokens WHERE token = $1", refreshToken)
	if err != nil {
		return fmt.Errorf("failed to delete refresh tokens: %w", err)
	}
	return nil
}

func (pr *postgresRepo) LoginUser(ctx context.Context, user *User) (*TokenResponse, error) {
	existUser := &User{}
	err := pr.db.QueryRowContext(ctx, "SELECT id, email, password_hash, role FROM users WHERE email = $1", user.Email).Scan(&existUser.ID, &existUser.Email, &existUser.Password, &existUser.Role)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("invalid credentials")
		}
		return nil, fmt.Errorf("failed to check users credentials: %w", err)
	}

	if !checkPassword(existUser.Password, user.Password) {
		return nil, fmt.Errorf("invalid email/password")
	}

	return pr.getAuthTokens(ctx, existUser.ID, existUser.Role)
}
func (pr *postgresRepo) GenerateTokens(ctx context.Context, userID int, oldRefreshToken string) (*TokenResponse, error) {
	var dbRefreshToken string
	err := pr.db.QueryRowContext(ctx, "SELECT token FROM refresh_tokens WHERE user_id = $1", userID).Scan(&dbRefreshToken)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("please login again")
		}
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}
	if dbRefreshToken != oldRefreshToken {
		return nil, fmt.Errorf("invalid token")
	}

	role, err := pr.GetUserRoleByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return pr.getAuthTokens(ctx, userID, role)
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

func (pr *postgresRepo) getAuthTokens(ctx context.Context, userID int, userRole string) (*TokenResponse, error) {
	accessToken, err := getAccessToken(userID, userRole, pr.jwtSecret, time.Now().Add(pr.AccessTokenTTL).Unix())
	if err != nil {
		return nil, fmt.Errorf("failed to get access token")
	}

	refreshTokenExpire := time.Now().Add(pr.RefreshTokenTTL)
	refreshToken, err := getRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("please try again later")
	}

	_, err = pr.db.ExecContext(ctx, "INSERT INTO refresh_tokens (user_id, token, expires_at) VALUES ($1, $2, $3)", userID, refreshToken, refreshTokenExpire)
	if err != nil {
		return nil, fmt.Errorf("failed to create new refresh token: %w", err)
	}
	return &TokenResponse{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func getAccessToken(userID int, role string, secret string, expireTime int64) (string, error) {
	claims := jwt.MapClaims{
		"userID": userID,
		"role":   role,
		"exp":    expireTime,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(secret))

	if err != nil {
		return "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return signedToken, nil
}

func getRefreshToken() (string, error) {
	b := make([]byte, 32)

	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s)

	if _, err := r.Read(b); err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}

func getPasswordHash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func checkPassword(hashPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashPassword), []byte(password))
	return err == nil
}
