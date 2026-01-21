package auth

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthServiceInterface interface {
	CreateUser(ctx context.Context, body *RegisterReqBody) (*IDResponse, *ErrorResponse)
	LogoutUser(ctx context.Context, refreshToken string) (*StatusResponse, *ErrorResponse)
	LoginUser(ctx context.Context, body *LoginReqBody) (*TokenResponse, *ErrorResponse)
	GenerateTokens(ctx context.Context, userID int, oldRefreshToken string) (*TokenResponse, *ErrorResponse)
	GetUserIDByRefreshToken(ctx context.Context, refreshToken string) (int, *ErrorResponse)
}

type AuthHandler struct {
	service AuthServiceInterface
}

func NewAuthHandler(service AuthServiceInterface) *AuthHandler {
	return &AuthHandler{
		service: service,
	}
}

// @Summary      Register new user
// @Description  Create a new user account
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      RegisterReqBody  true  "User registration info"
// @Success      200   {object}  IDResponse
// @Failure      400   {object}  ErrorResponse
// @Failure      500   {object}  ErrorResponse
// @Router       /auth/register [post]
func (ah *AuthHandler) Register(c *gin.Context) {
	var body RegisterReqBody

	if err := c.ShouldBindJSON(&body); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid input body")
		return
	}

	id, err := ah.service.CreateUser(c, &body)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Message)
		return
	}

	c.JSON(http.StatusOK, id)
}

// @Summary      Login user
// @Description  Authenticate user and return access + refresh tokens
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      LoginReqBody  true  "Login credentials"
// @Success      200   {object}  TokenResponse
// @Failure      400   {object}  ErrorResponse
// @Failure      500   {object}  ErrorResponse
// @Router       /auth/login [post]
func (ah *AuthHandler) Login(c *gin.Context) {
	var body LoginReqBody

	if err := c.ShouldBindJSON(&body); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid input body")
		return
	}

	tokens, err := ah.service.LoginUser(c, &body)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Message)
		return
	}

	c.JSON(http.StatusOK, tokens)
}

// @Summary      Logout user
// @Description  Delete refresh token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      LogoutReqBody  true  "Refresh token"
// @Success      200   {object}  StatusResponse
// @Failure      400   {object}  ErrorResponse
// @Router       /auth/logout [post]
func (ah *AuthHandler) Logout(c *gin.Context) {
	var body LogoutReqBody

	if err := c.ShouldBindJSON(&body); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid input body")
		return
	}

	status, err := ah.service.LogoutUser(c, body.RefreshToken)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Message)
		return
	}
	c.JSON(http.StatusOK, status)
}

// @Summary      Refresh JWT tokens
// @Description  Generate new tokens
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      LogoutReqBody  true  "Old refresh token"
// @Success  	 200   {object}  TokenResponse
// @Failure      401   {object}  ErrorResponse
// @Router       /auth/refresh [post]
func (ah *AuthHandler) RefreshToken(c *gin.Context) {
	var body LogoutReqBody

	if err := c.ShouldBindJSON(&body); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid input body")
		return
	}

	userID, err := ah.service.GetUserIDByRefreshToken(c, body.RefreshToken)
	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, "invalid refresh token")
		return
	}

	tokens, errResp := ah.service.GenerateTokens(c, userID, body.RefreshToken)
	if errResp != nil {
		newErrorResponse(c, http.StatusUnauthorized, errResp.Message)
		return
	}

	c.JSON(http.StatusOK, tokens)

}

func newErrorResponse(c *gin.Context, statusCode int, message string) {
	c.AbortWithStatusJSON(statusCode, ErrorResponse{message})
}
