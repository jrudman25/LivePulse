"use client";

import { useState, useEffect } from "react";
import { Heart, Users } from "lucide-react";
import { useAuth } from "@clerk/nextjs";
import { useFavorite } from "@/hooks/useFavorite";

export default function ArenaStatsTracker({ eventId }: { eventId: string }) {
  const { isFavorite, setIsFavorite, isLiking, notice, toggleFavorite } = useFavorite(eventId);
  const { getToken, isSignedIn } = useAuth();
  const [activeUsers, setActiveUsers] = useState<number | null>(null);

  useEffect(() => {
    const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

    // Poll Stats (skipped while the tab is hidden)
    const pollStats = () => {
      if (document.visibilityState === "hidden") {return;}
      fetch(`${API_URL}/api/sessions/stats?session_id=${eventId}`, { signal: AbortSignal.timeout(5000) })
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
        if (!token) {return;}
        fetch(`${API_URL}/api/favorites`, {
          headers: { "Authorization": `Bearer ${token}` },
          signal: AbortSignal.timeout(8000)
        })
          .then(r => r.json())
          .then(data => {
            if (Array.isArray(data) && data.includes(eventId)) {
              setIsFavorite(true);
            }
          })
          .catch(() => { });
      }).catch(() => { });
    }

    return () => clearInterval(interval);
  }, [eventId, isSignedIn, getToken, setIsFavorite]);

  return (
    <div className="mt-7 border-y border-[#45413c]">
      <div className="flex items-center justify-between gap-4 border-b border-[#45413c] py-4">
        <span className="flex items-center gap-3 font-mono text-[10px] font-semibold uppercase tracking-[0.12em] text-[#aaa49b]">
          <Users className="h-4 w-4" aria-hidden="true" />
          In room
        </span>
        <span className="frame-number font-heading text-3xl font-bold text-[#f2efe8]">{activeUsers ?? "--"}</span>
      </div>
      <button onClick={toggleFavorite} disabled={isLiking} className={`flex w-full items-center justify-between py-4 font-mono text-[10px] font-semibold uppercase tracking-[0.12em] transition-colors ${isFavorite ? "text-[#ed2f24]" : "text-[#aaa49b] hover:text-[#f2efe8]"}`}>
        <span>{isFavorite ? "Saved to your desk" : "Save this event"}</span>
        <Heart className="h-4 w-4" fill={isFavorite ? "currentColor" : "none"} aria-hidden="true" />
      </button>
      {notice && <p role="status" className="pb-4 font-mono text-[10px] font-semibold uppercase tracking-[0.12em] text-[#ed2f24]">{notice}</p>}
    </div>
  );
}
