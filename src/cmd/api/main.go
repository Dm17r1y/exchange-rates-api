package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"exchange-rates-service/src/config"
	"exchange-rates-service/src/internal"
	"exchange-rates-service/src/internal/model"
	"exchange-rates-service/src/internal/repository"
	"exchange-rates-service/src/internal/service"
	"exchange-rates-service/src/internal/storage"
	"log"
	"net/http"
	"time"

	_ "exchange-rates-service/src/docs"

	_ "github.com/lib/pq"

	httpSwagger "github.com/swaggo/http-swagger"
)

type HttpHandler struct {
	rateService *service.RateService
}

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
func (h *HttpHandler) startUpdateRate(w http.ResponseWriter, r *http.Request) {
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

	updateId, err := h.rateService.StartUpdateRate(request.From, request.To)
	if err != nil {
		handleError(w, err)
		return
	}

	response := model.StartUpdateRateResponse{
		UpdateId: updateId,
	}

	w.Header().Set("Content-Type", "application/json")
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
func (h *HttpHandler) getUpdateRate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.NotFound(w, r)
		return
	}

	updateId := r.URL.Query().Get("updateId")
	if updateId == "" {
		handleError(w, internal.NewBadRequestError("updateId is not set"))
		return
	}

	update, err := h.rateService.GetRateUpdate(updateId)
	if err != nil {
		handleError(w, err)
		return
	}

	if update.UpdateDateTime == nil {
		if err = json.NewEncoder(w).Encode(update); err != nil {
			handleError(w, err)
		}
		return
	}

	rateValue := update.Rate.String()
	updateValue := update.UpdateDateTime.Format(time.RFC3339Nano)
	response := model.GetRateResponse{
		Rate:       &rateValue,
		UpdateTime: &updateValue,
	}

	w.Header().Set("Content-Type", "application/json")
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
func (h *HttpHandler) getLastUpdateRate(w http.ResponseWriter, r *http.Request) {
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

	rate, err := h.rateService.GetLastRate(from, to)

	if err != nil {
		handleError(w, err)
		return
	}

	if rate.UpdateDateTime == nil {
		if err = json.NewEncoder(w).Encode(model.GetRateResponse{}); err != nil {
			handleError(w, err)
		}
		return
	}

	rateValue := rate.Rate.String()
	updateValue := rate.UpdateDateTime.Format(time.RFC3339Nano)
	response := model.GetRateResponse{
		Rate:       &rateValue,
		UpdateTime: &updateValue,
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(response); err != nil {
		handleError(w, err)
		return
	}
}

func handleError(w http.ResponseWriter, err error) {
	serviceError := &internal.ServiceError{}
	if errors.As(err, &serviceError) {
		http.Error(w, serviceError.ErrorMessage, int(serviceError.ErrorType))
		return
	}

	log.Println(err)
	http.Error(w, "Internal server error", http.StatusInternalServerError)

}

func main() {
	serviceConfig := config.NewConfig()
	db, err := sql.Open("postgres", serviceConfig.PostgresConnectionString)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	exchangeRateStorage := storage.NewExchangeRateStorage(db)
	exchangeRateUpdateStorage := storage.NewExchangeRateUpdateStorage(db)
	repo := repository.NewExchangeRateRepository(db, exchangeRateStorage, exchangeRateUpdateStorage)
	rateService := service.NewRateService(repo)
	handler := HttpHandler{rateService: rateService}

	http.HandleFunc("/api/rates/v1/update/start", handler.startUpdateRate)
	http.HandleFunc("/api/rates/v1/update", handler.getUpdateRate)
	http.HandleFunc("/api/rates/v1/update/last", handler.getLastUpdateRate)

	http.HandleFunc("/swagger/", httpSwagger.WrapHandler)

	log.Println("Starting server at port 8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Println("Error starting the server:", err)
	}
}
