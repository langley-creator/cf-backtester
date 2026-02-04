package uploader

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	
	"github.com/langley-creator/cf-backtester/internal/database"
	"github.com/langley-creator/cf-backtester/internal/models"
)

// CSVUploader handles CSV file uploads
type CSVUploader struct {
	db *database.PostgresDB
}

// NewCSVUploader creates a new CSV uploader
func NewCSVUploader(db *database.PostgresDB) *CSVUploader {
	return &CSVUploader{db: db}
}

// UploadCSV parses and uploads CSV file with candle data
// Expected CSV format: timestamp,open,high,low,close,volume,...
// Example: 1711929600000,128.32,128.32,128.07,128.09,223.266,...
func (u *CSVUploader) UploadCSV(filePath string, symbol string, interval string) (int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()
	
	reader := csv.NewReader(file)
	reader.Comma = ','
	reader.TrimLeadingSpace = true
	
	count := 0
	lineNum := 0
	
	for {
		lineNum++
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return count, fmt.Errorf("error reading line %d: %w", lineNum, err)
		}
		
		// Skip empty lines or lines with dashes (separators)
		if len(record) == 0 || strings.HasPrefix(record[0], "-") {
			continue
		}
		
		// Need at least 6 fields: timestamp, open, high, low, close, volume
		if len(record) < 6 {
			continue
		}
		
		// Parse candle data
		candle, err := u.parseCandle(record, symbol, interval)
		if err != nil {
			// Skip invalid lines but continue processing
			continue
		}
		
		// Insert into database (ON CONFLICT DO NOTHING)
		err = u.db.InsertCandle(candle)
		if err != nil {
			// Log but continue
			continue
		}
		
		count++
	}
	
	return count, nil
}

// parseCandle parses a CSV record into a Candle model
func (u *CSVUploader) parseCandle(record []string, symbol string, interval string) (*models.Candle, error) {
	// Parse timestamp (milliseconds)
	timestampMs, err := strconv.ParseInt(record[0], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid timestamp: %w", err)
	}
	
	// Convert to seconds
	timestamp := timestampMs / 1000
	
	// Parse OHLCV
	open, err := strconv.ParseFloat(record[1], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid open: %w", err)
	}
	
	high, err := strconv.ParseFloat(record[2], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid high: %w", err)
	}
	
	low, err := strconv.ParseFloat(record[3], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid low: %w", err)
	}
	
	close, err := strconv.ParseFloat(record[4], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid close: %w", err)
	}
	
	volume, err := strconv.ParseFloat(record[5], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid volume: %w", err)
	}
	
	return &models.Candle{
		Symbol:    symbol,
		Interval:  interval,
		Timestamp: timestamp,
		Open:      open,
		High:      high,
		Low:       low,
		Close:     close,
		Volume:    volume,
	}, nil
}
