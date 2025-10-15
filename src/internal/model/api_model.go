package model

import (
	"exchange-rates-service/src/internal"
)

type StartUpdateRateRequest struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type StartUpdateRateResponse struct {
	UpdateId string `json:"updateId"`
}

type GetRateResponse struct {
	Rate       *string `json:"rate"`
	UpdateTime *string `json:"updateTime"`
}

func (r *StartUpdateRateRequest) Validate() error {
	if r.From == "" {
		return internal.NewBadRequestError("from currency is not set")
	}
	if r.To == "" {
		return internal.NewBadRequestError("to currency is not set")
	}

	return nil
}
