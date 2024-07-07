package my_models

type DateRange struct {
	Start string `json:"start" db:"start"`
	End   string `json:"end" db:"end"`
}

const PocketTimeLayout = "2006-01-02 15:04:05.000Z"
