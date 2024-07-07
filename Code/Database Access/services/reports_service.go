package services

import (
	mongoInter "mongo-server/controllers/repointerfaces"
	"mongo-server/mongo_models"
	"pocketbase_go/logger"
	"pocketbase_go/my_models"
	interfaces "pocketbase_go/repos/interfaces"
	"strings"
	"time"
)

type ReportsService struct {
	ReservationRepo interfaces.IReservationRepo
	PropertiesRepo  interfaces.IPropertyRepo
	UsersRepo       interfaces.IUserRepo
	ReportsRepo     mongoInter.ReportsRepo
	SensorRepo      interfaces.ISensorRepo
}

func (c *ReportsService) GetLatestSensorReport(sensorId string) (mongo_models.SensorReport, error) {
	report, err := c.ReportsRepo.GetLatestSensorReport(sensorId)
	if err != nil {
		logger.Error("Service: error retrieving latest sensor report for id ", sensorId, ": ", err)
	}
	return report, err
}

func (c *ReportsService) ValidateSensorReport(report mongo_models.SensorReport) (mongo_models.SensorReport, error) {
	logger.Info("Service: Validating sensor report")
	if strings.HasPrefix(report.SensorId, "Security") {
		err := c.ValidateSecurityReport(report)
		if err != nil {
			return mongo_models.SensorReport{}, err
		} else {
			return report, nil
		}
	}

	sensor, err := c.SensorRepo.GetSensor(report.SensorId)
	if err != nil {
		logger.Error("Service: error retrieving sensor with id ", report.SensorId, ": ", err)
		return mongo_models.SensorReport{}, err
	}

	return sensor.ReportStructure.ValidateReport(report)
}

func (c *ReportsService) ValidateSecurityReport(report mongo_models.SensorReport) error {
	logger.Info("Service: Validating security report")
	parts := strings.Split(report.SensorId, "-")
	propertyId := parts[1]
	_, err := c.PropertiesRepo.GetPropertyById(propertyId)
	if err != nil {
		logger.Error("Service: error validating Security report from property ", propertyId, ": ", err)
		return err
	}

	return nil
}

func (c *ReportsService) GetPropertiesIncomes(property_id string, fromDate time.Time, untilDate time.Time) (my_models.IncomeReport, error) {
	logger.Info("Service: Getting properties incomes")

	property, err := c.PropertiesRepo.GetPropertyById(property_id)
	if err != nil {
		logger.Error("Service: error retrieving property with id ", property_id, ": ", err)
		return my_models.IncomeReport{}, err
	}

	fromDateStr := fromDate.Format(time.DateOnly)
	untilDateStr := untilDate.Format(time.DateOnly)

	filter := my_models.ReservationFilter{
		PropertyId:    &property_id,
		ReservedFrom:  &fromDateStr,
		ReservedUntil: &untilDateStr,
	}

	bookings, err := c.ReservationRepo.GetFilteredReservations(filter)
	if err != nil {
		logger.Error("Service: error retrieving reservations for property ", property_id, ": ", err)
		return my_models.IncomeReport{}, err
	}

	bookingsReports := make([]my_models.BookingIncomeReport, 0)
	for _, booking := range bookings {
		bookingsReports = append(bookingsReports, makeBookingIncomeReport(booking, float64(property.BookingPrice)))
	}

	logger.Info("Service: Got properties incomes successfully")
	return my_models.IncomeReport{
		PropertyId:      property_id,
		TotalIncome:     sumBookingsIncome(bookingsReports),
		FromDate:        fromDate,
		ToDate:          untilDate,
		BookingsReports: bookingsReports,
	}, nil
}

func (c *ReportsService) GetOccupations(fromDate time.Time, untilDate time.Time) ([]my_models.OccupationsReportItem, error) {
	logger.Info("Service: Getting occupations report")
	fromDateStr := fromDate.Format(time.DateOnly)
	untilDateStr := untilDate.Format(time.DateOnly)

	occupiedProperties, err := c.PropertiesRepo.GetOccupiedProperties(fromDateStr, untilDateStr)
	if err != nil {
		logger.Error("Service: error retrieving occupied properties: ", err)
		return nil, err
	}

	allProperties, err := c.PropertiesRepo.GetAllProperties()
	if err != nil {
		logger.Error("Service: error retrieving all properties: ", err)
		return nil, err
	}

	propertiesByNeighborhood := make(map[string]int)
	occupiedPropertiesByNeighborhood := make(map[string]int)

	for _, property := range allProperties {
		propertiesByNeighborhood[property.Neighborhood]++
	}

	for _, property := range occupiedProperties {
		occupiedPropertiesByNeighborhood[property.Neighborhood]++
	}

	occupationsReportItems := make([]my_models.OccupationsReportItem, 0)
	for neighborhood, propertiesAmount := range propertiesByNeighborhood {
		occupiedPropertiesAmount, ok := occupiedPropertiesByNeighborhood[neighborhood]
		if !ok {
			occupiedPropertiesAmount = 0
		}

		occupiedPercentage := float64(occupiedPropertiesAmount) / float64(propertiesAmount) * 100

		occupationsReportItems = append(occupationsReportItems, my_models.OccupationsReportItem{
			Neighborhood:             neighborhood,
			PropertiesAmount:         propertiesAmount,
			OccupiedPropertiesAmount: occupiedPropertiesAmount,
			OccupiedPercentage:       occupiedPercentage,
		})
	}

	logger.Info("Service: Got occupations report successfully")
	return occupationsReportItems, nil
}

func (c *ReportsService) GetPropertiesRanking(fromDate time.Time, untilDate time.Time) ([]mongo_models.RankingReportItem, error) {
	logger.Info("Service: Getting properties ranking")
	propertiesRanking, err := c.ReportsRepo.GetAllAppReports(fromDate, untilDate)
	if err != nil {
		logger.Error("Service: error retrieving all app reports: ", err)
		return nil, err
	}

	for i, rankingItem := range propertiesRanking {
		property, err := c.PropertiesRepo.GetPropertyById(rankingItem.Id)
		if err != nil {
			logger.Error("Service: error retrieving property with id ", rankingItem.Id, ": ", err)
			return nil, err
		}

		propertiesRanking[i].Name = property.Name
		propertiesRanking[i].Neighborhood = property.Neighborhood
		propertiesRanking[i].State = property.State

	}

	logger.Info("Service: Got properties ranking successfully")
	return propertiesRanking, nil
}

func makeBookingIncomeReport(booking my_models.ReservationModel, propertyPrice float64) my_models.BookingIncomeReport {
	fromDate, _ := time.Parse(my_models.PocketTimeLayout, booking.ReservedFrom)
	untilDate, _ := time.Parse(my_models.PocketTimeLayout, booking.ReservedUntil)

	daysCount := untilDate.Sub(fromDate).Hours() / 24

	return my_models.BookingIncomeReport{
		BookingId:      booking.ID,
		Income:         propertyPrice * daysCount,
		FromDate:       fromDate,
		ToDate:         untilDate,
		TenantEmail:    booking.Email,
		TenantName:     booking.Name,
		TenantLastName: booking.LastName,
	}
}

func sumBookingsIncome(bookings []my_models.BookingIncomeReport) float64 {
	totalIncome := 0.0
	for _, booking := range bookings {
		totalIncome += booking.Income
	}
	return totalIncome
}
