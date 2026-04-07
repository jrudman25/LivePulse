"use client";

import { useState, useEffect } from "react";
import { useRouter, usePathname, useSearchParams } from "next/navigation";
import { useDebounce } from "use-debounce";
import EventCard from "./EventCard";

export default function EventFeed({ initialEvents }: { initialEvents: any[] }) {
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();

  const [filterType, setFilterType] = useState<string>("All");
  const [filterCountry, setFilterCountry] = useState<string>("All");
  const [filterFavorites, setFilterFavorites] = useState<string>("All");
  const [searchQuery, setSearchQuery] = useState<string>(searchParams.get("q") || "");
  const [debouncedQuery] = useDebounce(searchQuery, 500);

  // Pagination State mappings natively decoupled from Server Component
  const [events, setEvents] = useState<any[]>(initialEvents);
  const [offset, setOffset] = useState<number>(initialEvents.length);
  const [hasMore, setHasMore] = useState<boolean>(initialEvents.length === 50);
  const [isLoading, setIsLoading] = useState<boolean>(false);

  // Synchronize dynamic Server Component payloads into local memory when SSR updates
  useEffect(() => {
    setEvents(initialEvents);
    setOffset(initialEvents.length);
    setHasMore(initialEvents.length === 50);
  }, [initialEvents]);

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
      router.replace(`${pathname}?${newQueryString}`, { scroll: false });
    }
  }, [debouncedQuery, pathname, router, searchParamsString]);

  const loadMoreEvents = async () => {
    if (isLoading || !hasMore) return;
    setIsLoading(true);

    try {
      const q = searchParamsString ? `&q=${encodeURIComponent(debouncedQuery)}` : "";
      const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
      const res = await fetch(`${API_URL}/api/events?offset=${offset}${q}`);
      const data = await res.json();
      
      if (!data || data.length < 50) {
        setHasMore(false);
      }
      
      if (data && data.length > 0) {
        setEvents(prev => [...prev, ...data]);
        setOffset(prev => prev + data.length);
      }
    } catch (e) {
      console.error("Pagination Fetch Error", e);
    } finally {
      setIsLoading(false);
    }
  };

  const types = ["All", ...Array.from(new Set(events.map(e => e.type).filter(Boolean)))];
  const countries = ["All", ...Array.from(new Set(events.map(e => e.country).filter(Boolean)))];

  const filteredEvents = events.filter(e => {
    if (filterType !== "All" && e.type !== filterType) return false;
    if (filterCountry !== "All" && e.country !== filterCountry) return false;
    if (filterFavorites === "Favorites" && !e.is_favorite) return false;
    return true;
  });

  const handleFavoriteUpdate = (id: string, isFav: boolean) => {
    setEvents(prev => prev.map(ev => ev.id === id ? { ...ev, is_favorite: isFav } : ev));
  };

  return (
    <div>
      <div className="flex flex-col sm:flex-row gap-4 mb-8">
         <div className="flex-1 flex flex-col gap-1">
           <label className="text-xs font-semibold uppercase tracking-wider text-slate-500 pl-1">Search Events</label>
           <input 
              type="text"
              placeholder="Search by title..."
              className="bg-white/5 border border-white/10 rounded-xl px-4 py-2.5 text-white focus:outline-none focus:border-fuchsia-500 focus:ring-1 focus:ring-fuchsia-500 transition-all font-medium w-full"
              value={searchQuery} onChange={e => setSearchQuery(e.target.value)}
           />
         </div>
         <div className="flex flex-col gap-1">
           <label className="text-xs font-semibold uppercase tracking-wider text-slate-500 pl-1">Event Type</label>
           <select 
              className="bg-white/5 border border-white/10 rounded-xl px-4 py-2.5 text-white focus:outline-none focus:border-fuchsia-500 focus:ring-1 focus:ring-fuchsia-500 transition-all font-medium min-w-[180px]"
              value={filterType} onChange={e => setFilterType(e.target.value)}>
               {types.map(t => <option key={t as string} value={t as string} className="bg-slate-900">{t as string}</option>)}
           </select>
         </div>
         <div className="flex flex-col gap-1">
           <label className="text-xs font-semibold uppercase tracking-wider text-slate-500 pl-1">Country</label>
           <select 
              className="bg-white/5 border border-white/10 rounded-xl px-4 py-2.5 text-white focus:outline-none focus:border-fuchsia-500 focus:ring-1 focus:ring-fuchsia-500 transition-all font-medium min-w-[180px]"
              value={filterCountry} onChange={e => setFilterCountry(e.target.value)}>
               {countries.map(c => <option key={c as string} value={c as string} className="bg-slate-900">{c === "All" ? "All Countries" : c as string}</option>)}
           </select>
         </div>
         <div className="flex flex-col gap-1">
           <label className="text-xs font-semibold uppercase tracking-wider text-slate-500 pl-1">Status</label>
           <select 
              className="bg-white/5 border border-white/10 rounded-xl px-4 py-2.5 text-white focus:outline-none focus:border-fuchsia-500 focus:ring-1 focus:ring-fuchsia-500 transition-all font-medium min-w-[160px]"
              value={filterFavorites} onChange={e => setFilterFavorites(e.target.value)}>
               <option value="All" className="bg-slate-900">All Events</option>
               <option value="Favorites" className="bg-slate-900">Favorites Only</option>
           </select>
         </div>
      </div>

      {filteredEvents.length === 0 ? (
        <div className="flex flex-col items-center justify-center p-12 bg-white/5 border border-white/10 rounded-2xl text-center shadow-lg backdrop-blur-sm">
          <p className="text-slate-400 font-medium">No events found matching those filters.</p>
          <button onClick={() => { setFilterType("All"); setFilterCountry("All"); setFilterFavorites("All"); setSearchQuery(""); }} className="mt-4 px-6 py-2 rounded-full border border-white/10 hover:bg-white/10 transition-colors text-white text-sm font-semibold">
            Clear Filters
          </button>
        </div>
      ) : (
        <>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
            {filteredEvents.map(event => (
              <EventCard key={event.id} event={event} onFavoriteToggle={handleFavoriteUpdate} />
            ))}
          </div>

          {/* Infinity Scroll Load More UX Mapping */}
          {hasMore && (
            <div className="mt-12 flex justify-center">
              <button 
                onClick={loadMoreEvents} 
                disabled={isLoading}
                className="bg-gradient-to-r from-fuchsia-600/80 to-blue-600/80 hover:from-fuchsia-500 hover:to-blue-500 border border-white/10 disabled:opacity-50 text-white rounded-full px-10 py-3 font-semibold transition-all shadow-[0_0_20px_rgba(217,70,239,0.2)] disabled:shadow-none min-w-[200px]"
              >
                {isLoading ? (
                  <span className="flex items-center justify-center gap-2">
                    <svg className="animate-spin h-4 w-4 text-white" fill="none" viewBox="0 0 24 24"><circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle><path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path></svg>
                    Loading...
                  </span>
                ) : "Load More Events"}
              </button>
            </div>
          )}
        </>
      )}
    </div>
  );
}
