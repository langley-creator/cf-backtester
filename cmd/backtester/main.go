package main

import (
	"fmt"
	"log"

	"github.com/langley-creator/cf-backtester/internal/database"
	"github.com/langley-creator/cf-backtester/internal/models"
)

func main() {
	fmt.Println("CF-Backtester Starting...")

	// Database connection string
	connStr := "host=localhost port=5432 user=cryptobot password=cryptobot123 dbname=cryptobot sslmode=disable"

	// Initialize database
	db, err := database.NewPostgresDB(connStr)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	fmt.Println("Connected to PostgreSQL successfully!")

	// Create tables if they don't exist
	err = db.CreateTables()
	if err != nil {
		log.Fatal("Failed to create tables:", err)
	}

	fmt.Println("Database tables ready")

	// Example: Insert sample candle data
	candle := &models.Candle{
		Symbol:    "BTCUSDT",
		Interval:  "1h",
		Timestamp: 1704067200, // Example timestamp
		Open:      42000.0,
		High:      42500.0,
		Low:       41800.0,
		Close:     42300.0,
		Volume:    1250.5,
	}

	err = db.InsertCandle(candle)
	if err != nil {
		log.Printf("Note: Could not insert sample candle (might already exist): %v\n", err)
	} else {
		fmt.Println("Sample candle inserted successfully")
	}

	// Fetch candles from database
	candles, err := db.GetCandles("BTCUSDT", "1h", 0, 9999999999)
	if err != nil {
		log.Fatal("Failed to fetch candles:", err)
	}

	fmt.Printf("\nFound %d candles in database\n", len(candles))
	if len(candles) > 0 {
		fmt.Printf("Latest candle: %s at %.2f\n", candles[0].Symbol, candles[0].Close)
	}

	fmt.Println("\nBacktester ready! Add your strategy logic here.")
}
