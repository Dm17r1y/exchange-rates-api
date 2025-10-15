package main

import (
	"encoding/json"
	"exchange-rates-service/src/config"
	"exchange-rates-service/src/internal/service"
	"log"
	"net/http"
	"time"
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
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	updateId, err := rateService.StartUpdateRate(request.From, request.To)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := startUpdateRateResponse{
		UpdateId: updateId,
	}

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

type getUpdateRateResponse struct {
	Rate       string `json:"rate"`
	UpdateTime string `json:"updateTime"`
}

func getUpdateRate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.NotFound(w, r)
		return
	}

	updateId := r.URL.Query().Get("updateId")
	if updateId == "" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	update, err := rateService.GetRateUpdate(updateId)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if update == nil {
		_ = json.NewEncoder(w).Encode(nil)
		return
	}

	response := getUpdateRateResponse{
		Rate:       update.Rate.String(),
		UpdateTime: update.UpdateDateTime.Format(time.RFC3339Nano),
	}

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

type getLastUpdateRateResponse struct {
	Rate       string `json:"rate"`
	UpdateTime string `json:"updateTime"`
}

func getLastUpdateRate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.NotFound(w, r)
		return
	}

	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	if from == "" || to == "" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	rate, err := rateService.GetLastRate(from, to)

	if err != nil {
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if rate == nil {
		_ = json.NewEncoder(w).Encode(nil)
		return
	}

	response := getLastUpdateRateResponse{
		Rate:       rate.Rate.String(),
		UpdateTime: rate.UpdateDateTime.Format(time.RFC3339Nano),
	}

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
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
