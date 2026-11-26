// Clerk's server package ships untranspiled ESM that Jest cannot load, and
// generateMetadata never calls auth(), so the boundary is stubbed at import.
jest.mock("@clerk/nextjs/server", () => ({ auth: jest.fn() }));
jest.mock("@clerk/nextjs", () => ({
  useUser: jest.fn(),
  useAuth: jest.fn(),
}));

import { generateMetadata } from "./page";

const params = (id: string) => ({ params: Promise.resolve({ id }) });

const mockFetch = (impl: jest.Mock) => {
  global.fetch = impl;
};

afterEach(() => {
  jest.restoreAllMocks();
});

describe("generateMetadata", () => {
  it("keeps resolvable event rooms out of the search index", async () => {
    mockFetch(jest.fn().mockResolvedValue({
      status: 200,
      ok: true,
      json: async () => ({ id: "abc", title: "Test Event" }),
    }));

    const meta = await generateMetadata(params("abc"));

    expect(meta.title).toBe("Test Event | LivePulse");
    expect(meta.robots).toMatchObject({ index: false, follow: false });
  });

  it("keeps unresolvable event rooms out of the search index", async () => {
    mockFetch(jest.fn().mockResolvedValue({ status: 404, ok: false }));

    const meta = await generateMetadata(params("missing"));

    expect(meta.title).toBe("Event room | LivePulse");
    expect(meta.robots).toMatchObject({ index: false, follow: false });
  });
});
