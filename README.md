# BetPro Server - Go Backend

A high-performance sports betting platform backend written in Go, refactored from Node.js for better concurrency and performance.

## Features

- **REST API**: Comprehensive betting platform API with authentication, betting, settlements, and more
- **WebSocket Support**: Real-time updates for odds, tickets, and account changes
- **Settlement System**: Distributed settlement worker with leader election and game-level locking
- **DOB Engine**: Live odds data orchestration with Swarm integration
- **Redis Coordination**: Distributed locking, rate limiting, and pub/sub
- **PostgreSQL**: High-performance database with connection pooling
- **Observability**: Metrics, logging, and traffic monitoring

## Architecture

```
server-go/
├── cmd/                    # Binary entry points
│   ├── api/               # Main API server
│   ├── settlement/        # Settlement worker
│   ├── presence/          # Presence worker
│   └── nginxtraffic/      # Nginx traffic worker
├── internal/              # Private application code
│   ├── config/           # Configuration
│   ├── db/               # Database layer
│   ├── models/           # Data models
│   ├── handlers/         # HTTP handlers
│   ├── middleware/       # HTTP middleware
│   ├── services/         # Business logic
│   ├── websocket/        # WebSocket servers
│   ├── dobengine/        # DOB/Swarm integration
│   └── workers/          # Worker implementations
├── pkg/                   # Public libraries
│   ├── logger/           # Logging utilities
│   ├── money/            # Money calculations
│   └── utils/            # Shared utilities
└── migrations/            # SQL migrations
```

## Prerequisites

- Go 1.21 or higher
- PostgreSQL 14+
- Redis 7+
- Swarm sports data provider access

## Installation

```bash
# Clone the repository
cd /home/keenness/Downloads/server-go

# Install dependencies
go mod download

# Copy environment variables
cp .env.example .env

# Edit .env with your configuration
nano .env
```

## Configuration

Key environment variables:

```bash
# Server
PORT=3001
JWT_SECRET=your-secret-key-change-in-production

# Database
DB_HOST=/var/run/postgresql
DB_PORT=5432
DB_NAME=betpro
DB_USER=postgres

# Redis
REDIS_URL=redis://localhost:6379
REDIS_ENABLED=true

# Swarm
SWARM_WS_URL=ws://your-swarm-provider:9999
SWARM_SITE_ID=4
SWARM_LANG=eng
```

See `.env.example` for all available options.

## Running

### Development

```bash
# Run API server
go run cmd/api/main.go

# Run settlement worker
go run cmd/settlement/main.go

# Run presence worker
go run cmd/presence/main.go
```

### Production

```bash
# Build all binaries
make build

# Or build individually
go build -o bin/api cmd/api/main.go
go build -o bin/settlement cmd/settlement/main.go
go build -o bin/presence cmd/presence/main.go
go build -o bin/nginxtraffic cmd/nginxtraffic/main.go

# Run
./bin/api
./bin/settlement
./bin/presence
```

## Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./internal/services -v

# Skip integration tests
go test -short ./...

# Run benchmarks
go test -bench=. ./...
```

## API Endpoints

### Authentication
- `POST /api/auth/register` - Register new user
- `POST /api/auth/login` - Login
- `POST /api/auth/refresh` - Refresh token

### Betting
- `POST /api/bets/place` - Place bet
- `GET /api/bets/tickets` - Get user tickets
- `POST /api/bets/cashout` - Cashout ticket
- `GET /api/bets/config` - Get betting configuration

### Wallet
- `GET /api/wallet/balance` - Get balance
- `GET /api/wallet/transactions` - Get transaction history

### Withdrawals
- `POST /api/withdrawals` - Request withdrawal
- `GET /api/withdrawals` - Get withdrawal history

### DOB (Data on Betting)
- `GET /api/dob/games` - Get games
- `GET /api/dob/live` - Get live games
- `GET /api/dob/search` - Search games

## WebSocket

### Application WebSocket
Connect to `ws://localhost:3001/ws` with JWT token:

```javascript
const ws = new WebSocket('ws://localhost:3001/ws?token=YOUR_JWT_TOKEN');

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('Received:', data);
};
```

Message types:
- `bet_tickets_update` - Ticket status updates
- `cashout_quote` - Cashout quote updates
- `account_snapshot` - Account balance updates
- `dashboard_snapshot` - Dashboard data updates

### DOB WebSocket
Connect to `ws://localhost:3001/dob-ws` for live odds updates.

## Performance

Expected improvements over Node.js version:

- **50-70% lower memory usage**: Go's efficient memory management
- **2-3x better concurrency**: Goroutines vs event loop
- **30-50% lower latency**: Compiled binary vs interpreted JavaScript
- **Better connection handling**: Native WebSocket support with goroutines

## Development

### Project Structure

- **cmd/**: Binary entry points (main.go files)
- **internal/**: Private application code (cannot be imported by other projects)
- **pkg/**: Public libraries (can be imported by other projects)
- **migrations/**: SQL migration files

### Adding a New Feature

1. Define models in `internal/models/`
2. Create service in `internal/services/`
3. Add handlers in `internal/handlers/`
4. Register routes in `cmd/api/main.go`
5. Write tests

### Code Style

- Follow standard Go conventions
- Use `gofmt` for formatting
- Run `go vet` for static analysis
- Write table-driven tests
- Use context.Context for cancellation

## Deployment

### Docker

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . .
RUN go build -o /api cmd/api/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /api /api
CMD ["/api"]
```

### Systemd

```ini
[Unit]
Description=BetPro API Server
After=network.target postgresql.service redis.service

[Service]
Type=simple
User=betpro
WorkingDirectory=/opt/betpro
EnvironmentFile=/opt/betpro/.env
ExecStart=/opt/betpro/bin/api
Restart=always

[Install]
WantedBy=multi-user.target
```

## Monitoring

- Health check: `GET /health`
- Metrics: `GET /metrics` (requires authorization)
- Logs: JSON structured logs to stdout

## License

Proprietary

## Support

For issues and questions, contact the development team.
