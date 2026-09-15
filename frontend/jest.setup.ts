import '@testing-library/jest-dom'

// jsdom does not provide fetch; stub it so components issuing background
// requests resolve to an empty payload unless a test overrides it.
global.fetch = jest.fn(() =>
  Promise.resolve({ ok: true, json: () => Promise.resolve({}) } as Response)
) as unknown as typeof fetch
