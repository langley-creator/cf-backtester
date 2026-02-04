package models

// Instrument represents a trading instrument
type Instrument struct {
	ID         int    `json:"id"`
	Symbol     string `json:"symbol"` // e.g., "BTCUSDT"
	Exchange   string `json:"exchange"` // e.g., "BINANCE"
	Kind       string `json:"kind"` // "CFBASE" or "CFATRADX"
	VolBucket  string `json:"vol_bucket"` // "low", "mid", or "high"
	CreatedAt  int64  `json:"created_at"`
}
