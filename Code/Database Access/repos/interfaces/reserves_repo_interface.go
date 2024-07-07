package repointerfaces

import (
	"pocketbase_go/my_models"
)

type IReservationRepo interface {
	CreateReservation(reservation my_models.ReservationModel) error
	ApproveReservation(reservationId string) error
	GetFilteredReservations(filter my_models.ReservationFilter) ([]my_models.ReservationModel, error)
	GetOwnReservation(email string, propertyId string) (my_models.ReservationModel, error)
	CancelReservation(reservationId string) error
	RemoveReservation(reservationId string) error
	GetReservationById(reservationId string) (my_models.ReservationModel, error)
	RegisterCheckIn(reservationId string) error
	RegisterCheckOut(reservationId string) error
	UpdateReservationStatus(id string, status string) error
	AutoCancelReservations(autoCancelDays int) ([]string, error)
}