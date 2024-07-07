package my_models

import (
	"strings"
)

type Property struct {
	Id               string      `json:"id" db:"id"`
	Name             string      `json:"name" db:"name"`
	AdultQuantity    int         `json:"adultQuantity" db:"adultQuantity"`
	KidQuantity      int         `json:"kidQuantity" db:"kidQuantity"`
	KingSizedBeds    int         `json:"kingSizedBeds" db:"kingSizedBeds"`
	SingleBeds       int         `json:"singleBeds" db:"singleBeds"`
	HasAC            string      `json:"hasAC" db:"hasAC"`
	HasWIFI          string      `json:"hasWIFI" db:"hasWIFI"`
	HasGarage        string      `json:"hasGarage" db:"hasGarage"`
	Type             int         `json:"type" db:"type"`
	BeachDistance    int         `json:"beachDistance" db:"beachDistance"`
	State            string      `json:"state" db:"state"`
	Resort           string      `json:"resort" db:"resort"`
	Neighborhood     string      `json:"neighborhood" db:"neighborhood"`
	UnavailableDates []DateRange `json:"unavailableDates" db:"unavailableDates"`
	IsPendingPayment bool        `json:"isPendingPayment" db:"isPendingPayment"`
	Paid             bool        `json:"paid" db:"paid"`
	Owner            string      `json:"owner" db:"owner"`
	BookingPrice     int         `json:"bookingPrice" db:"bookingPrice"`
	Images           []string    `json:"images" db:"images"`
}

type PropertyDBO struct {
	Id               string `json:"id" db:"id"`
	Name             string `json:"name" db:"name"`
	AdultQuantity    int    `json:"adultQuantity" db:"adultQuantity"`
	KidQuantity      int    `json:"kidQuantity" db:"kidQuantity"`
	KingSizedBeds    int    `json:"kingSizedBeds" db:"kingSizedBeds"`
	SingleBeds       int    `json:"singleBeds" db:"singleBeds"`
	HasAC            bool   `json:"hasAC" db:"hasAC"`
	HasWIFI          bool   `json:"hasWIFI" db:"hasWIFI"`
	HasGarage        bool   `json:"hasGarage" db:"hasGarage"`
	Type             int    `json:"type" db:"type"`
	BeachDistance    int    `json:"beachDistance" db:"beachDistance"`
	State            string `json:"state" db:"state"`
	Resort           string `json:"resort" db:"resort"`
	Neighborhood     string `json:"neighborhood" db:"neighborhood"`
	UnavailableDates string `json:"unavailableDates" db:"unavailableDates"`
	IsPendingPayment bool   `json:"isPendingPayment" db:"isPendingPayment"`
	Paid             bool   `json:"paid" db:"paid"`
	Owner            string `json:"owner" db:"owner"`
	BookingPrice     int    `json:"bookingPrice" db:"bookingPrice"`
}

type PropertyFilter struct {
	Page             *int    `json:"page"`
	Size             *int    `json:"size"`
	DateFrom         *string `json:"dateFrom"`
	DateTo           *string `json:"dateTo"`
	AdultQuantityMax *int    `json:"adultQuantityMax"`
	AdultQuantityMin *int    `json:"adultQuantityMin"`
	KidQuantityMax   *int    `json:"kidQuantityMax"`
	KidQuantityMin   *int    `json:"kidQuantityMin"`
	KingSizedBedsMax *int    `json:"kingSizedBedsMax"`
	KingSizedBedsMin *int    `json:"kingSizedBedsMin"`
	SingleBedsMax    *int    `json:"singleBedsMax"`
	SingleBedsMin    *int    `json:"singleBedsMin"`
	HasAC            *bool   `json:"hasAC"`
	HasWIFI          *bool   `json:"hasWIFI"`
	HasGarage        *bool   `json:"hasGarage"`
	Type             *int    `json:"type"`
	BeachDistanceMax *int    `json:"beachDistanceMax"`
	BeachDistanceMin *int    `json:"beachDistanceMin"`
	State            *string `json:"state"`
	Resort           *string `json:"resort"`
	Neighborhood     *string `json:"neighborhood"`
}

func (p *PropertyDBO) ToObject(unavailableDates []DateRange, images []string) Property {
	return Property{
		Id:               p.Id,
		Name:             p.Name,
		AdultQuantity:    p.AdultQuantity,
		KidQuantity:      p.KidQuantity,
		KingSizedBeds:    p.KingSizedBeds,
		SingleBeds:       p.SingleBeds,
		HasAC:            toString(p.HasAC),
		HasWIFI:          toString(p.HasWIFI),
		HasGarage:        toString(p.HasGarage),
		Type:             p.Type,
		BeachDistance:    p.BeachDistance,
		State:            p.State,
		Resort:           p.Resort,
		Neighborhood:     p.Neighborhood,
		UnavailableDates: unavailableDates,
		IsPendingPayment: p.IsPendingPayment,
		Paid:             p.Paid,
		Owner:            p.Owner,
		BookingPrice:     p.BookingPrice,
		Images:           images,
	}
}

func toString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func (d *UnavailableDatesDBO) ToObject() DateRange {
	startDate := strings.Split(d.DateFrom, " ")[0]
	endDate := strings.Split(d.DateTo, " ")[0]

	return DateRange{
		Start: startDate,
		End:   endDate,
	}
}

type ImagesDBO struct {
	PropertyId string `json:"propertyId" db:"propertyId"`
	FileName   string `json:"fileName" db:"fileName"`
}

type UnavailableDatesDBO struct {
	Id         string `json:"id" db:"id"`
	PropertyId string `json:"propertyId" db:"propertyId"`
	DateFrom   string `json:"dateFrom" db:"dateFrom"`
	DateTo     string `json:"dateTo" db:"dateTo"`
}

func (r *DateRange) ToMap(propertyId string) map[string]interface{} {
	return map[string]interface{}{
		"propertyId": propertyId,
		"dateFrom":   r.Start,
		"dateTo":     r.End,
	}
}

func (r *Property) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"name":             r.Name,
		"adultQuantity":    r.AdultQuantity,
		"kidQuantity":      r.KidQuantity,
		"kingSizedBeds":    r.KingSizedBeds,
		"singleBeds":       r.SingleBeds,
		"hasAC":            r.HasAC,
		"hasWIFI":          r.HasWIFI,
		"hasGarage":        r.HasGarage,
		"type":             r.Type,
		"beachDistance":    r.BeachDistance,
		"state":            r.State,
		"resort":           r.Resort,
		"neighborhood":     r.Neighborhood,
		"isPendingPayment": r.IsPendingPayment,
		"owner":            r.Owner,
		"bookingPrice":     r.BookingPrice,
	}
}
