"use client";

import { useState } from "react";
import Link from "next/link";
import { useAuth } from "@clerk/nextjs";

export default function EventCard({ event }: { event: any }) {
  const [isFavorite, setIsFavorite] = useState(false);
  const { getToken, isSignedIn } = useAuth();
  const [isLiking, setIsLiking] = useState(false);

  const toggleFavorite = async (e: React.MouseEvent) => {
    e.preventDefault(); 
    if (!isSignedIn) {
      alert("Please sign in to favorite events!");
      return; 
    }

    setIsLiking(true);
    try {
      const token = await getToken();
      const method = isFavorite ? "DELETE" : "POST";
      const res = await fetch("http://localhost:8080/api/favorites", {
        method,
        headers: {
          "Authorization": `Bearer ${token}`,
          "Content-Type": "application/json"
        },
        body: JSON.stringify({ event_id: event.id })
      });

      if (res.ok) setIsFavorite(!isFavorite);
    } catch(err) {
      console.error(err);
    } finally {
      setIsLiking(false);
    }
  };

  return (
    <Link href={`/events/${event.id}`}>
      <div className="group relative overflow-hidden rounded-2xl bg-white/5 border border-white/10 p-6 transition-all hover:bg-white/10 hover:border-fuchsia-500/50 hover:shadow-[0_0_30px_rgba(217,70,239,0.15)] cursor-pointer h-full flex flex-col justify-between">
        <div>
          <div className="absolute top-0 right-0 p-4 flex gap-2">
            <button 
              onClick={toggleFavorite}
              disabled={isLiking}
              className={`p-1.5 rounded-full z-10 transition-colors ${isFavorite ? "bg-red-500/20 text-red-400" : "bg-white/5 text-slate-400 hover:text-white"}`}
            >
              <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                <path fillRule="evenodd" d="M3.172 5.172a4 4 0 015.656 0L10 6.343l1.172-1.171a4 4 0 115.656 5.656L10 17.657l-6.828-6.829a4 4 0 010-5.656z" clipRule="evenodd" />
              </svg>
            </button>
            <span className="inline-flex items-center rounded-full bg-fuchsia-500/10 px-2.5 py-0.5 text-xs font-semibold text-fuchsia-400 line-clamp-1 max-w-[120px]">
              {event.type}
            </span>
          </div>
          <h3 className="text-xl md:text-2xl font-bold text-white mb-2 group-hover:bg-clip-text group-hover:text-transparent group-hover:bg-gradient-to-r group-hover:from-fuchsia-400 group-hover:to-blue-400 transition-all pr-[100px] line-clamp-2 leading-tight">{event.title}</h3>
          <p className="text-slate-400 text-sm">
            {new Date(event.start_time).toLocaleString(undefined, { weekday: 'short', month: 'short', day: 'numeric', hour: 'numeric', minute: 'numeric' })}
          </p>
        </div>
        
        <div className="mt-8 flex flex-row items-center justify-between">
          <span className="text-sm font-medium text-fuchsia-400 bg-fuchsia-500/10 rounded-full px-4 py-1.5 opacity-0 group-hover:opacity-100 transition-all translate-y-2 group-hover:translate-y-0 shadow-[0_0_15px_rgba(217,70,239,0.3)]">
            Join Arena &rarr;
          </span>
        </div>
      </div>
    </Link>
  );
}
