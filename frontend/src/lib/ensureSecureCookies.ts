export function ensureSecureCookies(headers: Headers): void {
  const cookies = headers.getSetCookie();
  if (cookies.length === 0) {
    return;
  }
  headers.delete("set-cookie");
  for (const cookie of cookies) {
    const hasSecure = cookie
      .split(";")
      .slice(1)
      .some((attribute) => attribute.trim().toLowerCase() === "secure");
    headers.append("set-cookie", hasSecure ? cookie : `${cookie}; Secure`);
  }
}
