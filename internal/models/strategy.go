package models

import "encoding/json"

// Strategy represents a backtesting strategy configuration
type Strategy struct {
	ID         int            `json:"id"`
	Name       string         `json:"name"`
	Config     StrategyConfig `json:"config"`     // Strategy parameters
	CreatedAt  int64          `json:"created_at"`
}

// StrategyConfig holds all strategy parameters
type StrategyConfig struct {
	// General
	InitialEquity    float64 `json:"initial_equity"`
	MaxOpenPositions int     `json:"max_open_positions"`
	Leverage         float64 `json:"leverage"`

	// CF Parameters
	TotalKlines      int     `json:"total_klines"`
	CustomAmplitude  float64 `json:"custom_amplitude"`
	MainCFStart      float64 `json:"main_cf_start"`
	SecondCFEnv      float64 `json:"second_cf_env"`
	Epsilon          float64 `json:"epsilon"`

	// Bounds
	NewTotalKlinesMin    int `json:"new_total_klines_min"`
	NewTotalKlinesMax    int `json:"new_total_klines_max"`
	SecondTotalKlinesMin int `json:"second_total_klines_min"`

	// ATR/ADX (for CFATRADX)
	ATRShortPeriod int     `json:"atr_short_period"`
	ATRLongPeriod  int     `json:"atr_long_period"`
	ADXPeriod      int     `json:"adx_period"`
	ADXThreshold   float64 `json:"adx_threshold"`
	ADXMin         float64 `json:"adx_min"`
	Kvol           float64 `json:"kvol"`

	// Fees
	FeeRate float64 `json:"fee_rate"`

	// Risk Management
	MaxDrawdownStop float64 `json:"max_drawdown_stop"`
}
