package handlers

import (
	"encoding/json"
	"net/http"
	"temperature-observability/service_orchestrator/handlers"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type Request struct {
	Cep string `json:"cep"`
}

const (
	INVALID_ZIPCODE       = "Invalid zipcode"
	INTERNAL_SERVER_ERROR = "internal server error"
	CAN_NOT_FIND_ZIPCODE  = "can not find zipcode"
)

var tracer trace.Tracer

func init() {
	tracer = otel.Tracer("serviceA")
}

func InputCepHandler(w http.ResponseWriter, r *http.Request) {
	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req.Cep) != 8 {
		http.Error(w, INVALID_ZIPCODE, http.StatusUnprocessableEntity)
		return
	}

	ctx, span := tracer.Start(r.Context(), "ServiceA:ValidateCEP")
	defer span.End()

	clientRequest, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://serviceb:8081/weather?cep="+req.Cep, nil)
	if err != nil {
		http.Error(w, INTERNAL_SERVER_ERROR, http.StatusInternalServerError)
		return
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(clientRequest)
	if err != nil {
		http.Error(w, INTERNAL_SERVER_ERROR, http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		w.WriteHeader(resp.StatusCode)
		w.Write([]byte(CAN_NOT_FIND_ZIPCODE))
		return
	}

	var response handlers.WeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		http.Error(w, INTERNAL_SERVER_ERROR, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
