package my_models

type CardInformation struct {
	CardNumber string `json:"cardNumber"`
	Name       string `json:"name"`
	CVV        string `json:"cvv"`
	ExpDate    string `json:"expDate"`
}