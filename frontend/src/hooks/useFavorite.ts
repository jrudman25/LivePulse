"use client";

import { useEffect, useState } from "react";
import { useAuth } from "@clerk/nextjs";

// Shared favorite-toggle behavior for EventCard and ArenaStatsTracker:
// token fetch, POST/DELETE against the Go API, loading state, and an
// auto-clearing inline notice for signed-out users.
export function useFavorite(eventId: string, initial = false) {
  const { getToken, isSignedIn } = useAuth();
  const [isFavorite, setIsFavorite] = useState(initial);
  const [isLiking, setIsLiking] = useState(false);
  const [notice, setNotice] = useState<string | null>(null);

  useEffect(() => {
    if (!notice) {return;}
    const timeoutId = setTimeout(() => setNotice(null), 3500);
    return () => clearTimeout(timeoutId);
  }, [notice]);

  const toggleFavorite = async (e?: React.MouseEvent) => {
    e?.preventDefault();
    if (!isSignedIn) {
      setNotice("Sign in to save events");
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
        body: JSON.stringify({ event_id: eventId }),
        signal: AbortSignal.timeout(8000)
      });
      if (res.ok) {
        setIsFavorite(!isFavorite);
        return true;
      }
    } catch (err) {
      console.error("Failed to update favorite:", err);
    } finally {
      setIsLiking(false);
    }
    return false;
  };

  return { isFavorite, setIsFavorite, isLiking, notice, toggleFavorite };
}
