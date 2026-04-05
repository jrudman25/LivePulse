import EventFeed from './EventFeed';
import { auth } from '@clerk/nextjs/server';

async function fetchEvents(userId: string | null) {
  try {
    const url = userId 
      ? `http://localhost:8080/api/events?user_id=${userId}`
      : 'http://localhost:8080/api/events';

    // Calling the Go Backend natively from a Next.js 15 Server Component!
    const res = await fetch(url, { cache: 'no-store' });
    if (!res.ok) return [];
    
    const data = await res.json();
    return data || [];
  } catch (err) {
    console.error("Failed to fetch events from Go backend:", err);
    return [];
  }
}

export default async function Home() {
  const { userId } = await auth();
  const events = await fetchEvents(userId);

  return (
    <div className="container mx-auto p-4 md:p-8 mt-10">
      <div className="mb-12 text-center md:text-left">
        <h1 className="text-4xl md:text-6xl font-bold tracking-tighter mb-4 text-transparent bg-clip-text bg-gradient-to-br from-white to-white/50">Live Events</h1>
        <p className="text-lg text-slate-400 max-w-2xl">Join the conversation. Tap into any real Ticketmaster event below to connect with fans around the world.</p>
      </div>

      {events.length === 0 ? (
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
