# Pitchlake WebSocket Server

A high-performance WebSocket server built in Go for real-time blockchain data streaming, specifically designed for gas price monitoring, vault state updates, and home dashboard data.

## 🚀 Features

- **Real-time WebSocket connections** for live data streaming
- **Gas price monitoring** with configurable time windows and TWAP calculations
- **Vault state management** with user-specific subscriptions
- **Home dashboard data** streaming
- **Concurrent subscriber management** with thread-safe operations
- **PostgreSQL integration** for data persistence
- **Health check endpoints** for monitoring
- **Comprehensive test coverage** for all API components

## ⚡ Quick Start

```bash
# Build and run
make build && make run
```

## 🏗️ Architecture

The server follows a clean, modular architecture:

```
server/
├── api/
│   ├── general/          # Gas price and general data endpoints
│   ├── home/             # Home dashboard data endpoints  
│   ├── vault/            # Vault state and user data endpoints
│   └── utils/            # Shared utilities
├── types/                # Type definitions and interfaces
├── db/                   # Database layer and repositories
├── models/               # Data models and structures
└── validations.go        # Request validation logic
```

## 📡 API Endpoints

### General Endpoints (`/general`)
- **`/health`** - Health check endpoint
- **`/subscribeGas`** - WebSocket endpoint for gas price data

### Home Endpoints (`/home`)  
- **`/subscribeHome`** - WebSocket endpoint for home dashboard data

### Vault Endpoints (`/vault`)
- **`/subscribeVault`** - WebSocket endpoint for vault state updates

## 🔌 WebSocket Subscriptions

### Gas Data Subscription
Subscribe to real-time gas price data with configurable parameters:

```json
{
  "startTimestamp": 1000,
  "endTimestamp": 2000,
  "roundDuration": 960
}
```

**Round Duration Options:**
- `960` - 12-minute TWAP
- `13200` - 3-hour TWAP  
- `2631600` - 30-day TWAP

### Vault Subscription
Subscribe to vault-specific updates:

```json
{
  "address": "0x...",
  "vaultAddress": "0x...",
  "userType": "user"
}
```

## 🛠️ Development

### Prerequisites
- Go 1.22.5+
- PostgreSQL database
- Docker (optional, for containerized development)

### Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd pitchlake-websocket
   ```

2. **Install dependencies**
   ```bash
   go mod tidy
   ```

3. **Set up environment variables**
   ```bash
   # Copy and configure environment file
   cp .env.example .env
   ```

4. **Run the server**
   ```bash
   go run .
   ```

### Docker Development

```bash
# Build and run with Docker Compose
docker-compose up --build

# Or build manually
docker build -t pitchlake-websocket .
docker run -p 8080:8080 pitchlake-websocket
```

## 🧪 Testing

The project includes comprehensive test coverage with both unit and integration tests. All testing commands are available via Makefile for easy development.

### Test Commands

#### **Run All Tests**
```bash
# Using Makefile (recommended)
make test

# Raw Go command
go test ./...
```

#### **Unit Tests Only** (Fast Development)
```bash
# Using Makefile (recommended)
make test-unit

# Raw Go command
go test ./server/api/... ./server/validations/...
```

#### **Integration Tests Only**
```bash
# Using Makefile (recommended)
make test-integration

# Raw Go command
go test ./server/...
```

#### **Run Tests by Package**
```bash
# Vault API tests
go test ./server/api/vault/...
```

#### **Test Coverage**
```bash
# Using Makefile (recommended)
make test-coverage

# Raw Go command
go test -cover ./...

# Coverage by specific package
go test -cover ./server/validations/...
```

### Test Structure
```
Unit Tests (Fast):
├── server/api/general/     # Handler tests (6 test cases)
├── server/api/home/        # Handler tests (6 test cases)  
├── server/api/vault/       # Handler tests (6 test cases)
└── server/validations/     # Validation tests (22 test cases)
Total: 40 unit tests

Integration Tests (Slower):
└── server/integration_test.go  # WebSocket tests (4 test cases)
```

## 📊 Data Models

### Core Types
- **`SubscriberGas`** - Gas price subscription data
- **`SubscriberVault`** - Vault subscription data  
- **`SubscriberHome`** - Home dashboard subscription data
- **`BlockResponse`** - Blockchain block data with TWAP values

### Database Models
- **`Block`** - Blockchain block information
- **`VaultState`** - Current vault state
- **`LiquidityProviderState`** - Liquidity provider status
- **`OptionBuyer`** - Option buyer information
- **`OptionRound`** - Option round details

## 🔒 Concurrency & Thread Safety

The server uses mutex-protected subscriber management to ensure thread-safe operations:

- **`SubscribersWithLock`** - Thread-safe subscriber collections
- **Concurrent subscriber addition/removal** - Safe for high-traffic scenarios
- **Message buffering** - Configurable buffer sizes for performance

## 📈 Performance Features

- **Message buffering** with configurable buffer sizes
- **Efficient WebSocket handling** using the `coder/websocket` library
- **Database connection pooling** with `pgx`
- **Graceful connection handling** with timeout management

## 🚨 Error Handling

- **Connection timeout management**
- **Graceful error recovery**
- **Comprehensive logging** for debugging

## 🔧 Configuration

Key configuration options:

```go
type GeneralRouter struct {
    subscriberMessageBuffer int           // Message buffer size
    Subscribers            SubscribersWithLock
    log                    log.Logger
    pool                   pgxpool.Pool
}
```
