import EventFeed from "../EventFeed";
import { auth } from "@clerk/nextjs/server";
import { LockKeyhole } from "lucide-react";

async function fetchEvents(userId: string | null, q: string) {
  try {
    const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
    let url = `${API_URL}/api/events`;
    const params = new URLSearchParams();
    if (userId) {params.append("user_id", userId);}
    if (q) {params.append("q", q);}

    const qs = params.toString();
    if (qs) {url += `?${qs}`;}

    const res = await fetch(url, { cache: "no-store" });
    if (!res.ok) {return [];}

    const data = await res.json();
    return data || [];
  } catch (err) {
    console.error("Failed to fetch events from Go backend:", err);
    return [];
  }
}

export default async function EventsPage({ searchParams }: { searchParams: Promise<{ [key: string]: string | string[] | undefined }> }) {
  const resolvedParams = await searchParams;
  const { userId } = await auth();

  const q = typeof resolvedParams.q === "string" ? resolvedParams.q : "";
  const events = await fetchEvents(userId, q);

  return (
    <div className="mx-auto w-full max-w-[1600px] border-x border-[#45413c]">
      <header className="grid border-b border-[#45413c] lg:grid-cols-[1fr_340px]">
        <div className="p-5 sm:p-8 lg:p-12">
          <p className="mb-4 flex items-center gap-3 font-mono text-[10px] font-semibold uppercase tracking-[0.18em] text-[#aaa49b] sm:text-xs">
            <span className="h-2.5 w-2.5 bg-[#ed2f24]" />
            Rolling 24-hour schedule
          </p>
          <h1 className="font-heading text-[clamp(4.5rem,10vw,9rem)] font-black uppercase leading-[0.8] tracking-[-0.055em] text-[#f2efe8]">Event desk</h1>
        </div>
        <div className="flex flex-col justify-end border-t border-[#45413c] bg-[#f2efe8] p-5 text-[#11100f] sm:p-8 lg:border-t-0 lg:border-l">
          <span className="frame-number font-heading text-7xl font-black leading-none tracking-[-0.06em]">{String(events.length).padStart(2, "0")}</span>
          <span className="mt-2 font-mono text-[10px] font-semibold uppercase tracking-[0.18em]">Events listed</span>
        </div>
      </header>

      <div className="p-4 sm:p-8 lg:p-12">
        {!userId && (
          <div className="mb-8 flex items-start gap-4 border border-[#67625b] p-4 text-[#d1cbc1]">
            <LockKeyhole className="mt-0.5 h-4 w-4 shrink-0 text-[#ed2f24]" aria-hidden="true" />
            <div>
              <p className="font-mono text-[11px] font-semibold uppercase tracking-[0.14em] text-[#f2efe8]">Viewing as guest</p>
              <p className="mt-1 text-sm text-[#aaa49b]">Browse every event. Sign in when you are ready to enter a room or save it.</p>
            </div>
          </div>
        )}

        {events.length === 0 && q === "" ? (
          <div className="grid min-h-[360px] place-items-center border border-[#45413c] bg-[#171614] p-8 text-center">
            <div>
              <p className="font-heading text-5xl font-bold uppercase tracking-[-0.035em] text-[#f2efe8]">The desk is quiet</p>
              <p className="mx-auto mt-3 max-w-md text-sm leading-relaxed text-[#aaa49b]">No upcoming events are on the wire yet. The schedule refreshes as new rooms become available.</p>
            </div>
          </div>
        ) : (
          <EventFeed initialEvents={events} />
        )}
      </div>
    </div>
  );
}
