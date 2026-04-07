import ChatRoom from "./ChatRoom";
import Link from "next/link";
import { notFound } from "next/navigation";

export default async function EventPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  
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
    console.error("Failed to fetch native event details", e);
  }

  // Native Next.js 404 routing MUST occur outside exception hierarchies to securely trigger!
  if (!isEventFound) {
      notFound();
  }

  return (
    <div className="container mx-auto p-4 md:p-8 h-[calc(100vh-80px)] flex flex-col">
      <div className="mb-6 flex items-start justify-between">
        <div className="flex flex-col gap-2">
          <Link href="/" className="text-fuchsia-400 hover:text-fuchsia-300 text-sm font-semibold flex items-center transition-colors w-fit group">
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
      
      {/* 
        This wrapper is the visual chassis for the chat room. 
        It leverages dark-themed translucent glassmorphism.
      */}
      <div className="flex-1 bg-black/40 backdrop-blur-md border border-white/10 rounded-2xl overflow-hidden shadow-[0_0_40px_rgba(0,0,0,0.5)]">
         <ChatRoom sessionId={id} />
      </div>
    </div>
  );
}
