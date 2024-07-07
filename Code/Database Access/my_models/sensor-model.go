package my_models

import (
	"fmt"
	"mongo-server/mongo_models"
	"pocketbase_go/logger"
	"regexp"
	"strconv"
)

type Item struct {
	Type  string `json:"type" db:"type"`
	Unit  string `json:"unit" db:"unit"`
	Value string `json:"value" db:"value"`
	Min   string `json:"min" db:"min"`
	Max   string `json:"max" db:"max"`
}

type ReportStructure struct {
	MeasureStructures []Item `json:"reports" db:"reports"`
	SensorId          string `json:"sensorId" db:"sensorId"`
}

func (r *ReportStructure) ValidateSelf(sensorId string) error {
	if r.SensorId != sensorId {
		return fmt.Errorf("sensor id must match the sensor id in the report structure")
	}

	for _, report := range r.MeasureStructures {
		if (report.Min != "" || report.Max != "") && report.Value != "" {
			return fmt.Errorf("provide either min/max or value, not both")
		}

		if report.Min == "" && report.Max != "" || report.Min != "" && report.Max == "" {
			return fmt.Errorf("provide both min and max values")
		}

		if report.Min != "" && report.Max != "" {
			minValue, errMin := strconv.ParseFloat(report.Min, 64)
			maxValue, errMax := strconv.ParseFloat(report.Max, 64)
			if errMin != nil || errMax != nil {
				return fmt.Errorf("min and max must be valid numbers")
			}

			if minValue > maxValue {
				return fmt.Errorf("min value must be less than max value")
			}
		}
	}

	return nil
}

func (r *ReportStructure) ValidateReport(report mongo_models.SensorReport) (mongo_models.SensorReport, error) {
	for _, item := range r.MeasureStructures {
		reportMeasure, ok := report.Reports[item.Type]
		if !ok {
			logger.Error("missing report for type ", item.Type)
			return mongo_models.SensorReport{}, fmt.Errorf("missing report for type %s", item.Type)
		}
		if item.Unit != reportMeasure.Unit {
			logger.Error("invalid unit for type ", item.Type)
			return mongo_models.SensorReport{}, fmt.Errorf("invalid unit for type %s", item.Type)
		}
		if item.Value != "" {
			regex, err := regexp.Compile(item.Value)
			if err != nil {
				logger.Error("invalid regex for type %s", item.Type)
				return mongo_models.SensorReport{}, fmt.Errorf("invalid regex for type %s", item.Type)
			}
			if !regex.MatchString(reportMeasure.Value) {
				logger.Error("invalid value for type %s", item.Type)
				return mongo_models.SensorReport{}, fmt.Errorf("invalid value for type %s", item.Type)
			}
		} else {
			minValue, _ := strconv.ParseFloat(item.Min, 64)
			maxValue, _ := strconv.ParseFloat(item.Max, 64)
			reportValue, err := strconv.ParseFloat(reportMeasure.Value, 64)
			if err != nil {
				logger.Error("invalid value for type %s", item.Type)
				return mongo_models.SensorReport{}, fmt.Errorf("invalid value for type %s", item.Type)
			}
			if reportValue < minValue || reportValue > maxValue {
				logger.Error("value for type %s is out of range", item.Type)
				return mongo_models.SensorReport{}, fmt.Errorf("value for type %s is out of range", item.Type)
			}
		}
	}

	return report, nil
}

type Sensor struct {
	Id                  string          `json:"id" db:"id"`
	Description         string          `json:"description" db:"description"`
	SerialNumber        string          `json:"serialNumber" db:"serialNumber"`
	Brand               string          `json:"brand" db:"brand"`
	Address             string          `json:"address" db:"address"`
	LastMaintenanceDate string          `json:"lastMaintenanceDate" db:"lastMaintenanceDate"`
	ServiceType         string          `json:"serviceType" db:"serviceType"`
	AssignedTo          string          `json:"assignedTo" db:"assignedTo"`
	ReportStructure     ReportStructure `json:"reportStructure" db:"reportStructure"`
}

type SensorDBO struct {
	Id                  string          `json:"id" db:"id"`
	Description         string          `json:"description" db:"description"`
	SerialNumber        string          `json:"serialNumber" db:"serialNumber"`
	Brand               string          `json:"brand" db:"brand"`
	Address             string          `json:"address" db:"address"`
	LastMaintenanceDate string          `json:"lastMaintenanceDate" db:"lastMaintenanceDate"`
	ServiceType         string          `json:"serviceType" db:"serviceType"`
	AssignedTo          string          `json:"assignedTo" db:"assignedTo"`
	ReportStructure     ReportStructure `json:"reportStructure" db:"reportStructure"`
}

func (p *SensorDBO) ToObject() Sensor {
	return Sensor{
		Id:                  p.Id,
		Description:         p.Description,
		SerialNumber:        p.SerialNumber,
		Brand:               p.Brand,
		Address:             p.Address,
		LastMaintenanceDate: p.LastMaintenanceDate,
		ServiceType:         p.ServiceType,
		AssignedTo:          p.AssignedTo,
		ReportStructure:     p.ReportStructure,
	}
}

func (sensor *Sensor) ToMap() map[string]interface{} {
	ret := map[string]interface{}{
		"id":                  sensor.Id,
		"description":         sensor.Description,
		"serialNumber":        sensor.SerialNumber,
		"brand":               sensor.Brand,
		"address":             sensor.Address,
		"lastMaintenanceDate": sensor.LastMaintenanceDate,
		"serviceType":         sensor.ServiceType,
		"reportStructure":     sensor.ReportStructure,
	}

	return ret
}
