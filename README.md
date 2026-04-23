# LivePulse

**LivePulse** is a high-performance, real-time engagement platform designed to act as a "live chat sidekick" for concerts, sports, and live television events. Users can seamlessly join an event room and engage in blazing-fast, decentralized chat with other attendees across the globe.

---

## 🏗️ System Architecture

LivePulse is constructed as a monorepo to maximize deployment velocity while decoupling the frontend and backend scaling concerns.

### 1. The Real-Time Engine (Go Backend)
At the core of the application is a high-availability **Go (Golang)** WebSockets server. 
- **Why Go?** Live events necessitate handling thousands of concurrent, active TCP connections. Go's native goroutines handle this with staggeringly low memory overhead compared to traditional Node.js/Python thread implementations.
- **WebSocket Hubs**: The Go server utilizes an efficient publisher/subscriber `Hub` model. When a user sends a chat, a dedicated goroutine parses the payload, applies robust JWT authentication verification, and broadcasts the message synchronously to all connected clients in the same `Session_ID`.

### 2. High-Speed Ephemeral Storage (Redis)
Chat messages are fired at a phenomenal rate during live events, representing a massive write-load.
- **Upstash (Serverless Redis)**: Rather than hammering a relational database with millions of ephemeral chat row-inserts, LivePulse leverages **Redis Lists**. Chat payloads are appended to an in-memory chronological list (`RPUSH`).
- **TTL Garbage Collection**: The application enforces a strict **1-Hour Time-To-Live (TTL)**. Exactly 60 minutes after a concert or sporting event concludes, Redis autonomously flushes the entire chat history. This prevents infinitely expanding memory queues and dramatically cuts cloud infrastructure overhead.

### 3. Persistent Data Layer (PostgreSQL)
- **Neon (Serverless Postgres)**: While chats are ephemeral, users and events are persistent. LivePulse uses Neon to reliably map Clerk User IDs and securely store Event start/end times.
- **Top-Tier Global Ingestion**: The Go backend utilizes an automated foreground/background fetch pooling constraint that pulls the 400 absolutely most relevant global events occurring within the next 24 hours across Ticketmaster globally!
- **On-Demand Search Engine**: The platform implements an intercepted infinite search. When a user queries for an obscure event currently outside the top 400, the Go Engine intercepts the HTTP request, hits Ticketmaster directly, and permanently weaves the unique event straight into the Postgres Database on the fly!
- **Automated DB Garbage Collection**: The Go backend deploys an automated background `Worker` pool via `robfig/cron` every 6 hours. Aside from replenishing the feed, this cron explicitly executes a powerful Garbage Collector query on the Neon database, automatically shredding any events that concluded more than 1 hour ago to seamlessly preserve free tier constraints natively! An initial fetch also runs on server startup so events are available immediately.

### 4. Interactive Client (Next.js Frontend)
- **Next.js 15 App Router**: The client maps heavily to React Server Components (RSC) when rendering the Event Dashboard, passing control off to Client Components exclusively for the Live Chat Arena.
- **Pagination & Query Mapping**: The platform supports infinite scrolling using native `OFFSET` database bounds synchronized to explicit URL query strings (`?q=...&offset=...`) allowing for copy-pasteable dashboard states.
- **Progressive Web App (PWA)**: Designed primarily for users holding their mobile phones at a concert, the frontend generates native `manifest.json` headers to allow consumers to install it seamlessly to their iOS or Android home screens.
- **Premium Aesthetics & Defenses**: Built via TailwindCSS, `shadcn/ui`, and Framer Motion to create a highly stylized layout prioritizing premium UI. React hooks (`useWebSocket`) maintain sub-millisecond sync with the Go array and actively intercept any Go backend Payload Refusal metrics dynamically transforming them into sleek "Toast" bounds to stop UX breaking.

### 5. Security Handshake (Clerk)
- All connections are guarded by **Clerk Auth**. The client establishes a WebSocket connection and immediately sends a `{ type: "authenticate", token }` message as its first payload over the encrypted channel. The Go server verifies the JWT before accepting the client into the session Hub — preventing token leakage in URLs, logs, and reverse proxies.
- **Origin Allowlist**: The WebSocket upgrader enforces a strict CORS origin check, only permitting connections from whitelisted production domains.
- **Auto-Reconnect**: The frontend implements exponential backoff reconnection (capped at 30s) to gracefully recover from temporary network drops — critical for mobile users at live events.

---

## 💻 Tech Stack Summary

**Frontend:** Next.js, React 19, TypeScript, TailwindCSS, shadcn/ui, Framer Motion
**Backend:** Go (Golang), WebSockets (`gorilla/websocket`), Go-Cron
**Databases:** Neon (PostgreSQL), Upstash (Redis)
**Authentication:** Clerk JWT Middleware
**Testing:** Jest, Go Test (Testify)
