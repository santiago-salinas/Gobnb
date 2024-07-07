package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

type CardInformation struct {
	CardNumber string `json:"cardNumber"`
	Name       string `json:"name"`
	CVV        string `json:"cvv"`
	ExpDate    string `json:"expDate"`
}

type PaymentRequest struct {
	CardInformation CardInformation `json:"cardInformation"`
	Price           int             `json:"price"`
}

func handlerFunc(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		var paymentRequest PaymentRequest

		err := json.NewDecoder(r.Body).Decode(&paymentRequest)
		if err != nil {
			http.Error(w, "Error", http.StatusBadRequest)
			return
		}

		cardInfo := paymentRequest.CardInformation
		if cardInfo.CardNumber == "" || cardInfo.Name == "" || cardInfo.CVV == "" || cardInfo.ExpDate == "" {
			http.Error(w, "Missing info", http.StatusBadRequest)
			return
		}

		if len(cardInfo.CardNumber) != 16 {
			http.Error(w, "Invalid card number", http.StatusBadRequest)
			return
		}
		number, err := strconv.Atoi(cardInfo.CardNumber)
		if err != nil || number < 0 {
			http.Error(w, "Invalid card number", http.StatusBadRequest)
			return
		}

		if len(cardInfo.CVV) != 3 {
			http.Error(w, "Invalid CVV", http.StatusBadRequest)
			return
		}
		cvv, err := strconv.Atoi(cardInfo.CVV)
		if err != nil || cvv < 0 {
			http.Error(w, "Invalid CVV", http.StatusBadRequest)
			return
		}

		expectedDateFormat := "2006-01"
		date, err := time.Parse(expectedDateFormat, cardInfo.ExpDate)
		if err != nil || date.Before(time.Now()) {
			http.Error(w, "Invalid date", http.StatusBadRequest)
			return
		}

		time.Sleep(time.Duration(rand.Intn(5)) * time.Second)
		fmt.Printf("Payment processed")
		fmt.Fprintf(w, "Payment processed")
	} else {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
	}
}

func refundHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		time.Sleep(time.Duration(rand.Intn(5)) * time.Second)

		fmt.Fprintf(w, "Refund processed")
	} else {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
	}

}

func main() {
	fmt.Printf("Service on")
	http.HandleFunc("/", handlerFunc)
	http.HandleFunc("/refund", refundHandler)
	log.Fatal(http.ListenAndServe(":8085", nil))
}