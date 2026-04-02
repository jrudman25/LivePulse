import Link from 'next/link';

export default function Home() {
  const events = [
    { id: "concert_123", name: "Astrophysical Tour 2026", type: "Concert", date: "Tonight, 8:00 PM" },
    { id: "sports_456", name: "Championship Finals", type: "Sports", date: "Tomorrow, 5:00 PM" },
    { id: "tv_789", name: "Live Awards Show", type: "TV Event", date: "Sunday, 7:00 PM" }
  ];

  return (
    <div className="container mx-auto p-4 md:p-8 mt-10">
      <div className="mb-12 text-center md:text-left">
        <h1 className="text-4xl md:text-6xl font-bold tracking-tighter mb-4 text-transparent bg-clip-text bg-gradient-to-br from-white to-white/50">Live Events</h1>
        <p className="text-lg text-slate-400 max-w-2xl">Join the conversation. Tap into any live event below to connect with fans around the world in real-time.</p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {events.map((e) => (
          <Link href={`/events/${e.id}`} key={e.id}>
            <div className="group relative overflow-hidden rounded-2xl bg-white/5 border border-white/10 p-6 transition-all hover:bg-white/10 hover:border-fuchsia-500/50 hover:shadow-[0_0_30px_rgba(217,70,239,0.15)] cursor-pointer">
              <div className="absolute top-0 right-0 p-4">
                <span className="inline-flex items-center rounded-full bg-fuchsia-500/10 px-2.5 py-0.5 text-xs font-semibold text-fuchsia-400">
                  {e.type}
                </span>
              </div>
              <h3 className="text-2xl font-bold text-white mb-2 group-hover:bg-clip-text group-hover:text-transparent group-hover:bg-gradient-to-r group-hover:from-fuchsia-400 group-hover:to-blue-400 transition-all">{e.name}</h3>
              <p className="text-slate-400">{e.date}</p>
              
              <div className="mt-8 flex items-center text-sm font-medium text-fuchsia-400 opacity-0 group-hover:opacity-100 transition-opacity translate-y-2 group-hover:translate-y-0">
                Join Chat Arena &rarr;
              </div>
            </div>
          </Link>
        ))}
      </div>
    </div>
  );
}
