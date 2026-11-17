/**
 * Image hosts that cannot be drawn onto a share canvas directly (no CORS), so their images go
 * through the API proxy. Exact hostnames only; the proxy itself enforces the same allowlist.
 */
export const SHARE_PROXY_HOSTS: readonly string[] = ["image.tmdb.org", "s4.anilist.co"]

/** Proxy route, relative to the API base URL. */
export const SHARE_PROXY_PATH = "/share/proxy-image"
