# Architecture Decisions & Interview Talking Points

This document tracks design decisions and how to honestly discuss them in interviews.

---

## 1. Shared SQLite Database

**Decision:** All services share a single SQLite database instead of each owning its own data store.

**Why:** Simplicity for a demo project. SQLite is file-based, no server needed, easy to deploy on Raspberry Pi.

**Tradeoff:** This is an anti-pattern for true microservices. Services are coupled at the data layer.

**How to talk about it:**
> "I chose a shared SQLite database for simplicity and to reduce infrastructure overhead on my Raspberry Pi. In a production microservices setup, each service would own its data store and communicate via APIs or events. This would provide better isolation and independent scalability, but adds operational complexity that wasn't necessary for this demo."

---

## 2. Embedded Auth vs Separate Auth Service

**Decision:** Auth is embedded in chat-api rather than being a separate service.

**Why:** Reduces number of processes running on Pi. Auth is tightly coupled to chat anyway (JWT validation happens on every WebSocket message).

**Tradeoff:** Less "micro" in microservices. Can't scale auth independently.

**How to talk about it:**
> "I combined auth with the chat API because they share the same scaling characteristics - every chat request needs auth validation anyway. In a larger system with different scaling needs (e.g., high login traffic during peak hours), I'd separate them. The code is structured so extraction would be straightforward."

---

## 3. Composable Config Pattern

**Decision:** Small config structs (DBConfig, RedisConfig, etc.) that compose into a full Config, with individual loaders.

**Why:** Each service only requires the config it needs. Migrate doesn't crash because JWT_SECRET isn't set.

**Tradeoff:** Slightly more boilerplate than a flat config struct.

**How to talk about it:**
> "I designed the config system to be composable so each service fails fast only on its actual dependencies. This follows the principle of least privilege for configuration and makes it clear what each service requires just by looking at its main.go."

---

## 4. SQLite over PostgreSQL

**Decision:** Using SQLite instead of PostgreSQL.

**Why:** Simpler deployment (no database server), sufficient for demo scale, reduces resource usage on Raspberry Pi.

**Tradeoff:** No concurrent write scaling, limited to single-node deployment.

**How to talk about it:**
> "SQLite was the right choice for this project's scale and deployment target. It eliminates a network hop and a separate process. For a production system with write-heavy loads or multiple app servers, I'd use PostgreSQL. The repository layer abstracts this - switching databases would only require changing the driver and connection logic."

---

## 5. Redis for Pub/Sub + Rate Limiting

**Decision:** Using Redis for message broadcasting AND rate limiting, not just queuing.

**Why:** Already have Redis for the moderation queue. Reusing it for rate limiting avoids adding another dependency.

**Tradeoff:** Redis becomes a critical dependency for multiple features.

**How to talk about it:**
> "Redis serves dual purposes: message pub/sub for real-time broadcasting and rate limiting to protect the infrastructure. This is a practical consolidation - Redis is already required for the moderation queue, so leveraging it for rate limiting adds capability without new dependencies. If Redis fails, both chat and rate limiting degrade, but for this scale that's an acceptable tradeoff."

---

## 6. Rate Limiting at Multiple Levels

**Decision:** Rate limiting on chat messages (10/10s), auth attempts (5/15min), and Perspective API calls (1 QPS).

**Why:** Protects both internal infrastructure (Raspberry Pi) and respects external API limits.

**Tradeoff:** Adds latency (Redis roundtrip per request).

**How to talk about it:**
> "I implemented rate limiting at three levels: user messages to prevent spam on my constrained hardware, login attempts for brute force protection, and API calls to respect Google's Perspective API quota. Each has different thresholds based on its use case. The Redis overhead is minimal compared to the protection it provides."

---

## 7. Pure Go SQLite Driver (modernc.org/sqlite)

**Decision:** Using pure Go SQLite driver instead of CGO-based mattn/go-sqlite3.

**Why:** No C compiler needed, easier cross-compilation, simpler build process.

**Tradeoff:** Slightly slower than native C implementation.

**How to talk about it:**
> "I chose the pure Go SQLite driver for build simplicity - no CGO means easier cross-compilation to ARM for the Raspberry Pi and simpler CI/CD. The performance difference is negligible at this scale."

---

## 8. Microservices-Ready Architecture

**Decision:** Structure code with clear service boundaries and repository interfaces, even though we use a shared database today.

**Why:** Allows easy evolution to true microservices without rewriting business logic.

**Design principles:**
- Repository pattern with interfaces (swap SQLite impl for API client later)
- Each domain only touches its own tables (auth → users, chat → rooms/messages, moderation → moderation_logs)
- No cross-boundary database joins
- Event-driven communication via Redis pub/sub

**Tradeoff:** Slightly more abstraction upfront. Could be seen as over-engineering for a demo.

**How to talk about it:**
> "I designed the system with clear service boundaries using the repository pattern. Each domain only accesses its own tables through interfaces. Today those interfaces are backed by SQLite queries, but to evolve to true microservices, I'd simply implement those same interfaces as HTTP/gRPC clients to separate services. The business logic wouldn't change at all."

**Evolution path:**
```
Today:                          Tomorrow:
┌─────────────────────┐         ┌─────────────────────┐
│ UserRepository      │         │ UserRepository      │
│ (SQLite impl)       │   →     │ (HTTP client impl)  │
└──────────┬──────────┘         └──────────┬──────────┘
           │                               │
     ┌─────┴─────┐                   ┌─────┴─────┐
     │  SQLite   │                   │ Auth API  │
     └───────────┘                   │ (own DB)  │
                                     └───────────┘
```

---

## Adding New Decisions

When you make a decision, add it here with:
1. **Decision:** What you chose
2. **Why:** Your reasoning
3. **Tradeoff:** What you gave up
4. **How to talk about it:** Honest framing for interviews
