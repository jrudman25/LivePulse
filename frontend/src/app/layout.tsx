import type { Metadata, Viewport } from "next";
import Link from "next/link";
import { Barlow_Condensed, IBM_Plex_Mono, IBM_Plex_Sans } from "next/font/google";
import { ClerkProvider, Show, UserButton, SignInButton } from "@clerk/nextjs";
import Footer from "./Footer";
import "./globals.css";

const body = IBM_Plex_Sans({
  variable: "--font-body",
  subsets: ["latin"],
  weight: ["400", "500", "600"],
});

const display = Barlow_Condensed({
  variable: "--font-display",
  subsets: ["latin"],
  weight: ["500", "600", "700", "800", "900"],
});

const mono = IBM_Plex_Mono({
  variable: "--font-plex-mono",
  subsets: ["latin"],
  weight: ["400", "500", "600"],
});

export const metadata: Metadata = {
  title: "LivePulse | The live event desk",
  description: "Find the room. Join the crowd. Follow live events as they happen.",
  manifest: "/manifest.json",
  appleWebApp: {
    capable: true,
    statusBarStyle: "black-translucent",
    title: "LivePulse",
  },
};

export const viewport: Viewport = {
  themeColor: "#11100f",
  width: "device-width",
  initialScale: 1,
  maximumScale: 1,
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" className="dark">
      <body className={`${body.variable} ${display.variable} ${mono.variable} min-h-screen bg-background text-foreground antialiased`}>
        <ClerkProvider>
          <div className="relative flex min-h-screen flex-col bg-background">
            {/* Ambient Background Glows */}
            <div className="pointer-events-none fixed inset-0 -z-10 bg-[#11100f]" />

            {/* Sticky Navigation / Header Placeholder */}
            <header className="sticky top-0 z-50 w-full border-b border-[#45413c] bg-[#11100f]/95 backdrop-blur-md">
              <div className="mx-auto grid h-[72px] w-full max-w-[1600px] grid-cols-[1fr_auto] items-center px-4 sm:px-6 lg:grid-cols-[1fr_auto_1fr] lg:px-10">
                <Link href="/" className="group flex w-fit items-center gap-3" aria-label="LivePulse home">
                  <span className="grid h-9 w-9 place-items-center bg-[#ed2f24] font-heading text-xl font-black text-[#fffaf2] transition-transform group-hover:-rotate-3">LP</span>
                  <span className="font-heading text-2xl font-extrabold uppercase tracking-[-0.03em] text-[#f2efe8]">LivePulse</span>
                </Link>
                <nav className="hidden items-center gap-8 font-mono text-[11px] font-medium uppercase tracking-[0.16em] text-[#aaa49b] lg:flex" aria-label="Primary navigation">
                  <Link href="/events" className="transition-colors hover:text-[#f2efe8]">Events</Link>
                  <Link href="/help" className="transition-colors hover:text-[#f2efe8]">Help & FAQ</Link>
                  <Link href="/privacy" className="transition-colors hover:text-[#f2efe8]">Privacy</Link>
                </nav>
                <div className="flex items-center justify-end gap-4">
                  <Show when="signed-out">
                    <SignInButton>
                      <button className="border border-[#67625b] px-4 py-2 font-mono text-[11px] font-semibold uppercase tracking-[0.14em] text-[#f2efe8] transition-colors hover:border-[#f2efe8] hover:bg-[#f2efe8] hover:text-[#11100f]">Sign in</button>
                    </SignInButton>
                  </Show>
                  <Show when="signed-in">
                    <UserButton appearance={{ elements: { avatarBox: "h-9 w-9 rounded-none ring-1 ring-[#67625b] transition-all hover:ring-[#f2efe8]" } }} />
                  </Show>
                </div>
              </div>
            </header>

            <main className="flex-1">{children}</main>
            <Footer />
          </div>
        </ClerkProvider>
      </body>
    </html>
  );
}
