# LivePulse

LivePulse is a web application for discovering upcoming concerts, sports, and entertainment events and joining a real-time chat room for each one. It pairs a Next.js frontend with a Go API and WebSocket server, using PostgreSQL for event data and Redis for chat history.

**Live site:** [livepulse-hq.vercel.app](https://livepulse-hq.vercel.app)

<img width="1597" height="992" alt="image" src="https://github.com/user-attachments/assets/5675c721-0b8f-46e3-bfb3-f1eb723ff953" />

---

## Features

- **Event discovery**: Browse Ticketmaster events happening in the next 24 hours, search by keyword, filter by type, country, and favorites, and load more results on demand.
- **Live event rooms**: Each event has a chat room backed by WebSockets. New joiners receive the last 100 messages so they can follow the conversation.
- **Favorites**: Signed-in users can save events; favorites are tied to their verified Clerk identity.
- **Moderation and limits**: Chat messages are profanity-filtered on the server and capped at 500 characters.
- **Resilient connections**: The client reconnects with exponential backoff and surfaces a manual reconnect control after repeated failures.
- **Installable**: A web app manifest provides install metadata and standalone display. There is no service worker or offline caching.

---

## Tech Stack

| Area | Technologies |
| --- | --- |
| Frontend | Next.js 16 (App Router), React 19, TypeScript, Tailwind CSS 4, shadcn, Framer Motion |
| Backend | Go 1.25, `gorilla/websocket`, `robfig/cron`, `TwiN/go-away` |
| Storage | PostgreSQL via `pgx` (Neon in production), Redis via `go-redis` (Upstash in production) |
| Auth and data | Clerk, Ticketmaster Discovery API |
| Hosting | Vercel (frontend), Northflank (backend) |
| Infrastructure | Docker, K3s/K3d local Kubernetes lab |
| Testing and CI | Jest, Testing Library, Go test, Testify, GitHub Actions, CodeQL, SonarCloud |

---

## Architecture

```text
 Browser (Next.js on Vercel)
   |  HTTPS: /api/events, /api/favorites, /api/sessions/stats
   |  WSS:   /ws?session_id=<event id>
   v
 Go API + WebSocket server
   |-- bounded event queue -> worker pool -> session hubs (broadcast)
   |-- PostgreSQL: events, users, favorites
   |-- Redis: per-room chat history with TTL
   '-- cron: Ticketmaster ingestion every 6 hours
```

The repository is a monorepo with separate `frontend` and `backend` applications.

### Go backend and WebSockets

- **Session hubs**: Each active room has an in-memory hub that owns its connected clients and broadcasts chat messages, reactions, and milestone events. Slow consumers are evicted instead of blocking the hub.
- **Event processing**: Incoming WebSocket events are placed on a bounded queue and processed by a configurable worker pool (`WORKER_COUNT`, `EVENT_QUEUE_SIZE`).
- **Connection lifecycle**: The server uses read limits, read deadlines, ping/pong handling, buffered outbound channels, and graceful shutdown that drains the worker pool.
- **Reactions and milestones**: The backend aggregates reaction counts per session and broadcasts milestone notifications when thresholds are reached. The current frontend uses chat and occupancy; it does not yet send reactions.

Hubs and aggregation state are local to a single backend process. There is no distributed fan-out or shared connection state across instances, so the backend runs as a single replica.

### Redis chat storage

Chat messages are stored in Redis lists with `RPUSH` and capped at 500 messages per room with `LTRIM`, both in one transaction pipeline. Each chat key expires one hour after the event's end time (about 25 hours for rooms with no matching event).

### PostgreSQL and Ticketmaster ingestion

PostgreSQL stores events, Clerk user IDs, and favorites, accessed through a `pgx` connection pool. The schema is created on startup.

- **Scheduled ingestion**: On startup and every six hours, the backend requests up to two 200-event pages from Ticketmaster for selected categories within the next 24 hours.
- **Search ingestion**: Searches also pull matching events from Ticketmaster into the database.
- **Event duration**: Ticketmaster results often lack an end time, so imported events use a three-hour duration.
- **Cleanup**: After each scheduled run, events that ended more than one hour earlier are deleted.

### Next.js frontend

- The events route is server-rendered and fetches its initial data from the Go API.
- Search terms live in the `q` URL parameter; type, country, and favorite filters are applied in client state.
- Event cards and event pages show live room occupancy; counts for all visible cards are fetched in one batched stats request.
- Help & FAQ and Privacy pages document product behavior and the implemented data flows.

### Authentication and security

- Clerk handles frontend authentication; the Go backend verifies Clerk JWTs. Endpoint requirements are listed in the API overview below.
- After opening a WebSocket, the client sends an `authenticate` message with its Clerk token, keeping the token out of the URL. Connections that do not authenticate within 15 seconds are closed.
- HTTP and WebSocket origins are allowlisted via `CORS_ALLOWED_ORIGINS` and `WS_ALLOWED_ORIGINS`. Production startup fails fast unless TLS database URLs, secrets, and origin allowlists are configured; see `backend/.env.example`.
- The frontend sets baseline security headers (`nosniff`, frame denial, referrer and permissions policies).

### API overview

| Endpoint | Auth | Purpose |
| --- | --- | --- |
| `GET /health` | None | Liveness check |
| `GET /health/ready` | None | Pings PostgreSQL and Redis |
| `GET /api/events` | Optional | Upcoming events, with `q` and `offset` |
| `GET /api/events/single?id=` | None | Single event |
| `/api/favorites` | Required | List or toggle favorites |
| `GET /api/sessions/stats` | None | Room stats by `session_id` or batched `session_ids` |
| `GET /api/sessions/milestones` | None | Milestones for a session |
| `POST /api/sessions`, `/api/sessions/join` | Required | Create or join a session |
| `/ws?session_id=` | Clerk token via first message | Real-time room |

---

## WebSocket Benchmark

The backend includes a WebSocket benchmark CLI (`backend/cmd/wsbench`) that measures Clerk authentication handshake latency and chat round-trip latency. One recorded run with 10 concurrent clients sending 50 messages each:

| Metric | p50 | p95 | p99 |
| --- | --- | --- | --- |
| Auth handshake | 279.5 ms | 286.8 ms | 286.8 ms |
| Chat round trip | 199.2 ms | 250.5 ms | 277.3 ms |

All 500 messages succeeded with no client failures. Full output is in [BENCHMARK.md](BENCHMARK.md).

To run it, from the `backend` directory:

```powershell
$env:LIVEPULSE_WS_TOKEN="your_clerk_jwt"
go run .\cmd\wsbench -url ws://localhost:8080 -session wsbench-local -clients 10 -messages 20 -warmup 2
```

Pass a `wss://` URL to benchmark a deployed backend. Other flags include `-origin`, `-timeout`, and `-author`.

---

## Getting Started

### Prerequisites

- Node.js 22 and npm
- Go 1.25 or newer
- PostgreSQL and Redis (local instances, the [K3d lab](#local-kubernetes-lab-k3d), or Neon and Upstash)
- A Clerk application and a Ticketmaster API key

### Configure environment

Copy the example files and fill in your values:

- `backend/.env.example` to `backend/.env`
- `frontend/.env.local.example` to `frontend/.env.local`

The backend defaults to port `8080` when `SERVER_PORT` is unset, and the frontend defaults to `http://localhost:8080` when `NEXT_PUBLIC_API_URL` is unset. Keep these aligned. `NEXT_PUBLIC_` values are embedded in the frontend at build time, so set production URLs in the deployment environment.

### Run locally

```powershell
# Backend (loads backend/.env)
cd .\backend
go run .\cmd\server

# Frontend, in another terminal
cd .\frontend
npm install
npm run dev
```

Open `http://localhost:3000`.

### Tests and lint

From `frontend`:

```powershell
npm run lint
npm run test-frontend
npm run test-backend   # runs go test ./... -v in backend
```

---

## Local Kubernetes Lab (K3d)

An optional local lab runs the Go backend, PostgreSQL, and Redis in a K3s cluster managed by K3d, using the `k3d-livepulse-lab` context. Docker Desktop supplies the container engine; its built-in Kubernetes is not needed. The lab uses in-cluster databases and never connects to the production Neon or Upstash endpoints.

See [k8s/README.md](k8s/README.md) for first-time cluster creation, local Secrets, and manifests.

### Start the lab

With Docker Desktop running, from the repository root in PowerShell:

```powershell
k3d cluster start livepulse-lab
kubectl config use-context k3d-livepulse-lab
kubectl rollout status deployment/livepulse-backend -n livepulse --timeout=180s
kubectl port-forward service/livepulse-backend 18081:8080 -n livepulse
```

Point the frontend at the lab in `frontend/.env.local`, then run `npm run dev`:

```dotenv
NEXT_PUBLIC_API_URL=http://localhost:18081
NEXT_PUBLIC_WS_URL=ws://localhost:18081
```

`curl.exe -i http://127.0.0.1:18081/health/ready` should report PostgreSQL and Redis as `ok`.

### Rebuild the backend image

```powershell
docker build -t livepulse-backend:dev -f backend/Dockerfile backend
k3d image import livepulse-backend:dev --cluster livepulse-lab
kubectl rollout restart deployment/livepulse-backend -n livepulse
kubectl rollout status deployment/livepulse-backend -n livepulse --timeout=180s
```

### Stop and resume

Stop the frontend and port-forward with `Ctrl+C`, then run `k3d cluster stop livepulse-lab`. PVC data survives a stop and start. Deleting the cluster or the `livepulse` namespace can delete lab data; to recreate it, follow the first-time steps in [k8s/README.md](k8s/README.md).

---

## Continuous Integration

GitHub Actions workflows in `.github/workflows`:

- **CI**: Frontend lint, Jest tests, and production build; backend `go build`, `go vet`, and `go test`.
- **CodeQL**: Static security analysis.
- **Dependency review**: Flags vulnerable dependency changes in pull requests.
- **SonarCloud**: Code quality analysis.

---

## Project Structure

```text
backend/
  cmd/server/       API and WebSocket server entry point
  cmd/wsbench/      WebSocket benchmark CLI
  cmd/testclient/   Manual API test client
  config/           Environment configuration and production validation
  internal/api/     HTTP handlers, middleware, Clerk auth, WebSocket hubs
  internal/events/  Event queue, worker pool, Ticketmaster fetcher
  internal/aggregation/  Per-session stats
  internal/milestones/   Milestone tracking
  internal/storage/ PostgreSQL and Redis clients
frontend/
  src/app/          Next.js routes and page components
  src/hooks/        WebSocket and favorites hooks
  src/components/ui/  shadcn components
k8s/                Local K3d manifests and secret helper
```
