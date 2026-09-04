import Link from "next/link";
import { ArrowUpRight } from "lucide-react";

export default function LandingPage() {
  return (
    <div className="overflow-hidden">
      <section className="relative mx-auto min-h-[calc(100vh-72px)] w-full max-w-[1600px] border-x border-[#45413c]">
        {/* Dynamic Gradients */}
        <div className="broadcast-grid pointer-events-none absolute inset-0 opacity-50" />

        <div className="relative grid min-h-[calc(100vh-72px)] lg:grid-cols-[minmax(0,1.45fr)_minmax(340px,0.55fr)]">
          <div className="flex flex-col justify-between border-b border-[#45413c] p-5 sm:p-8 lg:border-r lg:border-b-0 lg:p-12 xl:p-16">
            <div className="reveal-up flex items-center justify-between gap-4 font-mono text-[10px] font-medium uppercase tracking-[0.18em] text-[#aaa49b] sm:text-xs">
              <span>Concerts / sports / shows</span>
              <span>Event-based group chat</span>
            </div>

            <div className="py-16 sm:py-24 lg:py-12">
              <p className="reveal-up reveal-delay-1 mb-7 flex items-center gap-3 font-mono text-[11px] font-semibold uppercase tracking-[0.18em] text-[#f2efe8]">
                <span className="h-2.5 w-2.5 bg-[#ed2f24]" />
                Find an event. Join its room.
              </p>
              <h1 className="reveal-up reveal-delay-2 max-w-5xl font-heading text-[clamp(4.75rem,12vw,11rem)] font-black uppercase leading-[0.94] tracking-[-0.015em] text-[#f2efe8]">
                <span className="block">Be there.</span>
                <span className="mt-2 block text-[#ed2f24] sm:mt-3">While it</span>
                <span className="block">happens.</span>
              </h1>
            </div>

            <div className="reveal-up reveal-delay-3 grid gap-7 border-t border-[#67625b] pt-6 sm:grid-cols-[minmax(0,1fr)_auto] sm:items-end">
              <p className="max-w-xl text-base leading-relaxed text-[#aaa49b] sm:text-lg">
                One live desk for concerts, matches, and shows. Find the event, enter its room, and follow every reaction in real time.
              </p>
              <Link href="/events" className="group inline-flex w-fit items-center gap-5 bg-[#ed2f24] px-6 py-4 font-mono text-xs font-semibold uppercase tracking-[0.14em] text-[#fffaf2] transition-colors hover:bg-[#fffaf2] hover:text-[#11100f]">
                Open event desk
                <ArrowUpRight className="h-4 w-4 transition-transform group-hover:translate-x-1 group-hover:-translate-y-1" aria-hidden="true" />
              </Link>
            </div>
          </div>

          <aside className="relative flex flex-col bg-[#f2efe8] text-[#11100f]">
            <div className="flex items-center justify-between border-b border-[#11100f] px-5 py-4 font-mono text-[10px] font-semibold uppercase tracking-[0.18em] sm:px-8">
              <span>How it works</span>
              <span>3 steps</span>
            </div>
            <ol className="flex flex-1 flex-col justify-center">
              {[
                ["01", "Scan", "See what is happening in the next 24 hours."],
                ["02", "Enter", "Step into the room for your event."],
                ["03", "React", "Talk with the people watching alongside you."],
              ].map(([number, title, copy]) => (
                <li key={number} className="group grid grid-cols-[64px_1fr] border-b border-[#11100f] last:border-b-0 sm:grid-cols-[88px_1fr]">
                  <span className="frame-number border-r border-[#11100f] p-5 font-mono text-xs sm:p-7">{number}</span>
                  <div className="p-5 transition-colors group-hover:bg-[#ed2f24] group-hover:text-[#fffaf2] sm:p-7">
                    <h2 className="font-heading text-5xl font-extrabold uppercase leading-none tracking-[-0.04em]">{title}</h2>
                    <p className="mt-3 max-w-xs text-sm leading-relaxed opacity-70">{copy}</p>
                  </div>
                </li>
              ))}
            </ol>
            <div className="bg-[#ed2f24] px-5 py-5 font-heading text-xl font-bold uppercase tracking-tight text-[#fffaf2] sm:px-8">
              One room for each event
            </div>
          </aside>
        </div>
      </section>
    </div>
  );
}
