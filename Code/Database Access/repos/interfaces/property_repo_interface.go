package repointerfaces

import (
	"mime/multipart"
	"pocketbase_go/my_models"
)

type IPropertyRepo interface {
	AddProperty(property my_models.Property) (string, error)
	GetPropertyById(id string) (my_models.Property, error)
	GetFilteredProperties(filter my_models.PropertyFilter) ([]my_models.Property, error)
	GetUnavailableDates(propertyId string) ([]my_models.DateRange, error)
	GetPropertyImages(propertyId string) ([]string, error)
	GetAllProperties() ([]my_models.Property, error)
	GetOccupiedProperties(fromDate string, untilDate string) ([]my_models.Property, error)
	AddUnavailableDates(propertyId string, dates []my_models.DateRange) error
	RemoveUnavailableDate(propertyId string, date my_models.DateRange) error
	UpdatePropertyPaidStatus(id string) error
	UpdatePropertyPendingPaymentStatus(id string, status bool) error
	AddPropertyImage(id string, image multipart.File, fileExtension string) error
}
