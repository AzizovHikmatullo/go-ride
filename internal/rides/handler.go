package rides

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	userRole   = "USER"
	driverRole = "DRIVER"
)

type RideServiceInterface interface {
	CreateRide(ctx context.Context, userID int, start, end PointGeoJSON) (*CreateResponse, *ErrorResponse)
	GetRideByID(ctx context.Context, rideID int) (*Ride, *ErrorResponse)
	GetRideStatus(ctx context.Context, rideID int) (string, *ErrorResponse)
	TakeRide(ctx context.Context, rideID int, driverID int) (*ChangeRideResponse, *ErrorResponse)
	CompleteRide(ctx context.Context, rideID int) (*ChangeRideResponse, *ErrorResponse)
	CancelRide(ctx context.Context, rideID int) (*ChangeRideResponse, *ErrorResponse)
	GetSearchingRides(ctx context.Context) ([]Ride, *ErrorResponse)
	CheckAccess(rideID, userID int, role string) error
}

type RideHandler struct {
	service RideServiceInterface
}

func NewRideHandler(service RideServiceInterface) *RideHandler {
	return &RideHandler{
		service: service,
	}
}

func (rh *RideHandler) CreateRide(c *gin.Context) {
	var body CreateRequest

	if err := c.ShouldBindJSON(&body); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid input body")
		return
	}

	response, err := rh.service.CreateRide(c, c.GetInt("userID"), body.Start, body.End)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Message)
		return
	}
	c.JSON(http.StatusOK, response)

}

func (rh *RideHandler) GetRideByID(c *gin.Context) {
	id, ok := c.Params.Get("id")
	if !ok {
		newErrorResponse(c, http.StatusBadRequest, "invalid ride ID")
		return
	}

	rideID, convertErr := strconv.Atoi(id)
	if convertErr != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid ride ID")
		return
	}

	errResp := rh.service.CheckAccess(rideID, c.GetInt("userID"), userRole)
	if errResp != nil {
		newErrorResponse(c, http.StatusForbidden, errResp.Error())
		return
	}

	ride, err := rh.service.GetRideByID(c, rideID)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Message)
		return
	}
	c.JSON(http.StatusOK, ride)
}

func (rh *RideHandler) GetRideStatus(c *gin.Context) {
	id, ok := c.Params.Get("id")
	if !ok {
		newErrorResponse(c, http.StatusBadRequest, "invalid ride ID")
		return
	}

	rideID, convertErr := strconv.Atoi(id)
	if convertErr != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid ride ID")
		return
	}

	errResp := rh.service.CheckAccess(rideID, c.GetInt("userID"), userRole)
	if errResp != nil {
		newErrorResponse(c, http.StatusForbidden, errResp.Error())
		return
	}

	status, err := rh.service.GetRideStatus(c, rideID)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Message)
		return
	}
	c.JSON(http.StatusOK, map[string]interface{}{
		"order_id": rideID,
		"status":   status,
	})
}

func (rh *RideHandler) TakeRide(c *gin.Context) {
	id, ok := c.Params.Get("id")
	if !ok {
		newErrorResponse(c, http.StatusBadRequest, "invalid ride ID")
		return
	}

	idInt, convertErr := strconv.Atoi(id)
	if convertErr != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid ride ID")
		return
	}

	response, err := rh.service.TakeRide(c, idInt, c.GetInt("userID"))
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Message)
		return
	}
	c.JSON(http.StatusOK, response)
}

func (rh *RideHandler) CompleteRide(c *gin.Context) {
	id, ok := c.Params.Get("id")
	if !ok {
		newErrorResponse(c, http.StatusBadRequest, "invalid ride ID")
		return
	}

	rideID, convertErr := strconv.Atoi(id)
	if convertErr != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid ride ID")
		return
	}

	errResp := rh.service.CheckAccess(rideID, c.GetInt("userID"), userRole)
	if errResp != nil {
		newErrorResponse(c, http.StatusForbidden, errResp.Error())
		return
	}

	response, err := rh.service.CompleteRide(c, rideID)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Message)
		return
	}
	c.JSON(http.StatusOK, response)
}

func (rh *RideHandler) CancelRide(c *gin.Context) {
	id, ok := c.Params.Get("id")
	if !ok {
		newErrorResponse(c, http.StatusBadRequest, "invalid ride ID")
		return
	}

	rideID, convertErr := strconv.Atoi(id)
	if convertErr != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid ride ID")
		return
	}

	errResp := rh.service.CheckAccess(rideID, c.GetInt("userID"), userRole)
	if errResp != nil {
		newErrorResponse(c, http.StatusForbidden, errResp.Error())
		return
	}

	response, err := rh.service.CancelRide(c, rideID)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Message)
		return
	}
	c.JSON(http.StatusOK, response)
}

func (rh *RideHandler) GetSearchingRides(c *gin.Context) {
	rides, err := rh.service.GetSearchingRides(c)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Message)
		return
	}
	c.JSON(http.StatusOK, rides)
}

func newErrorResponse(c *gin.Context, statusCode int, message string) {
	c.AbortWithStatusJSON(statusCode, ErrorResponse{message})
}
