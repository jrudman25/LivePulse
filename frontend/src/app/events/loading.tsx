export default function EventsLoading() {
  return (
    <div className="mx-auto w-full max-w-[1600px] border-x border-[#45413c]" aria-busy="true" aria-label="Loading events">
      <header className="grid border-b border-[#45413c] lg:grid-cols-[1fr_340px]">
        <div className="p-5 sm:p-8 lg:p-12">
          <p className="mb-4 flex items-center gap-3 font-mono text-[10px] font-semibold uppercase tracking-[0.18em] text-[#aaa49b] sm:text-xs">
            <span className="h-2.5 w-2.5 animate-pulse bg-[#ed2f24]" />
            Rolling 24-hour schedule
          </p>
          <h1 className="font-heading text-[clamp(4.5rem,10vw,9rem)] font-black uppercase leading-[0.92] tracking-[-0.015em] text-[#f2efe8]">Event desk</h1>
        </div>
        <div className="flex flex-col justify-end border-t border-[#45413c] bg-[#f2efe8] p-5 text-[#11100f] sm:p-8 lg:border-t-0 lg:border-l">
          <span className="frame-number font-heading text-7xl font-black leading-none tracking-[-0.06em]">--</span>
          <span className="mt-2 font-mono text-[10px] font-semibold uppercase tracking-[0.18em]">Events listed</span>
        </div>
      </header>

      <div className="p-4 sm:p-8 lg:p-12">
        <div className="border-t border-[#45413c]">
          {Array.from({ length: 6 }).map((_, i) => (
            <div key={i} className="grid min-h-[190px] animate-pulse border-b border-[#45413c] md:grid-cols-[96px_160px_minmax(0,1fr)_220px_72px]">
              <div className="border-b border-[#45413c] md:border-r md:border-b-0" />
              <div className="border-b border-[#45413c] p-4 md:border-r md:border-b-0 md:p-5">
                <div className="h-12 w-14 bg-[#252321]" />
                <div className="mt-3 h-3 w-16 bg-[#252321]" />
              </div>
              <div className="p-5 sm:p-7">
                <div className="mb-4 h-3 w-20 bg-[#252321]" />
                <div className="h-9 w-3/4 bg-[#252321]" />
              </div>
              <div className="hidden border-l border-[#45413c] p-5 md:block">
                <div className="h-3 w-12 bg-[#252321]" />
                <div className="mt-3 h-4 w-4/5 bg-[#252321]" />
              </div>
              <div className="hidden border-l border-[#45413c] md:block" />
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
