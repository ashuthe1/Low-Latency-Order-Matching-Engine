package main

import (
	"log"
	"net/http"

	"github.com/ashuthe1/Low-Latency-Order-Matching-Engine/internal/api"
	"github.com/ashuthe1/Low-Latency-Order-Matching-Engine/internal/engine"
	"github.com/ashuthe1/Low-Latency-Order-Matching-Engine/internal/metrics"
)

func main() {
	// 1. Initialize Metrics
	appMetrics := metrics.NewMetrics()

	// 2. Initialize Engine (You may need to adjust NewEngine to accept appMetrics)
	matchingEngine := engine.NewEngine()

	// 3. Initialize API
	apiServer := api.NewAPI(matchingEngine)
	apiServer.Metrics = appMetrics // Ensure your API struct has this field

	mux := http.NewServeMux()

	// 4. Register Routes with Middleware
	mux.HandleFunc("POST /api/v1/orders", apiServer.LatencyMiddleware(apiServer.HandleSubmitOrder))
	mux.HandleFunc("DELETE /api/v1/orders/{order_id}", apiServer.LatencyMiddleware(apiServer.HandleCancelOrder))
	mux.HandleFunc("GET /api/v1/orderbook/{symbol}", apiServer.LatencyMiddleware(apiServer.HandleGetOrderBook))

	// Register Health & Metrics (No latency middleware needed for these)
	mux.HandleFunc("GET /health", apiServer.HandleHealth)
	mux.HandleFunc("GET /metrics", apiServer.HandleGetMetrics)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Println("🚀 Engine started on port 8080 with Metrics tracking")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
