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
  const [searchQuery, setSearchQuery] = useState<string>(searchParams.get("q") || "");
  const [debouncedQuery] = useDebounce(searchQuery, 500);

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

  const types = ["All", ...Array.from(new Set(initialEvents.map(e => e.type)))];
  const countries = ["All", ...Array.from(new Set(initialEvents.map(e => e.country).filter(Boolean)))];

  const filteredEvents = initialEvents.filter(e => {
    if (filterType !== "All" && e.type !== filterType) return false;
    if (filterCountry !== "All" && e.country !== filterCountry) return false;
    return true;
  });

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
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
        {filteredEvents.map((event) => (
          <EventCard key={event.id} event={event} />
        ))}
      </div>
      
      {filteredEvents.length === 0 && (
        <div className="py-20 text-center flex flex-col items-center justify-center border border-dashed border-white/10 rounded-2xl bg-white/5">
          <p className="text-xl font-semibold text-white mb-2">No events found</p>
          <p className="text-slate-500">Try adjusting your filters to find upcoming experiences.</p>
        </div>
      )}
    </div>
  );
}
