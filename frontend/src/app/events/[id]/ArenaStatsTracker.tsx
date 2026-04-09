"use client";

import { useState, useEffect } from "react";
import { useAuth } from "@clerk/nextjs";

export default function ArenaStatsTracker({ eventId }: { eventId: string }) {
  const [isFavorite, setIsFavorite] = useState(false);
  const { getToken, isSignedIn } = useAuth();
  const [isLiking, setIsLiking] = useState(false);
  const [activeUsers, setActiveUsers] = useState<number | null>(null);

  useEffect(() => {
    const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
    
    // Poll Stats
    const pollStats = () => {
        fetch(`${API_URL}/api/sessions/stats?session_id=${eventId}`)
        .then(r => r.json())
        .then(data => {
            if (data && data.active_user_count !== undefined) {
               setActiveUsers(data.active_user_count);
            }
        })
        .catch(() => { });
    };
    pollStats();
    const interval = setInterval(pollStats, 5000);

    // Fetch initial favorites if logged in
    if (isSignedIn) {
      getToken().then(token => {
          fetch(`${API_URL}/api/favorites`, {
              headers: { "Authorization": `Bearer ${token}` }
          })
          .then(r => r.json())
          .then(data => {
            if (Array.isArray(data) && data.includes(eventId)) {
                setIsFavorite(true);
            }
          })
          .catch(() => {});
      });
    }

    return () => clearInterval(interval);
  }, [eventId, isSignedIn, getToken]);

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
      const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
      const res = await fetch(`${API_URL}/api/favorites`, {
        method,
        headers: { "Authorization": `Bearer ${token}`, "Content-Type": "application/json" },
        body: JSON.stringify({ event_id: eventId })
      });
      if (res.ok) setIsFavorite(!isFavorite);
    } catch (err) {} finally {
      setIsLiking(false);
    }
  };

  return (
    <div className="flex flex-wrap items-center gap-3 mt-4">
      {activeUsers !== null && (
        <div className="flex items-center gap-2 bg-green-500/10 px-3 py-1.5 rounded-full border border-green-500/20 shadow-[0_0_15px_rgba(34,197,94,0.1)]">
          <span className="relative flex h-2 w-2">
            <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75"></span>
            <span className="relative inline-flex rounded-full h-2 w-2 bg-green-500"></span>
          </span>
          <span className="text-xs font-semibold text-green-400 tracking-wide">{activeUsers} active fans</span>
        </div>
      )}
      <button
        onClick={toggleFavorite}
        disabled={isLiking}
        className={`flex items-center gap-2 px-4 py-1.5 rounded-full transition-all font-medium text-sm border shadow-sm backdrop-blur ${isFavorite ? "bg-fuchsia-500/20 text-fuchsia-400 border-fuchsia-500/30 hover:bg-fuchsia-500/30" : "bg-white/5 text-slate-300 border-white/10 hover:bg-white/10 hover:text-white"}`}
      >
        <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
          <path fillRule="evenodd" d="M3.172 5.172a4 4 0 015.656 0L10 6.343l1.172-1.171a4 4 0 115.656 5.656L10 17.657l-6.828-6.829a4 4 0 010-5.656z" clipRule="evenodd" />
        </svg>
        {isFavorite ? "Favorited" : "Add to Favorites"}
      </button>
    </div>
  );
}
