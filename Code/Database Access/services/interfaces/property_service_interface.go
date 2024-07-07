package interfaces

import (
	"mime/multipart"
	"pocketbase_go/my_models"
)

type IPropertyService interface {
	AddUnavailableDates(propertyId string, dates []my_models.DateRange, userToken string) error
	RemoveUnavailableDates(propertyId string, date my_models.DateRange, userToken string) error
	AddProperty(property my_models.Property, userToken string) (string, error)
	AddPropertyImage(id string, image multipart.File, fileExtension string, userToken string) error
	GetFilteredProperties(filter my_models.PropertyFilter) ([]my_models.Property, error)
}
