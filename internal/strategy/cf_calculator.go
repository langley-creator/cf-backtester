package strategy

import (
	"math"
	
	"github.com/langley-creator/cf-backtester/internal/models"
)

// CFCalculator calculates Capital Flow indicators
type CFCalculator struct {
	config *models.StrategyConfig
}

// NewCFCalculator creates a new CF calculator
func NewCFCalculator(config *models.StrategyConfig) *CFCalculator {
	return &CFCalculator{config: config}
}

// CalculateMainCF calculates Main Capital Flow for time t
// According to TZ sections 5.1-5.3
func (c *CFCalculator) CalculateMainCF(candles []*models.Candle, t int, prevCF float64) (mainCF float64, newTotalKlines int) {
	// Need at least totalKlines candles
	if t < c.config.TotalKlines {
		return 0, c.config.TotalKlines
	}
	
	// Step 1: Calculate newTotalKlines (adaptive window)
	newTotalKlines = c.calculateNewTotalKlines(candles, t)
	
	// Step 2: Extract window
	window := candles[t-newTotalKlines+1 : t+1]
	
	// Step 3: Calculate MainCF
	mainCF = c.calculateCF(window, prevCF)
	
	return mainCF, newTotalKlines
}

// calculateNewTotalKlines calculates adaptive window size
// TZ Section 5.1
func (c *CFCalculator) calculateNewTotalKlines(candles []*models.Candle, t int) int {
	totalKlines := c.config.TotalKlines
	
	// Extract window for calculation
	window := candles[t-totalKlines+1 : t+1]
	
	// Find High and Low in window
	highWindow := c.findMaxHigh(window)
	lowWindow := c.findMinLow(window)
	
	// Calculate A_current
	if lowWindow == 0 {
		return totalKlines
	}
	aCurrent := (highWindow - lowWindow) / lowWindow
	
	// Calculate A_safe (protect from division by zero)
	aSafe := math.Max(aCurrent, c.config.Epsilon)
	
	// Calculate newTotalKlines
	// newTotalKlines = totalKlines * customAmplitude / A_safe
	newTotalKlinesFloat := float64(totalKlines) * c.config.CustomAmplitude / aSafe
	
	// Apply bounds
	newTotalKlines := int(newTotalKlinesFloat)
	newTotalKlines = c.clamp(newTotalKlines, c.config.NewTotalKlinesMin, c.config.NewTotalKlinesMax)
	
	return newTotalKlines
}

// calculateCF calculates CF value from window
// TZ Section 5.2
func (c *CFCalculator) calculateCF(window []*models.Candle, prevCF float64) float64 {
	if len(window) < 4 {
		return prevCF
	}
	
	// Split window into two halves
	half := len(window) / 2
	segment1 := window[:half]
	segment2 := window[half:]
	
	// Calculate highs and lows for each segment
	h1 := c.findMaxHigh(segment1)
	l1 := c.findMinLow(segment1)
	h2 := c.findMaxHigh(segment2)
	l2 := c.findMinLow(segment2)
	
	// Calculate H_cf and L_cf (percentage changes)
	var hCF, lCF float64
	if h1 != 0 {
		hCF = ((h2 - h1) / h1) * 100
	}
	if l1 != 0 {
		lCF = ((l2 - l1) / l1) * 100
	}
	
	// MainCF is the average of H_cf and L_cf
	mainCF := (hCF + lCF) / 2
	
	return mainCF
}

// CalculateSecondCF calculates Second Capital Flow
// TZ Sections 7.1-7.3
func (c *CFCalculator) CalculateSecondCF(candles []*models.Candle, entryIdx int, mainCFEntry float64, prevSecondTotalKlines int) (secondCF float64, secondTotalKlines int) {
	// Calculate secondTotalKlines based on mainCFEntry
	secondTotalKlines = c.calculateSecondTotalKlines(mainCFEntry, prevSecondTotalKlines)
	
	// Need enough candles
	if entryIdx+secondTotalKlines >= len(candles) {
		return 0, secondTotalKlines
	}
	
	// Extract window from entry point
	window := candles[entryIdx : entryIdx+secondTotalKlines]
	
	// Calculate SecondCF using same logic as MainCF
	secondCF = c.calculateCF(window, 0)
	
	return secondCF, secondTotalKlines
}

// calculateSecondTotalKlines calculates window size for SecondCF
// TZ Section 7.1
func (c *CFCalculator) calculateSecondTotalKlines(mainCFEntry float64, prevSecondTotalKlines int) int {
	// secondTotalKlines = newTotalKlines / (MainCF_entry * secondCFEnv)
	if mainCFEntry == 0 || c.config.SecondCFEnv == 0 {
		return c.config.SecondTotalKlinesMin
	}
	
	divisor := mainCFEntry * c.config.SecondCFEnv
	if divisor == 0 {
		return prevSecondTotalKlines
	}
	
	secondTotalKlinesFloat := float64(prevSecondTotalKlines) / math.Abs(divisor)
	secondTotalKlines := int(secondTotalKlinesFloat)
	
	// Apply minimum bound
	if secondTotalKlines < c.config.SecondTotalKlinesMin {
		secondTotalKlines = c.config.SecondTotalKlinesMin
	}
	
	return secondTotalKlines
}

// Helper functions

func (c *CFCalculator) findMaxHigh(candles []*models.Candle) float64 {
	if len(candles) == 0 {
		return 0
	}
	maxHigh := candles[0].High
	for _, candle := range candles[1:] {
		if candle.High > maxHigh {
			maxHigh = candle.High
		}
	}
	return maxHigh
}

func (c *CFCalculator) findMinLow(candles []*models.Candle) float64 {
	if len(candles) == 0 {
		return 0
	}
	minLow := candles[0].Low
	for _, candle := range candles[1:] {
		if candle.Low < minLow {
			minLow = candle.Low
		}
	}
	return minLow
}

func (c *CFCalculator) clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
