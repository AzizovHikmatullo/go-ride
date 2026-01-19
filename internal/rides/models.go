package rides

import "time"

type Point struct {
	Lat float64 `json:"lat" db:"lat"`
	Lon float64 `json:"lon" db:"lon"`
}

type Route struct {
	Path string `json:"path"`
}

type Ride struct {
	ID        int       `json:"id" db:"id"`
	UserID    int       `json:"user_id" db:"user_id"`
	DriverID  int       `json:"driver_id" db:"driver_id"`
	Status    string    `json:"status" db:"status"`
	Start     Point     `json:"start_point" db:"start_point"`
	End       Point     `json:"end_point" db:"end_point"`
	Route     Route     `json:"route" db:"route"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
