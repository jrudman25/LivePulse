import type { NextFetchEvent, NextRequest } from 'next/server'
import { clerkMiddleware } from '@clerk/nextjs/server'
import { ensureSecureCookies } from '@/lib/ensureSecureCookies'

const apiOrigin = new URL(process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080').origin
const wsOrigin = new URL(process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080').origin

const clerk = clerkMiddleware({
  contentSecurityPolicy: {
    strict: true,
    directives: {
      'connect-src': [apiOrigin, wsOrigin],
      'img-src': ['data:'],
      'frame-ancestors': ["'none'"],
      'base-uri': ["'self'"],
      'object-src': ["'none'"],
    },
  },
})

export default async function middleware(req: NextRequest, event: NextFetchEvent) {
  const response = await clerk(req, event)
  // Clerk's dev handshake sets __clerk_redirect_count without Secure; add it on HTTPS
  if (response && req.nextUrl.protocol === 'https:') {
    ensureSecureCookies(response.headers)
  }
  return response
}

export const config = {
  matcher: [
    '/((?!_next|[^?]*\\.(?:html?|css|js(?!on)|jpe?g|webp|png|gif|svg|ttf|woff2?|ico|csv|docx?|xlsx?|zip|webmanifest)).*)',
    '/(api|trpc)(.*)',
  ],
}
