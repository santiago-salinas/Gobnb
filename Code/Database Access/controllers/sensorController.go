package controllers

import (
	"fmt"
	"net/http"
	"pocketbase_go/logger"
	"pocketbase_go/my_models"
	"pocketbase_go/services/interfaces"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

type SensorController struct {
	Service     interfaces.ISensorService
	AuthService interfaces.IAuthService
}

func (controller *SensorController) InitSensorEndpoints(app core.App) {
	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {

		e.Router.POST("/sensor", func(c echo.Context) error {
			token := c.Request().Header.Get("auth")
			var req my_models.Sensor

			if err := c.Bind(&req); err != nil {
				logger.Error("Failed to read request data: \n", err)
				return apis.NewBadRequestError("Failed to read request data", err)
			}
			logger.Info("Request: \n", req)

			response := controller.PostSensor(token, req)
			if response != nil {
				return c.JSON(http.StatusNotAcceptable, map[string]string{"message": response.Error()})
			}
			return c.JSON(http.StatusCreated, map[string]string{"message": "Success"})
		})

		e.Router.GET("/sensor/:id", func(c echo.Context) error {
			id := c.PathParam("id")

			response, err := controller.GetSensor(id)
			if err != nil {
				logger.Error("Error retrieving sensor: \n", err)
				return c.JSON(http.StatusNotAcceptable, map[string]string{"message": err.Error()})
			}
			logger.Info("Success retrieving sensor: \n", response)
			return c.JSON(http.StatusCreated, response)
		})

		e.Router.POST("/sensor/assign", func(c echo.Context) error {
			token := c.Request().Header.Get("auth")

			type AssignBody struct {
				Sensor   string `json:"sensor"`
				Property string `json:"property"`
			}

			var req AssignBody

			if err := c.Bind(&req); err != nil {
				logger.Error("Failed to read request data: \n", err)
				return apis.NewBadRequestError("Failed to read request data", err)
			}
			err := controller.AssignSensorToProperty(token, req.Sensor, req.Property)
			if err != nil {
				logger.Error("Failed to assign sensor to property: \n", err)
				return c.JSON(http.StatusNotAcceptable, map[string]string{"message": err.Error()})
			}
			logger.Info("Success assigning sensor to property: \n", req)
			return c.JSON(http.StatusCreated, map[string]string{"message": "Success"})
		})
		return nil
	})
}

func (c *SensorController) PostSensor(token string, sensor my_models.Sensor) error {
	roles, _, err := c.AuthService.Login(token)
	if err != nil {
		logger.Error("Controller: Error in PostSensor: ", err)
		return err
	}

	for _, role := range roles {
		if role == "Admin" {
			return c.Service.AddSensor(sensor)
		}
	}

	err = fmt.Errorf("Controller: User is not an admin")
	logger.Error("Controller: Error in PostSensor: User is not an admin")
	return err
}

func (c *SensorController) GetSensor(id string) (my_models.Sensor, error) {
	sensor, err := c.Service.GetSensor(id)
	if err != nil {
		return my_models.Sensor{}, err
	}

	return sensor, nil
}

func (c *SensorController) AssignSensorToProperty(token string, sensorId string, propertyId string) error {
	roles, _, err := c.AuthService.Login(token)
	if err != nil {
		logger.Error("Controller: Error in AssignSensorToProperty: ", err)
		return err
	}

	for _, role := range roles {
		if role == "Admin" {
			return c.Service.AssignSensorToProperty(sensorId, propertyId)
		}
	}

	err = fmt.Errorf("Controller: User is not an admin")
	logger.Error("Controller: Error in AssignSensorToProperty: User is not an admin")
	return err
}
