package models

import (
	"encoding/json"
	"time"
)

// BacktestResult represents backtesting results
type BacktestResult struct {
	ID           int64              `json:"id"`
	InstrumentID int64              `json:"instrument_id"`
	StrategyID   int64              `json:"strategy_id"`
	StartTime    time.Time          `json:"start_time"`
	EndTime      time.Time          `json:"end_time"`
	TotalTrades  int                `json:"total_trades"`
	WinningTrades int               `json:"winning_trades"`
	LosingTrades int                `json:"losing_trades"`
	WinRate      float64            `json:"win_rate"`
	TotalPnL     float64            `json:"total_pnl"`
	TotalReturn  float64            `json:"total_return"`
	Metrics      map[string]float64 `json:"metrics"`
	CreatedAt    time.Time          `json:"created_at"`
}

// Trade represents a single trading position
type Trade struct {
	ID         int64     `json:"id"`
	Side       string    `json:"side"`       // LONG or SHORT
	EntryPrice float64   `json:"entry_price"`
	EntryTime  time.Time `json:"entry_time"`
	ExitPrice  float64   `json:"exit_price"`
	ExitTime   time.Time `json:"exit_time"`
	Size       float64   `json:"size"`
	PnL        float64   `json:"pnl"`
}

// StrategyRun represents a single backtest run
type StrategyRun struct {
	ID              int     `json:"id"`
	StrategyID      int     `json:"strategy_id"`
	InstrumentID    int     `json:"instrument_id"`
	From            int64   `json:"from"` // Unix timestamp
	To              int64   `json:"to"`   // Unix timestamp
	Status          string  `json:"status"` // PENDING, RUNNING, DONE, FAILED
	DebugMode       bool    `json:"debug_mode"`
	DebugMaxCandles int     `json:"debug_max_candles"`
	CreatedAt       int64   `json:"created_at"`
	CompletedAt     *int64  `json:"completed_at,omitempty"`

	// Results (filled after completion)
	TotalTrades   int     `json:"total_trades"`
	WinningTrades int     `json:"winning_trades"`
	LosingTrades  int     `json:"losing_trades"`
	TotalPnL      float64 `json:"total_pnl"`
	FinalEquity   float64 `json:"final_equity"`
	MaxDrawdown   float64 `json:"max_drawdown"`
	SharpeRatio   float64 `json:"sharpe_ratio"`
	ProfitFactor  float64 `json:"profit_factor"`
	WinRate       float64 `json:"win_rate"`
	Expectancy    float64 `json:"expectancy"`
}

// TradeDetails represents detailed trade information
type TradeDetails struct {
	ID          int     `json:"id"`
	StrategyRunID int   `json:"strategy_run_id"`
	InstrumentID int    `json:"instrument_id"`

	// Entry
	EntryTS    int64   `json:"entry_ts"` // Entry timestamp
	EntryPrice float64 `json:"entry_price"`

	// Exit
	ExitTS    int64   `json:"exit_ts"` // Exit timestamp
	ExitPrice float64 `json:"exit_price"`

	Side     string  `json:"side"` // LONG or SHORT
	Leverage float64 `json:"leverage"`

	// PnL
	PositionNotional float64 `json:"position_notional"`
	Fees             float64 `json:"fees"`
	PnLRaw           float64 `json:"pnl_raw"`   // Percentage
	PnLMoney         float64 `json:"pnl_money"` // Dollar amount

	// Metadata (CF values, reason, etc.)
	Meta json.RawMessage `json:"meta"`
}
