package strategy

import (
	"math"

	"github.com/langley-creator/cf-backtester/internal/models"
)

// ATRCalculator calculates Average True Range
type ATRCalculator struct {
	period int
}

// NewATRCalculator creates a new ATR calculator
func NewATRCalculator(period int) *ATRCalculator {
	return &ATRCalculator{period: period}
}

// Calculate computes ATR for the given candles
// ATR = EMA(TR, period) where TR = max(H-L, |H-PC|, |L-PC|)
func (a *ATRCalculator) Calculate(candles []*models.Candle) []float64 {
	if len(candles) < 2 {
		return make([]float64, len(candles))
	}

	atr := make([]float64, len(candles))
	tr := make([]float64, len(candles))

	// Calculate True Range for each candle
	tr[0] = candles[0].High - candles[0].Low
	for i := 1; i < len(candles); i++ {
		highLow := candles[i].High - candles[i].Low
		highPrevClose := math.Abs(candles[i].High - candles[i-1].Close)
		lowPrevClose := math.Abs(candles[i].Low - candles[i-1].Close)
		tr[i] = math.Max(highLow, math.Max(highPrevClose, lowPrevClose))
	}

	// Calculate EMA of TR
	alpha := 2.0 / float64(a.period+1)
	atr[0] = tr[0]
	for i := 1; i < len(candles); i++ {
		atr[i] = alpha*tr[i] + (1-alpha)*atr[i-1]
	}

	return atr
}

// ADXCalculator calculates Average Directional Index
type ADXCalculator struct {
	period int
}

// NewADXCalculator creates a new ADX calculator
func NewADXCalculator(period int) *ADXCalculator {
	return &ADXCalculator{period: period}
}

// Calculate computes ADX, +DI and -DI for the given candles
// Returns ADX, +DI, -DI as separate slices
func (a *ADXCalculator) Calculate(candles []*models.Candle) (adx, plusDI, minusDI []float64) {
	n := len(candles)
	if n < 2 {
		return make([]float64, n), make([]float64, n), make([]float64, n)
	}

	adx = make([]float64, n)
	plusDI = make([]float64, n)
	minusDI = make([]float64, n)

	// Calculate +DM and -DM
	plusDM := make([]float64, n)
	minusDM := make([]float64, n)
	tr := make([]float64, n)

	tr[0] = candles[0].High - candles[0].Low
	for i := 1; i < n; i++ {
		highDiff := candles[i].High - candles[i-1].High
		lowDiff := candles[i-1].Low - candles[i].Low

		if highDiff > lowDiff && highDiff > 0 {
			plusDM[i] = highDiff
		}
		if lowDiff > highDiff && lowDiff > 0 {
			minusDM[i] = lowDiff
		}

		// Calculate TR
		highLow := candles[i].High - candles[i].Low
		highPrevClose := math.Abs(candles[i].High - candles[i-1].Close)
		lowPrevClose := math.Abs(candles[i].Low - candles[i-1].Close)
		tr[i] = math.Max(highLow, math.Max(highPrevClose, lowPrevClose))
	}

	// Smooth +DM, -DM and TR using EMA
	alpha := 2.0 / float64(a.period+1)
	smoothedPlusDM := make([]float64, n)
	smoothedMinusDM := make([]float64, n)
	smoothedTR := make([]float64, n)

	smoothedPlusDM[0] = plusDM[0]
	smoothedMinusDM[0] = minusDM[0]
	smoothedTR[0] = tr[0]

	for i := 1; i < n; i++ {
		smoothedPlusDM[i] = alpha*plusDM[i] + (1-alpha)*smoothedPlusDM[i-1]
		smoothedMinusDM[i] = alpha*minusDM[i] + (1-alpha)*smoothedMinusDM[i-1]
		smoothedTR[i] = alpha*tr[i] + (1-alpha)*smoothedTR[i-1]
	}

	// Calculate +DI and -DI
	for i := 0; i < n; i++ {
		if smoothedTR[i] != 0 {
			plusDI[i] = (smoothedPlusDM[i] / smoothedTR[i]) * 100
			minusDI[i] = (smoothedMinusDM[i] / smoothedTR[i]) * 100
		}
	}

	// Calculate DX
	dx := make([]float64, n)
	for i := 0; i < n; i++ {
		diSum := plusDI[i] + minusDI[i]
		if diSum != 0 {
			diDiff := math.Abs(plusDI[i] - minusDI[i])
			dx[i] = (diDiff / diSum) * 100
		}
	}

	// Calculate ADX as EMA of DX
	adx[0] = dx[0]
	for i := 1; i < n; i++ {
		adx[i] = alpha*dx[i] + (1-alpha)*adx[i-1]
	}

	return adx, plusDI, minusDI
}

// CalculateATRPercent calculates ATR as percentage of close price
func CalculateATRPercent(candles []*models.Candle, atr []float64) []float64 {
	n := len(candles)
	atrPercent := make([]float64, n)
	for i := 0; i < n; i++ {
		if candles[i].Close != 0 {
			atrPercent[i] = (atr[i] / candles[i].Close) * 100
		}
	}
	return atrPercent
}
