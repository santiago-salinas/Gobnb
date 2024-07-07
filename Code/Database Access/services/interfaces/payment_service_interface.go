package interfaces

import (
	"pocketbase_go/my_models"
)

type IPaymentService interface {
	PayProperty(propertyId string, cardInformation my_models.CardInformation) error
	PayReservation(reservationId string, cardInformation my_models.CardInformation) error
}
