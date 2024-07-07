package my_models

import "time"

type IncomeReport struct {
	PropertyId      string                `json:"property_id"`
	Country         string                `json:"country"`
	City            string                `json:"city"`
	TotalIncome     float64               `json:"total_income"`
	FromDate        time.Time             `json:"from_date"`
	ToDate          time.Time             `json:"to_date"`
	BookingsReports []BookingIncomeReport `json:"bookings_reports"`
}

type BookingIncomeReport struct {
	BookingId      string    `json:"booking_id" db:"booking_id"`
	Income         float64   `json:"income" db:"income"`
	FromDate       time.Time `json:"from_date" db:"from_date"`
	ToDate         time.Time `json:"to_date" db:"to_date"`
	TenantEmail    string    `json:"tenant_email" db:"tenant_email"`
	TenantName     string    `json:"tenant_name" db:"tenant_name"`
	TenantLastName string    `json:"tenant_last_name" db:"tenant_last_name"`
}

type OccupationsReport struct {
	FromDate time.Time               `json:"from_date"`
	ToDate   time.Time               `json:"to_date"`
	Items    []OccupationsReportItem `json:"items"`
	City     string                  `json:"city"`
	Country  string                  `json:"country"`
}

type OccupationsReportItem struct {
	Neighborhood             string  `json:"neighborhood"`
	PropertiesAmount         int     `json:"properties_amount"`
	OccupiedPropertiesAmount int     `json:"occupied_properties_amount"`
	OccupiedPercentage       float64 `json:"occupied_percentage"`
}
