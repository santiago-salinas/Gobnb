package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"pocketbase_go/logger"
	"pocketbase_go/my_models"
	interfaces "pocketbase_go/repos/interfaces"
)

type PaymentService struct {
	PropertyRepo    interfaces.IPropertyRepo
	ReservationRepo interfaces.IReservationRepo
	UsersRepo       interfaces.IUserRepo
	paymentUrl string
}

func (p *PaymentService) SetConfigValues(paymentUrl string) {
	p.paymentUrl = paymentUrl
}

func (p *PaymentService) PayProperty(propertyId string, cardInformation my_models.CardInformation) error {
	logger.Info("Service: Paying property with id: ", propertyId)
	price := 1000

	property, err := p.PropertyRepo.GetPropertyById(propertyId)
	if err != nil {
		logger.Error("Service: Error in PayProperty: ", err)
		return err
	}

	if !property.IsPendingPayment {
		return fmt.Errorf("Property is not pending payment")
	}

	if property.Paid {
		return fmt.Errorf("Property has already been paid")
	}

	requestBody := map[string]interface{}{
		"cardInformation": cardInformation,
		"price":           price,
	}

	bodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		logger.Error("Service: Error in PayProperty: ", err)
		return err
	}

	req, err := http.NewRequest("POST", p.paymentUrl, bytes.NewBuffer(bodyJSON))
	if err != nil {
		logger.Error("Service: Error in PayProperty: ", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Service: Error in PayProperty: ", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Error("Service: Error in PayProperty: %d", resp.StatusCode)
		return fmt.Errorf("Something went wrong: %d", resp.StatusCode)
	}

	err = p.PropertyRepo.UpdatePropertyPaidStatus(propertyId)
	if err != nil {
		logger.Error("Service: Error in PayProperty: ", err)
		return err
	}

	logger.Info("Service: Property paid successfully")
	return nil
}

func (p *PaymentService) PayReservation(reservationId string, cardInformation my_models.CardInformation) error {
	logger.Info("Service: Paying reservation with id: ", reservationId)
	reservation, err := p.ReservationRepo.GetReservationById(reservationId)
	if err != nil {
		return err
	}

	property, err := p.PropertyRepo.GetPropertyById(reservation.PropertyId)
	if err != nil {
		return err
	}

	status := reservation.Status
	if status != "Approved" {
		return fmt.Errorf("Reservation is not approved")
	}

	price := property.BookingPrice

	reservationStartDate, err := time.Parse(my_models.PocketTimeLayout, reservation.ReservedFrom)
	if err != nil {
		return err
	}
	reservationEndDate, err := time.Parse(my_models.PocketTimeLayout, reservation.ReservedUntil)
	if err != nil {
		return err
	}

	timeDifference := reservationEndDate.Sub(reservationStartDate)
	days := int(timeDifference.Hours() / 24)

	totalPrice := price * days

	requestBody := map[string]interface{}{
		"cardInformation": cardInformation,
		"price":           totalPrice,
	}

	bodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		logger.Error("Service: Error in PayReservation: ", err)
		return err
	}

	req, err := http.NewRequest("POST", p.paymentUrl, bytes.NewBuffer(bodyJSON))
	if err != nil {
		logger.Error("Service: Error in PayReservation: ", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Service: Error in PayReservation: ", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Something went wrong: %d", resp.StatusCode)
	}

	err = p.ReservationRepo.UpdateReservationStatus(reservationId, "Paid")
	if err != nil {
		logger.Error("Service: Error in PayReservation: ", err)
		return err
	}

	admins, err := p.UsersRepo.GetUsersByRole("Admin")
	if err != nil {
		logger.Error("Service: Error in PayReservation: ", err)
		return err
	}
	for _, admin := range admins {
		logger.Info("Service: Notifying admin: ", admin, " about reservation: ", reservationId)
	}
	logger.Info("Service: Notifying owner of property: ", property.Owner, " about reservation: ", reservationId)
	logger.Info("Service: Reservation paid successfully")
	return nil
}