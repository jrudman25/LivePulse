import ChatRoom from "./ChatRoom";
import Link from "next/link";
import { notFound } from "next/navigation";
import { auth } from "@clerk/nextjs/server";
import { ArrowLeft, CalendarDays, LockKeyhole, MapPin } from "lucide-react";
import ArenaStatsTracker from "./ArenaStatsTracker";

type EventDetails = {
  id: string;
  title: string;
  type?: string;
  location?: string;
  country?: string;
  start_time?: string;
  end_time?: string;
};

export default async function EventPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const { userId } = await auth();

  let event: EventDetails = { id, title: "Live Session" };
  let isEventFound = true;
  try {
    const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
    const res = await fetch(`${API_URL}/api/events/single?id=${id}`, { cache: "no-store" });
    if (res.ok) {
      const data = await res.json();
      event = { ...event, ...data, title: data.title || "Live Session" };
    } else if (res.status === 404) {
      isEventFound = false;
    }
  } catch (e) {
    console.error("Failed to fetch event title:", e);
  }

  if (!isEventFound) {
    notFound();
  }

  const startTime = event.start_time ? new Date(event.start_time) : null;
  const hasValidStartTime = startTime && !Number.isNaN(startTime.getTime());

  return (
    <div className="mx-auto w-full max-w-[1600px] border-x border-[#45413c]">
      <div className="grid border-b border-[#45413c] lg:grid-cols-[minmax(0,1fr)_360px]">
        <div className="p-5 sm:p-8 lg:p-12">
          <Link href="/events" className="mb-8 flex w-fit items-center gap-3 font-mono text-[10px] font-semibold uppercase tracking-[0.16em] text-[#aaa49b] transition-colors hover:text-[#f2efe8]">
            <ArrowLeft className="h-3.5 w-3.5" aria-hidden="true" />
            Event desk
          </Link>
          <p className="mb-4 font-mono text-[10px] font-semibold uppercase tracking-[0.18em] text-[#aaa49b]">
            Event room
          </p>
          <h1 className="max-w-6xl font-heading text-[clamp(3.5rem,8vw,8rem)] font-black uppercase leading-[0.8] tracking-[-0.05em] text-[#f2efe8]">{event.title}</h1>
        </div>

        <aside className="border-t border-[#45413c] bg-[#f2efe8] text-[#11100f] lg:border-t-0 lg:border-l">
          <div className="border-b border-[#11100f] p-5 sm:p-7">
            <span className="font-mono text-[9px] font-semibold uppercase tracking-[0.16em] opacity-60">Transmission</span>
            <p className="mt-2 break-all font-mono text-[11px] font-medium uppercase tracking-[0.08em]">{id}</p>
          </div>
          <div className="space-y-5 p-5 sm:p-7">
            <div className="flex gap-3">
              <CalendarDays className="mt-0.5 h-4 w-4 shrink-0" aria-hidden="true" />
              <div>
                <span className="block font-mono text-[9px] font-semibold uppercase tracking-[0.14em] opacity-60">On air</span>
                <span className="mt-1 block text-sm font-medium">{hasValidStartTime ? startTime.toLocaleString(undefined, { weekday: "long", month: "short", day: "numeric", hour: "numeric", minute: "2-digit" }) : "Schedule to be announced"}</span>
              </div>
            </div>
            <div className="flex gap-3">
              <MapPin className="mt-0.5 h-4 w-4 shrink-0" aria-hidden="true" />
              <div>
                <span className="block font-mono text-[9px] font-semibold uppercase tracking-[0.14em] opacity-60">Location</span>
                <span className="mt-1 block text-sm font-medium">{event.location || "To be announced"}{event.country ? ` / ${event.country}` : ""}</span>
              </div>
            </div>
            {event.type && <p className="border-t border-[#11100f] pt-4 font-mono text-[10px] font-semibold uppercase tracking-[0.14em]">{event.type}</p>}
          </div>
        </aside>
      </div>

      <div className="grid lg:grid-cols-[300px_minmax(0,1fr)]">
        <aside className="border-b border-[#45413c] p-5 sm:p-8 lg:border-r lg:border-b-0">
          <p className="font-heading text-3xl font-bold uppercase tracking-[-0.03em] text-[#f2efe8]">Room details</p>
          <p className="mt-2 text-sm leading-relaxed text-[#aaa49b]">See how many people are connected and save this event for later.</p>
          <ArenaStatsTracker eventId={id} />
        </aside>

        <section className="min-h-[620px] bg-[#171614] lg:h-[76vh] lg:min-h-[620px]">
          {!userId ? (
            <div className="broadcast-grid grid h-full min-h-[620px] place-items-center p-8 text-center">
              <div className="max-w-lg border border-[#67625b] bg-[#11100f] p-8 sm:p-12">
                <LockKeyhole className="mx-auto h-8 w-8 text-[#ed2f24]" aria-hidden="true" />
                <p className="mt-7 font-mono text-[10px] font-semibold uppercase tracking-[0.18em] text-[#ed2f24]">Authentication required</p>
                <h2 className="mt-3 font-heading text-5xl font-black uppercase leading-none tracking-[-0.04em] text-[#f2efe8]">This room is restricted</h2>
                <p className="mt-4 text-sm leading-relaxed text-[#aaa49b]">Sign in to read the live feed and join the conversation with everyone following this event.</p>
                <Link href="/sign-in" className="mt-8 inline-flex bg-[#ed2f24] px-6 py-4 font-mono text-[10px] font-semibold uppercase tracking-[0.16em] text-[#fffaf2] transition-colors hover:bg-[#f2efe8] hover:text-[#11100f]">Sign in to enter</Link>
              </div>
            </div>
          ) : (
            <ChatRoom sessionId={id} />
          )}
        </section>
      </div>
    </div>
  );
}
