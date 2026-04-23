import EventFeed from '../EventFeed';
import { auth } from '@clerk/nextjs/server';

async function fetchEvents(userId: string | null, q: string) {
  try {
    const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
    let url = `${API_URL}/api/events`;
    const params = new URLSearchParams();
    if (userId) {params.append("user_id", userId);}
    if (q) {params.append("q", q);}
    
    const qs = params.toString();
    if (qs) {url += `?${qs}`;}

    const res = await fetch(url, { cache: 'no-store' });
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
  
  const q = typeof resolvedParams.q === 'string' ? resolvedParams.q : "";
  const events = await fetchEvents(userId, q);

  return (
    <div className="container mx-auto p-4 md:p-8 mt-10">
      <div className="mb-8 text-center md:text-left">
        <h1 className="text-4xl md:text-5xl font-bold tracking-tighter mb-4 text-transparent bg-clip-text bg-gradient-to-br from-white to-white/50">Live Arenas</h1>
        <p className="text-lg text-slate-400 max-w-2xl">Tap into any event below to connect with fans securely.</p>
      </div>

      {!userId && (
        <div className="mb-8 p-4 rounded-xl border border-fuchsia-500/30 bg-fuchsia-500/10 text-fuchsia-200 flex items-center justify-center font-medium shadow-[0_0_15px_rgba(217,70,239,0.1)]">
          🔒 You must be signed in to join the live arenas!
        </div>
      )}

      {events.length === 0 && q === "" ? (
        <div className="col-span-full py-20 flex flex-col items-center justify-center border border-white/5 rounded-3xl bg-white/5 relative overflow-hidden">
           <div className="absolute w-[200px] h-[200px] bg-fuchsia-500/20 blur-[60px] rounded-full pointer-events-none"></div>
           <p className="text-slate-400 font-medium tracking-wide z-10 text-lg">Waiting for Ticketmaster API ingestion...</p>
           <p className="text-slate-500 text-sm mt-2 z-10">Trigger <code className="bg-black/50 px-2 py-0.5 rounded ml-1">POST /api/admin/trigger-fetch</code> to populate.</p>
        </div>
      ) : (
        <EventFeed initialEvents={events} />
      )}
    </div>
  );
}
