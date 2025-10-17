package main

import (
	"encoding/json"
	"errors"
	"exchange-rates-service/src/config"
	"exchange-rates-service/src/internal"
	"exchange-rates-service/src/internal/model"
	"exchange-rates-service/src/internal/service"
	"log"
	"net/http"
	"time"
)

var serviceConfig = config.NewConfig()
var rateService *service.RateService

func startUpdateRate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.NotFound(w, r)
		return
	}

	var request model.StartUpdateRateRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		handleError(w, err)
		return
	}

	if err := request.Validate(); err != nil {
		handleError(w, err)
		return
	}

	updateId, err := rateService.StartUpdateRate(request.From, request.To)
	if err != nil {
		handleError(w, err)
		return
	}

	response := model.StartUpdateRateResponse{
		UpdateId: updateId,
	}

	if err = json.NewEncoder(w).Encode(response); err != nil {
		handleError(w, err)
		return
	}
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
		err = json.NewEncoder(w).Encode(model.GetRateResponse{})
		handleError(w, err)
		return
	}

	rateValue := update.Rate.String()
	updateValue := update.UpdateDateTime.Format(time.RFC3339Nano)
	response := model.GetRateResponse{
		Rate:       &rateValue,
		UpdateTime: &updateValue,
	}

	if err = json.NewEncoder(w).Encode(response); err != nil {
		handleError(w, err)
		return
	}
}

func getLastUpdateRate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.NotFound(w, r)
		return
	}

	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	if from == "" {
		handleError(w, internal.NewBadRequestError("from currency is not set"))
		return
	}

	if to == "" {
		handleError(w, internal.NewBadRequestError("to currency is not set"))
		return
	}

	rate, err := rateService.GetLastRate(from, to)

	if err != nil {
		handleError(w, err)
		return
	}

	if rate.UpdateDateTime == nil {
		err = json.NewEncoder(w).Encode(model.GetRateResponse{})
		handleError(w, err)
		return
	}

	rateValue := rate.Rate.String()
	updateValue := rate.UpdateDateTime.Format(time.RFC3339Nano)
	response := model.GetRateResponse{
		Rate:       &rateValue,
		UpdateTime: &updateValue,
	}

	if err = json.NewEncoder(w).Encode(response); err != nil {
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
	var err error

	rateService, err = service.NewRateService(serviceConfig)
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/api/rates/v1/update/start", startUpdateRate)
	http.HandleFunc("/api/rates/v1/update", getUpdateRate)
	http.HandleFunc("/api/rates/v1/update/last", getLastUpdateRate)

	log.Println("Starting server at port 8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Println("Error starting the server:", err)
	}
}
