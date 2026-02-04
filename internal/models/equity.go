package models

// Equity represents equity curve point for a strategy run
type Equity struct {
	ID             int     `json:"id"`
	StrategyRunID  int     `json:"strategy_run_id"`
	TS             int64   `json:"ts"` // Timestamp
	Equity         float64 `json:"equity"`
}

// DebugCandle stores CF calculation details for each candle
type DebugCandle struct {
	ID             int     `json:"id"`
	StrategyRunID  int     `json:"strategy_run_id"`
	InstrumentID   int     `json:"instrument_id"`
	
	// Candle data
	TS             int64   `json:"ts"`
	Open           float64 `json:"open"`
	High           float64 `json:"high"`
	Low            float64 `json:"low"`
	Close          float64 `json:"close"`
	Volume         float64 `json:"volume"`
	
	// CF values
	MainCF         float64 `json:"main_cf"`
	SecondCF       float64 `json:"second_cf"`
	
	// ATR/ADX (if applicable)
	ATRShort       float64 `json:"atr_short"`
	ATRLong        float64 `json:"atr_long"`
	ADX            float64 `json:"adx"`
	
	// Position state
	InPosition     bool    `json:"in_position"`
	Reason         string  `json:"reason"` // ENTER, EXIT, SKIP_ATR, SKIP_ADX, SKIP_CF, etc.
	
	// Window sizes
	WindowMain     int     `json:"window_main"`     // newTotalKlines
	WindowSecond   int     `json:"window_second"`   // secondTotalKlines
}
