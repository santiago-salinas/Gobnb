package repositories

import (
	"pocketbase_go/logger"

	"github.com/pocketbase/pocketbase"
)

const (
	cancellationSettings    = "cancellations_days_settings"
	refundSettings          = "refund_percentage_settings"
)

type PocketSettingsRepo struct {
	Db                      pocketbase.PocketBase
	defaultRefundPercentage float64
	defaultCancellationDays int
}

func (s *PocketSettingsRepo) SetConfigValues(defaultRefundPercentage float64, defaultCancellationDays int) {
	s.defaultCancellationDays = defaultCancellationDays
	s.defaultRefundPercentage = defaultRefundPercentage
}

func (s *PocketSettingsRepo) GetCancellationDays(countryCode string) (days int, err error) {
	logger.Info("Repo: Getting cancellation days for country: ", countryCode)
	record, err := s.Db.Dao().FindFirstRecordByData(cancellationSettings, "country", countryCode)
	if err != nil && err.Error() == "sql: no rows in result set" {
		logger.Warn("Repo: No cancellation days found for country: ", countryCode)
		return s.defaultCancellationDays, nil
	}

	if err != nil {
		logger.Error("Repo: ", err)
		return 0, err
	}

	logger.Info("Repo: Cancellation days found for country: ", countryCode)
	return record.GetInt("days"), nil
}

func (s *PocketSettingsRepo) GetRefundPercentage(countryCode string) (percentage float64, err error) {
	logger.Info("Repo: Getting refund percentage for country: ", countryCode)
	record, err := s.Db.Dao().FindFirstRecordByData(refundSettings, "country", countryCode)
	if err != nil && err.Error() == "sql: no rows in result set" {
		logger.Warn("Repo: No refund percentage found for country: ", countryCode)
		return s.defaultRefundPercentage, nil
	}

	if err != nil {
		logger.Error("Repo: ", err)
		return 0, err
	}

	logger.Info("Repo: Refund percentage found for country: ", countryCode)
	return record.GetFloat("value"), nil
}
