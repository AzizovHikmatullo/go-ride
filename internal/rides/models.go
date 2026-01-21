package rides

import (
	"encoding/json"
	"time"
)

type Ride struct {
	ID        int             `json:"id" db:"id"`
	UserID    int             `json:"user_id" db:"user_id"`
	DriverID  *int            `json:"driver_id,omitempty" db:"driver_id"`
	Status    string          `json:"status" db:"status"`
	Start     json.RawMessage `json:"start_point" db:"start_point"`
	End       json.RawMessage `json:"end_point" db:"end_point"`
	Route     json.RawMessage `json:"route" db:"route"`
	CreatedAt time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt time.Time       `json:"updated_at,omitempty" db:"updated_at"`
}

type APIResponse struct {
	Code   string  `json:"code"`
	Routes []Route `json:"routes"`
}

type Route struct {
	Geometry Geometry `json:"geometry"`
}

type Geometry struct {
	Type        string      `json:"type"`
	Coordinates [][]float64 `json:"coordinates"`
}

type PointGeoJSON struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
}

type CreateRequest struct {
	Start PointGeoJSON `json:"start_point"`
	End   PointGeoJSON `json:"end_point"`
}

type CreateResponse struct {
	ID     int             `json:"id" db:"id"`
	Status string          `json:"status" db:"status"`
	Route  json.RawMessage `json:"route" db:"route"`
}

type ChangeRideResponse struct {
	ID     int    `json:"id" db:"id"`
	Status string `json:"status" db:"status"`
}

type SearchRidesResponse struct {
	Rides []Ride `json:"rides" db:"ride"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

// Only  for Swagger
type CreateResponseSwagger struct {
	ID     int                    `json:"id"`
	Status string                 `json:"status"`
	Route  map[string]interface{} `json:"route"`
}

type RideSwagger struct {
	ID        int                    `json:"id" db:"id"`
	UserID    int                    `json:"user_id" db:"user_id"`
	DriverID  *int                   `json:"driver_id,omitempty" db:"driver_id"`
	Status    string                 `json:"status" db:"status"`
	Start     map[string]interface{} `json:"start_point" db:"start_point"`
	End       map[string]interface{} `json:"end_point" db:"end_point"`
	Route     map[string]interface{} `json:"route" db:"route"`
	CreatedAt time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt time.Time              `json:"updated_at,omitempty" db:"updated_at"`
}

type SearchRidesResponseSwagger struct {
	Rides []RideSwagger `json:"rides" db:"ride"`
}

func NewChangeRideResponse(id int, status string) *ChangeRideResponse {
	return &ChangeRideResponse{
		ID:     id,
		Status: status,
	}
}

func NewErrorResponse(err error) *ErrorResponse {
	return &ErrorResponse{
		Message: err.Error(),
	}
}
