package services

import (
	"fmt"
	"pocketbase_go/logger"
	"pocketbase_go/my_models"
	interfaces "pocketbase_go/repos/interfaces"
)

type SensorService struct {
	Repo interfaces.ISensorRepo
}

func (s *SensorService) AddSensor(sensor my_models.Sensor) error {
	logger.Info("Service: Adding sensor")
	if len(sensor.Id) > 15 {
		logger.Error("sensor ID must be less than 15 characters")
		return fmt.Errorf("sensor ID must be less than 15 characters")
	}

	err := sensor.ReportStructure.ValidateSelf(sensor.Id)
	if err != nil {
		logger.Error(err)
		return err
	}

	error := s.Repo.AddSensor(sensor)

	if error == nil {
		logger.Info("Service: Sensor added successfully")
	}
	return error
}

func (s *SensorService) GetSensor(id string) (my_models.Sensor, error) {
	logger.Info("Service: Getting sensor with id: ", id)
	sensor, err := s.Repo.GetSensor(id)
	if err != nil {
		return my_models.Sensor{}, err
	}

	logger.Info("Service: Got sensor successfully")
	return sensor, nil
}

func (s *SensorService) AssignSensorToProperty(sensorId string, propertyId string) error {
	logger.Info("Service: Assigning sensor with id: ", sensorId, "to property with id: ", propertyId)
	err := s.Repo.AssignSensorToProperty(sensorId, propertyId)
	if err != nil {
		return err
	}
	logger.Info("Service: Sensor assigned to property successfully")
	return nil
}
