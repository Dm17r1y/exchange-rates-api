package main

import (
	"encoding/json"
	"exchange-rates-service/src/config"
	"exchange-rates-service/src/internal"
	"exchange-rates-service/src/internal/service"
	"log"
	"net/http"
	"time"
	"errors"
)

var serviceConfig = config.NewConfig()
var rateService = service.NewRateService(serviceConfig)

type startUpdateRateRequest struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type startUpdateRateResponse struct {
	UpdateId string `json:"updateId"`
}

func startUpdateRate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.NotFound(w, r)
		return
	}

	var request startUpdateRateRequest
	err := json.NewDecoder(r.Body).Decode(&request)

	if err != nil {
		handleError(w, internal.NewBadRequestError(err.Error()))
		return
	}

	updateId, err := rateService.StartUpdateRate(request.From, request.To)
	if err != nil {
		handleError(w, err)
		return
	}

	response := startUpdateRateResponse{
		UpdateId: updateId,
	}

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		handleError(w, err)
		return
	}
}

type getUpdateRateResponse struct {
	Rate       *string `json:"rate"`
	UpdateTime *string `json:"updateTime"`
}

func getUpdateRate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.NotFound(w, r)
		return
	}

	updateId := r.URL.Query().Get("updateId")
	if updateId == "" {
		handleError(w, internal.NewBadRequestError("updateId is not set"))
		return
	}

	update, err := rateService.GetRateUpdate(updateId)
	if err != nil {
		handleError(w, err)
		return
	}

	if update.UpdateDateTime == nil {
		err = json.NewEncoder(w).Encode(getUpdateRateResponse{})
		handleError(w, err)
		return
	}

	rateValue := update.Rate.String()
	updateValue := update.UpdateDateTime.Format(time.RFC3339Nano)
	response := getUpdateRateResponse{
		Rate:       &rateValue,
		UpdateTime: &updateValue,
	}

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		handleError(w, err)
		return
	}
}

type getLastUpdateRateResponse struct {
	Rate       *string `json:"rate"`
	UpdateTime *string `json:"updateTime"`
}

func getLastUpdateRate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.NotFound(w, r)
		return
	}

	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	if from == "" {
		handleError(w, internal.NewBadRequestError("from is not set"))
		return
	}

	if to == "" {
		handleError(w, internal.NewBadRequestError("to is not set"))
		return
	}


	rate, err := rateService.GetLastRate(from, to)

	if err != nil {
		handleError(w, err)
		return
	}

	if rate.UpdateDateTime == nil {
		err = json.NewEncoder(w).Encode(getLastUpdateRateResponse{})
		handleError(w, err)
		return
	}

	rateValue := rate.Rate.String()
	updateValue := rate.UpdateDateTime.Format(time.RFC3339Nano)
	response := getUpdateRateResponse{
		Rate:       &rateValue,
		UpdateTime: &updateValue,
	}

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		handleError(w, err)
		return
	}
}

func handleError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	
	serviceError := &internal.ServiceError{} 
	if errors.As(err, &serviceError) {
		http.Error(w, serviceError.ErrorMessage, int(serviceError.ErrorType))
		return
	}

	log.Println(err)
	http.Error(w, "Internal server error", http.StatusInternalServerError)	

}

func main() {

	http.HandleFunc("/api/rates/v1/update/start", startUpdateRate)
	http.HandleFunc("/api/rates/v1/update", getUpdateRate)
	http.HandleFunc("/api/rates/v1/update/last", getLastUpdateRate)

	log.Println("Starting server at port 8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Println("Error starting the server:", err)
	}
}
