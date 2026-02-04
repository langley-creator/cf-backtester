package main

import (
	"flag"
	"log"
	"os"

	"github.com/langley-creator/cf-backtester/internal/api"
	"github.com/langley-creator/cf-backtester/internal/database"
)

func main() {
	// Command line flags
	dbConnString := flag.String("db", getEnv("DATABASE_URL", "postgres://user:password@localhost:5432/backtester?sslmode=disable"), "PostgreSQL connection string")
	port := flag.String("port", getEnv("PORT", "8080"), "API server port")
	flag.Parse()

	// Initialize database
	db, err := database.InitDB(*dbConnString)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	log.Println("Database initialized successfully")

	// Create and start API server
	server := api.NewServer(db, *port)
	log.Printf("Starting API server on http://localhost:%s", *port)
	log.Println("API endpoints:")
	log.Println("  GET  /api/health - Health check")
	log.Println("  GET  /api/instruments - List all instruments")
	log.Println("  POST /api/instruments - Create instrument")
	log.Println("  GET  /api/strategies - List strategies")
	log.Println("  POST /api/strategies - Create strategy")
	log.Println("  GET  /api/strategies/{name} - Get strategy")
	log.Println("  POST /api/backtests - Run backtest")
	log.Println("  GET  /api/backtests - List backtests")
	log.Println("  GET  /api/backtests/{id} - Get backtest details")
	log.Println("  GET  /api/backtests/{id}/trades - Get backtest trades")

	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start API server: %v", err)
	}
}

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
