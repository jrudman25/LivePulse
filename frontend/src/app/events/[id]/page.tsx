import ChatRoom from "./ChatRoom";
import Link from "next/link";
import { notFound } from "next/navigation";
import { auth } from "@clerk/nextjs/server";

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
      <div className="flex flex-col md:flex-row md:items-end justify-between border-b border-white/5 pb-4">
        <div className="flex flex-col relative">
          <Link href="/events" className="text-fuchsia-400 hover:text-fuchsia-300 text-sm font-semibold flex items-center transition-colors w-fit group">
            <span className="mr-1 group-hover:-translate-x-1 transition-transform">&larr;</span> Back to Events
          </Link>
          <h1 className="text-3xl md:text-4xl font-bold text-white mt-1 line-clamp-2">{eventTitle}</h1>
          <p className="text-slate-400 text-xs font-medium tracking-wide border border-white/10 bg-white/5 rounded-full px-3 py-1 w-fit mt-1">Session ID: {id}</p>
        </div>
        <div className="flex items-center gap-2 bg-green-500/10 px-3 py-1.5 rounded-full border border-green-500/20 shadow-[0_0_15px_rgba(34,197,94,0.1)] mt-8">
          <span className="relative flex h-2 w-2">
            <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75"></span>
            <span className="relative inline-flex rounded-full h-2 w-2 bg-green-500"></span>
          </span>
          <span className="text-sm font-semibold text-green-400 uppercase tracking-widest text-[10px]">Live</span>
        </div>
      </div>
      
      <div className="mt-8 flex-1 border border-white/10 rounded-2xl overflow-hidden bg-white/5 flex flex-col md:max-w-4xl mx-auto w-full shadow-2xl relative">
        <div className="p-4 border-b border-white/10 bg-black/40 backdrop-blur flex justify-between items-center z-10">
          <h2 className="text-xl font-bold">{eventTitle} <span className="opacity-50 font-normal">Arena</span></h2>
          <div className="flex gap-2">
            <span className="px-2 py-1 text-xs rounded-full bg-emerald-500/20 text-emerald-400 border border-emerald-500/30 drop-shadow-[0_0_10px_rgba(16,185,129,0.3)] flex items-center">
              <span className="w-1.5 h-1.5 rounded-full bg-emerald-500 mr-2 animate-pulse"></span>
              Live
            </span>
          </div>
        </div>

        {!userId ? (
          <div className="flex-1 flex flex-col items-center justify-center p-8 text-center border-t border-white/5 bg-black/20">
             <div className="w-16 h-16 rounded-full bg-fuchsia-500/10 flex items-center justify-center mb-4 border border-fuchsia-500/30 shadow-[0_0_20px_rgba(217,70,239,0.2)]">
                <span className="text-2xl">🔒</span>
             </div>
             <h3 className="text-2xl font-bold text-white mb-2">Restricted Arena</h3>
             <p className="text-slate-400 mb-6 max-w-md">You must be signed in to your LivePulse account to view messages and engage with other fans at this event.</p>
          </div>
        ) : (
          <ChatRoom sessionId={id} />
        )}
      </div>
    </div>
  );
}
