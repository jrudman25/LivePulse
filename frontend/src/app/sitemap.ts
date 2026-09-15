import type { MetadataRoute } from "next";

const BASE_URL = "https://livepulse-hq.vercel.app";

// Only stable, indexable routes are listed. Event rooms are ephemeral
// (24-hour window) and are intentionally excluded.
export default function sitemap(): MetadataRoute.Sitemap {
  const now = new Date();
  return [
    { url: BASE_URL, lastModified: now, changeFrequency: "daily", priority: 1 },
    { url: `${BASE_URL}/events`, lastModified: now, changeFrequency: "hourly", priority: 0.9 },
    { url: `${BASE_URL}/help`, lastModified: now, changeFrequency: "monthly", priority: 0.4 },
    { url: `${BASE_URL}/privacy`, lastModified: now, changeFrequency: "monthly", priority: 0.3 },
  ];
}
