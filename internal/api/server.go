package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/langley-creator/cf-backtester/internal/backtester"
	"github.com/langley-creator/cf-backtester/internal/database"
	"github.com/langley-creator/cf-backtester/internal/models"
)

// Server represents the API server
type Server struct {
	db     *database.DB
	router *mux.Router
	port   string
}

// NewServer creates a new API server
func NewServer(db *database.DB, port string) *Server {
	s := &Server{
		db:     db,
		router: mux.NewRouter(),
		port:   port,
	}
	s.setupRoutes()
	return s
}

// setupRoutes configures API routes
func (s *Server) setupRoutes() {
	// CORS middleware
	s.router.Use(corsMiddleware)

	// Health check
	s.router.HandleFunc("/api/health", s.healthCheck).Methods("GET")

	// Instruments
	s.router.HandleFunc("/api/instruments", s.getInstruments).Methods("GET")
	s.router.HandleFunc("/api/instruments", s.createInstrument).Methods("POST")

	// Strategies
	s.router.HandleFunc("/api/strategies", s.getStrategies).Methods("GET")
	s.router.HandleFunc("/api/strategies", s.createStrategy).Methods("POST")
	s.router.HandleFunc("/api/strategies/{name}", s.getStrategy).Methods("GET")

	// Backtests
	s.router.HandleFunc("/api/backtests", s.runBacktest).Methods("POST")
	s.router.HandleFunc("/api/backtests", s.getBacktests).Methods("GET")
	s.router.HandleFunc("/api/backtests/{id}", s.getBacktest).Methods("GET")
	s.router.HandleFunc("/api/backtests/{id}/trades", s.getBacktestTrades).Methods("GET")
}

// Start starts the API server
func (s *Server) Start() error {
	log.Printf("Starting API server on port %s", s.port)
	return http.ListenAndServe(":"+s.port, s.router)
}

// corsMiddleware adds CORS headers
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// healthCheck returns server health status
func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	responseJSON(w, http.StatusOK, map[string]string{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
	})
}

// getInstruments returns all instruments
func (s *Server) getInstruments(w http.ResponseWriter, r *http.Request) {
	instruments, err := s.db.GetAllInstruments()
	if err != nil {
		responseError(w, http.StatusInternalServerError, "Failed to fetch instruments")
		return
	}
	responseJSON(w, http.StatusOK, instruments)
}

// createInstrument creates a new instrument
func (s *Server) createInstrument(w http.ResponseWriter, r *http.Request) {
	var instrument models.Instrument
	if err := json.NewDecoder(r.Body).Decode(&instrument); err != nil {
		responseError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := s.db.SaveInstrument(&instrument); err != nil {
		responseError(w, http.StatusInternalServerError, "Failed to create instrument")
		return
	}

	responseJSON(w, http.StatusCreated, instrument)
}

// getStrategies returns all strategies
func (s *Server) getStrategies(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement GetAllStrategies in database layer
	responseJSON(w, http.StatusOK, []string{"capital_flow"})
}

// createStrategy creates a new strategy
func (s *Server) createStrategy(w http.ResponseWriter, r *http.Request) {
	var strategy models.Strategy
	if err := json.NewDecoder(r.Body).Decode(&strategy); err != nil {
		responseError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := s.db.SaveStrategy(&strategy); err != nil {
		responseError(w, http.StatusInternalServerError, "Failed to create strategy")
		return
	}

	responseJSON(w, http.StatusCreated, strategy)
}

// getStrategy returns a specific strategy
func (s *Server) getStrategy(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	strategy, err := s.db.GetStrategyByName(name)
	if err != nil {
		responseError(w, http.StatusNotFound, "Strategy not found")
		return
	}

	responseJSON(w, http.StatusOK, strategy)
}

// BacktestRequest represents a backtest request
type BacktestRequest struct {
	InstrumentID int64  `json:"instrument_id"`
	StrategyName string `json:"strategy_name"`
	StartDate    string `json:"start_date"`
	EndDate      string `json:"end_date"`
}

// runBacktest executes a backtest
func (s *Server) runBacktest(w http.ResponseWriter, r *http.Request) {
	var req BacktestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responseError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Parse dates
	startTime, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		responseError(w, http.StatusBadRequest, "Invalid start date format")
		return
	}

	endTime, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		responseError(w, http.StatusBadRequest, "Invalid end date format")
		return
	}

	// Run backtest
	engine := backtester.NewEngine(s.db, req.StrategyName)
	result, err := engine.Run(req.InstrumentID, startTime, endTime)
	if err != nil {
		responseError(w, http.StatusInternalServerError, fmt.Sprintf("Backtest failed: %v", err))
		return
	}

	responseJSON(w, http.StatusOK, result)
}

// getBacktests returns all backtest results
func (s *Server) getBacktests(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement GetAllBacktests in database layer
	responseJSON(w, http.StatusOK, []map[string]interface{}{})
}

// getBacktest returns a specific backtest result
func (s *Server) getBacktest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		responseError(w, http.StatusBadRequest, "Invalid backtest ID")
		return
	}

	// TODO: Implement GetBacktestByID in database layer
	responseJSON(w, http.StatusOK, map[string]interface{}{"id": id})
}

// getBacktestTrades returns trades for a specific backtest
func (s *Server) getBacktestTrades(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		responseError(w, http.StatusBadRequest, "Invalid backtest ID")
		return
	}

	// TODO: Implement GetTradesByBacktestID in database layer
	responseJSON(w, http.StatusOK, []map[string]interface{}{})
}

// responseJSON writes JSON response
func responseJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// responseError writes error response
func responseError(w http.ResponseWriter, status int, message string) {
	responseJSON(w, status, map[string]string{"error": message})
}
