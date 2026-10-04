export const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1";

const ACCESS_TOKEN_KEY = "campuscare_access_token";
const REFRESH_TOKEN_KEY = "campuscare_refresh_token";

export class APIError extends Error {
  status: number;
  details: string;

  constructor(status: number, message: string, details = "") {
    super(message);
    this.name = "APIError";
    this.status = status;
    this.details = details;
  }
}

export function getStoredToken(): string | null {
  if (typeof window === "undefined") return null;
  return window.localStorage.getItem(ACCESS_TOKEN_KEY);
}

export function getStoredRefreshToken(): string | null {
  if (typeof window === "undefined") return null;
  return window.localStorage.getItem(REFRESH_TOKEN_KEY);
}

export function setStoredTokens(access: string, refresh: string): void {
  if (typeof window === "undefined") return;
  window.localStorage.setItem(ACCESS_TOKEN_KEY, access);
  window.localStorage.setItem(REFRESH_TOKEN_KEY, refresh);
}

export function clearStoredTokens(): void {
  if (typeof window === "undefined") return;
  window.localStorage.removeItem(ACCESS_TOKEN_KEY);
  window.localStorage.removeItem(REFRESH_TOKEN_KEY);
}

// onSessionExpired lets AuthContext react to a refresh failure without api.ts
// having to import the context, which would create a cycle.
let sessionExpiredHandler: (() => void) | null = null;

export function setSessionExpiredHandler(handler: (() => void) | null): void {
  sessionExpiredHandler = handler;
}

type ErrorBody = { error?: string; message?: string; details?: string };

async function toError(response: Response): Promise<APIError> {
  const raw = await response.text();
  let body: ErrorBody = {};
  if (raw) {
    try {
      body = JSON.parse(raw) as ErrorBody;
    } catch {
      body = { error: raw };
    }
  }
  return new APIError(
    response.status,
    body.error || body.message || "An API error occurred",
    body.details || "",
  );
}

// refreshInFlight deduplicates concurrent refreshes so a page that fires several
// requests at once does not burn through the refresh token by racing itself.
let refreshInFlight: Promise<boolean> | null = null;

async function refreshAccessToken(): Promise<boolean> {
  if (refreshInFlight) return refreshInFlight;

  refreshInFlight = (async () => {
    const refreshToken = getStoredRefreshToken();
    if (!refreshToken) return false;

    try {
      const response = await fetch(`${API_BASE}/auth/refresh`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ refresh_token: refreshToken }),
      });

      if (!response.ok) return false;

      const payload = (await response.json()) as {
        data?: { access_token?: string; refresh_token?: string };
      };
      const access = payload.data?.access_token;
      const refresh = payload.data?.refresh_token;
      if (!access) return false;

      setStoredTokens(access, refresh || refreshToken);
      return true;
    } catch {
      return false;
    } finally {
      refreshInFlight = null;
    }
  })();

  return refreshInFlight;
}

export async function fetchAPI<T = unknown>(
  endpoint: string,
  options: RequestInit = {},
): Promise<T> {
  return request<T>(endpoint, options, true);
}

async function request<T>(
  endpoint: string,
  options: RequestInit,
  allowRetry: boolean,
): Promise<T> {
  const headers = new Headers(options.headers);
  if (!headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }

  const token = getStoredToken();
  if (token) {
    headers.set("Authorization", `Bearer ${token}`);
  }

  const response = await fetch(`${API_BASE}${endpoint}`, { ...options, headers });

  // An expired access token is recoverable with the refresh token, so the
  // request is retried once rather than logging the member out.
  if (response.status === 401 && allowRetry && getStoredRefreshToken()) {
    if (await refreshAccessToken()) {
      return request<T>(endpoint, options, false);
    }
    clearStoredTokens();
    sessionExpiredHandler?.();
  }

  if (!response.ok) {
    throw await toError(response);
  }

  const raw = await response.text();
  if (!raw) return null as T;

  try {
    return JSON.parse(raw) as T;
  } catch {
    throw new APIError(response.status, "Server returned a malformed response");
  }
}

/** Envelope returned by every successful CampusCare API response. */
export type ApiEnvelope<T> = {
  success: boolean;
  message: string;
  data: T;
};

/** Unwraps the standard `{ success, message, data }` envelope. */
export function unwrap<T>(envelope: ApiEnvelope<T>): T {
  return envelope.data;
}

/* ------------------------------------------------------------------ types */

export type CommunityMember = {
  id: string;
  email: string;
  display_name: string;
  role: string;
  status: string;
};

export type Club = {
  id: string;
  name: string;
  slug: string;
  description: string;
  category: string;
  verification_status: string;
  status: string;
  created_by: string;
  member_count?: number;
};

export type ClubMembership = {
  id: string;
  club_id: string;
  user_id: string;
  role: string;
  status: string;
  created_at: string;
};

export type Discussion = {
  id: string;
  club_id?: string | null;
  author_id: string;
  author_name?: string;
  title: string;
  body: string;
  category: string;
  status: string;
  is_pinned: boolean;
  reply_count: number;
  last_activity_at: string;
  created_at: string;
};

export type DiscussionReply = {
  id: string;
  discussion_id: string;
  author_id: string;
  author_name?: string;
  parent_id?: string | null;
  body: string;
  created_at: string;
};

export type CampusNotification = {
  id: string;
  user_id: string;
  type: string;
  title: string;
  body: string;
  link?: string | null;
  read_at?: string | null;
  created_at: string;
};

export type CampusEvent = {
  id: string;
  title: string;
  description: string;
  category: string;
  organizer_id: string;
  venue: string;
  is_online: boolean;
  start_time: string;
  end_time: string;
  registration_deadline: string;
  capacity: number;
  status: string;
  registration_status?: string;
};

export type Opportunity = {
  id: string;
  title: string;
  description: string;
  type: string;
  category: string;
  organizer_id: string;
  start_date: string;
  end_date: string;
  registration_deadline: string;
  location: string;
  delivery_mode: string;
  capacity: number;
  status: string;
};

export type CampusResource = {
  id: string;
  name: string;
  resource_type: string;
  description: string;
  location: string;
  capacity: number;
  is_bookable: boolean;
  requires_auth: boolean;
  status: string;
};

export type OpportunityApplication = {
  id: string;
  opportunity_id: string;
  user_id: string;
  title?: string;
  status: string;
  cover_note: string;
  resume_url?: string | null;
  applied_at: string;
  updated_at: string;
};

export type ResourceBooking = {
  id: string;
  resource_id: string;
  resource_nm?: string;
  start_time: string;
  end_time: string;
  note?: string;
  status: string;
};

export type SearchResult = {
  entity_type: string;
  entity_id: string;
  title: string;
  summary: string;
  category: string;
  url: string;
  event_at?: string | null;
};

/* ------------------------------------------------------------ endpoints */

type ListMeta = { count?: number; unread?: number; total?: number };

function list<T>(path: string): Promise<ApiEnvelope<T & ListMeta>> {
  return fetchAPI<ApiEnvelope<T & ListMeta>>(path);
}

export const api = {
  me: () => fetchAPI<ApiEnvelope<CommunityMember>>("/auth/me"),

  updateProfile: (displayName: string) =>
    fetchAPI<ApiEnvelope<CommunityMember>>("/auth/me", {
      method: "PATCH",
      body: JSON.stringify({ display_name: displayName }),
    }),

  discussions: (params?: { clubId?: string; limit?: number }) => {
    const search = new URLSearchParams();
    if (params?.clubId) search.set("club_id", params.clubId);
    if (params?.limit) search.set("limit", String(params.limit));
    const query = search.toString();
    return list<{ discussions: Discussion[] }>(`/community/discussions${query ? `?${query}` : ""}`);
  },

  discussion: (id: string) =>
    fetchAPI<ApiEnvelope<{ discussion: Discussion; replies: DiscussionReply[] }>>(
      `/community/discussions/${id}`,
    ),

  createDiscussion: (input: { title: string; body: string; category?: string; clubId?: string }) =>
    fetchAPI<ApiEnvelope<Discussion>>("/community/discussions", {
      method: "POST",
      body: JSON.stringify({
        title: input.title,
        body: input.body,
        category: input.category,
        club_id: input.clubId,
      }),
    }),

  replyToDiscussion: (id: string, body: string) =>
    fetchAPI<ApiEnvelope<DiscussionReply>>(`/community/discussions/${id}/replies`, {
      method: "POST",
      body: JSON.stringify({ body }),
    }),

  notifications: (unreadOnly = false) =>
    fetchAPI<ApiEnvelope<{ notifications: CampusNotification[]; unread: number }>>(
      `/community/notifications${unreadOnly ? "?unread=true" : ""}`,
    ),

  markNotificationRead: (id: string) =>
    fetchAPI<ApiEnvelope<null>>(`/community/notifications/${id}/read`, { method: "POST" }),

  markAllNotificationsRead: () =>
    fetchAPI<ApiEnvelope<{ updated: number }>>("/community/notifications/read-all", {
      method: "POST",
    }),

  clubs: (params?: { category?: string; q?: string }) => {
    const search = new URLSearchParams();
    if (params?.category) search.set("category", params.category);
    if (params?.q) search.set("q", params.q);
    const query = search.toString();
    return list<{ clubs: Club[] }>(`/clubs${query ? `?${query}` : ""}`);
  },

  club: (id: string) =>
    fetchAPI<ApiEnvelope<{ club: Club; member_count: number; members: ClubMembership[] }>>(
      `/clubs/${id}`,
    ),

  joinClub: (id: string) => fetchAPI<ApiEnvelope<ClubMembership>>(`/clubs/${id}/join`, { method: "POST" }),

  leaveClub: (id: string) => fetchAPI<ApiEnvelope<null>>(`/clubs/${id}/leave`, { method: "POST" }),

  myClubs: () => list<{ clubs: Club[] }>("/me/clubs"),

  events: (params?: { category?: string; q?: string }) => {
    const search = new URLSearchParams();
    if (params?.category) search.set("category", params.category);
    if (params?.q) search.set("q", params.q);
    const query = search.toString();
    return list<{ events: CampusEvent[] }>(`/events${query ? `?${query}` : ""}`);
  },

  event: (id: string) => fetchAPI<ApiEnvelope<{ event: CampusEvent }>>(`/events/${id}`),

  registerForEvent: (id: string) =>
    fetchAPI<ApiEnvelope<unknown>>(`/events/${id}/register`, { method: "POST" }),

  opportunities: (params?: { category?: string; q?: string; saved?: boolean }) => {
    const search = new URLSearchParams();
    if (params?.category) search.set("category", params.category);
    if (params?.q) search.set("q", params.q);
    if (params?.saved) search.set("saved", "true");
    const query = search.toString();
    return list<{ opportunities: Opportunity[] }>(`/opportunities${query ? `?${query}` : ""}`);
  },

  applyToOpportunity: (id: string, coverNote: string) =>
    fetchAPI<ApiEnvelope<unknown>>(`/community/opportunities/${id}/apply`, {
      method: "POST",
      body: JSON.stringify({ cover_note: coverNote }),
    }),

  myApplications: () => fetchAPI<ApiEnvelope<{ applications: OpportunityApplication[] }>>("/me/opportunities"),

  withdrawApplication: (id: string) =>
    fetchAPI<ApiEnvelope<null>>(`/community/opportunities/${id}/apply`, { method: "DELETE" }),

  services: (params?: { type?: string; q?: string }) => {
    const search = new URLSearchParams();
    if (params?.type) search.set("type", params.type);
    if (params?.q) search.set("q", params.q);
    const query = search.toString();
    return list<{ resources: CampusResource[] }>(`/community/services${query ? `?${query}` : ""}`);
  },

  bookService: (id: string, startTime: string, endTime: string, note?: string) =>
    fetchAPI<ApiEnvelope<unknown>>(`/community/services/${id}/bookings`, {
      method: "POST",
      body: JSON.stringify({ start_time: startTime, end_time: endTime, note }),
    }),

  myBookings: (upcoming = false) =>
    fetchAPI<ApiEnvelope<{ bookings: ResourceBooking[] }>>(
      `/community/bookings${upcoming ? "?upcoming=true" : ""}`,
    ),

  cancelBooking: (id: string) =>
    fetchAPI<ApiEnvelope<null>>(`/community/bookings/${id}`, { method: "DELETE" }),

  search: (q: string, type?: string) => {
    const search = new URLSearchParams({ q });
    if (type) search.set("type", type);
    return fetchAPI<
      ApiEnvelope<{ query: string; total: number; results: SearchResult[]; by_type: Record<string, number> }>
    >(`/community/search?${search.toString()}`);
  },
};