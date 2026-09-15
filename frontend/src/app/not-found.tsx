import Link from "next/link";
import { ArrowLeft } from "lucide-react";

export default function NotFound() {
  return (
    <div className="mx-auto grid min-h-[calc(100vh-72px)] w-full max-w-[1600px] place-items-center border-x border-[#45413c] p-5">
      <div className="w-full max-w-2xl border border-[#45413c] bg-[#171614] p-8 text-center sm:p-12">
        <p className="font-mono text-[10px] font-semibold uppercase tracking-[0.18em] text-[#ed2f24]">404 / Off the wire</p>
        <h1 className="mt-3 font-heading text-5xl font-bold uppercase tracking-[-0.015em] text-[#f2efe8]">Page not found</h1>
        <p className="mx-auto mt-3 max-w-md text-sm leading-relaxed text-[#aaa49b]">This address does not match a live desk. Head back to the event schedule.</p>
        <Link href="/events" className="mt-7 inline-flex items-center gap-3 border border-[#67625b] px-5 py-3 font-mono text-[10px] font-semibold uppercase tracking-[0.14em] text-[#f2efe8] transition-colors hover:bg-[#f2efe8] hover:text-[#11100f]">
          <ArrowLeft className="h-3.5 w-3.5" aria-hidden="true" />
          Back to events
        </Link>
      </div>
    </div>
  );
}
