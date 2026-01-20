package rides

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

type RepositoryInterface interface {
	CreateRide(ctx context.Context, userID int, start, end []byte, route json.RawMessage) (*CreateResponse, error)
	GetRideByID(ctx context.Context, rideID int) (*Ride, error)
	GetRideStatus(ctx context.Context, rideID int) (string, error)
	TakeRide(ctx context.Context, rideID int, driverID int) (*ChangeRideResponse, error)
	CompleteRide(ctx context.Context, rideID int) (*ChangeRideResponse, error)
	CancelRide(ctx context.Context, rideID int) (*ChangeRideResponse, error)
	GetSearchingRides(ctx context.Context) (*SearchRidesResponse, error)
}

type RideService struct {
	repo   RepositoryInterface
	logger *slog.Logger
}

func NewRideService(repository RepositoryInterface, logger *slog.Logger) RideServiceInterface {
	return &RideService{
		repo:   repository,
		logger: logger,
	}
}

func (rs *RideService) CreateRide(ctx context.Context, userID int, start, end PointGeoJSON) (*CreateResponse, *ErrorResponse) {
	route, err := fetchRideFromAPI(ctx, start, end)
	if err != nil {
		rs.logger.Error("failed to fetch route from API",
			slog.String("error", err.Error()),
		)
		return nil, NewErrorResponse(err)
	}

	routeJSON, err := json.Marshal(route)
	if err != nil {
		return nil, NewErrorResponse(err)
	}

	startJSON, err := json.Marshal(start)
	if err != nil {
		return nil, NewErrorResponse(err)
	}

	endJSON, err := json.Marshal(end)
	if err != nil {
		return nil, NewErrorResponse(err)
	}

	response, err := rs.repo.CreateRide(ctx, userID, startJSON, endJSON, routeJSON)
	if err != nil {
		return nil, NewErrorResponse(err)
	}

	rs.logger.Info("ride created",
		slog.Int("user_id", userID),
		slog.Int("ride_id", response.ID),
	)

	return response, nil
}

func (rs *RideService) GetRideByID(ctx context.Context, rideID int) (*Ride, *ErrorResponse) {
	ride, err := rs.repo.GetRideByID(ctx, rideID)
	if err != nil {
		return nil, NewErrorResponse(err)
	}
	return ride, nil
}

func (rs *RideService) GetRideStatus(ctx context.Context, rideID int) (string, *ErrorResponse) {
	status, err := rs.repo.GetRideStatus(ctx, rideID)
	if err != nil {
		return "", NewErrorResponse(err)
	}
	return status, nil
}

func (rs *RideService) TakeRide(ctx context.Context, rideID int, driverID int) (*ChangeRideResponse, *ErrorResponse) {
	response, err := rs.repo.TakeRide(ctx, rideID, driverID)
	if err != nil {
		return nil, NewErrorResponse(err)
	}

	rs.logger.Info("driver took ride",
		slog.Int("ride_id", rideID),
		slog.Int("driver_id", driverID),
	)

	return response, nil
}

func (rs *RideService) CompleteRide(ctx context.Context, rideID int) (*ChangeRideResponse, *ErrorResponse) {
	response, err := rs.repo.CompleteRide(ctx, rideID)
	if err != nil {
		return nil, NewErrorResponse(err)
	}

	rs.logger.Info("driver completed ride",
		slog.Int("ride_id", rideID),
	)

	return response, nil
}

func (rs *RideService) CancelRide(ctx context.Context, rideID int) (*ChangeRideResponse, *ErrorResponse) {
	response, err := rs.repo.CancelRide(ctx, rideID)
	if err != nil {
		return nil, NewErrorResponse(err)
	}

	rs.logger.Info("user canceled ride",
		slog.Int("ride_id", rideID),
	)

	return response, nil
}

func (rs *RideService) GetSearchingRides(ctx context.Context) (*SearchRidesResponse, *ErrorResponse) {
	rides, err := rs.repo.GetSearchingRides(ctx)
	if err != nil {
		return nil, NewErrorResponse(err)
	}
	return rides, nil
}

func (s *RideService) CheckAccess(rideID, userID int, role string) error {
	ride, errResp := s.GetRideByID(context.Background(), rideID)
	if errResp != nil {
		return fmt.Errorf("failed to get ride: %s", errResp.Message)
	}

	if role == "USER" && ride.UserID != userID {
		return fmt.Errorf("this ride does not belong to you")
	}

	if role == "DRIVER" && ride.DriverID != nil && *ride.DriverID != userID {
		return fmt.Errorf("you are not assigned to this ride")
	}

	return nil
}

func fetchRideFromAPI(ctx context.Context, start, end PointGeoJSON) (*Geometry, error) {
	url := fmt.Sprintf(
		"http://router.project-osrm.org/route/v1/driving/%f,%f;%f,%f?geometries=geojson",
		start.Coordinates[0], start.Coordinates[1], end.Coordinates[0], end.Coordinates[1],
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("bad response: %s, body: %s", resp.Status, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	if apiResp.Code != "Ok" {
		return nil, fmt.Errorf("failed to fetch route from API")
	}

	if len(apiResp.Routes) == 0 {
		return nil, fmt.Errorf("no routes found")
	}

	return &apiResp.Routes[0].Geometry, nil
}
