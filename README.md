# LivePulse

LivePulse is a web application for discovering upcoming concerts, sports, and entertainment events and joining event-specific chat rooms. The application combines a Next.js frontend with a Go API and WebSocket server.

---

## System Architecture

LivePulse is organized as a monorepo with separate `frontend` and `backend` applications.

### 1. Go Backend and WebSockets

The Go backend provides HTTP endpoints and a WebSocket endpoint for event rooms.

- **Session hubs**: Each active session has an in-memory hub that tracks connected clients and broadcasts chat messages, reactions, and milestone events.
- **Event processing**: Incoming WebSocket events are placed on a bounded queue and processed by a configurable worker pool.
- **Connection lifecycle**: The server uses read deadlines, ping/pong handling, buffered outbound channels, and graceful HTTP shutdown.

The session hubs and aggregation state are local to a single backend process. The current implementation does not provide distributed fan-out or shared connection state across multiple backend instances.

### 2. Redis Chat Storage

The backend stores chat messages in Redis lists using `RPUSH` and retains at most 500 messages per session with `LTRIM`; both run in a transaction pipeline.

Each chat key is scheduled for deletion one hour after the event's end time via `SetChatTTL`. When a client authenticates over WebSocket, the server replays the last 100 persisted messages so new joiners can follow the conversation.

### 3. PostgreSQL and Ticketmaster Data

PostgreSQL stores events, Clerk user IDs, and user favorites. The database client uses `pgx` connection pooling.

- **Scheduled ingestion**: On startup and every six hours, the backend requests up to two 200-event pages from Ticketmaster for selected categories within the next 24 hours.
- **Search ingestion**: A non-empty event search requests up to 50 Ticketmaster results and upserts returned events into PostgreSQL before querying the database.
- **Event duration**: Because Ticketmaster results may not include an end time, imported events currently use a three-hour duration from their start time.
- **Cleanup**: After a scheduled ingestion run, events whose stored end time passed more than one hour earlier are deleted.

### 4. Next.js Frontend

The frontend uses the Next.js 16 App Router with React 19 and TypeScript.

- The events route is server-rendered and fetches its initial data from the Go API.
- Search terms are stored in the `q` URL parameter. Type, country, and favorite filters are applied in client state.
- Additional event results are requested through a **Load more events** control using an `offset` query parameter.
- A web app manifest provides install metadata and standalone display settings. The project does not currently include a service worker or offline caching.
- Tailwind CSS and shadcn-based components provide the broadcast-inspired interface styling.
- Help & FAQ and Privacy routes document product behavior and the application’s implemented data flows.

### 5. Authentication and Connection Controls

Clerk provides frontend authentication and JWT verification in the Go backend.

- After opening a WebSocket, the client sends an `authenticate` message containing its Clerk token. The server registers the client with the session only after verifying that token, and closes connections that do not authenticate within 15 seconds.
- Sending the token in the first WebSocket message keeps it out of the connection URL. Transport encryption depends on using `wss://` in deployment.
- The WebSocket upgrader accepts origins from `WS_ALLOWED_ORIGINS` (comma-separated). When unset it falls back to `http://localhost:3000`, `https://livepulse-hq.vercel.app`, and clients that omit the `Origin` header.
- The client reconnects after a closed connection using exponential backoff capped at 30 seconds, up to 6 attempts before surfacing a manual reconnect control.
- The favorites API and session create/join endpoints require a Clerk bearer token. Event listing is public but accepts an optional token to decorate results with the caller's favorites.
- HTTP CORS follows `CORS_ALLOWED_ORIGINS`; when unset it allows all origins for local development. `GET /health` is a liveness check and `GET /health/ready` pings PostgreSQL and Redis for readiness probes.
- `POST /api/admin/trigger-fetch` triggers a manual Ticketmaster ingestion. It accepts `POST` only, is throttled to one trigger per minute, requires `X-Admin-Key` when `ADMIN_API_KEY` is configured, and never overlaps an in-progress run.
- Setting `APP_ENV=production` makes startup fail fast unless `CLERK_SECRET_KEY`, `EXTERNAL_API_KEY`, a TLS `DATABASE_URL`, a `rediss://` `REDIS_URL`, and both origin allowlists are configured.

---

## Tech Stack

- **Frontend:** Next.js 16, React 19, TypeScript, Tailwind CSS, shadcn, Framer Motion
- **Backend:** Go, `gorilla/websocket`, `robfig/cron`
- **Storage:** PostgreSQL through `pgx`, Redis through `go-redis`
- **External services:** Clerk, Ticketmaster, Neon PostgreSQL, Upstash Redis
- **Testing:** Jest, Testing Library, Go test, Testify

---

## WebSocket Benchmarking

The backend includes a standalone WebSocket benchmark CLI that measures Clerk authentication handshake latency and chat-message round-trip latency. [BENCHMARK.md](BENCHMARK.md) contains one recorded benchmark.

Run the benchmark from the backend directory:

```powershell
cd .\backend
$env:LIVEPULSE_WS_TOKEN="your_clerk_jwt"
go run .\cmd\wsbench -url ws://localhost:8080 -session wsbench-local -clients 10 -messages 20 -warmup 2
```

For a deployed backend, pass its WebSocket base URL:

```powershell
cd .\backend
$env:LIVEPULSE_WS_TOKEN="your_clerk_jwt"
go run .\cmd\wsbench -url wss://your-backend.example.com -session wsbench-prod -clients 10 -messages 20 -warmup 2
```
