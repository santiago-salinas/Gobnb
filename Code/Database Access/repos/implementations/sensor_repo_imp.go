package repositories

import (
	"fmt"
	"pocketbase_go/logger"
	"pocketbase_go/my_models"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/forms"
	"github.com/pocketbase/pocketbase/models"

	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

const (
	sensorsCollection = "sensors"
)

type PocketSensorRepo struct {
	Db    core.App
	Cache *redis.Client
}

func (r *PocketSensorRepo) AddSensor(sensor my_models.Sensor) error {
	logger.Info("Repo: Adding sensor")
	collection, err := r.Db.Dao().FindCollectionByNameOrId(sensorsCollection)
	if err != nil {
		logger.Error("Repo: ", err)
		return err
	}

	record := models.NewRecord(collection)
	sensorId := sensor.Id
	sensorLessThanFifteen := len(sensorId) < 15
	if sensorLessThanFifteen {
		sensor.Id = sensorId
		record.SetId(sensorId)
		if err := r.Db.Dao().SaveRecord(record); err != nil {
			logger.Error("Repo: ", err)
			return err
		}
	}
	sensor.Id = ""
	record.MarkAsNotNew()
	form := forms.NewRecordUpsert(r.Db, record)
	form.LoadData(sensor.ToMap())
	if err := form.Submit(); err != nil {
		r.Db.Dao().DeleteRecord(record)
		logger.Error("Repo: ", err)
		return err
	}

	fmt.Printf("Record: %v\n", record)

	logger.Info("Repo: Sensor added successfully")
	return nil
}

func (r *PocketSensorRepo) GetSensor(id string) (my_models.Sensor, error) {
	logger.Info("Repo: Getting sensor")
	if r.Cache != nil {
		val, err := r.Cache.Get(ctx, id).Result()
		if err == redis.Nil || err != nil {
			model, err := r.getSensorFromDB(id)
			if err != nil {
				logger.Error("Repo: ", err)
				return my_models.Sensor{}, err
			}

			err = r.storeSensorInCache(id, model)
			if err != nil {
				logger.Warn("Could not store sensor in cache: ", err)
			}

			return model, nil
		} else {
			var sensor my_models.Sensor

			err := json.Unmarshal([]byte(val), &sensor)
			if err != nil {
				logger.Error("Repo: ", err)
				return my_models.Sensor{}, err
			}

			logger.Info("Repo: Sensor retrieved from redis succesfully")
			return sensor, nil
		}
	} else {
		return r.getSensorFromDB(id)
	}
}

func (r *PocketSensorRepo) getSensorFromDB(id string) (my_models.Sensor, error) {
	query := fmt.Sprintf(`
        SELECT *
        FROM sensors
        WHERE id = '%s'`, id)

	var sensor my_models.SensorDBO
	err := r.Db.Dao().DB().NewQuery(query).One(&sensor)
	if err != nil && err.Error() == "sql: no rows in result set" {
		logger.Error("Repo: Sensor not found")
		return my_models.Sensor{}, fmt.Errorf("Repo: Sensor not found")
	}

	if err != nil {
		logger.Error("Repo: ", err)
		return my_models.Sensor{}, err
	}
	record, err := r.Db.Dao().FindRecordById(sensorsCollection, id)
	if err != nil {
		logger.Error("Repo: ", err)
		return my_models.Sensor{}, err
	}
	record.UnmarshalJSONField("reportStructure", &sensor.ReportStructure)

	sensorObject := sensor.ToObject()

	logger.Info("Repo: Sensor retrieved successfully from pocketbase")
	return sensorObject, nil
}

func (r *PocketSensorRepo) storeSensorInCache(id string, sensorObject my_models.Sensor) error {
	ttlInMinutes := 2 * time.Minute

	sensorJSON, err := json.Marshal(sensorObject)
	if err != nil {
		logger.Error("Repo: ", err)
		return err
	}

	if r.Cache != nil {
		err = r.Cache.Set(ctx, id, sensorJSON, ttlInMinutes).Err()
		if err != nil {
			logger.Error("Repo: ", err)
			return err
		}
	}
	return nil
}

func (r *PocketSensorRepo) AssignSensorToProperty(sensorId string, propertyId string) error {
	logger.Info("Repo: Assigning sensor to property")

	query := fmt.Sprintf(`
        UPDATE sensors
        SET assignedTo = '%s'
        WHERE id = '%s'`, propertyId, sensorId)

	_, err := r.Db.Dao().DB().NewQuery(query).Execute()
	if err != nil {
		logger.Error("Repo: ", err)
		return err
	}

	logger.Info("Repo: Sensor assigned to property successfully")
	return nil
}
