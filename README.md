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
- **Automated Data Ingestion**: The Go backend utilizes `robfig/cron` to deploy an automated background `Worker` pool at 2:00 AM every day. This autonomously hits external APIs (e.g. Ticketmaster), parses incoming daily events, and commits them to the Postgres storage schema automatically.

### 4. Interactive Client (Next.js Frontend)
- **Next.js 15 App Router**: The client maps heavily to React Server Components (RSC) when rendering the Event Dashboard, passing control off to Client Components exclusively for the Live Chat Arena.
- **Pagination & Query Mapping**: The platform supports infinite scrolling using native `OFFSET` database bounds synchronized to explicit URL query strings (`?q=...&offset=...`) allowing for copy-pasteable dashboard states.
- **Progressive Web App (PWA)**: Designed primarily for users holding their mobile phones at a concert, the frontend generates native `manifest.json` headers to allow consumers to install it seamlessly to their iOS or Android home screens.
- **Premium Aesthetics & Defenses**: Built via TailwindCSS, `shadcn/ui`, and Framer Motion to create a highly stylized layout prioritizing premium UI. React hooks (`useWebSocket`) maintain sub-millisecond sync with the Go array and actively intercept any Go backend Payload Refusal metrics dynamically transforming them into sleek "Toast" bounds to stop UX breaking.

### 5. Security Handshake (Clerk)
- All connections are guarded by **Clerk Auth**. Instead of native standard headers (which native WebSockets do not support), Next.js encrypts the user's active session JSON Web Token (JWT) into the secure `WSS://` connection string. The Go server catches the token, decrypts it manually, and rigidly maps the payload strictly to the proven sender.

---

## 💻 Tech Stack Summary

**Frontend:** Next.js, React 19, TypeScript, TailwindCSS, shadcn/ui, Framer Motion
**Backend:** Go (Golang), WebSockets (`gorilla/websocket`), Go-Cron
**Databases:** Neon (PostgreSQL), Upstash (Redis)
**Authentication:** Clerk JWT Middleware
**Testing:** Jest, Go Test (Testify)
