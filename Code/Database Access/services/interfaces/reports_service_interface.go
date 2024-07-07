package interfaces

import (
	"mongo-server/mongo_models"
	"pocketbase_go/my_models"
	"time"
)

type IReportsService interface {
	GetLatestSensorReport(sensorId string) (mongo_models.SensorReport, error)
	ValidateSensorReport(report mongo_models.SensorReport) (mongo_models.SensorReport, error)
	ValidateSecurityReport(report mongo_models.SensorReport) error
	GetPropertiesIncomes(property_id string, fromDate time.Time, untilDate time.Time) (my_models.IncomeReport, error)
	GetOccupations(fromDate time.Time, untilDate time.Time) ([]my_models.OccupationsReportItem, error)
	GetPropertiesRanking(fromDate time.Time, untilDate time.Time) ([]mongo_models.RankingReportItem, error)
}
