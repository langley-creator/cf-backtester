package database

import (
	"database/sql"
	"fmt"

	"github.com/langley-creator/cf-backtester/internal/models"
	_ "github.com/lib/pq"
)

type PostgresDB struct {
	db *sql.DB
}

// NewPostgresDB creates a new PostgreSQL connection
func NewPostgresDB(connStr string) (*PostgresDB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresDB{db: db}, nil
}

// Close closes the database connection
func (p *PostgresDB) Close() error {
	return p.db.Close()
}

// CreateTables creates necessary tables
func (p *PostgresDB) CreateTables() error {
	query := `
	CREATE TABLE IF NOT EXISTS candles (
		id SERIAL PRIMARY KEY,
		symbol VARCHAR(20) NOT NULL,
		interval VARCHAR(10) NOT NULL,
		timestamp BIGINT NOT NULL,
		open DECIMAL(20, 8) NOT NULL,
		high DECIMAL(20, 8) NOT NULL,
		low DECIMAL(20, 8) NOT NULL,
		close DECIMAL(20, 8) NOT NULL,
		volume DECIMAL(20, 8) NOT NULL,
		UNIQUE(symbol, interval, timestamp)
	);

	CREATE INDEX IF NOT EXISTS idx_candles_symbol_interval ON candles(symbol, interval);
	CREATE INDEX IF NOT EXISTS idx_candles_timestamp ON candles(timestamp);
	`

	_, err := p.db.Exec(query)
	return err
}

// InsertCandle inserts a new candle
func (p *PostgresDB) InsertCandle(c *models.Candle) error {
	query := `
		INSERT INTO candles (symbol, interval, timestamp, open, high, low, close, volume)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (symbol, interval, timestamp) DO NOTHING
	`

	_, err := p.db.Exec(query, c.Symbol, c.Interval, c.Timestamp, c.Open, c.High, c.Low, c.Close, c.Volume)
	return err
}

// GetCandles retrieves candles for a symbol and interval
func (p *PostgresDB) GetCandles(symbol, interval string, startTime, endTime int64) ([]*models.Candle, error) {
	query := `
		SELECT id, symbol, interval, timestamp, open, high, low, close, volume
		FROM candles
		WHERE symbol = $1 AND interval = $2 AND timestamp >= $3 AND timestamp <= $4
		ORDER BY timestamp DESC
	`

	rows, err := p.db.Query(query, symbol, interval, startTime, endTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var candles []*models.Candle
	for rows.Next() {
		var c models.Candle
		err := rows.Scan(&c.ID, &c.Symbol, &c.Interval, &c.Timestamp, &c.Open, &c.High, &c.Low, &c.Close, &c.Volume)
		if err != nil {
			return nil, err
		}
		candles = append(candles, &c)
	}

	return candles, rows.Err()
}
