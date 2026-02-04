package backtester

import (
	"time"

	"github.com/langley-creator/cf-backtester/internal/models"
	"github.com/langley-creator/cf-backtester/internal/strategy"
)

// Engine manages backtesting execution
type Engine struct {
	db           *models.DB
	strategyName string
	cfCalculator *strategy.CFCalculator
	atrCalc      *strategy.ATRCalculator
	adxCalc      *strategy.ADXCalculator
}

// NewEngine creates a new backtesting engine
func NewEngine(db *models.DB, strategyName string) *Engine {
	return &Engine{
		db:           db,
		strategyName: strategyName,
	}
}

// Run executes backtesting for the specified instrument and time range
func (e *Engine) Run(instrumentID int64, startTime, endTime time.Time) (*models.BacktestResult, error) {
	// Load strategy configuration
	strategy, err := e.db.GetStrategyByName(e.strategyName)
	if err != nil {
		return nil, err
	}

	// Initialize calculators
	e.cfCalculator = strategy.NewCFCalculator(strategy.config)
	e.atrCalc = strategy.NewATRCalculator(14) // Standard ATR period
	e.adxCalc = strategy.NewADXCalculator(14) // Standard ADX period

	// Load candles from database
	candles, err := e.db.GetCandlesByTimeRange(instrumentID, startTime, endTime)
	if err != nil {
		return nil, err
	}

	// Calculate indicators
	mainCF, secondCF := e.calculateCFIndicators(candles)
	atr := e.atrCalc.Calculate(candles)
	atrPercent := strategy.CalculateATRPercent(candles, atr)
	adx, plusDI, minusDI := e.adxCalc.Calculate(candles)

	// Execute trading logic
	result := e.executeBacktest(candles, mainCF, secondCF, atr, atrPercent, adx, plusDI, minusDI, strategy)

	// Save result to database
	result.InstrumentID = instrumentID
	result.StrategyID = strategy.ID
	result.StartTime = startTime
	result.EndTime = endTime
	result.CreatedAt = time.Now()

	if err := e.db.SaveBacktestResult(result); err != nil {
		return nil, err
	}

	return result, nil
}

// calculateCFIndicators computes Capital Flow indicators
func (e *Engine) calculateCFIndicators(candles []*models.Candle) (mainCF, secondCF []float64) {
	mainCF = make([]float64, len(candles))
	secondCF = make([]float64, len(candles))

	for i := range candles {
		if i == 0 {
			continue
		}
		// Calculate MainCF
		mainCF[i] = e.cfCalculator.CalculateMainCF(candles[:i+1], i, 0, 0)

		// Calculate SecondCF (for visualization)
		secondCF[i] = e.cfCalculator.CalculateSecondCF(candles[:i+1], i, 0)
	}

	return mainCF, secondCF
}

// executeBacktest runs trading simulation
func (e *Engine) executeBacktest(
	candles []*models.Candle,
	mainCF, secondCF, atr, atrPercent, adx, plusDI, minusDI []float64,
	strategy *models.Strategy,
) *models.BacktestResult {
	result := &models.BacktestResult{
		Metrics: make(map[string]float64),
	}

	var trades []*models.Trade
	var currentPosition *Position
	initialBalance := 10000.0
	currentBalance := initialBalance

	// Trading loop
	for i := 1; i < len(candles); i++ {
		candle := candles[i]

		// Check for entry signal
		if currentPosition == nil {
			signal := e.checkEntrySignal(i, mainCF, secondCF, adx, plusDI, minusDI, atrPercent, strategy)
			if signal != "" {
				currentPosition = e.openPosition(signal, candle, atr[i], currentBalance)
				if currentPosition != nil {
					currentBalance -= currentPosition.Size * candle.Close
				}
			}
		} else {
			// Check for exit signal
			if e.checkExitSignal(i, currentPosition, candle, atr[i]) {
				trade := e.closePosition(currentPosition, candle)
				currentBalance += trade.ExitPrice * currentPosition.Size
				trade.PnL = (trade.ExitPrice - trade.EntryPrice) * currentPosition.Size
				if currentPosition.Side == "SHORT" {
					trade.PnL = -trade.PnL
				}
				trades = append(trades, trade)
				currentPosition = nil
			}
		}
	}

	// Close any open position at the end
	if currentPosition != nil {
		lastCandle := candles[len(candles)-1]
		trade := e.closePosition(currentPosition, lastCandle)
		currentBalance += trade.ExitPrice * currentPosition.Size
		trade.PnL = (trade.ExitPrice - trade.EntryPrice) * currentPosition.Size
		if currentPosition.Side == "SHORT" {
			trade.PnL = -trade.PnL
		}
		trades = append(trades, trade)
	}

	// Calculate metrics
	e.calculateMetrics(result, trades, initialBalance, currentBalance)

	return result
}

// Position represents an open trading position
type Position struct {
	Side       string
	EntryPrice float64
	EntryTime  time.Time
	Size       float64
	StopLoss   float64
	TakeProfit float64
}

// checkEntrySignal determines if entry conditions are met
func (e *Engine) checkEntrySignal(
	idx int,
	mainCF, secondCF, adx, plusDI, minusDI, atrPercent []float64,
	strategy *models.Strategy,
) string {
	if idx < 2 {
		return ""
	}

	// Long signal: MainCF crosses above 0, ADX > threshold, +DI > -DI
	if mainCF[idx] > 0 && mainCF[idx-1] <= 0 {
		if adx[idx] > strategy.Config.ADXThreshold && plusDI[idx] > minusDI[idx] {
			return "LONG"
		}
	}

	// Short signal: MainCF crosses below 0, ADX > threshold, -DI > +DI
	if mainCF[idx] < 0 && mainCF[idx-1] >= 0 {
		if adx[idx] > strategy.Config.ADXThreshold && minusDI[idx] > plusDI[idx] {
			return "SHORT"
		}
	}

	return ""
}

// openPosition creates a new trading position
func (e *Engine) openPosition(signal string, candle *models.Candle, atr, balance float64) *Position {
	position := &Position{
		Side:       signal,
		EntryPrice: candle.Close,
		EntryTime:  candle.Timestamp,
		Size:       balance * 0.1 / candle.Close, // Risk 10% of balance
	}

	// Set stop loss and take profit based on ATR
	if signal == "LONG" {
		position.StopLoss = candle.Close - 2*atr
		position.TakeProfit = candle.Close + 3*atr
	} else {
		position.StopLoss = candle.Close + 2*atr
		position.TakeProfit = candle.Close - 3*atr
	}

	return position
}

// checkExitSignal determines if exit conditions are met
func (e *Engine) checkExitSignal(idx int, position *Position, candle *models.Candle, atr float64) bool {
	if position.Side == "LONG" {
		// Exit long: hit stop loss or take profit
		if candle.Low <= position.StopLoss || candle.High >= position.TakeProfit {
			return true
		}
	} else {
		// Exit short: hit stop loss or take profit
		if candle.High >= position.StopLoss || candle.Low <= position.TakeProfit {
			return true
		}
	}
	return false
}

// closePosition closes an open position
func (e *Engine) closePosition(position *Position, candle *models.Candle) *models.Trade {
	exitPrice := candle.Close

	// Determine actual exit price based on SL/TP
	if position.Side == "LONG" {
		if candle.Low <= position.StopLoss {
			exitPrice = position.StopLoss
		} else if candle.High >= position.TakeProfit {
			exitPrice = position.TakeProfit
		}
	} else {
		if candle.High >= position.StopLoss {
			exitPrice = position.StopLoss
		} else if candle.Low <= position.TakeProfit {
			exitPrice = position.TakeProfit
		}
	}

	return &models.Trade{
		Side:       position.Side,
		EntryPrice: position.EntryPrice,
		EntryTime:  position.EntryTime,
		ExitPrice:  exitPrice,
		ExitTime:   candle.Timestamp,
		Size:       position.Size,
	}
}

// calculateMetrics computes performance metrics
func (e *Engine) calculateMetrics(result *models.BacktestResult, trades []*models.Trade, initialBalance, finalBalance float64) {
	result.TotalTrades = len(trades)
	if result.TotalTrades == 0 {
		return
	}

	var winningTrades, losingTrades int
	var totalProfit, totalLoss float64

	for _, trade := range trades {
		if trade.PnL > 0 {
			winningTrades++
			totalProfit += trade.PnL
		} else {
			losingTrades++
			totalLoss += trade.PnL
		}
	}

	result.WinningTrades = winningTrades
	result.LosingTrades = losingTrades
	result.WinRate = float64(winningTrades) / float64(result.TotalTrades) * 100

	if winningTrades > 0 {
		result.Metrics["avg_win"] = totalProfit / float64(winningTrades)
	}
	if losingTrades > 0 {
		result.Metrics["avg_loss"] = totalLoss / float64(losingTrades)
	}

	result.TotalPnL = finalBalance - initialBalance
	result.TotalReturn = (finalBalance/initialBalance - 1) * 100

	if result.Metrics["avg_loss"] != 0 {
		result.Metrics["profit_factor"] = totalProfit / (-totalLoss)
	}
}
