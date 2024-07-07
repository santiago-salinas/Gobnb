package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mongo-server/mongo_models"
	"net/http"
	"pocketbase_go/logger"
	"pocketbase_go/my_models"
	pipes_and_filters "pocketbase_go/pipes-and-filters"
	"pocketbase_go/services/interfaces"
	"pocketbase_go/workers"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

type ReportsController struct {
	AuthService          interfaces.IAuthService
	ReportsService       interfaces.IReportsService
	SensorService        interfaces.ISensorService
	NotificationsService interfaces.INotificationService
	SensorPipeline       pipes_and_filters.SensorReportPipeline
	AppPipeline          pipes_and_filters.AppReportPipeline
	worker               workers.Worker
}

func NewReportsController(
	authService interfaces.IAuthService,
	reportsService interfaces.IReportsService,
	notificationsService interfaces.INotificationService,
	worker workers.Worker,
) *ReportsController {

	controller := &ReportsController{
		AuthService:          authService,
		ReportsService:       reportsService,
		NotificationsService: notificationsService,
		worker:               worker,
	}

	sensorPipeline := pipes_and_filters.SensorReportPipeline{}
	sensorPipeline.Use(
		pipes_and_filters.ValidateSensorReport,
		reportsService.ValidateSensorReport,
		controller._checkSensorIsAssigned,
		controller._sendSensorReportToMongo,
		controller._publishSensorReport,
	)

	appPipeline := pipes_and_filters.AppReportPipeline{}
	appPipeline.Use(pipes_and_filters.ValidateAppReport, controller._sendAppReportToMongo)

	controller.SensorPipeline = sensorPipeline
	controller.AppPipeline = appPipeline

	return controller
}

func (controller *ReportsController) InitReportsEndpoints(app core.App) {
	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.GET("/reports/incomes", func(c echo.Context) error {
			token := c.Request().Header.Get("auth")

			country := c.Request().Header.Get("country")
			city := c.Request().Header.Get("city")
			if country == "" || city == "" {
				return c.JSON(http.StatusNotAcceptable, map[string]string{"message": "country and city headers are required"})
			}

			propertyId := c.QueryParam("propertyId")
			fromDate := c.QueryParam("fromDate")
			untilDate := c.QueryParam("untilDate")

			response, err := controller.GetPropertiesIncomes(token, propertyId, fromDate, untilDate)
			if err != nil {
				return c.JSON(http.StatusNotAcceptable, map[string]string{"message": err.Error()})
			}

			response.City = city
			response.Country = country

			return c.JSON(http.StatusOK, response)
		})

		e.Router.GET("/reports/occupations", func(c echo.Context) error {
			token := c.Request().Header.Get("auth")

			country := c.Request().Header.Get("country")
			city := c.Request().Header.Get("city")
			if country == "" || city == "" {
				return c.JSON(http.StatusNotAcceptable, map[string]string{"message": "country and city headers are required"})
			}

			fromDate := c.QueryParam("fromDate")
			untilDate := c.QueryParam("untilDate")

			response := my_models.OccupationsReport{}
			items, err := controller.GetOccupations(token, fromDate, untilDate)
			if err != nil {
				return c.JSON(http.StatusNotAcceptable, map[string]string{"message": err.Error()})
			}
			response.Items = items
			response.City = city
			response.Country = country
			response.FromDate, _ = time.Parse(time.DateOnly, fromDate)
			response.ToDate, _ = time.Parse(time.DateOnly, untilDate)

			return c.JSON(http.StatusOK, response)
		})

		e.Router.POST("/reports/app", func(c echo.Context) error {
			var req mongo_models.AppReport
			if err := c.Bind(&req); err != nil {
				return apis.NewBadRequestError("Failed to read request data", err)
			}

			err := controller.InputAppReportInPipeline(req)

			if err != nil {
				return c.JSON(http.StatusNotAcceptable, map[string]string{"message": err.Error()})
			}

			return c.JSON(http.StatusOK, map[string]string{"message": "Success"})
		})

		e.Router.POST("/reports/sensor", func(c echo.Context) error {
			var req mongo_models.SensorReport
			if err := c.Bind(&req); err != nil {
				return apis.NewBadRequestError("Failed to read request data", err)
			}

			err := controller.InputSensorReportInPipeline(req)

			if err != nil {
				return c.JSON(http.StatusNotAcceptable, map[string]string{"message": err.Error()})
			}

			return c.JSON(http.StatusOK, map[string]string{"message": "Success"})
		})

		e.Router.GET("/reports/properties-ranking", func(c echo.Context) error {
			token := c.Request().Header.Get("auth")
			country := c.Request().Header.Get("country")
			city := c.Request().Header.Get("city")
			if country == "" || city == "" {
				return c.JSON(http.StatusNotAcceptable, map[string]string{"message": "country and city headers are required"})
			}

			fromDate := c.QueryParam("fromDate")
			untilDate := c.QueryParam("untilDate")

			propertiesRanking, err := controller.GetPropertiesRanking(token, fromDate, untilDate)
			if err != nil {
				return c.JSON(http.StatusNotAcceptable, map[string]string{"message": err.Error()})
			}

			response := mongo_models.RankingReport{Items: propertiesRanking}

			return c.JSON(http.StatusOK, response)
		})

		e.Router.GET("/sensor/:id/state", func(c echo.Context) error {
			id := c.PathParam("id")
			token := c.Request().Header.Get("auth")

			response, err := controller.GetLatestSensorReport(id, token)
			if err != nil {
				logger.Error("Error retrieving sensor state: \n", err)
				return c.JSON(http.StatusNotAcceptable, map[string]string{"message": err.Error()})
			}
			logger.Info("Success retrieving sensor state: \n", response)
			return c.JSON(http.StatusCreated, response)
		})
		return nil
	})

}

func (c *ReportsController) GetLatestSensorReport(sensorId string, token string) (mongo_models.SensorReport, error) {
	roles, _, err := c.AuthService.Login(token)
	if err != nil {
		logger.Error("Controller: error login with token ", token, ": ", err)
		return mongo_models.SensorReport{}, err
	}

	for _, role := range roles {
		if role == "Admin" {
			report, err := c.ReportsService.GetLatestSensorReport(sensorId)
			if err != nil {
				logger.Error("Controller: error retrieving latest sensor report for id ", sensorId, ": ", err)
			}
			return report, err
		}
	}
	err = fmt.Errorf("user does not have the role Admin")
	logger.Error("Controller: ", err)
	return mongo_models.SensorReport{}, err
}

func (c *ReportsController) GetPropertiesIncomes(userToken string, property_id string, fromDateStr string, untilDateStr string) (my_models.IncomeReport, error) {
	logger.Info("Controller: Getting properties incomes")
	roles, _, err := c.AuthService.Login(userToken)

	if err != nil {
		return my_models.IncomeReport{}, err
	}

	for _, role := range roles {
		if role == "Admin" || role == "Operator" {

			fromDate, err := time.Parse(time.DateOnly, fromDateStr)
			if err != nil {
				return my_models.IncomeReport{}, err
			}

			untilDate, err := time.Parse(time.DateOnly, untilDateStr)
			if err != nil {
				return my_models.IncomeReport{}, err
			}

			if fromDate.After(untilDate) {
				return my_models.IncomeReport{}, fmt.Errorf("from date must be before until date")
			}

			report, err := c.ReportsService.GetPropertiesIncomes(property_id, fromDate, untilDate)
			if err != nil {
				return my_models.IncomeReport{}, err
			}

			logger.Info("Controller: Got properties incomes successfully")
			return report, nil
		}
	}

	logger.Error("Controller: User is not authorized to getPropertiesIncome")
	return my_models.IncomeReport{}, fmt.Errorf("user is not authorized to perform this action")
}

func (c *ReportsController) GetOccupations(userToken string, fromDateStr string, untilDateStr string) ([]my_models.OccupationsReportItem, error) {
	logger.Info("Controller: Getting occupations")
	roles, _, err := c.AuthService.Login(userToken)

	if err != nil {
		return []my_models.OccupationsReportItem{}, err
	}

	for _, role := range roles {
		if role == "Admin" || role == "Operator" {
			fromDate, err := time.Parse(time.DateOnly, fromDateStr)
			if err != nil {
				return nil, err
			}

			untilDate, err := time.Parse(time.DateOnly, untilDateStr)
			if err != nil {
				return nil, err
			}

			if fromDate.After(untilDate) {
				return nil, fmt.Errorf("from date must be before until date")
			}

			reports, err := c.ReportsService.GetOccupations(fromDate, untilDate)
			if err != nil {
				return []my_models.OccupationsReportItem{}, err
			}

			logger.Info("Controller: Got occupations successfully")
			return reports, nil
		}
	}

	logger.Error("Controller: User is not authorized to getOccupations")
	return []my_models.OccupationsReportItem{}, fmt.Errorf("user is not authorized to perform this action")
}

func (c *ReportsController) GetPropertiesRanking(userToken string, fromDateStr string, untilDateStr string) ([]mongo_models.RankingReportItem, error) {
	logger.Info("Controller: Getting properties ranking")
	roles, _, err := c.AuthService.Login(userToken)

	if err != nil {
		return []mongo_models.RankingReportItem{}, err
	}

	for _, role := range roles {
		if role == "Admin" || role == "Operator" {
			fromDate, err := time.Parse(time.DateOnly, fromDateStr)
			if err != nil {
				return nil, err
			}

			untilDate, err := time.Parse(time.DateOnly, untilDateStr)
			if err != nil {
				return nil, err
			}

			if fromDate.After(untilDate) {
				return nil, fmt.Errorf("from date must be before until date")
			}

			reports, err := c.ReportsService.GetPropertiesRanking(fromDate, untilDate)
			if err != nil {
				return []mongo_models.RankingReportItem{}, err
			}

			logger.Info("Controller: Got properties ranking successfully")
			return reports, nil
		}
	}

	logger.Error("Controller: User is not authorized to getPropertiesRanking")
	return []mongo_models.RankingReportItem{}, fmt.Errorf("user is not authorized to perform this action")
}

func (c *ReportsController) InputSensorReportInPipeline(report mongo_models.SensorReport) error {
	return c.SensorPipeline.Run(report)
}

func (c *ReportsController) InputAppReportInPipeline(report mongo_models.AppReport) error {
	return c.AppPipeline.Run(report)
}

func (c *ReportsController) _sendAppReportToMongo(report mongo_models.AppReport) (mongo_models.AppReport, error) {
	err := c.worker.Health()
	if err != nil {
		logger.Error("Failed to send app report to mongo: ", err)
		return mongo_models.AppReport{}, err
	}

	reqBodyBytes := new(bytes.Buffer)
	json.NewEncoder(reqBodyBytes).Encode(report)

	c.worker.Send("appReports", []byte(reqBodyBytes.Bytes()))

	return report, nil
}

func (c *ReportsController) _sendSensorReportToMongo(report mongo_models.SensorReport) (mongo_models.SensorReport, error) {
	err := c.worker.Health()
	if err != nil {
		logger.Error("Failed to send sensor report to mongo: ", err)
		return mongo_models.SensorReport{}, err
	}

	reqBodyBytes := new(bytes.Buffer)
	json.NewEncoder(reqBodyBytes).Encode(report)

	c.worker.Send("sensorReports", []byte(reqBodyBytes.Bytes()))

	return report, nil
}

func (c *ReportsController) _publishSensorReport(report mongo_models.SensorReport) (mongo_models.SensorReport, error) {
	for key, reportItem := range report.Reports {
		var channel string
		message := fmt.Sprintf("Sensor %s: %s - %s %s", report.SensorId, key, reportItem.Value, reportItem.Unit)
		if key == "Security" {
			channel = report.SensorId
		} else {
			channel = key
		}
		err := c.NotificationsService.PublishToChannel(channel, message)
		if err != nil {
			logger.Error("Failed to publish sensor report to channel: ", err)
			return mongo_models.SensorReport{}, err
		}
	}

	return report, nil
}

func (c *ReportsController) _checkSensorIsAssigned(report mongo_models.SensorReport) (mongo_models.SensorReport, error) {
	sensor, err := c.SensorService.GetSensor(report.SensorId)
	if err != nil {
		logger.Error("Sensor not found: ", err)
		return mongo_models.SensorReport{}, err
	}

	if sensor.AssignedTo == "" {
		err = fmt.Errorf("Sensor %s is not assigned to any property", report.SensorId)
		logger.Error(err)
		return mongo_models.SensorReport{}, err
	}

	return report, nil
}
