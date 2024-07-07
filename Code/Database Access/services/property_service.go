package services

import (
	"fmt"
	"mime/multipart"
	"pocketbase_go/logger"
	"pocketbase_go/my_models"
	interfaces "pocketbase_go/repos/interfaces"
)

type PropertyService struct {
	Repo     interfaces.IPropertyRepo
	UserRepo interfaces.IUserRepo
}

func (r *PropertyService) AddUnavailableDates(propertyId string, dates []my_models.DateRange, userToken string) error {
	logger.Info("Service: Adding unavailable dates")
	roles, userId, err := r.UserRepo.Login(userToken)
	if err != nil {
		return err
	}

	property, err := r.Repo.GetPropertyById(propertyId)
	if err != nil {
		return err
	}

	if property.Owner != userId {
		logger.Error("Service: User is not the owner of the property")
		return fmt.Errorf("user is not authorized to add unavailable dates for this property")
	}

	for _, role := range roles {
		if role == "Owner" {
			return r.Repo.AddUnavailableDates(propertyId, dates)
		}
	}

	logger.Error("Service: User is not an owner")
	return fmt.Errorf("provided token does not belong to an owner user")
}

func (r *PropertyService) RemoveUnavailableDates(propertyId string, date my_models.DateRange, userToken string) error {
	logger.Info("Service: Removing unavailable dates")
	roles, userId, err := r.UserRepo.Login(userToken)
	if err != nil {
		return err
	}

	property, err := r.Repo.GetPropertyById(propertyId)
	if err != nil {
		return err
	}

	if property.Owner != userId {
		logger.Error("Service: User is not the owner of the property")
		return fmt.Errorf("user is not authorized to remove unavailable date for this property")
	}

	for _, role := range roles {
		if role == "Owner" {
			return r.Repo.RemoveUnavailableDate(propertyId, date)
		}
	}

	logger.Error("Service: User is not an owner")
	return fmt.Errorf("provided token does not belong to an owner user")
}

func (r *PropertyService) AddProperty(property my_models.Property, userToken string) (string, error) {
	logger.Info("Service: Adding property")
	roles, userId, err := r.UserRepo.Login(userToken)
	if err != nil {
		return "", err
	}

	for _, role := range roles {
		if role == "Owner" {
			property.Owner = userId
			property.IsPendingPayment = false
			return r.Repo.AddProperty(property)
		}
	}

	logger.Error("Service: User is not an admin")
	return "", fmt.Errorf("provided token does not belong to an Owner user")
}

func (r *PropertyService) AddPropertyImage(id string, image multipart.File, fileExtension string, userToken string) error {
	logger.Info("Service: Adding image to property with id: ", id)
	roles, _, err := r.UserRepo.Login(userToken)
	if err != nil {
		return err
	}

	for _, role := range roles {
		if role == "Owner" {
			return r.Repo.AddPropertyImage(id, image, fileExtension)
		}
	}

	logger.Error("Service: User is not an admin")
	return fmt.Errorf("provided token does not belong to an Owner user")
}

func (r *PropertyService) GetFilteredProperties(filter my_models.PropertyFilter) ([]my_models.Property, error) {
	logger.Info("Service: Getting filtered properties")
	hasFromDate := filter.DateFrom != nil
	hasUntilDate := filter.DateTo != nil
	if hasFromDate != hasUntilDate {
		logger.Error("Service: Provide full date range or do not provide any dates")
		return nil, fmt.Errorf("both from and until dates must be provided")
	}

	properties, err := r.Repo.GetFilteredProperties(filter)
	if err != nil {
		return nil, err
	}

	logger.Info("Service: Got filtered properties succesfully")
	return properties, nil
}
