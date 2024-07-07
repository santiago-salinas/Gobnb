package controllers

import (
	"fmt"
	"net/http"
	"pocketbase_go/logger"
	"pocketbase_go/my_models"
	"pocketbase_go/services/interfaces"

	"strings"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

type NotificationsController struct {
	NotificationsService interfaces.INotificationService
	ReservationsService  interfaces.IReservationService
}

const (
	FireReportChannel        = "Fire"
	ElectricityReportChannel = "Electricity"
	GasReportChannel         = "Gas"
)

func NewNotificationsController(notificationsService interfaces.INotificationService, reservationsService interfaces.IReservationService) *NotificationsController {
	notificationsService.OpenChannel(FireReportChannel)
	notificationsService.OpenChannel(ElectricityReportChannel)
	notificationsService.OpenChannel(GasReportChannel)

	notificationsService.SubscribeToChannel(FireReportChannel, "FireDepartment", notificationsService.MailMethod("FireDepartment"))
	notificationsService.SubscribeToChannel(ElectricityReportChannel, "ElectricityCompany", notificationsService.MailMethod("ElectricityCompany"))
	notificationsService.SubscribeToChannel(GasReportChannel, "GasCompany", notificationsService.MailMethod("GasCompany"))

	return &NotificationsController{
		NotificationsService: notificationsService,
		ReservationsService:  reservationsService,
	}
}

func (controller *NotificationsController) InitNotificationsEndpoints(app core.App) {
	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.POST("/notifications/subscribe/security-notifications", func(c echo.Context) error {
			type SubscribeBody struct {
				Email      string `json:"email"`
				PropertyId string `json:"propertyId"`
			}

			var req SubscribeBody
			if err := c.Bind(&req); err != nil {
				return apis.NewBadRequestError("Failed to read request data", err)
			}

			err := controller.SubscribeTenantToPropertyChannel(req.Email, req.PropertyId)
			if err != nil {
				return c.JSON(http.StatusNotAcceptable, map[string]string{"message": err.Error()})
			}
			return c.JSON(http.StatusCreated, map[string]string{"message": "Success"})
		})

		e.Router.POST("/notifications/subscribe", func(c echo.Context) error {
			type SubscribeBody struct {
				Subscriber string `json:"subscriber"`
				Channel    string `json:"channel"`
			}

			var req SubscribeBody
			if err := c.Bind(&req); err != nil {
				return apis.NewBadRequestError("Failed to read request data", err)
			}

			err := controller.SubscribeToChannel(req.Subscriber, req.Channel)
			if err != nil {
				return c.JSON(http.StatusNotAcceptable, map[string]string{"message": err.Error()})
			}
			return c.JSON(http.StatusCreated, map[string]string{"message": "Success"})
		})

		e.Router.POST("/notifications/unsubscribe", func(c echo.Context) error {
			type UnsubscribeBody struct {
				Subscriber string `json:"subscriber"`
				Channel    string `json:"channel"`
			}

			var req UnsubscribeBody
			if err := c.Bind(&req); err != nil {
				return apis.NewBadRequestError("Failed to read request data", err)
			}

			err := controller.UnsubscribeFromChannel(req.Subscriber, req.Channel)
			if err != nil {
				return c.JSON(http.StatusNotAcceptable, map[string]string{"message": err.Error()})
			}
			return c.JSON(http.StatusOK, map[string]string{"message": "Success"})
		})

		return nil
	})
}

func (c *NotificationsController) SubscribeToChannel(subscriber string, channel string) error {
	logger.Info("Controller: Subscribing to channel: ", channel)
	if strings.Contains(subscriber, ":") || strings.Contains(channel, ":") {
		logger.Error("Controller: ':' is not allowed in subscriber or channel name")
		return fmt.Errorf("':' is not allowed in subscriber or channel name")
	}

	err := c.NotificationsService.SubscribeToChannel(channel, subscriber, c.NotificationsService.MailMethod(subscriber))
	if err != nil {
		return err
	}
	logger.Info("Controller: ", subscriber, " subscribed to channel: ", channel)
	return nil
}

func (c *NotificationsController) UnsubscribeFromChannel(subscriber string, channel string) error {
	logger.Info("Controller: Unsubscribing from channel: ", channel)
	err := c.NotificationsService.UnsubscribeFromChannel(subscriber, channel)
	if err != nil {
		return err
	}
	logger.Info("Controller: ", subscriber, " unsubscribed from channel: ", channel)
	return nil
}

func (c *NotificationsController) SubscribeTenantToPropertyChannel(tenantEmail string, propertyId string) error {
	logger.Info("Controller: Subscribing tenant to property channel")

	reservation, err := c.ReservationsService.GetOwnReservation(tenantEmail, propertyId)
	if err != nil {
		logger.Error("Controller: Error getting tenant reservation")
		return err
	}
	if reservation.Status != "Approved" {
		logger.Error("Controller: Tenant reservation is not approved")
		return fmt.Errorf("Tenant reservation is not approved")
	}

	fromDate, _ := time.Parse(my_models.PocketTimeLayout, reservation.ReservedFrom)
	untilDate, _ := time.Parse(my_models.PocketTimeLayout, reservation.ReservedUntil)
	if time.Now().Before(fromDate) || time.Now().After(untilDate) {
		logger.Error("Controller: Tenant is not allowed to subscribe to property channel")
		return fmt.Errorf("Tenant is not allowed to subscribe to property channel")
	}

	channel := "Security-" + propertyId
	err = c.NotificationsService.OpenChannel(channel)
	if err != nil {
		logger.Error("Controller: Error opening channel: ", channel)
		return err
	}
	err = c.NotificationsService.SubscribeToChannel(channel, tenantEmail, c.NotificationsService.MailMethod(tenantEmail))
	if err != nil {
		logger.Error("Controller: Error subscribing tenant to property channel")
		return err
	}
	logger.Info("Controller: Tenant subscribed to property channel")
	return nil
}
