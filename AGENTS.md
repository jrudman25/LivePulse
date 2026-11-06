# AGENTS.md

## Frontend Notes
- This is a Next.js app with TypeScript
- Use Tailwind CSS for styling
- Use shadcn for UI components
- The CSP is Clerk's strict nonce mode, set in `frontend/src/proxy.ts`. Any new external origin (API, images, frames, scripts) has to be added to the `directives` there. Baseline security headers live in `frontend/next.config.ts`.
- `<ClerkProvider dynamic>` is required so nonces reach every page. Pages render per request.

## Backend Notes
- This is a Go app
- Use Neon PostgreSQL for database
- Use Upstash for Redis caching
- Use Ticketmaster API for event data
- Use Clerk for authentication

## Setup commands
- Install deps: `npm install`
- Start dev server: `npm run dev`
- Lint: `npx eslint .`
- Run frontend tests (from `/frontend`): `npm run test-frontend`
- Run backend tests (from `/frontend`): `npm run test-backend`
- Run backend tests (from `/backend`): `go test ./... -v`

## Code style
- TypeScript strict mode
