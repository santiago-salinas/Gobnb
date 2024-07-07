package interfaces

import (
	"pocketbase_go/my_models"
)

type IReservationService interface {
	CreateReservation(reservation my_models.ReservationModel) error
	GetFilteredReservations(filter my_models.ReservationFilter) ([]my_models.ReservationModel, error)
	NotifyValidReservation(reservation my_models.ReservationModel, ownerEmail string) error
	GetOwnReservation(email string, propertyId string) (my_models.ReservationModel, error)
	ApproveReservation(reservationId string) error
	RemoveReservation(reservationId string) error
	CancelReservation(email string, reservationId string) (refundPercentage float64, err error)
	DoCheckIn(reservationId string) error
	DoCheckOut(reservationId string) error
	GetReservationById(reservationId string) (my_models.ReservationModel, error)
	AutoCancelReservations() error
}
