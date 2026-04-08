import Link from 'next/link';

export default function LandingPage() {
  return (
    <div className="relative flex flex-col items-center justify-center min-h-[85vh] text-center px-4 overflow-hidden">
      {/* Dynamic Gradients */}
      <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[80vw] h-[80vw] md:w-[600px] md:h-[600px] bg-gradient-to-tr from-fuchsia-600/20 to-blue-600/20 blur-[100px] rounded-full pointer-events-none"></div>

      <div className="z-10 max-w-4xl space-y-8">

        <h1 className="text-5xl md:text-8xl font-black tracking-tighter text-transparent bg-clip-text bg-gradient-to-b from-white to-white/60 drop-shadow-sm leading-tight">
          The heartbeat of <br className="hidden md:block" /> every live event.
        </h1>

        <p className="text-lg md:text-2xl text-slate-400 max-w-2xl mx-auto font-light leading-relaxed">
          Sync up with thousands of fans at the exact same venue in real-time. Drop into live arenas powered by Ticketmaster.
        </p>

        <div className="pt-8">
          <Link href="/events" className="group relative inline-flex items-center justify-center px-8 py-4 font-bold text-white transition-all bg-fuchsia-600 rounded-full hover:bg-fuchsia-500 hover:scale-105 hover:shadow-[0_0_40px_rgba(217,70,239,0.4)] active:scale-95">
            <span className="absolute inset-0 w-full h-full rounded-full border border-white/20"></span>
            Explore Live Arenas
            <svg className="w-5 h-5 ml-2 transition-transform group-hover:translate-x-1" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2.5} d="M13 5l7 7-7 7M5 12h14" /></svg>
          </Link>
        </div>
      </div>
    </div>
  );
}
