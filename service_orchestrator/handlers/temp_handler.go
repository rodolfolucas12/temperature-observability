package handlers

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

const (
	INVALID_ZIPCODE       = "Invalid zipcode"
	INTERNAL_SERVER_ERROR = "internal server error"
	CAN_NOT_FIND_ZIPCODE  = "can not find zipcode"
	CEP_NOT_FOUND         = "CEP not found"
)

type WeatherResponse struct {
	City  string  `json:"city"`
	TempC float64 `json:"temp_C"`
	TempF float64 `json:"temp_F"`
	TempK float64 `json:"temp_K"`
}

var tracer trace.Tracer

func init() {
	tracer = otel.Tracer("serviceB")
}

func GetTemperature(w http.ResponseWriter, r *http.Request) {
	cep := r.URL.Query().Get("cep")
	if len(cep) != 8 {
		http.Error(w, INVALID_ZIPCODE, http.StatusUnprocessableEntity)
		return
	}

	ctx, span := tracer.Start(r.Context(), "ServiceB:FetchWeather")
	defer span.End()

	city, err := fetchCityByCEP(ctx, cep)
	if err != nil {
		if err.Error() == CEP_NOT_FOUND {
			http.Error(w, CAN_NOT_FIND_ZIPCODE, http.StatusNotFound)
		} else {
			http.Error(w, INTERNAL_SERVER_ERROR, http.StatusInternalServerError)
		}
		return
	}

	tempC, err := fetchTemperature(ctx, city)
	if err != nil {
		http.Error(w, INTERNAL_SERVER_ERROR, http.StatusInternalServerError)
		return
	}

	response := WeatherResponse{
		City:  city,
		TempC: tempC,
		TempF: tempC*1.8 + 32,
		TempK: tempC + 273,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func fetchCityByCEP(ctx context.Context, cep string) (string, error) {
	ctx, span := tracer.Start(ctx, "ServiceB:fetchCityByCEP")
	defer span.End()

	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Get(fmt.Sprintf("https://viacep.com.br/ws/%s/json/", cep))
	if err != nil {
		log.Printf("Erro ao realizar requisição para ViaCEP: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Resposta inválida da ViaCEP: %d", resp.StatusCode)
		return "", fmt.Errorf(CEP_NOT_FOUND)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("Erro ao decodificar resposta da ViaCEP: %v", err)
		return "", err
	}

	if _, exists := result["erro"]; exists {
		log.Printf("CEP não encontrado: %s", cep)
		return "", fmt.Errorf(CEP_NOT_FOUND)
	}

	city, ok := result["localidade"].(string)
	if !ok {
		log.Printf("Erro ao extrair cidade da resposta da ViaCEP")
		return "", fmt.Errorf(CEP_NOT_FOUND)
	}

	log.Printf("CEP %s corresponde à cidade %s", cep, city)
	return city, nil
}

func fetchTemperature(ctx context.Context, city string) (float64, error) {
	ctx, span := tracer.Start(ctx, "ServiceB:fetchTemperature")
	defer span.End()

	apiKey := os.Getenv("WEATHER_API_KEY")
	if apiKey == "" {
		log.Println("WEATHER_API_KEY não está definida")
		return 0, fmt.Errorf("WEATHER_API_KEY não definida")
	}

	encodedCyty := url.QueryEscape(city)
	url := fmt.Sprintf("http://api.weatherapi.com/v1/current.json?key=%s&q=%s", apiKey, encodedCyty)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		log.Printf("Erro ao criar requisição para WeatherAPI: %v", err)
		return 0, err
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Erro ao realizar requisição para WeatherAPI: %v", err)
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("erro ao buscar temperatura")
	}

	var weatherData struct {
		Current struct {
			TempC float64 `json:"temp_c"`
		} `json:"current"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&weatherData); err != nil {
		log.Printf("Erro ao decodificar resposta da WeatherAPI: %v", err)
		return 0, err
	}

	log.Printf("Temperatura atual em %s: %.2f°C", city, weatherData.Current.TempC)
	return weatherData.Current.TempC, nil
}
