"use client";

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  useSyncExternalStore,
  type ReactNode,
} from "react";
import {
  APIError,
  clearStoredTokens,
  fetchAPI,
  setSessionExpiredHandler,
  setStoredTokens,
} from "@/lib/api";

export type User = {
  id: string;
  email: string;
  displayName: string;
  role: string;
};

type ApiEnvelope<T> = {
  success: boolean;
  message?: string;
  data: T;
};

type LoginPayload = {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  user: {
    id: string;
    email: string;
    display_name?: string;
    role: string;
  };
};

type AuthContextType = {
  user: User | null;
  login: (email: string, password: string) => Promise<User>;
  logout: () => void;
  /** Re-reads the signed-in member from the API, used after a profile change. */
  refreshProfile: () => Promise<User | null>;
  isLoading: boolean;
  isReady: boolean;
};

const USER_KEY = "campuscare_user";
const ACCESS_TOKEN_KEY = "campuscare_access_token";
const REFRESH_TOKEN_KEY = "campuscare_refresh_token";

const AuthContext = createContext<AuthContextType | undefined>(undefined);

const listeners = new Set<() => void>();

function notify() {
  for (const listener of listeners) listener();
}

function subscribe(listener: () => void): () => void {
  listeners.add(listener);
  window.addEventListener("storage", listener);
  return () => {
    listeners.delete(listener);
    window.removeEventListener("storage", listener);
  };
}

function getServerSnapshot(): null {
  return null;
}

function subscribeToHydration(): () => void {
  return () => {};
}

function getClientReadySnapshot(): boolean {
  return true;
}

function getServerReadySnapshot(): boolean {
  return false;
}

let cachedKey: string | null = null;
let cachedUser: User | null = null;

function readUser(): { key: string; user: User | null } {
  if (typeof window === "undefined") return { key: "", user: null };

  const access = window.localStorage.getItem(ACCESS_TOKEN_KEY);
  const refresh = window.localStorage.getItem(REFRESH_TOKEN_KEY);
  const raw = window.localStorage.getItem(USER_KEY);
  const key = `${access ?? ""}::${refresh ?? ""}::${raw ?? ""}`;

  if (!raw) return { key, user: null };

  try {
    return { key, user: JSON.parse(raw) as User };
  } catch {
    return { key, user: null };
  }
}

function getSnapshot(): User | null {
  const { key, user } = readUser();
  if (key !== cachedKey) {
    cachedKey = key;
    cachedUser = user;
  }
  return cachedUser;
}

function persistSession(tokens: { access: string; refresh: string }, user: User) {
  setStoredTokens(tokens.access, tokens.refresh);
  window.localStorage.setItem(USER_KEY, JSON.stringify(user));
  notify();
}

function clearSession() {
  clearStoredTokens();
  window.localStorage.removeItem(USER_KEY);
  notify();
}

function toUser(account: LoginPayload["user"]): User {
  return {
    id: account.id,
    email: account.email,
    displayName: account.display_name?.trim() || account.email.split("@")[0],
    role: account.role,
  };
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const user = useSyncExternalStore(subscribe, getSnapshot, getServerSnapshot);
  const isReady = useSyncExternalStore(
    subscribeToHydration,
    getClientReadySnapshot,
    getServerReadySnapshot,
  );
  const [isPending, setIsPending] = useState(false);

  const login = useCallback(async (email: string, password: string): Promise<User> => {
    setIsPending(true);
    try {
      const res = await fetchAPI<ApiEnvelope<LoginPayload>>("/auth/login", {
        method: "POST",
        body: JSON.stringify({ email, password }),
      });

      const payload = res.data;
      if (!payload?.access_token || !payload?.refresh_token || !payload?.user?.id) {
        throw new APIError(500, "Malformed login response from server");
      }

      const nextUser = toUser(payload.user);
      persistSession({ access: payload.access_token, refresh: payload.refresh_token }, nextUser);
      return nextUser;
    } finally {
      setIsPending(false);
    }
  }, []);

  const logout = useCallback(() => {
    clearSession();
  }, []);

  const refreshProfile = useCallback(async (): Promise<User | null> => {
    try {
      const res = await fetchAPI<ApiEnvelope<{
        id: string;
        email: string;
        display_name?: string;
        role: string;
      }>>("/auth/me");

      const nextUser = toUser({
        id: res.data.id,
        email: res.data.email,
        display_name: res.data.display_name,
        role: res.data.role,
      });
      window.localStorage.setItem(USER_KEY, JSON.stringify(nextUser));
      notify();
      return nextUser;
    } catch {
      return null;
    }
  }, []);

  // A failed token refresh in api.ts means the session is unrecoverable, so the
  // cached profile is dropped here rather than left pointing at a dead token.
  useEffect(() => {
    setSessionExpiredHandler(() => {
      window.localStorage.removeItem(USER_KEY);
      notify();
    });
    return () => setSessionExpiredHandler(null);
  }, []);

  const value = useMemo<AuthContextType>(
    () => ({ user, login, logout, refreshProfile, isLoading: isPending, isReady }),
    [user, login, logout, refreshProfile, isPending, isReady],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthContextType {
  const context = useContext(AuthContext);
  if (context === undefined) throw new Error("useAuth must be used within an AuthProvider");
  return context;
}