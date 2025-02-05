package main

import (
	"context"

	"log"
	"net/http"
	"os"

	"temperature-observability/service_input/handlers"

	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

func main() {
	// Configura o tracer com Zipkin
	exporter, err := zipkin.New(
		os.Getenv("ZIPKIN_ENDPOINT"),
		zipkin.WithLogger(log.Default()),
	)
	if err != nil {
		log.Fatalf("Failed to initialize Zipkin exporter: %v", err)
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("serviceA"),
		)),
	)
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Fatalf("Error shutting down tracer provider: %v", err)
		}
	}()
	otel.SetTracerProvider(tp)

	r := chi.NewRouter()
	r.Post("/cep", handlers.InputCepHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Service A running on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
