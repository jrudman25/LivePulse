import { metadata } from "./layout";

describe("root metadata", () => {
  it("carries the Google Search Console verification token", () => {
    expect(metadata.verification).toEqual({
      google: "3Ah_oDSBF5hKvgd8-4h1ZuJBrc1ecSVJMY2F10BYPV4",
    });
  });

  it("defines Open Graph and Twitter card metadata for link previews", () => {
    expect(metadata.openGraph).toMatchObject({
      type: "website",
      siteName: "LivePulse",
      title: "LivePulse | The live event desk",
      url: "/",
    });
    expect(metadata.twitter).toMatchObject({ card: "summary_large_image" });
  });
});
