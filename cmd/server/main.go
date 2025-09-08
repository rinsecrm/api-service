package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"

	"github.com/your-org/api-service/internal/canaryctx"
	"github.com/your-org/api-service/internal/client"
	"github.com/your-org/api-service/internal/server"
)

func main() {
	// Get configuration from environment
	port := getEnvOrDefault("PORT", "8080")
	storeServiceAddr := getEnvOrDefault("STORE_SERVICE_ADDR", "store-service:8080")

	// Initialize store client
	storeClient, err := client.NewStoreClient(storeServiceAddr)
	if err != nil {
		log.Fatalf("Failed to create store client: %v", err)
	}
	defer storeClient.Close()

	// Create server
	srv := server.NewServer(storeClient)

	// Setup routes
	r := mux.NewRouter()

	// API routes
	api := r.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/items", srv.CreateItem).Methods("POST")
	api.HandleFunc("/items", srv.ListItems).Methods("GET")
	api.HandleFunc("/items/{id}", srv.GetItem).Methods("GET")
	api.HandleFunc("/items/{id}", srv.UpdateItem).Methods("PUT")
	api.HandleFunc("/items/{id}", srv.DeleteItem).Methods("DELETE")
	api.HandleFunc("/items/{id}/inventory", srv.UpdateInventory).Methods("PATCH")

	// Health check
	r.HandleFunc("/health", srv.HealthCheck).Methods("GET")

	// Setup CORS with X-Canary header support
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // Configure this properly for production
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*", "X-Canary"},
		ExposedHeaders:   []string{"X-Canary-Echo"},
		AllowCredentials: true,
	})

	// Apply middleware
	handler := c.Handler(canaryctx.HTTPMiddleware(r))

	// Create HTTP server
	httpServer := &http.Server{
		Addr:    ":" + port,
		Handler: handler,
	}

	// Start server in goroutine
	go func() {
		log.Printf("API service listening on port %s", port)
		log.Printf("Store service address: %s", storeServiceAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down API service...")

	// Give outstanding requests 30 seconds to finish
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("API service stopped")
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
