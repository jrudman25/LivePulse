"use client";

import { useState, useEffect } from "react";
import Link from "next/link";
import { ArrowUpRight, Heart, Users } from "lucide-react";
import { useAuth } from "@clerk/nextjs";

export type EventItem = {
  id: string;
  type?: string;
  title: string;
  country?: string;
  location?: string;
  start_time?: string;
  end_time?: string;
  is_favorite?: boolean;
};

// Add optional callback updating parent DOM states cleanly
export default function EventCard({ event, index = 0, onFavoriteToggle }: { event: EventItem, index?: number, onFavoriteToggle?: (_id: string, _isFav: boolean) => void }) {
  const [isFavorite, setIsFavorite] = useState(event.is_favorite || false);
  const { getToken, isSignedIn } = useAuth();
  const [isLiking, setIsLiking] = useState(false);
  const [activeUsers, setActiveUsers] = useState<number | null>(null);

  useEffect(() => {
    setIsFavorite(Boolean(event.is_favorite));
  }, [event.is_favorite]);

  useEffect(() => {
    // Poll the Go WebSocket Hub for active connected users passively
    const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
    fetch(`${API_URL}/api/sessions/stats?session_id=${event.id}`)
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
      const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
      const res = await fetch(`${API_URL}/api/favorites`, {
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

  const startTime = new Date(event.start_time || "");
  const eventNumber = String(index + 1).padStart(2, "0");

  return (
    <article className="group relative grid min-h-[190px] border-b border-[#45413c] bg-[#11100f] transition-colors hover:bg-[#171614] md:grid-cols-[96px_160px_minmax(0,1fr)_220px_72px]">
      <div className="frame-number border-b border-[#45413c] p-4 font-mono text-[10px] text-[#67625b] md:border-r md:border-b-0 md:p-5">{eventNumber}</div>

      <div className="border-b border-[#45413c] p-4 md:border-r md:border-b-0 md:p-5">
        {!Number.isNaN(startTime.getTime()) ? (
          <>
            <span className="block font-heading text-5xl font-black uppercase leading-none tracking-[-0.05em] text-[#f2efe8]">{startTime.toLocaleDateString(undefined, { day: "2-digit" })}</span>
            <span className="mt-1 block font-mono text-[10px] font-semibold uppercase tracking-[0.14em] text-[#aaa49b]">{startTime.toLocaleDateString(undefined, { month: "short", weekday: "short" })}</span>
            <span className="mt-5 block font-mono text-[11px] text-[#f2efe8]">{startTime.toLocaleTimeString(undefined, { hour: "numeric", minute: "2-digit" })}</span>
          </>
        ) : (
          <span className="font-mono text-xs uppercase text-[#aaa49b]">Time TBA</span>
        )}
      </div>

      <div className="flex flex-col justify-between p-5 sm:p-7">
        <div>
          <div className="mb-4 flex flex-wrap items-center gap-3 font-mono text-[9px] font-semibold uppercase tracking-[0.16em]">
            <span className="text-[#aaa49b]">{event.type || "Live event"}</span>
          </div>
          <h2 className="max-w-3xl font-heading text-3xl font-bold uppercase leading-[0.95] tracking-[-0.035em] text-[#f2efe8] transition-colors group-hover:text-white sm:text-4xl">{event.title}</h2>
        </div>
        <Link href={`/events/${event.id}`} className="after:absolute after:inset-0 focus-visible:after:outline-2 focus-visible:after:outline-offset-[-3px] focus-visible:after:outline-[#ed2f24]" aria-label={`Open ${event.title}`} />
      </div>

      <div className="flex flex-col justify-between border-t border-[#45413c] p-5 md:border-t-0 md:border-l">
        <div>
          <span className="font-mono text-[9px] font-semibold uppercase tracking-[0.14em] text-[#67625b]">Venue</span>
          <p className="mt-2 text-sm leading-snug text-[#d1cbc1]">{event.location || "Location to be announced"}</p>
          {event.country && <p className="mt-1 font-mono text-[10px] uppercase tracking-[0.12em] text-[#67625b]">{event.country}</p>}
        </div>
        {activeUsers !== null && (
          <div className="mt-5 flex items-center gap-2 font-mono text-[10px] uppercase tracking-[0.1em] text-[#aaa49b]">
            <Users className="h-3.5 w-3.5" aria-hidden="true" />
            {activeUsers} in room
          </div>
        )}
      </div>

      <div className="relative z-10 flex items-center justify-between border-t border-[#45413c] px-4 py-3 md:flex-col md:border-t-0 md:border-l md:px-0 md:py-5">
        <button onClick={toggleFavorite} disabled={isLiking} className={`grid h-10 w-10 place-items-center border transition-colors ${isFavorite ? "border-[#ed2f24] bg-[#ed2f24] text-[#fffaf2]" : "border-[#45413c] text-[#aaa49b] hover:border-[#f2efe8] hover:text-[#f2efe8]"}`} aria-label={isFavorite ? "Remove from favorites" : "Add to favorites"}>
          <Heart className="h-4 w-4" fill={isFavorite ? "currentColor" : "none"} aria-hidden="true" />
        </button>
        <ArrowUpRight className="h-5 w-5 text-[#67625b] transition-all group-hover:-translate-y-1 group-hover:translate-x-1 group-hover:text-[#f2efe8]" aria-hidden="true" />
      </div>
    </article>
  );
}
