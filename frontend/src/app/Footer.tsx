export default function Footer() {
  return (
    <footer className="w-full border-t border-white/5 bg-black/50 py-8 mt-12 w-full max-w-[100vw] overflow-hidden mix-blend-screen px-4">
      <div className="container mx-auto px-4 flex flex-col items-center justify-center space-y-2">
        <p className="text-slate-500 font-medium tracking-wide text-sm flex items-center justify-center gap-1.5 flex-wrap text-center">
          Powered by <span className="text-fuchsia-400 font-bold bg-fuchsia-500/10 px-2 py-0.5 rounded-md mx-1 tracking-wider uppercase text-[11px] shadow-[0_0_10px_rgba(217,70,239,0.1)]">LivePulse</span>
          | Source: <span className="text-blue-400 font-bold tracking-wider mx-1 flex items-center gap-1"><a href="https://github.com/jrudman25/livepulse">GitHub</a></span>
        </p>
        <p className="text-slate-600 text-xs font-medium text-center">
          Built securely on Next.js, Go, Neon PostgreSQL & Upstash Redis by Jordan
        </p>
        <p className="text-slate-600 text-xs font-medium text-center">
          Hosted on Vercel and Northflank
        </p>
      </div>
    </footer>
  );
}
