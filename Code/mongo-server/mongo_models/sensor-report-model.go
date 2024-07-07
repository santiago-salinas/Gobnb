package mongo_models

import "time"

type SensorTypeItem struct {
	Value string `json:"value" bson:"value"`
	Unit  string `json:"unit" bson:"unit"`
}

type SensorReport struct {
	SensorId string                    `json:"sensorId" bson:"sensorId"`
	Date     string                    `json:"date" bson:"date"`
	Reports  map[string]SensorTypeItem `json:"reports" bson:"reports"`
}

type SensorReportDBO struct {
	SensorId string                    `json:"sensorId" bson:"sensorId"`
	Date     time.Time                 `json:"date" bson:"date"`
	Reports  map[string]SensorTypeItem `json:"reports" bson:"reports"`
}

func (report *SensorReport) ToDBO() SensorReportDBO {
	date, _ := time.Parse(time.DateOnly, report.Date)

	return SensorReportDBO{
		SensorId: report.SensorId,
		Date:     date,
		Reports:  report.Reports,
	}
}

func (report *SensorReportDBO) ToObject() SensorReport {
	date := report.Date.String()

	return SensorReport{
		SensorId: report.SensorId,
		Date:     date,
		Reports:  report.Reports,
	}
}
