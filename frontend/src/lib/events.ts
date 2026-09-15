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

export type EventDetails = {
  id: string;
  title: string;
  type?: string;
  location?: string;
  country?: string;
  start_time?: string;
  end_time?: string;
};

const OPTIONAL_STRING_FIELDS = ["type", "country", "location", "start_time", "end_time"] as const;

export function isEventItem(value: unknown): value is EventItem {
  if (typeof value !== "object" || value === null) {return false;}
  const e = value as Record<string, unknown>;
  if (typeof e.id !== "string" || e.id.length === 0) {return false;}
  if (typeof e.title !== "string" || e.title.length === 0) {return false;}
  for (const key of OPTIONAL_STRING_FIELDS) {
    if (e[key] !== undefined && e[key] !== null && typeof e[key] !== "string") {return false;}
  }
  if (e.is_favorite !== undefined && e.is_favorite !== null && typeof e.is_favorite !== "boolean") {return false;}
  return true;
}

export function isEventList(value: unknown): value is EventItem[] {
  return Array.isArray(value) && value.every(isEventItem);
}

function optionalString(value: unknown): string | undefined {
  return typeof value === "string" && value.length > 0 ? value : undefined;
}

export function parseEventDetails(value: unknown, fallbackId: string): EventDetails | null {
  if (typeof value !== "object" || value === null) {return null;}
  const v = value as Record<string, unknown>;
  return {
    id: optionalString(v.id) ?? fallbackId,
    title: optionalString(v.title) ?? "Live Session",
    type: optionalString(v.type),
    location: optionalString(v.location),
    country: optionalString(v.country),
    start_time: optionalString(v.start_time),
    end_time: optionalString(v.end_time),
  };
}
