package auth

type User struct {
	ID       int    `json:"id" db:"id"`
	Name     string `json:"name" db:"name"`
	Email    string `json:"email" db:"email"`
	Password string `json:"password" db:"password_hash"`
	Role     string `json:"role" db:"role"`
}

type IDResponse struct {
	ID int `json:"id"`
}

type RegisterReqBody struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type LoginReqBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LogoutReqBody struct {
	RefreshToken string `json:"refresh_token"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type StatusResponse struct {
	Status string `json:"status"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

func NewErrorResponse(err error) *ErrorResponse {
	return &ErrorResponse{
		Message: err.Error(),
	}
}

func NewStatusResponse(status string) *StatusResponse {
	return &StatusResponse{
		Status: status,
	}
}
