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

	_ "exchange-rates-service/src/docs"

	httpSwagger "github.com/swaggo/http-swagger"
)

var serviceConfig = config.NewConfig()
var rateService *service.RateService

// StartUpdateRate godoc
//
//	@Summary		Start exchange rate update
//	@Description	Start exchange rate update. Returns updateId, which can be used in GetUpdateRate
//	@Tags			exchange-rate-api
//	@Accept			json
//	@Produce		json
//	@Param			request	body		model.StartUpdateRateRequest	true	"Update request"
//	@Success		200		{object}	model.StartUpdateRateResponse	"OK"
//	@Router			/api/rates/v1/update/start [post]
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

// GetUpdateRate godoc
//
//	@Summary		Get exchange rate update
//	@Description	Get the exchange rate update by updateId. Returns rate and updateTime. Both will be null if the update was not performed
//	@Tags			exchange-rate-api
//	@Accept			json
//	@Produce		json
//	@Param			updateId	query		string					true	"Update id"
//	@Success		200			{object}	model.GetRateResponse	"OK"
//	@Failure		404			{string}	error					"NotFound"
//	@Failure		400			{string}	error					"BadRequest"
//	@Router			/api/rates/v1/update [get]
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

// GetLastUpdateRate godoc
//
//	@Summary		Get last exchange rate update
//	@Description	Get exchange rate update. Returns rate and updateTime
//	@Tags			exchange-rate-api
//	@Accept			json
//	@Produce		json
//	@Param			from	query		string					true	"From currency"
//	@Param			to		query		string					true	"To currency"
//	@Success		200		{object}	model.GetRateResponse	"OK"
//	@Failure		404		{string}	error					"NotFound"
//	@Failure		400		{string}	error					"BadRequest"
//	@Router			/api/rates/v1/update/last [get]
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

	http.HandleFunc("/swagger/", httpSwagger.WrapHandler)

	log.Println("Starting server at port 8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Println("Error starting the server:", err)
	}
}
