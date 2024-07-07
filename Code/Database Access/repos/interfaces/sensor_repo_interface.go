package repointerfaces

import (
	"pocketbase_go/my_models"
)

type ISensorRepo interface {
	AddSensor(sensor my_models.Sensor) error
	GetSensor(id string) (my_models.Sensor, error)
	AssignSensorToProperty(sensorId string, propertyId string) error
}
