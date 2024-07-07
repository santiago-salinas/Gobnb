package repointerfaces

import (
	"mongo-server/mongo_models"
	"time"
)

type ReportsRepo interface {
	AddAppReport(report mongo_models.AppReport) error
	AddSensorReport(report mongo_models.SensorReport) error
	GetAllAppReports(startDate time.Time, endDate time.Time) ([]mongo_models.RankingReportItem, error)
	GetLatestSensorReport(sensorId string) (mongo_models.SensorReport, error)
}
