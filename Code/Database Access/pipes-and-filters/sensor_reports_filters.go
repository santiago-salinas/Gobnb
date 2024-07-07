package pipes_and_filters

import (
	"fmt"
	"mongo-server/mongo_models"
	"pocketbase_go/logger"
	"strings"
	"time"
)

func ValidateSensorReport(report mongo_models.SensorReport) (mongo_models.SensorReport, error) {
	logger.Info("Controller: Validating sensor report")
	if report.SensorId == "" {
		logger.Error("Controller: sensor id is required")
		return mongo_models.SensorReport{}, fmt.Errorf("sensor id is required")
	}

	_, err := time.Parse(time.DateOnly, report.Date)
	if err != nil {
		logger.Error("Controller: invalid date format")
		return mongo_models.SensorReport{}, fmt.Errorf("invalid date format")
	}

	return report, nil
}

func ValidateAppReport(report mongo_models.AppReport) (mongo_models.AppReport, error) {
	logger.Info("Controller: Validating app report")

	if report.SensorId == "" {
		logger.Error("Controller: app id is required")
		return mongo_models.AppReport{}, fmt.Errorf("app id is required")
	}

	if !strings.HasPrefix(report.SensorId, "APP") {
		logger.Error("Controller: invalid app id")
		return mongo_models.AppReport{}, fmt.Errorf("invalid app id")
	}

	_, err := time.Parse(time.DateOnly, report.Date)
	if err != nil {
		logger.Error("Controller: invalid date format")
		return mongo_models.AppReport{}, fmt.Errorf("invalid date format")
	}

	logger.Info("Controller: App report validated successfully")
	return report, nil
}


