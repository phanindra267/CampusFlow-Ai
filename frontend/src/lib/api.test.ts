import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  clearStoredTokens,
  fetchAPI,
  getStoredToken,
  setSessionExpiredHandler,
  setStoredTokens,
} from "./api";

function createStorage() {
  const values = new Map<string, string>();

  return {
    getItem: (key: string) => values.get(key) ?? null,
    setItem: (key: string, value: string) => values.set(key, value),
    removeItem: (key: string) => values.delete(key),
  };
}

function jsonResponse(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

describe("API authentication", () => {
  let storage: ReturnType<typeof createStorage>;

  beforeEach(() => {
    storage = createStorage();
    vi.stubGlobal("window", { localStorage: storage });
    vi.stubGlobal("fetch", vi.fn<typeof fetch>());
  });

  afterEach(() => {
    clearStoredTokens();
    setSessionExpiredHandler(null);
    vi.unstubAllGlobals();
  });

  it("stores and clears access and refresh tokens", () => {
    setStoredTokens("access-token", "refresh-token");

    expect(getStoredToken()).toBe("access-token");

    clearStoredTokens();

    expect(getStoredToken()).toBeNull();
  });

  it("refreshes an expired token and retries the original request", async () => {
    setStoredTokens("expired-access", "refresh-token");
    const fetchMock = vi.mocked(fetch);
    fetchMock
      .mockResolvedValueOnce(jsonResponse({ error: "expired" }, 401))
      .mockResolvedValueOnce(
        jsonResponse({
          data: { access_token: "new-access", refresh_token: "new-refresh" },
        }),
      )
      .mockResolvedValueOnce(jsonResponse({ data: { id: "member-1" } }));

    await expect(fetchAPI("/auth/me")).resolves.toEqual({
      data: { id: "member-1" },
    });

    expect(fetchMock).toHaveBeenCalledTimes(3);
    expect(
      new Headers(fetchMock.mock.calls[2][1]?.headers).get("Authorization"),
    ).toBe("Bearer new-access");
    expect(getStoredToken()).toBe("new-access");
  });

  it("clears the session and reports an API error when refresh fails", async () => {
    setStoredTokens("expired-access", "invalid-refresh");
    const fetchMock = vi.mocked(fetch);
    const onExpired = vi.fn();
    setSessionExpiredHandler(onExpired);
    fetchMock
      .mockResolvedValueOnce(jsonResponse({ error: "expired" }, 401))
      .mockResolvedValueOnce(jsonResponse({ error: "invalid refresh" }, 401));

    await expect(fetchAPI("/auth/me")).rejects.toMatchObject({
      name: "APIError",
      status: 401,
    });

    expect(getStoredToken()).toBeNull();
    expect(storage.getItem("campuscare_refresh_token")).toBeNull();
    expect(onExpired).toHaveBeenCalledOnce();
  });
});
