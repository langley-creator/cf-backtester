package models

// Candle represents a single candlestick data point
type Candle struct {
	ID        int64   `json:"id"`
	Symbol    string  `json:"symbol"`    // e.g., "BTCUSDT"
	Interval  string  `json:"interval"`  // e.g., "1h", "4h", "1d"
	Timestamp int64   `json:"timestamp"` // Unix timestamp
	Open      float64 `json:"open"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Close     float64 `json:"close"`
	Volume    float64 `json:"volume"`
}
