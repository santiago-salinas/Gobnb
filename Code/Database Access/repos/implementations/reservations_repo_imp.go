package repositories

import (
	"fmt"
	"pocketbase_go/logger"
	"pocketbase_go/my_models"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/forms"
	"github.com/pocketbase/pocketbase/models"

	dr "github.com/felixenescu/date-range"
)

const (
	reservationsCollectionName = "reservations"
	propertiesCollectionName   = "properties"
	usersCollectionName        = "users"
)

type PocketReservationRepo struct {
	Db pocketbase.PocketBase
}

func (r *PocketReservationRepo) CreateReservation(reservation my_models.ReservationModel) error {
	logger.Info("Repo: Creating reservation")
	reservationsCollection, err := r.Db.Dao().FindCollectionByNameOrId(reservationsCollectionName)
	if err != nil {
		logger.Error("Repo: ", err)
		return err
	}

	propertyRecord, err := r.Db.Dao().FindRecordById(propertiesCollectionName, reservation.PropertyId)
	if err != nil {
		logger.Error("Repo: ", err)
		return err
	}

	var unavailableDates []my_models.DateRange
	propertyRecord.UnmarshalJSONField("unavailableDates", &unavailableDates)
	propertyAdultQuantity := propertyRecord.GetInt("adultQuantity")
	propertyKidQuantity := propertyRecord.GetInt("kidQuantity")

	if reservation.Adults > propertyAdultQuantity || reservation.Minors > propertyKidQuantity {
		logger.Error("Repo: property does not have enough capacity for the given number of tenants")
		return fmt.Errorf("property does not have enough capacity for the given number of tenants")
	}

	if err := _checkExistingReservations(reservation, r.Db); err != nil {
		logger.Error("Repo: ", err)
		return err
	}

	if err := _checkPropertyAvailableDates(reservation, unavailableDates); err != nil {
		logger.Error("Repo: ", err)
		return err
	}

	record := models.NewRecord(reservationsCollection)
	form := forms.NewRecordUpsert(r.Db, record)
	reservation.Status = "Pending"

	form.LoadData(reservation.ToMap())

	fmt.Printf("Record: %v\n", record)

	if err := form.Submit(); err != nil {
		logger.Error("Repo: ", err)
		return err
	}

	logger.Info("Repo: Reservation created succesfully")
	return nil
}

func _checkExistingReservations(reservation my_models.ReservationModel, db pocketbase.PocketBase) error {
	query := fmt.Sprintf(`
		SELECT *
		FROM %s
		WHERE property = '%s'
		AND status = 'Approved'
		`, reservationsCollectionName, reservation.PropertyId)

	var existingreservations []my_models.ReservationModel
	err := db.Dao().DB().NewQuery(query).All(&existingreservations)
	if err != nil {
		return err
	}

	myreservationDateRange, err := _createDateRange(reservation.ReservedFrom, reservation.ReservedUntil, time.DateOnly)
	if err != nil {
		return err
	}

	for _, r := range existingreservations {
		rDateRange, err := _createDateRange(r.ReservedFrom, r.ReservedUntil, my_models.PocketTimeLayout)
		if err != nil {
			return err
		}
		if myreservationDateRange.Overlaps(rDateRange) {
			return fmt.Errorf("property is not available for the given dates")
		}
	}

	return nil
}

func _checkPropertyAvailableDates(reservation my_models.ReservationModel, unavailableDates []my_models.DateRange) error {
	myReservationDateRange, err := _createDateRange(reservation.ReservedFrom, reservation.ReservedUntil, time.DateOnly)
	if err != nil {
		return err
	}

	for _, elem := range unavailableDates {
		dateRange, err := _createDateRange(elem.Start, elem.End, time.DateOnly)
		if err != nil {
			return err
		}

		if myReservationDateRange.Overlaps(dateRange) {
			return fmt.Errorf("property is not available for the given dates")
		}
	}

	return nil
}

func (r *PocketReservationRepo) ApproveReservation(reservationId string) error {
	logger.Info("Repo: Approving reservation")

	_, err := r.Db.Dao().FindCollectionByNameOrId(reservationsCollectionName)
	if err != nil {
		logger.Error("Repo: ", err)
		return err
	}

	record, err := r.Db.Dao().FindRecordById(reservationsCollectionName, reservationId)
	if err != nil {
		logger.Error("Repo: ", err)
		return err
	}

	record.Set("status", "Approved")
	record.Set("approved_date", time.Now())
	err = r.Db.Dao().SaveRecord(record)
	if err != nil {
		logger.Error("Repo: ", err)
		return err
	}

	logger.Info("Repo: Reservation approved succesfully")
	return nil
}

func _createDateRange(from string, until string, dateLayout string) (dr.DateRange, error) {
	fromDate, err := time.Parse(dateLayout, from)
	if err != nil {
		return dr.DateRange{}, err
	}

	untilDate, err := time.Parse(dateLayout, until)
	if err != nil {
		return dr.DateRange{}, err
	}

	myDateRange := dr.NewDateRange(fromDate, untilDate)

	return myDateRange, nil
}

func (r *PocketReservationRepo) GetFilteredReservations(filter my_models.ReservationFilter) ([]my_models.ReservationModel, error) {
	logger.Info("Repo: Getting filtered reservations")

	// Start building the SQL query
	query := fmt.Sprintf(`
		SELECT *
		FROM %s
		WHERE 1 = 1
		`, reservationsCollectionName)

	// Add conditions based on filter values
	if filter.ReservedFrom != nil {
		query += fmt.Sprintf(` AND reserved_from >= %s`, *filter.ReservedFrom)
	}

	if filter.ReservedUntil != nil {
		query += fmt.Sprintf(` AND reserved_until <= %s`, *filter.ReservedUntil)
	}

	if filter.Status != nil {
		query += fmt.Sprintf(` AND status = '%s'`, *filter.Status)
	}

	if filter.PropertyId != nil {
		query += fmt.Sprintf(` AND property = '%s'`, *filter.PropertyId)
	}

	if filter.TenantEmail != nil {
		query += fmt.Sprintf(` AND email = '%s'`, *filter.TenantEmail)
	}

	if filter.TenantName != nil {
		query += fmt.Sprintf(` AND name = '%s'`, *filter.TenantName)
	}

	if filter.TenantLastName != nil {
		query += fmt.Sprintf(` AND lastName = '%s'`, *filter.TenantLastName)
	}

	// Execute the query
	var reservations []my_models.ReservationModel
	err := r.Db.Dao().DB().NewQuery(query).All(&reservations)
	if err != nil {
		logger.Error("Repo: ", err)
		return nil, err
	}

	logger.Info("Repo: Got filtered reservations succesfully")
	return reservations, nil
}

func (r *PocketReservationRepo) GetOwnReservation(email string, propertyId string) (my_models.ReservationModel, error) {
	logger.Info("Repo: Getting own reservation")

	query := fmt.Sprintf(`
		SELECT *
		FROM %s
		WHERE email = '%s' AND property = '%s'
		`, reservationsCollectionName, email, propertyId)

	var reservation my_models.ReservationModel
	err := r.Db.Dao().DB().NewQuery(query).One(&reservation)
	if err != nil {
		logger.Error("Repo: ", err)
		return my_models.ReservationModel{}, err
	}

	logger.Info("Repo: Got own reservation succesfully")
	return reservation, nil
}

func (r *PocketReservationRepo) RemoveReservation(reservationId string) error {
	logger.Info("Repo: Removing reservation")

	record, err := r.Db.Dao().FindRecordById(reservationsCollectionName, reservationId)
	if err != nil {
		logger.Error("Repo: Failed to find record: ", err)
		return err
	}

	err = r.Db.Dao().DeleteRecord(record)
	if err != nil {
		logger.Error("Repo: Failed to delete record: ", err)
		return err
	}

	logger.Info("Repo: Successfully removed reservation")
	return nil
}

func (r *PocketReservationRepo) CancelReservation(reservationId string) error {
	logger.Info("Repo: Cancelling reservation")

	record, err := r.Db.Dao().FindRecordById(reservationsCollectionName, reservationId)
	if err != nil {
		logger.Error("Repo: ", err)
		return err
	}

	currentStatus := record.GetString("status")

	if currentStatus == "Cancelled" {
		logger.Warn("Repo: reservation is already cancelled")
		return fmt.Errorf("reservation is already cancelled")
	}

	if currentStatus == "Pending" {
		logger.Error("Repo: reservation is still pending")
		return fmt.Errorf("reservation is still pending")
	}

	record.Set("status", "Cancelled")
	err = r.Db.Dao().SaveRecord(record)
	if err != nil {
		logger.Error("Repo: ", err)
		return err
	}

	logger.Info("Repo: Reservation cancelled succesfully")
	return nil
}

func (r *PocketReservationRepo) GetReservationById(reservationId string) (my_models.ReservationModel, error) {
	logger.Info("Repo: Getting reservation by id")

	query := fmt.Sprintf(`
		SELECT *
		FROM %s
		WHERE id = '%s'
		`, reservationsCollectionName, reservationId)

	var reservation my_models.ReservationModel
	err := r.Db.Dao().DB().NewQuery(query).One(&reservation)
	if err != nil {
		logger.Error("Repo: ", err)
		return my_models.ReservationModel{}, err
	}

	logger.Info("Repo: Got reservation by id succesfully")
	return reservation, nil
}

func (r *PocketReservationRepo) RegisterCheckIn(reservationId string) error {
	logger.Info("Repo: Registering check in")

	record, err := r.Db.Dao().FindRecordById(reservationsCollectionName, reservationId)
	if err != nil {
		logger.Error("Repo: ", err)
		return err
	}

	currentStatus := record.GetString("status")

	if currentStatus == "Cancelled" {
		logger.Error("Repo: reservation is already cancelled")
		return fmt.Errorf("reservation is already cancelled")
	}

	if currentStatus == "Pending" {
		logger.Error("Repo: reservation is still pending")
		return fmt.Errorf("reservation is still pending")
	}

	record.Set("check_in", time.Now())
	err = r.Db.Dao().SaveRecord(record)
	if err != nil {
		logger.Error("Repo: ", err)
		return err
	}

	logger.Info("Repo: Check in registered succesfully")
	return nil
}

func (r *PocketReservationRepo) RegisterCheckOut(reservationId string) error {
	logger.Info("Repo: Doing check out")

	record, err := r.Db.Dao().FindRecordById(reservationsCollectionName, reservationId)
	if err != nil {
		logger.Error("Repo: ", err)
		return err
	}

	currentStatus := record.GetString("status")

	if currentStatus == "Cancelled" {
		logger.Error("Repo: reservation is already cancelled")
		return fmt.Errorf("reservation is already cancelled")
	}

	if currentStatus == "Pending" {
		logger.Error("Repo: reservation is still pending")
		return fmt.Errorf("reservation is still pending")
	}

	record.Set("check_out", time.Now())
	err = r.Db.Dao().SaveRecord(record)
	if err != nil {
		logger.Error("Repo: ", err)
		return err
	}

	logger.Info("Repo: Check out registered succesfully")
	return nil
}

func (r *PocketReservationRepo) UpdateReservationStatus(id string, status string) error {
	logger.Info("Repo: Updating reservation status")
	record, err := r.Db.Dao().FindRecordById("reservations", id)
	if err != nil {
		logger.Error("Repo: ", err)
		return err
	}

	record.Set("status", status)

	if err := r.Db.Dao().SaveRecord(record); err != nil {
		logger.Error("Repo: ", err)
		return err
	}

	logger.Info("Repo: Reservation status updated succesfully")
	return nil
}

func (r *PocketReservationRepo) AutoCancelReservations(autoCancelDays int) ([]string, error) {
	logger.Info("Repo: Auto cancelling reservations")
	today := time.Now()
	expirationDate := today.AddDate(0, 0, -autoCancelDays)

	query := fmt.Sprintf(`
		SELECT *
		FROM %s
		WHERE status = 'Approved'
		AND approved_date < '%s'
		`, reservationsCollectionName, expirationDate.Format(time.DateOnly))

	var reservations []my_models.ReservationModel
	err := r.Db.Dao().DB().NewQuery(query).All(&reservations)
	if err != nil {
		logger.Error("Repo: ", err)
		return nil, err
	}

	propertiesIds := []string{}
	for _, reservation := range reservations {
		if err := r.CancelReservation(reservation.ID); err != nil {
			logger.Error("Repo: ", err)
			return nil, err
		}
		propertiesIds = append(propertiesIds, reservation.PropertyId)
		logger.Info("Notification: Sending email to Tenant ", reservation.Email, " about reservation being canceled :", reservation.ID)
	}

	logger.Info("Repo: Auto cancelled reservations succesfully")
	return propertiesIds, nil
}
