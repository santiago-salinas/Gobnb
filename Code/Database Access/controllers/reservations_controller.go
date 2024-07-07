package controllers

import (
	"fmt"
	"net/http"
	"os"
	"pocketbase_go/logger"
	"pocketbase_go/my_models"
	"pocketbase_go/services/interfaces"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/prometheus/client_golang/prometheus"
)

type ReservationsController struct {
	ReservationsService interfaces.IReservationService
	AuthService         interfaces.IAuthService
	PaymentService      interfaces.IPaymentService
}

func (controller *ReservationsController) InitReservationEndpoints(app core.App, monitorReservations, monitorReservationPaymentSuccess, monitorReservationPaymentFailure prometheus.Counter) {
	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {

		e.Router.GET("/reservations", func(c echo.Context) error {
			token := c.Request().Header.Get("auth")

			var filter my_models.ReservationFilter

			if reservedFromQuery := c.QueryParam("reservedFrom"); reservedFromQuery != "" {
				filter.ReservedFrom = &reservedFromQuery
			}
			if reservedUntilQuery := c.QueryParam("reservedUntil"); reservedUntilQuery != "" {
				filter.ReservedUntil = &reservedUntilQuery
			}
			if statusQuery := c.QueryParam("status"); statusQuery != "" {
				filter.Status = &statusQuery
			}
			if propertyIdQuery := c.QueryParam("propertyId"); propertyIdQuery != "" {
				filter.PropertyId = &propertyIdQuery
			}
			if tenantEmailQuery := c.QueryParam("email"); tenantEmailQuery != "" {
				filter.TenantEmail = &tenantEmailQuery
			}
			if tenantNameQuery := c.QueryParam("name"); tenantNameQuery != "" {
				filter.TenantName = &tenantNameQuery
			}
			if tenantLastNameQuery := c.QueryParam("lastName"); tenantLastNameQuery != "" {
				filter.TenantLastName = &tenantLastNameQuery
			}

			response, err := controller.GetFilteredReservations(filter, token)
			if err != nil {
				return c.JSON(http.StatusNotAcceptable, map[string]string{"message": err.Error()})
			}
			return c.JSON(http.StatusCreated, response)
		})

		e.Router.POST("/reservations", func(c echo.Context) error {
			var req my_models.ReservationModel
			token := c.Request().Header.Get("auth")
			if token == "" {
				return c.JSON(http.StatusNotAcceptable, map[string]string{"message": "failed to read auth token"})
			}

			if err := c.Bind(&req); err != nil {
				return apis.NewBadRequestError("Failed to read request data", err)
			}
			fmt.Fprintf(os.Stdout, "Request: %v\n", req)
			response := controller.PostReservation(req, token)
			if response != nil {
				return c.JSON(http.StatusNotAcceptable, map[string]string{"message": response.Error()})
			}
			monitorReservations.Inc()
			return c.JSON(http.StatusCreated, map[string]string{"message": "Success"})
		})

		e.Router.GET("/reservations/:email/:propertyId", func(c echo.Context) error {
			token := c.Request().Header.Get("auth")
			email := c.PathParam("email")
			propertyId := c.PathParam("propertyId")

			response, err := controller.GetOwnReservation(token, email, propertyId)
			if err != nil {
				return c.JSON(http.StatusNotAcceptable, map[string]string{"message": err.Error()})
			}
			return c.JSON(http.StatusOK, response)
		})

		e.Router.POST("/reservations/:reservationId/approve", func(c echo.Context) error {
			reservationId := c.PathParam("reservationId")
			token := c.Request().Header.Get("auth")

			err := controller.ApproveReservation(reservationId, token)
			if err != nil {
				return c.JSON(http.StatusNotAcceptable, map[string]string{"message": err.Error()})
			}
			return c.JSON(http.StatusOK, map[string]string{"message": "Success"})
		})

		e.Router.POST("/reservations/:reservationId/cancel", func(c echo.Context) error {
			reservationId := c.PathParam("reservationId")
			email := c.Request().Header.Get("email")
			token := c.Request().Header.Get("auth")

			refundPercentage, err := controller.CancelReservation(email, reservationId, token)
			if err != nil {
				return c.JSON(http.StatusNotAcceptable, map[string]string{"message": err.Error()})
			}
			return c.JSON(http.StatusOK, map[string]string{"message": "Success", "refundPercentage": fmt.Sprintf("%f", refundPercentage)})
		})

		e.Router.POST("/reservations/:reservationId/remove", func(c echo.Context) error {
			reservationId := c.PathParam("reservationId")
			token := c.Request().Header.Get("auth")

			err := controller.RemoveReservation(reservationId, token)
			if err != nil {
				return c.JSON(http.StatusNotAcceptable, map[string]string{"message": err.Error()})
			}
			return c.JSON(http.StatusOK, map[string]string{"message": "Success"})
		})

		e.Router.POST("/reservations/:reservationId/check_in", func(c echo.Context) error {
			reservationId := c.PathParam("reservationId")
			token := c.Request().Header.Get("auth")

			err := controller.DoCheckIn(reservationId, token)
			if err != nil {
				return c.JSON(http.StatusNotAcceptable, map[string]string{"message": err.Error()})
			}
			return c.JSON(http.StatusOK, map[string]string{"message": "Success"})
		})

		e.Router.POST("/reservations/:reservationId/check_out", func(c echo.Context) error {
			reservationId := c.PathParam("reservationId")
			token := c.Request().Header.Get("auth")

			err := controller.DoCheckOut(reservationId, token)
			if err != nil {
				return c.JSON(http.StatusNotAcceptable, map[string]string{"message": err.Error()})
			}
			return c.JSON(http.StatusOK, map[string]string{"message": "Success"})
		})

		e.Router.POST("/reservations/pay", func(c echo.Context) error {
			token := c.Request().Header.Get("auth")
			type PayBody struct {
				ReservationId string `json:"reservationId"`
				CardInfo      my_models.CardInformation
			}

			var req PayBody

			if err := c.Bind(&req); err != nil {
				return apis.NewBadRequestError("Failed to read request data", err)
			}
			fmt.Fprintf(os.Stdout, "Request: %v\n", req)
			response := controller.PayReservation(req.ReservationId, req.CardInfo, token)
			if response != nil {
				monitorReservationPaymentFailure.Inc()
				return c.JSON(http.StatusNotAcceptable, map[string]string{"message": response.Error()})
			}
			monitorReservationPaymentSuccess.Inc()
			return c.JSON(http.StatusCreated, map[string]string{"message": "Success"})
		})

		return nil
	})
}

func (c *ReservationsController) GetFilteredReservations(filter my_models.ReservationFilter, userToken string) ([]my_models.ReservationModel, error) {
	roles, _, err := c.AuthService.Login(userToken)
	if err != nil {
		logger.Error("Controller: Error in GetFilteredReservations: ", err)
		return nil, err
	}

	for _, role := range roles {
		if role == "Admin" || role == "Operator" {
			reservations, err := c.ReservationsService.GetFilteredReservations(filter)
			if err != nil {
				return nil, err
			}

			return reservations, nil
		}
	}

	logger.Error("Controller: Error in GetFilteredReservations: provided token does not belong to an Admin or Operator user")
	return nil, fmt.Errorf("provided token does not belong to an Admin or Operator user")
}

func (c *ReservationsController) PostReservation(reservation my_models.ReservationModel, userToken string) error {
	roles, _, err := c.AuthService.Login(userToken)
	if err != nil {
		logger.Error("Controller: Error in PostReservation: ", err)
		return err
	}

	for _, role := range roles {
		if role == "Tenant" {
			return c.ReservationsService.CreateReservation(reservation)
		}
	}

	logger.Error("Controller: Error in PostReservation: provided token does not belong to a Tenant user")
	return fmt.Errorf("provided token does not belong to a Client user")
}

func (c *ReservationsController) GetOwnReservation(token string, email string, propertyId string) (my_models.ReservationModel, error) {
	roles, userId, err := c.AuthService.Login(token)
	if err != nil {
		logger.Error("Controller: Error in GetOwnReservation: ", err)
		return my_models.ReservationModel{}, err
	}

	user, err := c.AuthService.GetUserById(userId)
	if err != nil {
		logger.Error("Controller: Error in GetOwnReservation: ", err)
		return my_models.ReservationModel{}, err
	}

	if user.Email != email {
		logger.Error("Controller: Error in GetOwnReservation: provided token does not match the user email")
		return my_models.ReservationModel{}, fmt.Errorf("provided token does not match the user email")
	}

	for _, role := range roles {
		if role == "Tenant" {
			return c.ReservationsService.GetOwnReservation(email, propertyId)
		}
	}

	err = fmt.Errorf("Controller: User is not a Tenant")
	logger.Error("Controller: Error in GetOwnReservation: User is not a Tenant")
	return my_models.ReservationModel{}, err
}

func (c *ReservationsController) ApproveReservation(reservationId string, userToken string) error {
	roles, _, err := c.AuthService.Login(userToken)
	if err != nil {
		return err
	}

	for _, role := range roles {
		if role == "Admin" {
			return c.ReservationsService.ApproveReservation(reservationId)
		}
	}

	err = fmt.Errorf("provided token does not belong to an Admin")
	logger.Error("Controller: Error in ApproveReservation: ", err)
	return err
}

func (c *ReservationsController) CancelReservation(email string, reservationId string, userToken string) (refundPercentage float64, err error) {
	roles, _, err := c.AuthService.Login(userToken)
	if err != nil {
		return 0, err
	}

	for _, role := range roles {
		if role == "Tenant" {
			return c.ReservationsService.CancelReservation(email, reservationId)
		}
	}

	err = fmt.Errorf("Controller: provided token does not belong to a Tenant user")
	logger.Error("Controller: Error in CancelReservation: ", err)
	return 0, err
}

func (c *ReservationsController) RemoveReservation(reservationId string, userToken string) (err error) {
	roles, _, err := c.AuthService.Login(userToken)
	if err != nil {
		return err
	}

	for _, role := range roles {
		if role == "Admin" || role == "Operator" {
			return c.ReservationsService.RemoveReservation(reservationId)
		}
	}

	err = fmt.Errorf("Controller: provided token does not belong to an Admin or Operator user")
	logger.Error("Controller: Error in RemoveReservation: ", err)
	return err
}

func (c *ReservationsController) DoCheckIn(reservationId string, token string) error {
	_, userId, err := c.AuthService.Login(token)
	if err != nil {
		logger.Error("Controller: Error in DoCheckIn: ", err)
		return err
	}

	user, err := c.AuthService.GetUserById(userId)
	if err != nil {

		return err
	}

	userEmail := user.Email

	reservation, err := c.ReservationsService.GetReservationById(reservationId)
	if err != nil {
		return err
	}

	if reservation.CheckIn != "" {
		return fmt.Errorf("reservation is already checked in")
	}

	if reservation.Email != userEmail {
		return fmt.Errorf("provided token does not match the reservation email")
	}

	reservationInitDate, err := time.Parse(my_models.PocketTimeLayout, reservation.ReservedFrom)
	if err != nil {
		return err
	}
	reservationEndDate, err := time.Parse(my_models.PocketTimeLayout, reservation.ReservedUntil)
	if err != nil {
		return err
	}

	if reservation.Status != "Approved" {
		return fmt.Errorf("reservation is not approved")
	}

	today := time.Now()

	if reservationInitDate.After(today) {
		return fmt.Errorf("reservation start date is not for today")
	}

	if reservationEndDate.Before(today) {
		return fmt.Errorf("reservation period already passed")
	}

	return c.ReservationsService.DoCheckIn(reservationId)

}

func (c *ReservationsController) DoCheckOut(reservationId string, token string) error {
	_, userId, err := c.AuthService.Login(token)
	if err != nil {
		return err
	}

	user, err := c.AuthService.GetUserById(userId)
	if err != nil {
		return err
	}

	userEmail := user.Email

	reservation, err := c.ReservationsService.GetReservationById(reservationId)
	if err != nil {
		return err
	}

	if reservation.CheckOut != "" {
		return fmt.Errorf("reservation is already checked out")
	}

	if reservation.Email != userEmail {
		return fmt.Errorf("provided token does not match the reservation email")
	}

	reservationEndDate, err := time.Parse(my_models.PocketTimeLayout, reservation.ReservedUntil)
	if err != nil {
		return err
	}

	if reservation.CheckIn == "" {
		return fmt.Errorf("reservation is not checked in")
	}

	today := time.Now()

	if reservationEndDate.Before(today) {
		return fmt.Errorf("checkout is late, reservation period already passed")
	}

	return c.ReservationsService.DoCheckOut(reservationId)
}

func (c *ReservationsController) PayReservation(reservationId string, cardInformation my_models.CardInformation, token string) error {
	logger.Info("Controller: Paying reservation with id: ", reservationId)

	roles, _, err := c.AuthService.Login(token)
	if err != nil {
		return err
	}

	for _, role := range roles {
		if role == "Tenant" {
			return c.PaymentService.PayReservation(reservationId, cardInformation)
		}
	}

	err = fmt.Errorf("provided token does not belong to a Tenant user")
	logger.Error("Controller: Error in PayReservation: ", err)
	return err
}

func (c *ReservationsController) AutoCancelReservations() error {
	logger.Info("Controller: AutoCancelReservations")
	err := c.ReservationsService.AutoCancelReservations()
	if err != nil {
		logger.Error("Controller: Error in AutoCancelReservations: ", err)
	} else {
		logger.Info("Controller: AutoCancelReservations done")
	}
	return err
}
