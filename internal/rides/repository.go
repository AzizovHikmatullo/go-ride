package rides

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/jmoiron/sqlx"
)

const (
	searchingStatus  = "SEARCHING"
	inProgressStatus = "IN_PROGRESS"
	completedStatus  = "COMPLETED"
	canceledStatus   = "CANCELED"
)

type postgresRepo struct {
	db     *sqlx.DB
	logger *slog.Logger
}

func NewRepository(db *sqlx.DB, logger *slog.Logger) RepositoryInterface {
	return &postgresRepo{db, logger}
}

func (pr *postgresRepo) CreateRide(ctx context.Context, userID int, start, end []byte, route json.RawMessage) (*CreateResponse, error) {
	var id int

	err := pr.db.QueryRowContext(ctx, "INSERT INTO rides (user_id, status, start_point, end_point, route) VALUES ($1, $2, $3, $4, $5) RETURNING id", userID, searchingStatus, start, end, route).Scan(&id)
	if err != nil {
		pr.logger.Error("failed to create ride",
			slog.Int("user_id", userID),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("failed to create ride: %w", err)
	}

	return &CreateResponse{id, searchingStatus, route}, nil
}

func (pr *postgresRepo) GetRideByID(ctx context.Context, rideID int) (*Ride, error) {
	var ride Ride

	err := pr.db.QueryRowContext(ctx, "SELECT id, user_id, driver_id, status, start_point, end_point, route, created_at, updated_at FROM rides WHERE id = $1", rideID).Scan(&ride.ID, &ride.UserID, &ride.DriverID, &ride.Status, &ride.Start, &ride.End, &ride.Route, &ride.CreatedAt, &ride.UpdatedAt)
	if err != nil {
		pr.logger.Error("failed to get ride by ID",
			slog.Int("ride_id", rideID),
			slog.String("error", err.Error()),
		)
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("ride with this id not found")
		}
		return nil, err
	}

	return &ride, nil
}

func (pr *postgresRepo) GetRideStatus(ctx context.Context, rideID int) (string, error) {
	var status string

	err := pr.db.QueryRowContext(ctx, "SELECT status FROM rides WHERE id = $1", rideID).Scan(&status)
	if err != nil {
		pr.logger.Error("failed to get ride status",
			slog.Int("ride_id", rideID),
			slog.String("error", err.Error()),
		)
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("ride with this id not found")
		}
		return "", err
	}

	return status, nil
}

func (pr *postgresRepo) TakeRide(ctx context.Context, rideID int, driverID int) (*ChangeRideResponse, error) {
	tx, err := pr.db.BeginTxx(ctx, nil)
	if err != nil {
		pr.logger.Error("failed to take ride",
			slog.Int("driver_id", driverID),
			slog.Int("ride_id", rideID),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("failed to take ride %w", err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var status string
	err = tx.Get(&status, "SELECT status FROM rides WHERE id = $1", rideID)
	if status != searchingStatus {
		return nil, fmt.Errorf("ride already taken")
	}

	_, err = tx.ExecContext(ctx, "UPDATE rides SET driver_id = $1, status = $2, updated_at = now() WHERE id = $3", driverID, inProgressStatus, rideID)
	if err != nil {
		pr.logger.Error("failed to update ride",
			slog.Int("driver_id", driverID),
			slog.Int("ride_id", rideID),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("failed to update ride: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		pr.logger.Error("failed to update ride",
			slog.Int("driver_id", driverID),
			slog.Int("ride_id", rideID),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("failed to update ride: %w", err)
	}

	return NewChangeRideResponse(rideID, inProgressStatus), nil
}

func (pr *postgresRepo) CompleteRide(ctx context.Context, rideID int) (*ChangeRideResponse, error) {
	_, err := pr.db.ExecContext(ctx, "UPDATE rides SET status = $1, updated_at = now() WHERE id = $2", completedStatus, rideID)
	if err != nil {
		pr.logger.Error("failed to complete ride",
			slog.Int("ride_id", rideID),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("failed to complete ride: %w", err)
	}

	return NewChangeRideResponse(rideID, completedStatus), nil
}

func (pr *postgresRepo) CancelRide(ctx context.Context, rideID int) (*ChangeRideResponse, error) {
	_, err := pr.db.ExecContext(ctx, "UPDATE rides SET status = $1, updated_at = now() WHERE id = $2", canceledStatus, rideID)
	if err != nil {
		pr.logger.Error("failed to cancel ride",
			slog.Int("ride_id", rideID),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("failed to cancel ride: %w", err)
	}

	return NewChangeRideResponse(rideID, canceledStatus), nil
}

func (pr *postgresRepo) GetSearchingRides(ctx context.Context) ([]Ride, error) {
	var rides []Ride

	err := pr.db.SelectContext(ctx, &rides, "SELECT * FROM rides WHERE status = $1", searchingStatus)
	if err != nil {
		pr.logger.Error("failed to get rides",
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("failed to get rides: %w", err)
	}

	return rides, nil
}
