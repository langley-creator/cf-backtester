package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/langley-creator/cf-backtester/internal/models"
)

// DB wraps database connection and provides data access methods
type DB struct {
	conn *sql.DB
}

// InitDB initializes database connection and creates tables
func InitDB(connString string) (*DB, error) {
	conn, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db := &DB{conn: conn}
	if err := db.createTables(); err != nil {
		return nil, err
	}

	return db, nil
}

// Close closes database connection
func (db *DB) Close() error {
	return db.conn.Close()
}

// createTables creates all necessary database tables
func (db *DB) createTables() error {
	schema := `
	CREATE TABLE IF NOT EXISTS instruments (
		id SERIAL PRIMARY KEY,
		symbol VARCHAR(50) NOT NULL,
		timeframe VARCHAR(10) NOT NULL,
		exchange VARCHAR(50) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(symbol, timeframe, exchange)
	);

	CREATE TABLE IF NOT EXISTS candles (
		id SERIAL PRIMARY KEY,
		instrument_id INTEGER NOT NULL REFERENCES instruments(id),
		timestamp BIGINT NOT NULL,
		open NUMERIC(20, 8) NOT NULL,
		high NUMERIC(20, 8) NOT NULL,
		low NUMERIC(20, 8) NOT NULL,
		close NUMERIC(20, 8) NOT NULL,
		volume NUMERIC(20, 8) NOT NULL,
		UNIQUE(instrument_id, timestamp)
	);

	CREATE INDEX IF NOT EXISTS idx_candles_instrument_timestamp 
		ON candles(instrument_id, timestamp);

	CREATE TABLE IF NOT EXISTS strategies (
		id SERIAL PRIMARY KEY,
		name VARCHAR(100) NOT NULL UNIQUE,
		config JSONB NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS backtest_results (
		id SERIAL PRIMARY KEY,
		instrument_id INTEGER NOT NULL REFERENCES instruments(id),
		strategy_id INTEGER NOT NULL REFERENCES strategies(id),
		start_time TIMESTAMP NOT NULL,
		end_time TIMESTAMP NOT NULL,
		total_trades INTEGER NOT NULL,
		winning_trades INTEGER NOT NULL,
		losing_trades INTEGER NOT NULL,
		win_rate NUMERIC(5, 2) NOT NULL,
		total_pnl NUMERIC(20, 8) NOT NULL,
		total_return NUMERIC(10, 4) NOT NULL,
		metrics JSONB,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS trades (
		id SERIAL PRIMARY KEY,
		backtest_id INTEGER NOT NULL REFERENCES backtest_results(id),
		side VARCHAR(10) NOT NULL,
		entry_price NUMERIC(20, 8) NOT NULL,
		entry_time TIMESTAMP NOT NULL,
		exit_price NUMERIC(20, 8) NOT NULL,
		exit_time TIMESTAMP NOT NULL,
		size NUMERIC(20, 8) NOT NULL,
		pnl NUMERIC(20, 8) NOT NULL
	);
	`

	_, err := db.conn.Exec(schema)
	return err
}

// SaveInstrument saves an instrument to database
func (db *DB) SaveInstrument(inst *models.Instrument) error {
	query := `
		INSERT INTO instruments (symbol, timeframe, exchange) 
		VALUES ($1, $2, $3) 
		ON CONFLICT (symbol, timeframe, exchange) DO UPDATE SET symbol = EXCLUDED.symbol
		RETURNING id`
	return db.conn.QueryRow(query, inst.Symbol, inst.Timeframe, inst.Exchange).Scan(&inst.ID)
}

// GetInstrumentBySymbol retrieves an instrument by symbol and timeframe
func (db *DB) GetInstrumentBySymbol(symbol, timeframe string) (*models.Instrument, error) {
	query := `SELECT id, symbol, timeframe, exchange FROM instruments WHERE symbol = $1 AND timeframe = $2`
	inst := &models.Instrument{}
	err := db.conn.QueryRow(query, symbol, timeframe).Scan(&inst.ID, &inst.Symbol, &inst.Timeframe, &inst.Exchange)
	if err != nil {
		return nil, err
	}
	return inst, nil
}

// GetAllInstruments retrieves all instruments from database
func (db *DB) GetAllInstruments() ([]*models.Instrument, error) {
	query := `SELECT id, symbol, timeframe, exchange FROM instruments ORDER BY symbol`
	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var instruments []*models.Instrument
	for rows.Next() {
		inst := &models.Instrument{}
		if err := rows.Scan(&inst.ID, &inst.Symbol, &inst.Timeframe, &inst.Exchange); err != nil {
			return nil, err
		}
		instruments = append(instruments, inst)
	}
	return instruments, nil
}

// SaveCandle saves a candle to database
func (db *DB) SaveCandle(candle *models.Candle) error {
	query := `
		INSERT INTO candles (instrument_id, timestamp, open, high, low, close, volume)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (instrument_id, timestamp) DO NOTHING
		RETURNING id`
	err := db.conn.QueryRow(query, candle.InstrumentID, candle.Timestamp.UnixMilli(), 
		candle.Open, candle.High, candle.Low, candle.Close, candle.Volume).Scan(&candle.ID)
	if err == sql.ErrNoRows {
		return nil // Duplicate, ignore
	}
	return err
}

// GetCandlesByTimeRange retrieves candles for an instrument within a time range
func (db *DB) GetCandlesByTimeRange(instrumentID int64, startTime, endTime time.Time) ([]*models.Candle, error) {
	query := `
		SELECT id, instrument_id, timestamp, open, high, low, close, volume
		FROM candles
		WHERE instrument_id = $1 AND timestamp >= $2 AND timestamp <= $3
		ORDER BY timestamp ASC`
	
	rows, err := db.conn.Query(query, instrumentID, startTime.UnixMilli(), endTime.UnixMilli())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var candles []*models.Candle
	for rows.Next() {
		candle := &models.Candle{}
		var timestamp int64
		if err := rows.Scan(&candle.ID, &candle.InstrumentID, &timestamp, 
			&candle.Open, &candle.High, &candle.Low, &candle.Close, &candle.Volume); err != nil {
			return nil, err
		}
		candle.Timestamp = time.UnixMilli(timestamp)
		candles = append(candles, candle)
	}
	return candles, nil
}

// SaveStrategy saves a strategy to database
func (db *DB) SaveStrategy(strategy *models.Strategy) error {
	configJSON, err := json.Marshal(strategy.Config)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO strategies (name, config) 
		VALUES ($1, $2) 
		ON CONFLICT (name) DO UPDATE SET config = EXCLUDED.config, updated_at = CURRENT_TIMESTAMP
		RETURNING id`
	return db.conn.QueryRow(query, strategy.Name, configJSON).Scan(&strategy.ID)
}

// GetStrategyByName retrieves a strategy by name
func (db *DB) GetStrategyByName(name string) (*models.Strategy, error) {
	query := `SELECT id, name, config FROM strategies WHERE name = $1`
	strategy := &models.Strategy{}
	var configJSON []byte
	err := db.conn.QueryRow(query, name).Scan(&strategy.ID, &strategy.Name, &configJSON)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(configJSON, &strategy.Config); err != nil {
		return nil, err
	}
	return strategy, nil
}

// SaveBacktestResult saves backtest result to database
func (db *DB) SaveBacktestResult(result *models.BacktestResult) error {
	metricsJSON, err := json.Marshal(result.Metrics)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO backtest_results 
		(instrument_id, strategy_id, start_time, end_time, total_trades, winning_trades, losing_trades, 
		win_rate, total_pnl, total_return, metrics)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id`
	
	return db.conn.QueryRow(query, result.InstrumentID, result.StrategyID, result.StartTime, result.EndTime,
		result.TotalTrades, result.WinningTrades, result.LosingTrades, result.WinRate, 
		result.TotalPnL, result.TotalReturn, metricsJSON).Scan(&result.ID)
}

// SaveTrade saves a trade to database
func (db *DB) SaveTrade(backtestID int64, trade *models.Trade) error {
	query := `
		INSERT INTO trades (backtest_id, side, entry_price, entry_time, exit_price, exit_time, size, pnl)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`
	
	return db.conn.QueryRow(query, backtestID, trade.Side, trade.EntryPrice, trade.EntryTime,
		trade.ExitPrice, trade.ExitTime, trade.Size, trade.PnL).Scan(&trade.ID)
}
