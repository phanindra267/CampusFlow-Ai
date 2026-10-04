"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { APIError } from "@/lib/api";

type AsyncDataState<T> = {
  data: T | null;
  error: string | null;
  isLoading: boolean;
  /** Re-runs the fetcher, clearing the error first. */
  refresh: () => void;
  /** Replaces the loaded data locally, for optimistic updates. */
  setData: React.Dispatch<React.SetStateAction<T | null>>;
};

function describe(err: unknown): string {
  if (err instanceof APIError) {
    return err.details ? `${err.message} — ${err.details}` : err.message;
  }
  if (err instanceof Error) return err.message;
  return "Something went wrong.";
}

/**
 * useAsyncData runs an async fetcher on mount and whenever the deps change.
 *
 * Every state update happens after an await, so the fetch never calls setState
 * synchronously during the effect body, and results that arrive after unmount
 * are discarded instead of warning.
 */
export function useAsyncData<T>(
  fetcher: () => Promise<T>,
  deps: React.DependencyList,
): AsyncDataState<T> {
  const [data, setData] = useState<T | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [nonce, setNonce] = useState(0);

  const active = useRef(true);

  useEffect(() => {
    active.current = true;
    return () => {
      active.current = false;
    };
  }, []);

  useEffect(() => {
    let cancelled = false;

    void (async () => {
      try {
        const result = await fetcher();
        if (cancelled || !active.current) return;
        setData(result);
        setError(null);
      } catch (err) {
        if (cancelled || !active.current) return;
        setError(describe(err));
      } finally {
        if (!cancelled && active.current) setIsLoading(false);
      }
    })();

    return () => {
      cancelled = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [...deps, nonce]);

  const refresh = useCallback(() => {
    setIsLoading(true);
    setNonce((n) => n + 1);
  }, []);

  return { data, error, isLoading, refresh, setData };
}