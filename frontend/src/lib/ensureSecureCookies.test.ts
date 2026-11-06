/** @jest-environment node */
import { ensureSecureCookies } from "./ensureSecureCookies";

describe("ensureSecureCookies", () => {
  it("adds Secure to a cookie that lacks it", () => {
    const headers = new Headers();
    headers.append("set-cookie", "session=abc; Path=/; HttpOnly");

    ensureSecureCookies(headers);

    expect(headers.getSetCookie()).toEqual(["session=abc; Path=/; HttpOnly; Secure"]);
  });

  it("does not duplicate an existing Secure attribute", () => {
    const headers = new Headers();
    headers.append("set-cookie", "a=1; Secure; HttpOnly");
    headers.append("set-cookie", "b=2; Path=/; secure");

    ensureSecureCookies(headers);

    expect(headers.getSetCookie()).toEqual([
      "a=1; Secure; HttpOnly",
      "b=2; Path=/; secure",
    ]);
  });

  it("preserves all cookies when multiple are set", () => {
    const headers = new Headers();
    headers.append("set-cookie", "a=1; HttpOnly");
    headers.append("set-cookie", "b=2; Secure");
    headers.append("set-cookie", "c=3; SameSite=Lax");

    ensureSecureCookies(headers);

    expect(headers.getSetCookie()).toEqual([
      "a=1; HttpOnly; Secure",
      "b=2; Secure",
      "c=3; SameSite=Lax; Secure",
    ]);
  });

  it("leaves headers unchanged when there is no Set-Cookie", () => {
    const headers = new Headers({ "content-type": "text/html" });

    ensureSecureCookies(headers);

    expect(headers.getSetCookie()).toEqual([]);
    expect(headers.get("content-type")).toBe("text/html");
  });
});
