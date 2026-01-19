package auth

import (
	"context"
)

type RepositoryInterface interface {
	CreateUser(ctx context.Context, user *User) (int, error)
	LogoutUser(ctx context.Context, refreshToken string) error
	LoginUser(ctx context.Context, user *User) (*TokenResponse, error)
	GenerateTokens(ctx context.Context, userID int, oldRefreshToken string) (*TokenResponse, error)
	GetUserIDByRefreshToken(ctx context.Context, refreshToken string) (int, error)
	getAuthTokens(ctx context.Context, userID int, role string) (*TokenResponse, error)
	GetUserRoleByID(ctx context.Context, userID int) (string, error)
}

type AuthService struct {
	repo RepositoryInterface
}

func NewAuthService(repository RepositoryInterface) AuthServiceInterface {
	return &AuthService{
		repo: repository,
	}
}

func (as *AuthService) CreateUser(ctx context.Context, body *RegisterReqBody) (*IDResponse, *ErrorResponse) {
	user := &User{Name: body.Name, Email: body.Email, Password: body.Password, Role: body.Role}
	id, err := as.repo.CreateUser(ctx, user)
	if err != nil {
		return nil, NewErrorResponse(err)
	}
	return &IDResponse{id}, nil
}

func (as *AuthService) LogoutUser(ctx context.Context, refreshToken string) (*StatusResponse, *ErrorResponse) {
	err := as.repo.LogoutUser(ctx, refreshToken)
	if err != nil {
		return nil, NewErrorResponse(err)
	}
	return NewStatusResponse("loged out"), nil
}

func (as *AuthService) LoginUser(ctx context.Context, body *LoginReqBody) (*TokenResponse, *ErrorResponse) {
	user := &User{
		Email:    body.Email,
		Password: body.Password,
	}
	tokens, err := as.repo.LoginUser(ctx, user)
	if err != nil {
		return nil, NewErrorResponse(err)
	}
	return tokens, nil
}

func (as *AuthService) GenerateTokens(ctx context.Context, userID int, oldRefreshToken string) (*TokenResponse, *ErrorResponse) {
	tokens, err := as.repo.GenerateTokens(ctx, userID, oldRefreshToken)
	if err != nil {
		return nil, NewErrorResponse(err)
	}
	return tokens, nil
}

func (as *AuthService) GetUserIDByRefreshToken(ctx context.Context, refreshToken string) (int, *ErrorResponse) {
	userID, err := as.repo.GetUserIDByRefreshToken(ctx, refreshToken)
	if err != nil {
		return 0, NewErrorResponse(err)
	}
	return userID, nil
}
