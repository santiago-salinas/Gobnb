package mongo_models

import "time"

type AppReport struct {
	PropertyId string `json:"propertyId" bson:"propertyId"`
	SensorId   string `json:"sensorId" bson:"sensorId"`
	Date       string `json:"date" bson:"date"`
	Type       string `json:"type" bson:"type"`
	Value      string `json:"value" bson:"value"`
}

type AppReportDBO struct {
	PropertyId string    `json:"propertyId" bson:"propertyId"`
	SensorId   string    `json:"sensorId" bson:"sensorId"`
	Date       time.Time `json:"date" bson:"date"`
	Type       string    `json:"type" bson:"type"`
	Value      string    `json:"value" bson:"value"`
}

func (report *AppReport) ToDBO() AppReportDBO {
	date, _ := time.Parse(time.DateOnly, report.Date)

	return AppReportDBO{
		PropertyId: report.PropertyId,
		SensorId:   report.SensorId,
		Date:       date,
		Type:       report.Type,
		Value:      report.Value,
	}
}

func (reportDBO *AppReportDBO) ToReport() AppReport {
	return AppReport{
		PropertyId: reportDBO.PropertyId,
		SensorId:   reportDBO.SensorId,
		Date:       reportDBO.Date.Format(time.DateOnly),
		Type:       reportDBO.Type,
		Value:      reportDBO.Value,
	}
}
