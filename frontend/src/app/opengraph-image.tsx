import { ImageResponse } from "next/og";

export const alt = "LivePulse | The live event desk";
export const size = { width: 1200, height: 630 };
export const contentType = "image/png";

export default function OpengraphImage() {
  return new ImageResponse(
    (
      <div
        style={{
          width: "100%",
          height: "100%",
          display: "flex",
          flexDirection: "column",
          alignItems: "center",
          justifyContent: "center",
          background: "#11100f",
          border: "24px solid #45413c",
        }}
      >
        <div
          style={{
            display: "flex",
            alignItems: "center",
            gap: 32,
          }}
        >
          <div
            style={{
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
              width: 120,
              height: 120,
              background: "#ed2f24",
              color: "#fffaf2",
              fontSize: 64,
              fontWeight: 800,
            }}
          >
            LP
          </div>
          <div
            style={{
              display: "flex",
              color: "#f2efe8",
              fontSize: 128,
              fontWeight: 800,
              letterSpacing: "-0.03em",
              textTransform: "uppercase",
            }}
          >
            LivePulse
          </div>
        </div>
        <div
          style={{
            display: "flex",
            marginTop: 40,
            color: "#aaa49b",
            fontSize: 32,
            letterSpacing: "0.16em",
            textTransform: "uppercase",
          }}
        >
          Find the room. Join the crowd.
        </div>
      </div>
    ),
    { ...size },
  );
}
