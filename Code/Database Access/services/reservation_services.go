package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"pocketbase_go/logger"
	"pocketbase_go/my_models"
	interfaces "pocketbase_go/repos/interfaces"
	"time"
)

const autoCancelDays = 3

type ReservationService struct {
	ReservationRepo interfaces.IReservationRepo
	UserRepo        interfaces.IUserRepo
	SettingsRepo    interfaces.ISettingsRepo
	PropertiesRepo  interfaces.IPropertyRepo
	refundUrl       string
}

func (s *ReservationService) SetConfigValues(refundUrl string) {
	s.refundUrl = refundUrl
}

func (s *ReservationService) CreateReservation(reservation my_models.ReservationModel) error {
	if err := reservation.ValidateFields(); err != nil {
		return err
	}

	property, err := s.PropertiesRepo.GetPropertyById(reservation.PropertyId)
	if err != nil {
		return err
	}
	if property.IsPendingPayment {
		return fmt.Errorf("property %s is pending payment, reservation cannot be made", reservation.PropertyId)
	}

	if err := s.ReservationRepo.CreateReservation(reservation); err != nil {
		return err
	}

	ownerEmail, err := s.UserRepo.GetPropertyOwner(reservation.PropertyId)
	if err != nil {
		return err
	}

	if err := s.NotifyValidReservation(reservation, ownerEmail); err != nil {
		return err
	}

	return nil
}

func (s *ReservationService) GetFilteredReservations(filter my_models.ReservationFilter) ([]my_models.ReservationModel, error) {
	return s.ReservationRepo.GetFilteredReservations(filter)
}

func (s *ReservationService) NotifyValidReservation(reservation my_models.ReservationModel, ownerEmail string) error {
	admins, err := s.UserRepo.GetUsersByRole("Admin")
	if err != nil {
		return err
	}

	// Send email to admins
	for _, admin := range admins {
		logger.Info("Notification: Sending email to Admin ", admin, "about reservation ", reservation.ID)
	}

	// Send email to owner
	logger.Info("Notification: Sending email to Owner ", ownerEmail, " about reservation ", reservation.ID)

	return nil
}

func (s *ReservationService) GetOwnReservation(email string, propertyId string) (my_models.ReservationModel, error) {
	return s.ReservationRepo.GetOwnReservation(email, propertyId)
}

func (s *ReservationService) ApproveReservation(reservationId string) error {
	return s.ReservationRepo.ApproveReservation(reservationId)
}

func (s *ReservationService) RemoveReservation(reservationId string) error {
	return s.ReservationRepo.RemoveReservation(reservationId)
}

func (s *ReservationService) CancelReservation(email string, reservationId string) (refundPercentage float64, err error) {
	reservation, err := s.ReservationRepo.GetReservationById(reservationId)
	if err != nil {
		return 0, err
	}

	if reservation.Email != email {
		return 0, fmt.Errorf("user %s is not allowed to cancel reservation %s", email, reservationId)
	}

	cancellationDaysLimit, err := s.SettingsRepo.GetCancellationDays(reservation.Country)
	if err != nil {
		return 0, err
	}
	reservationStartDate, err := time.Parse(my_models.PocketTimeLayout, reservation.ReservedFrom)
	if err != nil {
		return 0, err
	}

	timeDifference := time.Until(reservationStartDate)

	if timeDifference <= 0 {
		return 0, fmt.Errorf("reservation %s starting date already passed, it cannot be cancelled", reservationId)
	}

	refundPercentage = 100.0
	if timeDifference.Hours() < float64(cancellationDaysLimit*24) {
		refundPercentage, err = s.SettingsRepo.GetRefundPercentage(reservation.Country)
		if err != nil {
			return 0, err
		}
		logger.Error("Service: cancellation date is beyond permmitted date, only ", refundPercentage, " percent will be refunded")
	}

	property, err := s.PropertiesRepo.GetPropertyById(reservation.PropertyId)
	if err != nil {
		return 0, err
	}

	pricePerDay := property.BookingPrice
	totalPrice := pricePerDay * int(timeDifference.Hours()/24)

	refund := float64(totalPrice) * refundPercentage / 100
	requestBody := map[string]interface{}{
		"amount": refund,
	}

	bodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequest("POST", s.refundUrl, bytes.NewBuffer(bodyJSON))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Error("Could not refund: %d", resp.StatusCode)
		return 0, fmt.Errorf("Could not refund: %d", resp.StatusCode)
	}

	logger.Info("Service: User ", email, "is trying to cancel reservation ", reservationId)
	err = s.ReservationRepo.CancelReservation(reservationId)
	if err != nil {
		return 0, err
	} else {
		return refundPercentage, nil
	}
}

func (s *ReservationService) DoCheckIn(reservationId string) error {
	return s.ReservationRepo.RegisterCheckIn(reservationId)
}

func (s *ReservationService) DoCheckOut(reservationId string) error {
	return s.ReservationRepo.RegisterCheckOut(reservationId)
}

func (s *ReservationService) GetReservationById(reservationId string) (my_models.ReservationModel, error) {
	return s.ReservationRepo.GetReservationById(reservationId)
}

func (s *ReservationService) AutoCancelReservations() error {
	propertiesIds, err := s.ReservationRepo.AutoCancelReservations(autoCancelDays)
	if err != nil {
		return err
	}

	for _, propertyId := range propertiesIds {
		owner, err := s.UserRepo.GetPropertyOwner(propertyId)
		if err != nil {
			return err
		}

		logger.Info("Notification: Sending email to Owner ", owner, " about reservation being canceled :", propertyId)
	}

	return nil
}
