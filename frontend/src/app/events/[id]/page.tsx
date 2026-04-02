import ChatRoom from "./ChatRoom";

export default async function EventPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  
  return (
    <div className="container mx-auto p-4 md:p-8 h-[calc(100vh-80px)] flex flex-col">
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-white">Live Session</h1>
          <p className="text-slate-400 text-sm">Session ID: {id}</p>
        </div>
        <div className="flex items-center gap-2 bg-green-500/10 px-3 py-1.5 rounded-full border border-green-500/20 shadow-[0_0_15px_rgba(34,197,94,0.1)]">
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
