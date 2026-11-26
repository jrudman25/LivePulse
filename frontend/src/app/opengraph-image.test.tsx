import OpengraphImage, { alt, size, contentType } from "./opengraph-image";

// next/og renders via WASM at request time; the mock captures the args our
// component passes so the test stays fast and environment-independent.
jest.mock("next/og", () => ({
  ImageResponse: jest.fn(function ImageResponseMock(this: object, element: unknown, options: unknown) {
    return { element, options };
  }),
}));

import { ImageResponse } from "next/og";

describe("opengraph-image", () => {
  it("declares the standard 1200x630 OG card format", () => {
    expect(size).toEqual({ width: 1200, height: 630 });
    expect(contentType).toBe("image/png");
    expect(alt).toBe("LivePulse | The live event desk");
  });

  it("renders an ImageResponse at the declared size", () => {
    const result = OpengraphImage() as unknown as { element: unknown; options: unknown };

    expect(ImageResponse).toHaveBeenCalledTimes(1);
    expect(result.options).toEqual({ width: 1200, height: 630 });
    expect(result.element).toBeTruthy();
  });
});
