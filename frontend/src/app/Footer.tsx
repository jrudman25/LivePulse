import Link from "next/link";
import { ArrowUpRight } from "lucide-react";

export default function Footer() {
  return (
    <footer className="mt-20 w-full border-t border-[#45413c] bg-[#11100f] text-[#f2efe8]">
      <div className="mx-auto grid w-full max-w-[1600px] border-x border-[#45413c] lg:grid-cols-[1fr_auto]">
        <div className="p-6 sm:p-10">
          <Link href="/" className="font-heading text-5xl font-black uppercase tracking-[-0.045em] sm:text-7xl">LivePulse</Link>
          <p className="mt-3 max-w-md text-sm leading-relaxed text-[#aaa49b]">Event-based group chat for concerts, sports, and shows.</p>
        </div>
        <div className="grid border-t border-[#45413c] sm:grid-cols-2 lg:min-w-[520px] lg:border-t-0 lg:border-l">
          <nav className="flex flex-col gap-4 p-6 font-mono text-[10px] font-semibold uppercase tracking-[0.14em] text-[#aaa49b] sm:p-10" aria-label="Footer navigation">
            <Link href="/events" className="transition-colors hover:text-[#f2efe8]">Events</Link>
            <Link href="/help" className="transition-colors hover:text-[#f2efe8]">Help & FAQ</Link>
            <Link href="/privacy" className="transition-colors hover:text-[#f2efe8]">Privacy</Link>
          </nav>
          <div className="flex flex-col justify-between gap-8 border-t border-[#45413c] p-6 sm:border-t-0 sm:border-l sm:p-10">
            <a href="https://github.com/jrudman25/livepulse" className="group flex items-center justify-between font-mono text-xs font-semibold uppercase tracking-[0.14em] transition-colors hover:text-[#ed2f24]">
              View source
              <ArrowUpRight className="h-4 w-4" aria-hidden="true" />
            </a>
            <div className="space-y-2 font-mono text-[10px] uppercase tracking-[0.12em] text-[#67625b]">
              <p>Next.js / Go / Neon / Upstash</p>
              <p>Built by Jordan</p>
            </div>
          </div>
        </div>
      </div>
    </footer>
  );
}
