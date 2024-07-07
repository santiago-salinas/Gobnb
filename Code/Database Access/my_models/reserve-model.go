package my_models

import (
	"fmt"
	"time"
)

type ReservationModel struct {
	ID            string `json:"id" db:"id"`
	Document      string `json:"document" db:"document"`
	Name          string `json:"name" db:"name"`
	LastName      string `json:"last_name" db:"last_name"`
	Email         string `json:"email" db:"email"`
	Phone         string `json:"phone" db:"phone"`
	Address       string `json:"address" db:"address"`
	Nationality   string `json:"nationality" db:"nationality"`
	Country       string `json:"country" db:"country"`
	Adults        int    `json:"adults" db:"adults"`
	Minors        int    `json:"minors" db:"minors"`
	PropertyId    string `json:"property" db:"property"`
	ReservedFrom  string `json:"reserved_from" db:"reserved_from"`
	ReservedUntil string `json:"reserved_until" db:"reserved_until"`
	Status        string `json:"status" db:"status"`
	CheckIn       string `json:"check_in" db:"check_in"`
	CheckOut      string `json:"check_out" db:"check_out"`
}


type ReservationFilter struct {
	ReservedFrom   *string `json:"reserved_from"`
	ReservedUntil  *string `json:"reserved_until"`
	Status         *string `json:"status"`
	PropertyId     *string `json:"propertyId"`
	TenantEmail    *string `json:"email"`
	TenantName     *string `json:"name"`
	TenantLastName *string `json:"lastName"`
}

func (r *ReservationModel) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id":             r.ID,
		"document":       r.Document,
		"name":           r.Name,
		"last_name":      r.LastName,
		"email":          r.Email,
		"phone":          r.Phone,
		"address":        r.Address,
		"nationality":    r.Nationality,
		"country":        r.Country,
		"adults":         r.Adults,
		"minors":         r.Minors,
		"property":       r.PropertyId,
		"reserved_from":  r.ReservedFrom,
		"reserved_until": r.ReservedUntil,
		"status":         r.Status,
	}
}

func (r *ReservationModel) ValidateFields() error {
	if r.Adults < 1 {
		return fmt.Errorf("at least one adult must be present")
	}

	if r.ReservedFrom == "" || r.ReservedUntil == "" {
		return fmt.Errorf("dates must be provided")
	}

	fromDate, err := time.Parse(time.DateOnly, r.ReservedFrom)
	if err != nil {
		return err
	}

	untilDate, err := time.Parse(time.DateOnly, r.ReservedUntil)
	if err != nil {
		return err
	}

	if fromDate.After(untilDate) {
		return fmt.Errorf("start date must not be after end date")
	}

	if fromDate.Before(time.Now()) {
		return fmt.Errorf("start date must not be in the past")
	}

	return nil
}
