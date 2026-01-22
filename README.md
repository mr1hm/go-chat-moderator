# Go Chat Moderator

A real-time chat application with AI-powered content moderation using Go, React, and Mistral AI.

## Features

- Real-time messaging via WebSocket
- AI content moderation (Mistral AI)
- User authentication (JWT)
- Multi-room support
- Cross-instance messaging (Redis pub/sub)
- React TypeScript frontend

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

## Tech Stack

**Backend:**
- Go 1.21+
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

- Go 1.21+
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

## License

MIT
