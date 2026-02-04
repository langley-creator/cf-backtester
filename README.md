# CF-Backtester

Crypto trading bot backtester with Capital Flow strategy implementation.

## Features

- **Capital Flow Strategy**: Main strategy based on volume and price momentum analysis
- **PostgreSQL Integration**: Stores historical candlestick data
- **Backtesting Engine**: Tests strategies on historical data
- **Multiple Timeframes**: Support for different candlestick intervals

## Prerequisites

- **Go** 1.21+ ([Download](https://go.dev/dl/))
- **Docker Desktop** for Mac ([Download](https://www.docker.com/products/docker-desktop/))
- **Git** (usually pre-installed on Mac)

## Installation (Mac)

### 1. Clone the repository

```bash
git clone https://github.com/langley-creator/cf-backtester.git
cd cf-backtester
```

### 2. Start PostgreSQL with Docker

```bash
docker compose up -d
```

This will start PostgreSQL on `localhost:5432` with:
- Database: `cryptobot`
- User: `cryptobot`
- Password: `cryptobot123`

### 3. Install Go dependencies

```bash
go mod download
```

### 4. Run the backtester

```bash
go run cmd/backtester/main.go
```

## Project Structure

```
cf-backtester/
├── cmd/
│   └── backtester/
│       └── main.go          # Entry point
├── internal/
│   ├── database/
│   │   └── postgres.go      # Database operations
│   ├── models/
│   │   └── candle.go        # Data models
│   └── strategies/
│       ├── capital_flow.go  # Main CF strategy
│       └── interface.go     # Strategy interface
├── docker-compose.yml       # PostgreSQL setup
├── go.mod                   # Go dependencies
└── README.md
```

## Configuration

Database connection settings can be modified in `cmd/backtester/main.go`:

```go
connStr := "host=localhost port=5432 user=cryptobot password=cryptobot123 dbname=cryptobot sslmode=disable"
```

## Usage

### Run backtests

```bash
go run cmd/backtester/main.go
```

### Stop PostgreSQL

```bash
docker compose down
```

### View PostgreSQL data

```bash
docker exec -it cf-postgres psql -U cryptobot -d cryptobot
```

Then you can run SQL queries:
```sql
SELECT * FROM candles ORDER BY timestamp DESC LIMIT 10;
```

## Development

### Build binary

```bash
go build -o backtester cmd/backtester/main.go
./backtester
```

### Run tests

```bash
go test ./...
```

## Strategy Details

### Capital Flow (CF)

The Capital Flow strategy identifies trading opportunities based on:
- Volume momentum
- Price action confirmation
- Entry/exit signals

## Troubleshooting

### PostgreSQL connection error

Make sure Docker is running:
```bash
docker ps
```

You should see `cf-postgres` container running.

### Port already in use

If port 5432 is already taken, modify `docker-compose.yml`:
```yaml
ports:
  - "5433:5432"  # Use 5433 instead
```

And update connection string accordingly.

## License

MIT

## Contributing

Pull requests are welcome!
