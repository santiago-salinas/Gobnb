package repositories

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"pocketbase_go/logger"
	"pocketbase_go/my_models"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/forms"
	"github.com/pocketbase/pocketbase/models"
	ffmpeg "github.com/u2takey/ffmpeg-go"

	"encoding/json"

	"github.com/go-redis/redis/v8"
)

const (
	propertiesCollection = "properties"
)

type PocketPropertyRepo struct {
	Db                     pocketbase.PocketBase
	Cache                  *redis.Client
	imagesUrl              string
	imagesDir              string
	imagesCompressionScale string
}

func (r *PocketPropertyRepo) SetConfigValues(url, dir, compresisonScale string) {
	r.imagesUrl = url
	r.imagesDir = dir
	r.imagesCompressionScale = compresisonScale
}

func (r *PocketPropertyRepo) AddProperty(property my_models.Property) (string, error) {
	logger.Info("Repo: Adding property")

	boolean := property.HasAC
	if boolean != "true" && boolean != "false" && boolean != "0" && boolean != "1" {
		logger.Error("Repo: Invalid hasAC value, valid values are true, false, 0, 1")
		return "", fmt.Errorf("invalid hasAC value: %s, valid values are true, false, 0, 1", boolean)
	}

	boolean = property.HasWIFI
	if boolean != "true" && boolean != "false" && boolean != "0" && boolean != "1" {
		logger.Error("Repo: Invalid hasAC value, valid values are true, false, 0, 1")
		return "", fmt.Errorf("invalid hasAC value: %s, valid values are true, false, 0, 1", boolean)
	}

	boolean = property.HasGarage
	if boolean != "true" && boolean != "false" && boolean != "0" && boolean != "1" {
		logger.Error("Repo: Invalid hasAC value, valid values are true, false, 0, 1")
		return "", fmt.Errorf("invalid hasAC value: %s, valid values are true, false, 0, 1", boolean)
	}

	for _, date := range property.UnavailableDates {
		if err := validateDate(date.Start); err != nil {
			logger.Error("Repo: ", err)
			return "", fmt.Errorf("invalid start date %v: %v", date.Start, err)
		}
		if err := validateDate(date.End); err != nil {
			logger.Error("Repo: ", err)
			return "", fmt.Errorf("invalid end date %v: %v", date.End, err)
		}
	}

	// Find the properties collection
	collection, err := r.Db.Dao().FindCollectionByNameOrId(propertiesCollection)
	if err != nil {
		logger.Error("Repo: ", err)
		return "", err
	}

	// Convert property to a map
	propertyMap := property.ToMap()
	unavailableDates := property.UnavailableDates

	// Create a new record for the property
	record := models.NewRecord(collection)
	form := forms.NewRecordUpsert(r.Db, record)
	form.LoadData(propertyMap)

	// Submit the form to save the property record
	if err := form.Submit(); err != nil {
		logger.Error("Repo: ", err)
		return "", err
	}

	// Capture the assigned ID from the record
	property.Id = record.GetId()

	// Find the unavailable dates collection
	collection, err = r.Db.Dao().FindCollectionByNameOrId("unavailableDates")
	if err != nil {
		logger.Error("Repo: ", err)
		return "", err
	}

	// Add new records for each unavailable date
	for _, date := range unavailableDates {
		record := models.NewRecord(collection)
		form := forms.NewRecordUpsert(r.Db, record)
		form.LoadData(date.ToMap(property.Id))
		if err := form.Submit(); err != nil {
			logger.Error("Repo: ", err)
			return "", err
		}
	}

	logger.Info("Record saved with ID: ", property.Id)
	fmt.Printf("Record saved with ID: %s", property.Id)
	return property.Id, nil
}

func (r *PocketPropertyRepo) GetPropertyById(id string) (my_models.Property, error) {
	logger.Info("Repo: Getting property by id")

	if r.Cache != nil {
		val, err := r.Cache.Get(ctx, id).Result()
		if err == redis.Nil || err != nil {
			model, err := r.getPropertyFromDB(id)
			if err != nil {
				return my_models.Property{}, err
			}

			err = r.storePropertyInCache(id, model)
			if err != nil {
				logger.Warn("Could not store property in cache: ", err)
			}

			return model, nil
		} else {
			var property my_models.Property

			err := json.Unmarshal([]byte(val), &property)
			if err != nil {
				return my_models.Property{}, err
			}

			return property, nil
		}
	} else {
		return r.getPropertyFromDB(id)
	}
}

func (r *PocketPropertyRepo) getPropertyFromDB(id string) (my_models.Property, error) {
	query := fmt.Sprintf(`
		SELECT *
		FROM '%s'	
		WHERE id = '%s'`, propertiesCollection, id)

	var property my_models.PropertyDBO

	err := r.Db.Dao().DB().NewQuery(query).One(&property)
	if err != nil && err.Error() == "sql: no rows in result set" {
		logger.Error("Repo: property with provided id not found")
		return my_models.Property{}, errors.New("property with provided id not found")
	}
	if err != nil {
		logger.Error("Repo: ", err)
		return my_models.Property{}, err
	}

	unavailableDates, err := r.GetUnavailableDates(property.Id)
	if err != nil {
		logger.Error("Repo: ", err)
		return my_models.Property{}, err
	}

	imagesPaths, err := r.GetPropertyImages(property.Id)
	if err != nil {
		logger.Error("Repo: ", err)
		return my_models.Property{}, err
	}

	propertyObject := property.ToObject(unavailableDates, imagesPaths)
	logger.Info("Repo: Property retrieved successfully from pocketbase")
	return propertyObject, nil
}

func (r *PocketPropertyRepo) storePropertyInCache(id string, propertyObject my_models.Property) error {
	ttlInMinutes := 2 * time.Minute

	propertyJSON, err := json.Marshal(propertyObject)
	if err != nil {
		logger.Error("Repo: ", err)
		return err
	}
	if r.Cache != nil {
		err = r.Cache.Set(ctx, id, propertyJSON, ttlInMinutes).Err()
		if err != nil {
			logger.Error("Repo: ", err)
			return err
		}
	}
	logger.Info("Repo: Property stored successfully in redis")
	return nil
}

func (r *PocketPropertyRepo) GetFilteredProperties(filter my_models.PropertyFilter) ([]my_models.Property, error) {
	logger.Info("Repo: Getting filtered properties")

	query := (`
		SELECT *
		FROM properties
		WHERE 1=1
		`)

	if filter.AdultQuantityMax != nil {
		query += fmt.Sprintf(` AND AdultQuantity <= %d`, *filter.AdultQuantityMax)
	}

	if filter.AdultQuantityMin != nil {
		query += fmt.Sprintf(` AND AdultQuantity >= %d`, *filter.AdultQuantityMin)
	}

	if filter.KidQuantityMax != nil {
		query += fmt.Sprintf(`AND KidQuantity <= %d`, *filter.KidQuantityMax)
	}

	if filter.KidQuantityMin != nil {
		query += fmt.Sprintf(` AND KidQuantity >= %d`, *filter.KidQuantityMin)
	}

	if filter.KingSizedBedsMax != nil {
		query += fmt.Sprintf(` AND KingSizedBeds <= %d`, *filter.KingSizedBedsMax)
	}

	if filter.KingSizedBedsMin != nil {
		query += fmt.Sprintf(` AND KingSizedBeds >= %d`, *filter.KingSizedBedsMin)
	}

	if filter.SingleBedsMax != nil {
		query += fmt.Sprintf(` AND SingleBeds <= %d`, *filter.SingleBedsMax)
	}

	if filter.SingleBedsMin != nil {
		query += fmt.Sprintf(` AND SingleBeds >= %d`, *filter.SingleBedsMin)
	}

	if filter.HasAC != nil {
		query += fmt.Sprintf(` AND HasAC = %t`, *filter.HasAC)
	}

	if filter.HasWIFI != nil {
		query += fmt.Sprintf(`AND HasWIFI = %t`, *filter.HasWIFI)
	}

	if filter.HasGarage != nil {
		query += fmt.Sprintf(` AND HasGarage = %t`, *filter.HasGarage)
	}

	if filter.Type != nil {
		query += fmt.Sprintf(` AND Type = %d`, *filter.Type)
	}

	if filter.BeachDistanceMax != nil {
		query += fmt.Sprintf(` AND BeachDistance <= %d`, *filter.BeachDistanceMax)
	}

	if filter.BeachDistanceMin != nil {
		query += fmt.Sprintf(` AND BeachDistance >= %d`, *filter.BeachDistanceMin)
	}

	if filter.State != nil {
		query += fmt.Sprintf(` AND State = %s`, *filter.State)
	}

	if filter.Resort != nil {
		query += fmt.Sprintf(` AND Resort = %s`, *filter.Resort)
	}

	if filter.Neighborhood != nil {
		query += fmt.Sprintf(` AND Neighborhood = %s`, *filter.Neighborhood)
	}

	query += " AND paid = true"

	var startDate string
	var endDate string

	if filter.DateFrom == nil && filter.DateTo == nil {
		startDate = time.Now().Format(time.DateOnly)
		endDate = time.Now().AddDate(0, 0, 30).Format(time.DateOnly)
	} else if filter.DateFrom != nil && filter.DateTo != nil {
		startDate = *filter.DateFrom
		endDate = *filter.DateTo
	}

	if startDate != "" && endDate != "" {
		query += " AND id NOT IN ("
		query += fmt.Sprintf(`
			SELECT property
			FROM reservations
			WHERE (status = 'Approved' OR status = 'Paid')
			AND NOT (
				reserved_until <= '%s'
				OR reserved_from >= '%s'
			)
		`, startDate, endDate)
		query += ")"

		query += " AND id NOT IN ("
		query += fmt.Sprintf(`
			SELECT propertyId
			FROM unavailableDates
			WHERE NOT (
				dateTo <= '%s'
				OR dateFrom >= '%s'
			)
		`, startDate, endDate)
		query += ")"
	}

	quantity := *filter.Size
	offset := ((*filter.Page - 1) * quantity)
	query += fmt.Sprintf(` LIMIT %d OFFSET %d`, quantity, offset)

	var properties []my_models.PropertyDBO
	err := r.Db.Dao().DB().NewQuery(query).All(&properties)
	logger.Info("Query: ", query, properties)
	if err != nil {
		logger.Error("Repo: ", err)
		return nil, err
	}

	var newProperties []my_models.Property
	for _, value := range properties {
		unavailableDates, err := r.GetUnavailableDates(value.Id)
		if err != nil {
			logger.Error("Repo: ", err)
			return nil, err
		}
		imagesPaths, err := r.GetPropertyImages(value.Id)
		if err != nil {
			logger.Error("Repo: ", err)
			return nil, err
		}
		property := value.ToObject(unavailableDates, imagesPaths)
		newProperties = append(newProperties, property)
	}

	logger.Info("Repo: Got filtered properties succesfully")
	return newProperties, nil
}

func (r *PocketPropertyRepo) GetUnavailableDates(propertyId string) ([]my_models.DateRange, error) {
	logger.Info("Repo: Getting unavailable dates")

	query := fmt.Sprintf("SELECT dateFrom, dateTo FROM unavailableDates WHERE propertyId = '%s'", propertyId)

	var unavailableDatesDBOs []my_models.UnavailableDatesDBO
	err := r.Db.Dao().DB().NewQuery(query).All(&unavailableDatesDBOs)
	if err != nil {
		logger.Error("Repo: ", err)
		return nil, err
	}

	var unavailableDates []my_models.DateRange
	for _, dbo := range unavailableDatesDBOs {
		unavailableDates = append(unavailableDates, dbo.ToObject())
	}

	logger.Info("Repo: Got unavailable dates succesfully")
	return unavailableDates, nil
}

func (r *PocketPropertyRepo) GetPropertyImages(propertyId string) ([]string, error) {
	logger.Info("Repo: Getting property images")
	query := fmt.Sprintf("SELECT fileName FROM images WHERE propertyId = '%s'", propertyId)

	var imagesDBOs []my_models.ImagesDBO
	err := r.Db.Dao().DB().NewQuery(query).All(&imagesDBOs)
	if err != nil {
		logger.Error("Repo: ", err)
		return nil, err
	}
	var imagesPaths []string
	for _, dbo := range imagesDBOs {
		url := fmt.Sprintf(r.imagesUrl, dbo.FileName)
		imagesPaths = append(imagesPaths, url)
	}

	logger.Info("Repo: Got property images succesfully")
	return imagesPaths, nil
}

func (r *PocketPropertyRepo) GetAllProperties() ([]my_models.Property, error) {
	logger.Info("Repo: Getting all properties")
	query := fmt.Sprintf("SELECT * FROM %s", propertiesCollection)

	var properties []my_models.PropertyDBO
	err := r.Db.Dao().DB().NewQuery(query).All(&properties)
	if err != nil {
		logger.Error("Repo: ", err)
		return nil, err
	}

	var newProperties []my_models.Property
	for _, value := range properties {
		property := value.ToObject([]my_models.DateRange{}, []string{})
		newProperties = append(newProperties, property)
	}

	logger.Info("Repo: Got all properties succesfully")
	return newProperties, nil
}

func (r *PocketPropertyRepo) GetOccupiedProperties(fromDate string, untilDate string) ([]my_models.Property, error) {
	logger.Info("Repo: Getting occupied properties")
	query := fmt.Sprintf(`
		SELECT *
		FROM properties
		WHERE id IN (
			SELECT property
			FROM reservations
			WHERE NOT (
				reserved_until <= '%s'
				OR reserved_from >= '%s'
			)
		)
	`, fromDate, untilDate)

	var properties []my_models.PropertyDBO
	err := r.Db.Dao().DB().NewQuery(query).All(&properties)
	if err != nil {
		logger.Error("Repo: ", err)
		return nil, err
	}

	var newProperties []my_models.Property
	for _, value := range properties {
		unavailableDates, err := r.GetUnavailableDates(value.Id)
		if err != nil {
			return nil, err
		}
		imagesPaths, err := r.GetPropertyImages(value.Id)
		if err != nil {
			return nil, err
		}
		property := value.ToObject(unavailableDates, imagesPaths)
		newProperties = append(newProperties, property)
	}

	logger.Info("Repo: Got occupied properties succesfully")
	return newProperties, nil
}

func validateDate(dateStr string) error {
	_, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return errors.New("invalid date format, must be YYYY-MM-DD")
	}
	return nil
}

func (r *PocketPropertyRepo) AddUnavailableDates(propertyId string, dates []my_models.DateRange) error {
	logger.Info("Repo: Adding unavailable dates for property ID: %d", propertyId)

	// Find the properties collection
	collection, err := r.Db.Dao().FindCollectionByNameOrId(propertiesCollection)
	if err != nil {
		logger.Error("Repo: %s", err)
		return err
	}

	// Check if there is a property with that ID
	_, err = r.Db.Dao().FindFirstRecordByData(collection.Name, "id", propertyId)
	if err != nil {
		logger.Error("There is no property with that id to add unavailable dates. Id: %d", propertyId)
		return fmt.Errorf("no property found with id %d: %v", propertyId, err)
	}

	// Find the unavailable dates collection
	collection, err = r.Db.Dao().FindCollectionByNameOrId("unavailableDates")
	if err != nil {
		logger.Error("Repo: %s", err)
		return err
	}

	// Add new records for each unavailable date range
	for _, date := range dates {
		// Validate start and end dates
		if err := validateDate(date.Start); err != nil {
			logger.Error("Repo: invalid start date %v: %v", date.Start, err)
			return fmt.Errorf("invalid start date %v: %v", date.Start, err)
		}
		if err := validateDate(date.End); err != nil {
			logger.Error("Repo: invalid end date %v: %v", date.End, err)
			return fmt.Errorf("invalid end date %v: %v", date.End, err)
		}

		// Create a new record for the date range
		record := models.NewRecord(collection)
		form := forms.NewRecordUpsert(r.Db, record)
		form.LoadData(date.ToMap(propertyId))

		// Submit the form to save the date range record
		if err := form.Submit(); err != nil {
			logger.Error("Repo: %s", err)
			return err
		}

	}

	logger.Info("Unavailable dates added for property ID: %d", propertyId)
	return nil
}

func (r *PocketPropertyRepo) RemoveUnavailableDate(propertyId string, date my_models.DateRange) error {
	logger.Info("Repo: Removing unavailable dates for property ID:", propertyId)

	propertiesCollection := "properties"
	collection, err := r.Db.Dao().FindCollectionByNameOrId(propertiesCollection)
	if err != nil {
		logger.Error("Repo: Error finding properties collection", err)
		return err
	}

	_, err = r.Db.Dao().FindFirstRecordByData(collection.Name, "id", propertyId)
	if err != nil {
		logger.Error("Repo: There is no property with that id. Id: ", propertyId)
		return fmt.Errorf("no property found with id %s: %v", propertyId, err)
	}

	unavailableDatesCollection := "unavailableDates"
	collection, err = r.Db.Dao().FindCollectionByNameOrId(unavailableDatesCollection)
	if err != nil {
		logger.Error("Repo: Error finding unavailable dates collection - ", err)
		return err
	}

	query := fmt.Sprintf("SELECT * FROM %s WHERE propertyId = '%s'", unavailableDatesCollection, propertyId)
	var records []my_models.UnavailableDatesDBO
	err = r.Db.Dao().DB().NewQuery(query).All(&records)
	if err != nil {
		logger.Error("Repo: Error querying unavailable dates - %s", err)
		return err
	}
	if len(records) == 0 {
		logger.Error("Repo: No records found with that property id")
		return errors.New("no records found with that property id")
	}

	atLeastOneDeletion := false
	for _, dates := range records {
		logger.Debug(dates.DateFrom)
		if err := validateDate(date.Start); err != nil {
			logger.Error("Repo: ", err)
			return fmt.Errorf("invalid start date %v: %v", date.Start, err)
		}
		if err := validateDate(date.End); err != nil {
			logger.Error("Repo: ", err)
			return fmt.Errorf("invalid end date %v: %v", date.End, err)
		}
		if dates.DateFrom != date.Start+" 00:00:00.000Z" || dates.DateTo != date.End+" 00:00:00.000Z" {
			continue
		}
		record, err := r.Db.Dao().FindRecordById(unavailableDatesCollection, dates.Id)
		if err != nil {
			logger.Error("Repo: Failed to find record - %s", err)
			return err
		}
		err = r.Db.Dao().DeleteRecord(record)
		if err != nil {
			logger.Error("Repo: Failed to delete record - %s", err)
			return err
		}
	}
	if atLeastOneDeletion == false {
		logger.Error("Repo: No records found with that date range")
		return errors.New("no records found with that date range")
	}
	logger.Info("Repo: Successfully deleted records for property ID:", propertyId)
	return nil
}

func (r *PocketPropertyRepo) UpdatePropertyPaidStatus(id string) error {
	logger.Info("Repo: Updating property paid status")
	record, err := r.Db.Dao().FindRecordById("properties", id)
	if err != nil {
		logger.Error("Repo: ", err)
		return err
	}
	//Only update if property is pending payment
	if !record.GetBool("isPendingPayment") {
		logger.Warn("Repo: Property is not pending payment")
		return nil
	}

	record.Set("paid", true)
	record.Set("isPendingPayment", false)

	if err := r.Db.Dao().SaveRecord(record); err != nil {
		logger.Error("Repo: ", err)
		return err
	}

	_, err = r.Cache.Del(ctx, id).Result()
	if err != nil {
		logger.Error("Repo: ", err)
		return err
	}

	logger.Info("Repo: Property paid status updated successfully")
	return nil
}

func (r *PocketPropertyRepo) UpdatePropertyPendingPaymentStatus(id string, status bool) error {
	logger.Info("Repo: Updating property pending payment status")
	record, err := r.Db.Dao().FindRecordById("properties", id)
	if err != nil {
		logger.Error("Repo: ", err)
		return err
	}

	record.Set("isPendingPayment", status)

	if err := r.Db.Dao().SaveRecord(record); err != nil {
		logger.Error("Repo: ", err)
		return err
	}

	_, err = r.Cache.Del(ctx, id).Result()
	if err != nil {
		logger.Error("Repo: Error deleting property from cache:", err)
		return err
	}

	logger.Info("Repo: Property pending payment status updated successfully")
	return nil
}

func (r *PocketPropertyRepo) AddPropertyImage(id string, image multipart.File, fileExtension string) error {
	logger.Info("Repo: Adding image to property with id: ", id)
	collection, err := r.Db.Dao().FindCollectionByNameOrId("images")
	if err != nil {
		logger.Error("Repo: Error finding images collection:", err)
		return err
	}

	// Create a new record in the images collection
	record := models.NewRecord(collection)
	form := forms.NewRecordUpsert(r.Db, record)

	// Ensure the "images" directory exists
	if err := os.MkdirAll(r.imagesDir, os.ModePerm); err != nil {
		logger.Error("Repo: Error creating images directory:", err)
		return err
	}

	// Generate a random name for the temporary file
	randomFileName := fmt.Sprintf("%d%s.tmp", time.Now().UnixNano(), fileExtension)
	tempFilePath := filepath.Join(r.imagesDir, randomFileName)
	tempFile, err := os.Create(tempFilePath)
	if err != nil {
		logger.Error("Repo: Error creating temporary image file:", err)
		return err
	}
	defer tempFile.Close()

	// Copy the image data to the temporary file
	_, err = io.Copy(tempFile, image)
	if err != nil {
		logger.Error("Repo: Error copying image data to temporary file: ", err)
		return err
	}

	// Close the temporary file before processing with ffmpeg
	tempFile.Close()

	// Generate a random name for the final file
	finalFileName := fmt.Sprintf("%d%s", time.Now().UnixNano(), fileExtension)
	finalFilePath := filepath.Join(r.imagesDir, finalFileName)

	// Run ffmpeg to resize the image from the temporary file to the final file
	err = ffmpeg.Input(tempFilePath).
		Output(finalFilePath, ffmpeg.KwArgs{"vf": r.imagesCompressionScale}).
		OverWriteOutput().ErrorToStdOut().Run()
	if err != nil {
		logger.Error("Repo: Error resizing image:", err)
		os.Remove(tempFilePath) // Remove the temporary file
		return err
	}

	// Remove the temporary file
	os.Remove(tempFilePath)

	// Load data into the form
	form.LoadData(map[string]interface{}{
		"propertyId": id,
		"fileName":   finalFileName, // Provide the relative path to the saved image file
	})

	// Submit the form
	if err := form.Submit(); err != nil {
		logger.Error("Repo: Error submitting form:", err)
		os.Remove(finalFilePath) // Remove the final file
		return err
	}

	query := fmt.Sprintf("SELECT * FROM images WHERE propertyId = '%s'", id)
	var imagesDBO []my_models.ImagesDBO
	err = r.Db.Dao().DB().NewQuery(query).All(&imagesDBO)
	if err != nil {
		logger.Error("Repo: ", err)
		os.Remove(finalFilePath) // Remove the final file
		return err
	}

	// If there are more than 4 images, set the property as paid
	if len(imagesDBO) >= 4 {
		err = r.UpdatePropertyPendingPaymentStatus(id, true)
		if err != nil {
			logger.Error("Repo: ", err)
			os.Remove(finalFilePath) // Remove the final file
			return fmt.Errorf("error updating property pending payment status: %w", err)
		}
	}

	logger.Info("Repo: Image added to property with id: ", id)
	return nil
}
