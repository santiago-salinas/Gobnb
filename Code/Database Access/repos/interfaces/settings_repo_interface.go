package repointerfaces

type ISettingsRepo interface {
	GetCancellationDays(countryCode string) (int, error)
	GetRefundPercentage(countryCode string) (float64, error)
}
