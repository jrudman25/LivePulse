# LivePulse - Real-Time Event Engagement Platform

A high-performance Go backend system for powering live events with real-time user engagement, reactions, and collective milestone tracking.

## 🌟 Features

- **Real-Time Event Processing**: High-throughput event ingestion using buffered channels
- **Worker Pool Architecture**: Configurable goroutine pool for concurrent event processing
- **In-Memory Aggregation**: Thread-safe, atomic counters for real-time statistics
- **Milestone Tracking**: Automatic detection and notification of crowd goals
- **WebSocket Support**: Real-time bidirectional communication with clients
- **DynamoDB Integration**: Persistent storage for events, sessions, and milestones
- **Graceful Shutdown**: Proper cleanup and event draining on shutdown

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    HTTP/WebSocket API                   │
├─────────────────────────────────────────────────────────┤
│              Event Ingestion Pipeline                   │
│   ┌──────────────┐      ┌────────────────┐            │
│   │ Event Queue  │─────▶│  Worker Pool   │            │
│   │  (Channel)   │      │  (Goroutines)  │            │
│   └──────────────┘      └────────┬───────┘            │
├──────────────────────────────────┼──────────────────────┤
│         In-Memory Aggregation    │                      │
│   ┌──────────────────────────────▼────────────┐        │
│   │  Session Stats (Thread-Safe Maps)         │        │
│   │  - Reaction Counts (Atomic)                │        │
│   │  - Active Users                            │        │
│   │  - Peak Concurrent Users                   │        │
│   └──────────────────┬─────────────────────────┘        │
├──────────────────────┼──────────────────────────────────┤
│   Milestone Tracker  │                                  │
│   ┌──────────────────▼────────────┐                    │
│   │  Goal Detection & Notify      │                    │
│   └──────────────────┬────────────┘                    │
├──────────────────────┼──────────────────────────────────┤
│   DynamoDB Storage   │                                  │
│   ┌──────────────────▼────────────┐                    │
│   │  Events, Sessions, Milestones │                    │
│   └───────────────────────────────┘                    │
└─────────────────────────────────────────────────────────┘
```

## 📋 Prerequisites

- **Go 1.21+** installed
- **AWS Account** (optional, for DynamoDB)
- **DynamoDB Local** (optional, for local development)

## 🚀 Quick Start

### 1. Clone and Setup

```bash
cd c:\Projects\LivePulse

# Copy environment configuration
copy .env.example .env

# Edit .env with your settings
notepad .env
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Run the Server

```bash
# Run with default configuration (in-memory only)
go run cmd/server/main.go

# Or build and run
go build -o livepulse.exe cmd/server/main.go
.\livepulse.exe
```

The server will start on `http://localhost:8080` by default.

## 🔧 Configuration

Edit `.env` file to configure the application:

```env
# Server Settings
SERVER_PORT=8080
SERVER_READ_TIMEOUT=15s
SERVER_WRITE_TIMEOUT=15s

# Worker Pool
WORKER_COUNT=10
EVENT_QUEUE_SIZE=10000

# DynamoDB (leave DYNAMODB_ENDPOINT empty for AWS)
DYNAMODB_ENDPOINT=http://localhost:8000  # For local DynamoDB
AWS_REGION=us-east-1

# AWS Credentials (only for AWS DynamoDB)
# AWS_ACCESS_KEY_ID=your_key
# AWS_SECRET_ACCESS_KEY=your_secret

# Table Names
DYNAMODB_EVENTS_TABLE=LivePulse_Events
DYNAMODB_SESSIONS_TABLE=LivePulse_Sessions
DYNAMODB_MILESTONES_TABLE=LivePulse_Milestones

# Milestone Thresholds
MILESTONE_THRESHOLDS=100,500,1000,5000,10000
```

## 📡 API Endpoints

### REST API

#### Create Session
```bash
POST /api/sessions
Content-Type: application/json

{
  "name": "My Awesome Event",
  "milestones": [100, 500, 1000]
}

Response:
{
  "session_id": "uuid",
  "name": "My Awesome Event",
  "created_at": "2026-01-07T18:00:00Z"
}
```

#### Join Session
```bash
POST /api/sessions/join?session_id=<session_id>&user_id=<user_id>

Response:
{
  "status": "joined",
  "session_id": "uuid",
  "user_id": "user123"
}
```

#### Get Session Statistics
```bash
GET /api/sessions/stats?session_id=<session_id>

Response:
{
  "session_id": "uuid",
  "active_user_count": 42,
  "peak_concurrent_users": 100,
  "total_reactions": 1523,
  "reaction_counts": {
    "like": 500,
    "love": 300,
    "cheer": 200,
    "applause": 523
  },
  "start_time": "2026-01-07T18:00:00Z",
  "duration_seconds": 3600
}
```

#### Get Milestones
```bash
GET /api/sessions/milestones?session_id=<session_id>

Response:
{
  "session_id": "uuid",
  "milestones": [
    {
      "id": "milestone_1",
      "type": "total_reactions",
      "threshold": 100,
      "progress": 150,
      "achieved": true,
      "achieved_at": "2026-01-07T18:05:00Z"
    }
  ]
}
```

#### Health Check
```bash
GET /health

Response:
{
  "status": "healthy",
  "time": "2026-01-07T18:00:00Z"
}
```

### WebSocket API

Connect to `ws://localhost:8080/ws?session_id=<session_id>&user_id=<user_id>`

#### Send Reaction
```json
{
  "type": "reaction",
  "reaction_type": "like"
}
```

Available reaction types: `like`, `love`, `cheer`, `applause`, `fire`, `heart`

#### Receive Events

**Reaction Event:**
```json
{
  "type": "reaction",
  "user_id": "user123",
  "reaction_type": "like",
  "timestamp": "2026-01-07T18:00:00Z"
}
```

**Milestone Achievement:**
```json
{
  "type": "milestone_achieved",
  "milestone": {
    "id": "milestone_1",
    "type": "total_reactions",
    "threshold": 100,
    "achieved": true
  },
  "achieved_at": "2026-01-07T18:05:00Z"
}
```

## 🗄️ DynamoDB Setup

### Option 1: AWS DynamoDB

1. **Create Tables** using AWS Console or CLI:

```bash
# Events Table
aws dynamodb create-table \
  --table-name LivePulse_Events \
  --attribute-definitions \
    AttributeName=session_id,AttributeType=S \
    AttributeName=event_id,AttributeType=S \
  --key-schema \
    AttributeName=session_id,KeyType=HASH \
    AttributeName=event_id,KeyType=RANGE \
  --billing-mode PAY_PER_REQUEST

# Sessions Table
aws dynamodb create-table \
  --table-name LivePulse_Sessions \
  --attribute-definitions \
    AttributeName=session_id,AttributeType=S \
  --key-schema \
    AttributeName=session_id,KeyType=HASH \
  --billing-mode PAY_PER_REQUEST

# Milestones Table
aws dynamodb create-table \
  --table-name LivePulse_Milestones \
  --attribute-definitions \
    AttributeName=session_id,AttributeType=S \
    AttributeName=milestone_id,AttributeType=S \
  --key-schema \
    AttributeName=session_id,KeyType=HASH \
    AttributeName=milestone_id,KeyType=RANGE \
  --billing-mode PAY_PER_REQUEST
```

2. **Configure AWS Credentials**:
   - Set `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` in `.env`
   - Or use AWS CLI configuration (`aws configure`)

### Option 2: DynamoDB Local

1. **Download and Run DynamoDB Local**:
```bash
# Using Docker
docker run -p 8000:8000 amazon/dynamodb-local

# Or download from AWS
# https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/DynamoDBLocal.html
```

2. **Set Endpoint in `.env`**:
```env
DYNAMODB_ENDPOINT=http://localhost:8000
```

3. **Create Tables** (same commands as above, but add `--endpoint-url http://localhost:8000`)

## 🧪 Testing

### Manual Testing with cURL

```bash
# Create a session
curl -X POST http://localhost:8080/api/sessions \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Event","milestones":[10,50,100]}'

# Join session (replace session_id)
curl -X POST "http://localhost:8080/api/sessions/join?session_id=YOUR_SESSION_ID&user_id=user1"

# Get stats
curl "http://localhost:8080/api/sessions/stats?session_id=YOUR_SESSION_ID"

# Get milestones
curl "http://localhost:8080/api/sessions/milestones?session_id=YOUR_SESSION_ID"
```

### WebSocket Testing

Use a WebSocket client like [websocat](https://github.com/vi/websocat):

```bash
# Connect to session
websocat "ws://localhost:8080/ws?session_id=YOUR_SESSION_ID&user_id=user1"

# Send reactions (type and press Enter)
{"type":"reaction","reaction_type":"like"}
{"type":"reaction","reaction_type":"love"}
```

## 🎯 Performance Characteristics

- **Event Throughput**: 10,000+ events/second (with 10 workers)
- **Latency**: Sub-millisecond in-memory aggregation
- **Concurrency**: Thread-safe with atomic operations
- **Memory**: O(n) where n = number of active sessions
- **Graceful Shutdown**: Processes all queued events before exit

## 🛠️ Development

### Project Structure

```
LivePulse/
├── cmd/
│   └── server/          # Main application entry point
│       └── main.go
├── config/              # Configuration management
│   └── config.go
├── internal/
│   ├── aggregation/     # In-memory statistics
│   │   ├── manager.go
│   │   └── session.go
│   ├── api/             # HTTP/WebSocket handlers
│   │   ├── handlers.go
│   │   ├── middleware.go
│   │   └── websocket.go
│   ├── events/          # Event processing
│   │   ├── queue.go
│   │   ├── types.go
│   │   └── worker.go
│   ├── milestones/      # Milestone tracking
│   │   ├── tracker.go
│   │   └── types.go
│   └── storage/         # DynamoDB persistence
│       ├── dynamodb.go
│       ├── events.go
│       ├── milestones.go
│       └── sessions.go
├── .env.example         # Example configuration
├── .gitignore
├── go.mod
└── README.md
```

### Building

```bash
# Build for current platform
go build -o livepulse.exe cmd/server/main.go

# Build for Linux
GOOS=linux GOARCH=amd64 go build -o livepulse cmd/server/main.go

# Build with optimizations
go build -ldflags="-s -w" -o livepulse.exe cmd/server/main.go
```

## 📝 License

MIT License - feel free to use this for your projects!

## 🤝 Contributing

Contributions are welcome! Feel free to open issues or submit pull requests.

## 📧 Support

For questions or issues, please open a GitHub issue.

---

**Built with ❤️ using Go**
