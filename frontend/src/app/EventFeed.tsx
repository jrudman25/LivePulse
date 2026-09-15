"use client";

import { useState, useEffect } from "react";
import { useRouter, usePathname, useSearchParams } from "next/navigation";
import { useAuth } from "@clerk/nextjs";
import { Search, SlidersHorizontal, X } from "lucide-react";
import { useDebounce } from "use-debounce";
import EventCard from "./EventCard";
import { isEventList, type EventItem } from "@/lib/events";

// Matches the LIMIT the Go backend applies in HandleGetLiveEvents
const EVENT_PAGE_SIZE = 200;

export default function EventFeed({ initialEvents }: { initialEvents: EventItem[] }) {
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();

  const [filterType, setFilterType] = useState<string>("All");
  const [filterCountry, setFilterCountry] = useState<string>("All");
  const [filterFavorites, setFilterFavorites] = useState<string>("All");
  const [searchQuery, setSearchQuery] = useState<string>(searchParams.get("q") || "");
  const [debouncedQuery] = useDebounce(searchQuery, 500);

  // Pagination State mappings natively decoupled from Server Component
  const [events, setEvents] = useState<EventItem[]>(initialEvents);
  const [offset, setOffset] = useState<number>(initialEvents.length);
  const [hasMore, setHasMore] = useState<boolean>(initialEvents.length === EVENT_PAGE_SIZE);
  const [isLoading, setIsLoading] = useState<boolean>(false);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [roomCounts, setRoomCounts] = useState<Record<string, number>>({});
  const { getToken, isSignedIn } = useAuth();

  // Reset paginated state when the server payload changes (new search, refresh),
  // while preserving appended pages when the same events refetch unchanged.
  const eventsSignature = initialEvents.map(e => `${e.id}${e.is_favorite ? "f" : ""}`).join("|");
  const [prevSignature, setPrevSignature] = useState(eventsSignature);
  if (prevSignature !== eventsSignature) {
    setPrevSignature(eventsSignature);
    setEvents(initialEvents);
    setOffset(initialEvents.length);
    setHasMore(initialEvents.length === EVENT_PAGE_SIZE);
    setLoadError(null);
  }

  const searchParamsString = searchParams.toString();

  useEffect(() => {
    const params = new URLSearchParams(searchParamsString);
    if (debouncedQuery) {
      params.set("q", debouncedQuery);
    } else {
      params.delete("q");
    }

    // Extracted into a string, if the new URL parameters match the exact parameters currently live, we abort replacing
    // to prevent circular useEffect infinite API polling!
    const newQueryString = params.toString();
    if (newQueryString !== searchParamsString) {
      router.replace(newQueryString ? `${pathname}?${newQueryString}` : pathname, { scroll: false });
    }
  }, [debouncedQuery, pathname, router, searchParamsString]);

  // Fetch room counts for all loaded events in one batched request
  const eventIdsKey = events.map(e => e.id).join(",");
  useEffect(() => {
    if (!eventIdsKey) {return;}
    const ids = eventIdsKey.split(",").map(encodeURIComponent).join(",");
    const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
    const controller = new AbortController();
    fetch(`${API_URL}/api/sessions/stats?session_ids=${ids}`, { signal: controller.signal })
      .then(r => (r.ok ? r.json() : null))
      .then((data: unknown) => {
        if (data && typeof data === "object" && !Array.isArray(data)) {
          const counts = data as Record<string, unknown>;
          setRoomCounts(prev => {
            let changed = false;
            const next = { ...prev };
            for (const [id, count] of Object.entries(counts)) {
              if (typeof count === "number" && next[id] !== count) {
                next[id] = count;
                changed = true;
              }
            }
            return changed ? next : prev;
          });
        }
      })
      .catch(() => { });
    return () => controller.abort();
  }, [eventIdsKey]);

  const loadMoreEvents = async () => {
    if (isLoading || !hasMore) {return;}
    setIsLoading(true);

    try {
      setLoadError(null);
      const q = debouncedQuery ? `&q=${encodeURIComponent(debouncedQuery)}` : "";
      const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
      // Send the Clerk token so later pages keep favorite decoration.
      const headers: Record<string, string> = {};
      if (isSignedIn) {
        const token = await getToken();
        if (token) {headers["Authorization"] = `Bearer ${token}`;}
      }
      const res = await fetch(`${API_URL}/api/events?offset=${offset}${q}`, { headers, signal: AbortSignal.timeout(8000) });
      if (!res.ok) {throw new Error(`Event request failed with status ${res.status}`);}

      const data: unknown = await res.json();
      if (!isEventList(data)) {throw new Error("Event request returned an invalid response");}

      if (data.length < EVENT_PAGE_SIZE) {
        setHasMore(false);
      }

      if (data.length > 0) {
        setEvents(prev => {
          const seen = new Set(prev.map(e => e.id));
          return [...prev, ...data.filter(d => !seen.has(d.id))];
        });
        setOffset(prev => prev + data.length);
      }
    } catch (e) {
      console.error("Pagination Fetch Error", e);
      setLoadError("More events could not be loaded. Please try again.");
    } finally {
      setIsLoading(false);
    }
  };

  const types = ["All", ...Array.from(new Set(events.flatMap(e => e.type ? [e.type] : [])))];
  const countries = ["All", ...Array.from(new Set(events.flatMap(e => e.country ? [e.country] : [])))];

  const filteredEvents = events.filter(e => {
    if (filterType !== "All" && e.type !== filterType) {return false;}
    if (filterCountry !== "All" && e.country !== filterCountry) {return false;}
    if (filterFavorites === "Favorites" && !e.is_favorite) {return false;}
    return true;
  });

  const handleFavoriteUpdate = (id: string, isFav: boolean) => {
    setEvents(prev => prev.map(ev => ev.id === id ? { ...ev, is_favorite: isFav } : ev));
  };

  const clearFilters = () => {
    setFilterType("All");
    setFilterCountry("All");
    setFilterFavorites("All");
    setSearchQuery("");
  };

  const hasActiveFilters = filterType !== "All" || filterCountry !== "All" || filterFavorites !== "All" || searchQuery !== "";

  return (
    <div>
      <div className="mb-8 border-y border-[#67625b]">
        <div className="grid lg:grid-cols-[minmax(260px,1fr)_repeat(3,minmax(150px,0.4fr))]">
          <label className="group relative border-b border-[#45413c] lg:border-r lg:border-b-0">
            <span className="sr-only">Search Events</span>
            <Search className="absolute left-4 top-1/2 h-4 w-4 -translate-y-1/2 text-[#67625b] transition-colors group-focus-within:text-[#ed2f24]" aria-hidden="true" />
            <input
              type="search"
              placeholder="Search the event wire"
              className="h-14 w-full bg-transparent pr-4 pl-12 text-sm text-[#f2efe8] placeholder:text-[#67625b] focus:bg-[#171614] focus:outline-none"
              value={searchQuery}
              onChange={e => setSearchQuery(e.target.value)}
            />
          </label>
          <label className="grid grid-cols-[1fr_auto] items-center border-b border-[#45413c] px-4 lg:border-r lg:border-b-0">
            <span className="font-mono text-[9px] font-semibold uppercase tracking-[0.14em] text-[#67625b]">Type</span>
            <select aria-label="Filter by event type" className="h-14 max-w-[120px] bg-transparent text-right text-sm font-medium text-[#f2efe8] focus:outline-none" value={filterType} onChange={e => setFilterType(e.target.value)}>
              {types.map(t => <option key={t as string} value={t as string} className="bg-[#171614]">{t as string}</option>)}
            </select>
          </label>
          <label className="grid grid-cols-[1fr_auto] items-center border-b border-[#45413c] px-4 lg:border-r lg:border-b-0">
            <span className="font-mono text-[9px] font-semibold uppercase tracking-[0.14em] text-[#67625b]">Region</span>
            <select aria-label="Filter by country" className="h-14 max-w-[140px] bg-transparent text-right text-sm font-medium text-[#f2efe8] focus:outline-none" value={filterCountry} onChange={e => setFilterCountry(e.target.value)}>
              {countries.map(c => <option key={c as string} value={c as string} className="bg-[#171614]">{c === "All" ? "All Countries" : c as string}</option>)}
            </select>
          </label>
          <label className="grid grid-cols-[1fr_auto] items-center px-4">
            <span className="font-mono text-[9px] font-semibold uppercase tracking-[0.14em] text-[#67625b]">Saved</span>
            <select aria-label="Filter by saved events" className="h-14 max-w-[140px] bg-transparent text-right text-sm font-medium text-[#f2efe8] focus:outline-none" value={filterFavorites} onChange={e => setFilterFavorites(e.target.value)}>
              <option value="All" className="bg-[#171614]">All Events</option>
              <option value="Favorites" className="bg-[#171614]">Favorites Only</option>
            </select>
          </label>
        </div>
      </div>

      <div className="mb-4 flex items-center justify-between gap-4 font-mono text-[10px] font-semibold uppercase tracking-[0.14em] text-[#aaa49b]">
        <span className="flex items-center gap-2"><SlidersHorizontal className="h-3.5 w-3.5" aria-hidden="true" />{filteredEvents.length} results</span>
        {hasActiveFilters && (
          <button onClick={clearFilters} className="flex items-center gap-2 transition-colors hover:text-[#f2efe8]"><X className="h-3.5 w-3.5" aria-hidden="true" />Clear filters</button>
        )}
      </div>

      {loadError && (
        <div role="alert" className="mb-4 border border-[#ed2f24] bg-[#2c1513] px-4 py-3 text-sm text-[#ff8d86]">{loadError}</div>
      )}

      {filteredEvents.length === 0 ? (
        <div className="grid min-h-[280px] place-items-center border border-[#45413c] bg-[#171614] p-8 text-center">
          <div>
            <p className="font-heading text-4xl font-bold uppercase tracking-[-0.03em] text-[#f2efe8]">No events found</p>
            <p className="mt-2 text-sm text-[#aaa49b]">Nothing on the wire matches those filters.</p>
            <button onClick={clearFilters} className="mt-6 border border-[#67625b] px-5 py-3 font-mono text-[10px] font-semibold uppercase tracking-[0.14em] text-[#f2efe8] transition-colors hover:bg-[#f2efe8] hover:text-[#11100f]">Reset the desk</button>
          </div>
        </div>
      ) : (
        <>
          <div className="border-t border-[#45413c]">
            {filteredEvents.map((event, index) => (
              <EventCard key={event.id} event={event} index={index} activeUsers={roomCounts[event.id]} onFavoriteToggle={handleFavoriteUpdate} />
            ))}
          </div>

          {/* Infinity Scroll Load More UX Mapping */}
          {hasMore && (
            <div className="mt-10 flex justify-center">
              <button onClick={loadMoreEvents} disabled={isLoading} className="min-w-[220px] bg-[#ed2f24] px-7 py-4 font-mono text-[11px] font-semibold uppercase tracking-[0.14em] text-[#fffaf2] transition-colors hover:bg-[#f2efe8] hover:text-[#11100f] disabled:cursor-wait disabled:bg-[#45413c] disabled:text-[#aaa49b]">
                {isLoading ? "Loading..." : "Load more events"}
              </button>
            </div>
          )}
        </>
      )}
    </div>
  );
}
