# Go Chat Moderator

A real-time chat application with AI-powered content moderation using Go, React, and Mistral AI.

## Features

- Real-time messaging via WebSocket
- AI content moderation (Mistral AI)
- User authentication (JWT)
- Multi-room support
- Cross-instance messaging (Redis pub/sub)
- React TypeScript frontend

## Quick Start

```bash
# 1. Clone and setup
git clone https://github.com/mr1hm/go-chat-moderator.git
cd go-chat-moderator
cat > .env << EOF
PORT=:8080
DB_PATH=data/chat.db
REDIS_ADDR=localhost:6379
JWT_SECRET=secret
MISTRALAI_API_KEY=your-mistral-api-key
EOF

# 2. Start Redis
docker-compose up -d

# 3. Run migrations
go run ./cmd/migrate

# 4. Start backend (2 terminals)
export $(cat .env | xargs) && go run ./cmd/api
export $(cat .env | xargs) && go run ./cmd/moderation-service

# 5. Start frontend
cd frontend && npm install && npm run dev
```

Open http://localhost:5173

## Docker Deployment

Deploy everything with a single command:

```bash
# Set your Mistral API key
export MISTRALAI_API_KEY=your-mistral-api-key

# Build and run
docker-compose up -d --build
```

Open http://localhost:3000

This starts:
- **Redis** - message queue and pub/sub
- **Backend** - API server + moderation worker
- **Frontend** - React app served via nginx

## Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Frontend  │────▶│   Backend   │────▶│   SQLite    │
│   (React)   │◀────│    (Gin)    │     │             │
└─────────────┘     └──────┬──────┘     └─────────────┘
                           │
                     ┌─────▼─────┐
                     │   Redis   │
                     │  Pub/Sub  │
                     └─────┬─────┘
                           │
                     ┌─────▼─────┐
                     │ Moderation│
                     │  Worker   │
                     └─────┬─────┘
                           │
                     ┌─────▼─────┐
                     │ Mistral AI│
                     └───────────┘
```

## Design Decisions

| Decision | Why |
|----------|-----|
| **SQLite over PostgreSQL** | No server needed, file-based, simple deployment. Repository pattern allows easy swap to Postgres later. |
| **Redis for pub/sub + queue** | Already needed for moderation queue, reuse for real-time broadcasting avoids extra dependencies. |
| **Embedded auth in API** | Auth and chat share scaling characteristics - every chat request needs JWT validation anyway. |
| **Repository interfaces** | Business logic depends on abstractions, not implementations. Swap SQLite for Postgres or mocks without changing code. |
| **Broadcast-first, persist async** | Real-time UX is priority. Messages broadcast instantly via Redis, then saved to DB in a goroutine. |
| **Pure Go SQLite driver** | No CGO means easier cross-compilation and simpler builds. |

## Tech Stack

**Backend:**
- Go 1.25+
- Gin (HTTP framework)
- Gorilla WebSocket
- SQLite (database)
- Redis (pub/sub & queue)
- JWT (authentication)

**Frontend:**
- React 19
- TypeScript
- Vite
- React Router

**AI:**
- Mistral AI (content moderation)

## Prerequisites

- Go 1.25+
- Node.js 20+
- Docker (for Redis)
- Mistral AI API key

## Setup

### 1. Clone and configure

```bash
git clone https://github.com/mr1hm/go-chat-moderator.git
cd go-chat-moderator

# Create .env file
cat > .env << EOF
JWT_SECRET=your-secret-key-here
MISTRAL_API_KEY=your-mistral-api-key
EOF
```

### 2. Start Redis

```bash
docker-compose up -d
```

### 3. Run the backend

```bash
# Terminal 1: API server
export $(cat .env | xargs) && go run ./cmd/api

# Terminal 2: Moderation worker
export $(cat .env | xargs) && go run ./cmd/moderation-service
```

### 4. Run the frontend

```bash
cd frontend
npm install
npm run dev
```

### 5. Open the app

Visit http://localhost:5173

## Project Structure

```
.
├── cmd/
│   ├── api/                 # HTTP server & WebSocket
│   ├── migrate/             # Database migrations
│   └── moderation-service/  # AI moderation worker
├── internal/
│   ├── auth/                # JWT authentication
│   ├── chat/                # Chat logic, hub, client
│   ├── moderation/          # Mistral AI integration
│   └── shared/
│       ├── redis/           # Redis client
│       └── sqlite/          # Database setup
├── frontend/
│   └── src/
│       ├── api/             # API client
│       ├── components/      # React components
│       ├── hooks/           # Custom hooks
│       ├── pages/           # Page components
│       └── types/           # TypeScript types
├── data/                    # SQLite database
└── docker-compose.yml       # Redis container
```

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/register` | Register new user |
| POST | `/login` | Login, returns JWT |
| GET | `/rooms` | List all rooms |
| POST | `/rooms` | Create a room |
| GET | `/rooms/:id/messages` | Get room messages |
| WS | `/ws/:roomId` | WebSocket connection |

## WebSocket Messages

**Incoming (from server):**
```json
{"type": "message", "payload": {"id": "...", "content": "...", "moderation_status": "pending"}}
{"type": "moderation_update", "payload": {"message_id": "...", "status": "approved"}}
```

**Outgoing (to server):**
```json
{"content": "Hello world"}
```

## Moderation Flow

1. User sends message via WebSocket
2. Message saved with `pending` status
3. Message queued in Redis for moderation
4. Worker sends content to Mistral AI
5. If toxic (score > 0.7), status = `flagged`
6. Update broadcast to all room clients
7. Frontend hides flagged messages

## Scaling to Microservices

The backend is designed for easy conversion to a microservice architecture. Here's how each component is already decoupled:

### Current Architecture Benefits

| Component | Why It's Ready |
|-----------|----------------|
| **API Server** | Stateless - no in-memory session state, JWT for auth |
| **Moderation Worker** | Already a separate process, communicates via Redis queue |
| **Redis Pub/Sub** | Handles cross-instance WebSocket broadcasts |
| **Redis Queue** | Multiple workers can compete for moderation jobs |

### Step-by-Step Conversion

**1. Replace SQLite with PostgreSQL/MySQL**

```go
// internal/shared/postgres/db.go
import "database/sql"
import _ "github.com/lib/pq"

var DB *sql.DB

func Init(connStr string) error {
    var err error
    DB, err = sql.Open("postgres", connStr)
    return err
}
```

Update connection string in `.env`:
```
DATABASE_URL=postgres://user:pass@localhost:5432/chatmod
```

**2. Containerize Each Service**

```dockerfile
# Dockerfile.api
FROM golang:1.25-alpine
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o api ./cmd/api
CMD ["./api"]
```

```dockerfile
# Dockerfile.moderation
FROM golang:1.25-alpine
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o worker ./cmd/moderation-service
CMD ["./worker"]
```

**3. Scale with Docker Compose**

- Add `deploy.replicas: 3` to run multiple API instances
- Add `deploy.replicas: 2` for multiple moderation workers
- Workers automatically compete for queue items (no code changes needed)

**4. Add Load Balancer**

Use nginx or Traefik in front of API instances:
- Enable `ip_hash` for sticky sessions (required for WebSocket)
- Proxy WebSocket upgrade headers

**5. Extract Services Further (Optional)**

For full microservices, split into separate repos/modules:

```
chat-platform/
├── api-gateway/          # Auth, routing, rate limiting
├── chat-service/         # WebSocket handling, message storage
├── moderation-service/   # AI moderation (already separate)
├── user-service/         # User management, profiles
└── shared/               # Protobuf definitions, shared types
```

Use gRPC or message queues for inter-service communication.

## License

MIT
