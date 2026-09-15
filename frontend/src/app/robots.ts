import type { MetadataRoute } from "next";

const BASE_URL = "https://livepulse-hq.vercel.app";

export default function robots(): MetadataRoute.Robots {
  return {
    rules: [
      {
        userAgent: "*",
        allow: "/",
        disallow: ["/sign-in", "/sign-up"],
      },
    ],
    sitemap: `${BASE_URL}/sitemap.xml`,
  };
}
