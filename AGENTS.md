# AGENTS.md

## Frontend Notes
- This is a Next.js app with TypeScript
- Use Tailwind CSS for styling
- Use shadcn for UI components

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
