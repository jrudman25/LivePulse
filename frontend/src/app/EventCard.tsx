"use client";

import { useState, useEffect } from "react";
import Link from "next/link";
import { useAuth } from "@clerk/nextjs";

// Add optional callback updating parent DOM states cleanly
export default function EventCard({ event, onFavoriteToggle }: { event: any, onFavoriteToggle?: (id: string, isFav: boolean) => void }) {
  const [isFavorite, setIsFavorite] = useState(event.is_favorite || false);
  const { getToken, isSignedIn } = useAuth();
  const [isLiking, setIsLiking] = useState(false);
  const [activeUsers, setActiveUsers] = useState<number | null>(null);

  useEffect(() => {
    // Poll the Go WebSocket Hub for active connected users passively
    fetch(`http://localhost:8080/api/sessions/stats?session_id=${event.id}`)
      .then(r => r.json())
      .then(data => {
        if (data && data.active_user_count !== undefined) {
          setActiveUsers(data.active_user_count);
        }
      })
      .catch(() => { });
  }, [event.id]);

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

      if (res.ok) {
        setIsFavorite(!isFavorite);
        if (onFavoriteToggle) {
          onFavoriteToggle(event.id, !isFavorite);
        }
      }
    } catch (err) {
      console.error(err);
    } finally {
      setIsLiking(false);
    }
  };

  const typeColors: Record<string, string> = {
    Music: "bg-fuchsia-500/10 text-fuchsia-400",
    Sports: "bg-blue-500/10 text-blue-400",
    Arts: "bg-amber-500/10 text-amber-400",
    Default: "bg-slate-500/10 text-slate-300"
  };
  const colorClass = typeColors[event.type] || typeColors.Default;

  return (
    <Link href={`/events/${event.id}`}>
      <div className="group relative overflow-hidden rounded-2xl bg-white/5 border border-white/10 p-6 transition-all hover:bg-white/10 hover:border-fuchsia-500/50 hover:shadow-[0_0_30px_rgba(217,70,239,0.15)] cursor-pointer h-full flex flex-col justify-between">
        <div>
          <div className="flex gap-2 items-center justify-end mb-3">
            <span className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-semibold line-clamp-1 max-w-[120px] ${colorClass}`}>
              {event.type}
            </span>
            <button
              onClick={toggleFavorite}
              disabled={isLiking}
              className={`p-1.5 rounded-full transition-colors ${isFavorite ? "bg-red-500/20 text-red-400" : "bg-white/5 text-slate-400 hover:text-white"}`}
            >
              <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                <path fillRule="evenodd" d="M3.172 5.172a4 4 0 015.656 0L10 6.343l1.172-1.171a4 4 0 115.656 5.656L10 17.657l-6.828-6.829a4 4 0 010-5.656z" clipRule="evenodd" />
              </svg>
            </button>
          </div>
          <h3 className="text-xl md:text-2xl font-bold text-white mb-2 group-hover:bg-clip-text group-hover:text-transparent group-hover:bg-gradient-to-r group-hover:from-fuchsia-400 group-hover:to-blue-400 transition-all line-clamp-2 leading-tight">{event.title}</h3>

          <div className="text-slate-400 text-sm mt-3 pt-3 border-t border-white/5">
            {event.location && <span className="block text-white/80 font-medium mb-1 line-clamp-1">{event.location}</span>}
            <span className="block text-slate-500">
              {new Date(event.start_time).toLocaleString(undefined, { weekday: 'short', month: 'short', day: 'numeric', hour: 'numeric', minute: 'numeric' })}
            </span>
          </div>
        </div>

        <div className="mt-8 flex flex-row items-center justify-between">
          <div className="flex items-center">
            {activeUsers !== null && (
              <div className="flex items-center gap-2 bg-green-500/10 px-3 py-1.5 rounded-full border border-green-500/20 shadow-[0_0_15px_rgba(34,197,94,0.1)]">
                <span className="relative flex h-2 w-2">
                  <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75"></span>
                  <span className="relative inline-flex rounded-full h-2 w-2 bg-green-500"></span>
                </span>
                <span className="text-xs font-semibold text-green-400 tracking-wide">{activeUsers} chatters</span>
              </div>
            )}
          </div>
          <span className="text-sm font-medium text-fuchsia-400 bg-fuchsia-500/10 rounded-full px-4 py-1.5 opacity-0 group-hover:opacity-100 transition-all translate-y-2 group-hover:translate-y-0 shadow-[0_0_15px_rgba(217,70,239,0.3)]">
            Join Arena &rarr;
          </span>
        </div>
      </div>
    </Link>
  );
}
