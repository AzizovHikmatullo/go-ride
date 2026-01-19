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
