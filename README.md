# LivePulse

LivePulse is a web application for discovering upcoming concerts, sports, and entertainment events and joining event-specific chat rooms. The application combines a Next.js frontend with a Go API and WebSocket server.

---

## Local Kubernetes Lab (K3d)

The local lab runs the Go backend, PostgreSQL, and Redis in a K3s cluster managed by K3d. Docker Desktop supplies the Docker Engine that runs the K3d node containers and builds the backend image. This setup uses the `k3d-livepulse-lab` Kubernetes context; Docker Desktop's optional built-in Kubernetes cluster is separate and is not needed.

See [k8s/README.md](k8s/README.md) for first-time cluster setup, local Secrets, database and Redis manifests, and deployment details. The lab avoids the production Neon and Upstash endpoints.

### Start the lab

Start Docker Desktop and wait for its Linux container engine to be ready. Then run these commands from the repository root in PowerShell:

```powershell
docker info --format '{{.OSType}}' # should print linux
k3d cluster start livepulse-lab
kubectl config use-context k3d-livepulse-lab
kubectl get nodes
kubectl get pods,services,pvc -n livepulse
kubectl rollout status deployment/livepulse-backend -n livepulse --timeout=180s
```

Keep the port-forward running in its own PowerShell window:

```powershell
kubectl port-forward service/livepulse-backend 18081:8080 -n livepulse
```

For local frontend development, set or update these values in `frontend/.env.local` while keeping the existing Clerk development credentials:

```dotenv
NEXT_PUBLIC_API_URL=http://localhost:18081
NEXT_PUBLIC_WS_URL=ws://localhost:18081
```

These localhost values are for `npm run dev`. Configure the production API and WebSocket URLs in the deployment environment; `NEXT_PUBLIC_` values are embedded in the frontend when it is built.

In another PowerShell window, run the Next.js development server:

```powershell
cd .\frontend
npm run dev
```

Open `http://localhost:3000`. To verify backend and dependency health independently, run `curl.exe -i http://127.0.0.1:18081/health/ready`; the response should report PostgreSQL and Redis as `ok`.

### Stop and resume

Stop the frontend and port-forward with `Ctrl+C` in their PowerShell windows. To stop the K3d cluster and free its compute resources while retaining it for later use:

```powershell
k3d cluster stop livepulse-lab
```

Start it again later with `k3d cluster start livepulse-lab`, then select the `k3d-livepulse-lab` context and repeat the port-forward and frontend steps. The local PVC data is intended to survive a cluster stop/start. Deleting the cluster or the `livepulse` namespace can delete the lab data. If the cluster has been deleted, create it with `k3d cluster create livepulse-lab` and follow the first-time deployment steps in [k8s/README.md](k8s/README.md).

### Rebuild the backend image

After changing backend code, build and import the image into the K3d cluster, then restart the Deployment to use it:

```powershell
docker build -t livepulse-backend:dev -f backend/Dockerfile backend
k3d image import livepulse-backend:dev --cluster livepulse-lab
kubectl rollout restart deployment/livepulse-backend -n livepulse
kubectl rollout status deployment/livepulse-backend -n livepulse --timeout=180s
```

The current backend keeps WebSocket hubs and aggregation state in process memory, so the lab uses one backend replica.

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
