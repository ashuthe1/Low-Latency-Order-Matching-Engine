package main

import (
	"log"
	"net/http"

	"github.com/ashuthe1/Low-Latency-Order-Matching-Engine/internal/api"
	"github.com/ashuthe1/Low-Latency-Order-Matching-Engine/internal/engine"
)

func main() {
	// Initialize the core matching engine
	matchingEngine := engine.NewEngine()

	// Initialize the API layer
	apiServer := api.NewAPI(matchingEngine)

	// Create a new ServeMux (Go 1.22+ required for wildcard paths)
	mux := http.NewServeMux()

	// Register Routes
	mux.HandleFunc("POST /api/v1/orders", apiServer.HandleSubmitOrder)
	mux.HandleFunc("DELETE /api/v1/orders/{order_id}", apiServer.HandleCancelOrder)
	mux.HandleFunc("GET /api/v1/orderbook/{symbol}", apiServer.HandleGetOrderBook)
	
	// Optional: GET /api/v1/orders/{order_id} can be implemented similarly using the OrderToSymbol map
	// mux.HandleFunc("GET /api/v1/orders/{order_id}", apiServer.HandleGetOrderStatus)

	// Server configurations
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Println("🚀 Low-Latency Matching Engine started on port 8080")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}