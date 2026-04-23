import ChatRoom from "./ChatRoom";
import Link from "next/link";
import { notFound } from "next/navigation";
import { auth } from "@clerk/nextjs/server";
import ArenaStatsTracker from "./ArenaStatsTracker";

export default async function EventPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const { userId } = await auth();
  
  let eventTitle = "Live Session";
  let isEventFound = true;
  try {
    const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
    const res = await fetch(`${API_URL}/api/events/single?id=${id}`, { cache: 'no-store' });
    if (res.ok) {
        const data = await res.json();
        eventTitle = data.title || "Live Session";
    } else if (res.status === 404) {
        isEventFound = false;
    }
  } catch (e) {
    console.error("Failed to fetch event title:", e);
  }

  if (!isEventFound) {
      notFound();
  }

  return (
    <div className="container mx-auto p-4 md:p-8 h-[85vh] flex flex-col">
      <div className="flex flex-col md:flex-row md:items-end justify-between border-b border-white/5 pb-6">
        <div className="flex flex-col relative w-full">
          <Link href="/events" className="text-fuchsia-400 hover:text-fuchsia-300 text-sm font-semibold flex items-center transition-colors w-fit group mb-1">
            <span className="mr-1 group-hover:-translate-x-1 transition-transform">&larr;</span> Back to Events
          </Link>
          <h1 className="text-4xl md:text-5xl font-black text-transparent bg-clip-text bg-gradient-to-r from-white to-white/70 py-1 drop-shadow-sm line-clamp-2">{eventTitle}</h1>
          <p className="text-slate-400 text-xs font-bold tracking-widest border border-white/10 bg-white/5 rounded-md px-3 py-1 w-fit mt-2 uppercase shadow-inner">Session: {id}</p>
          <ArenaStatsTracker eventId={id} />
        </div>
      </div>
      
      <div className="mt-6 flex-1 border border-white/10 rounded-2xl overflow-hidden bg-white/5 flex flex-col w-full shadow-[0_0_50px_rgba(0,0,0,0.5)] relative">
        {!userId ? (
          <div className="flex-1 flex flex-col items-center justify-center p-8 text-center bg-black/40 backdrop-blur-sm relative overflow-hidden">
             <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[300px] h-[300px] bg-fuchsia-500/10 blur-[100px] rounded-full pointer-events-none"></div>
             <div className="w-20 h-20 rounded-full bg-fuchsia-500/10 flex items-center justify-center mb-6 border border-fuchsia-500/30 shadow-[0_0_30px_rgba(217,70,239,0.2)] z-10">
                <span className="text-3xl">🔒</span>
             </div>
             <h3 className="text-3xl font-bold text-white mb-3 z-10 tracking-tight">Restricted Arena</h3>
             <p className="text-slate-400 mb-6 max-w-md text-lg z-10 leading-relaxed text-balance">You must be signed in to your LivePulse account to view messages and engage with other fans.</p>
          </div>
        ) : (
          <ChatRoom sessionId={id} />
        )}
      </div>
    </div>
  );
}
