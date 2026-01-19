package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

type RepositoryInterface interface {
	CreateUser(ctx context.Context, user *User, passwordHash string) (int, error)
	DeleteRefreshToken(ctx context.Context, refreshToken string) error
	GetUserIDByRefreshToken(ctx context.Context, refreshToken string) (int, error)
	GetUserRoleByID(ctx context.Context, userID int) (string, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	SaveRefreshToken(ctx context.Context, userID int, refreshToken string, refreshTokenExpire time.Time) error
}

type AuthService struct {
	repo            RepositoryInterface
	jwtSecret       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

func NewAuthService(repository RepositoryInterface, jwtSecret string, accessTokenTTL, refreshTokenTTL time.Duration) AuthServiceInterface {
	return &AuthService{
		repo: repository,
	}
}

func (as *AuthService) CreateUser(ctx context.Context, body *RegisterReqBody) (*IDResponse, *ErrorResponse) {
	user := &User{Name: body.Name, Email: body.Email, Password: body.Password, Role: body.Role}

	passwordHash, err := getPasswordHash(user.Password)
	if err != nil {
		return nil, NewErrorResponse(err)
	}

	id, err := as.repo.CreateUser(ctx, user, passwordHash)
	if err != nil {
		return nil, NewErrorResponse(err)
	}
	return &IDResponse{id}, nil
}

func (as *AuthService) LogoutUser(ctx context.Context, refreshToken string) (*StatusResponse, *ErrorResponse) {
	err := as.repo.DeleteRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, NewErrorResponse(err)
	}
	return NewStatusResponse("loged out"), nil
}

func (as *AuthService) LoginUser(ctx context.Context, body *LoginReqBody) (*TokenResponse, *ErrorResponse) {
	user, err := as.repo.GetUserByEmail(ctx, body.Email)
	if err != nil {
		return nil, NewErrorResponse(err)
	}

	if !checkPassword(user.Password, body.Password) {
		return nil, NewErrorResponse(fmt.Errorf("invalid credentials"))
	}

	accessToken, _ := getAccessToken(user.ID, user.Role, as.jwtSecret, time.Now().Add(as.AccessTokenTTL).Unix())
	refreshToken, _ := getRefreshToken()

	err = as.repo.SaveRefreshToken(ctx, user.ID, refreshToken, time.Now().Add(as.RefreshTokenTTL))
	if err != nil {
		return nil, NewErrorResponse(err)
	}

	return &TokenResponse{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func (as *AuthService) GenerateTokens(ctx context.Context, userID int, oldRefreshToken string) (*TokenResponse, *ErrorResponse) {
	userID, err := as.repo.GetUserIDByRefreshToken(ctx, oldRefreshToken)
	if err != nil {
		return nil, NewErrorResponse(fmt.Errorf("invalid refresh token"))
	}

	role, err := as.repo.GetUserRoleByID(ctx, userID)
	if err != nil {
		return nil, NewErrorResponse(err)
	}

	accessToken, err := getAccessToken(userID, role, as.jwtSecret, time.Now().Add(as.AccessTokenTTL).Unix())
	if err != nil {
		return nil, NewErrorResponse(err)
	}

	refreshToken, err := getRefreshToken()
	if err != nil {
		return nil, NewErrorResponse(err)
	}

	err = as.repo.SaveRefreshToken(ctx, userID, refreshToken, time.Now().Add(as.RefreshTokenTTL))
	if err != nil {
		return nil, NewErrorResponse(err)
	}

	return &TokenResponse{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func (as *AuthService) GetUserIDByRefreshToken(ctx context.Context, refreshToken string) (int, *ErrorResponse) {
	userID, err := as.repo.GetUserIDByRefreshToken(ctx, refreshToken)
	if err != nil {
		return 0, NewErrorResponse(err)
	}
	return userID, nil
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
	if _, err := rand.Read(b); err != nil {
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
