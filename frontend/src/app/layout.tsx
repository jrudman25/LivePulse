import type { Metadata, Viewport } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import { ClerkProvider, Show, UserButton, SignInButton } from "@clerk/nextjs";
import "./globals.css";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "LivePulse",
  description: "Real-time sidekick for live events.",
  manifest: "/manifest.json",
  appleWebApp: {
    capable: true,
    statusBarStyle: "black-translucent",
    title: "LivePulse",
  },
};

export const viewport: Viewport = {
  themeColor: "#000000",
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
      <body className={`${geistSans.variable} ${geistMono.variable} min-h-screen bg-black text-slate-50 antialiased selection:bg-fuchsia-500/30`}>
        <ClerkProvider>
          <div className="relative flex min-h-screen flex-col bg-background">
            {/* Ambient Background Glows */}
            <div className="fixed top-0 -z-10 h-full w-full bg-black">
              <div className="absolute bottom-auto left-auto right-0 top-0 h-[500px] w-[500px] -translate-x-[30%] translate-y-[20%] rounded-full bg-[rgba(173,109,244,0.15)] opacity-50 blur-[80px] pointer-events-none"></div>
              <div className="absolute bottom-0 left-0 right-auto top-auto h-[500px] w-[500px] translate-x-[10%] -translate-y-[20%] rounded-full bg-[rgba(45,115,255,0.15)] opacity-50 blur-[80px] pointer-events-none"></div>
            </div>
            
            {/* Sticky Navigation / Header Placeholder */}
            <header className="sticky top-0 z-50 w-full border-b border-white/10 bg-black/50 backdrop-blur-xl">
              <div className="container mx-auto flex h-16 items-center justify-between px-4">
                <div className="flex items-center gap-2">
                  <div className="h-8 w-8 rounded-full bg-gradient-to-br from-fuchsia-500 to-blue-500 flex items-center justify-center shadow-[0_0_15px_rgba(217,70,239,0.5)]">
                    <span className="text-white font-bold text-lg leading-none">L</span>
                  </div>
                  <span className="font-bold text-xl tracking-tight bg-clip-text text-transparent bg-gradient-to-r from-white to-white/60">LivePulse</span>
                </div>
                <div className="flex items-center justify-end gap-4">
                  <Show when="signed-out">
                    <SignInButton>
                      <button className="rounded-full bg-white/10 px-5 py-2 text-sm font-semibold text-white transition-colors hover:bg-white/20">Sign In</button>
                    </SignInButton>
                  </Show>
                  <Show when="signed-in">
                    <UserButton appearance={{ elements: { avatarBox: "h-9 w-9 ring-2 ring-fuchsia-500/50 hover:ring-fuchsia-400 transition-all" } }} />
                  </Show>
                </div>
              </div>
            </header>

            <main className="flex-1">{children}</main>
          </div>
        </ClerkProvider>
      </body>
    </html>
  );
}
