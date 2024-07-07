package interfaces

import (
	"pocketbase_go/my_models"
)

type ISensorService interface {
	AddSensor(sensor my_models.Sensor) error
	GetSensor(id string) (my_models.Sensor, error)
	AssignSensorToProperty(sensorId string, propertyId string) error
}
