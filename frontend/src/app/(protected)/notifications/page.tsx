"use client";
import { useState } from "react";
import Link from "next/link";
import { useAsyncData } from "@/hooks/useAsyncData";
import { Badge } from "@/components/ui/Badge";
import { api, unwrap } from "@/lib/api";
import { formatRelativeDate } from "@/lib/utils";

const ICONS: Record<string, string> = {
  discussion: "💬",
  reply: "↩️",
  event: "🎉",
  club: "👥",
  opportunity: "💼",
  application: "📄",
  resource: "🏛️",
  booking: "📅",
  mention: "🔔",
  system: "⚙️",
};

const ACCENT: Record<string, "success" | "warning" | "lpu" | "default"> = {
  event: "lpu",
  club: "success",
  opportunity: "success",
  application: "warning",
};

export default function NotificationsPage() {
  const [filter, setFilter] = useState<"all" | "unread">("all");

  const { data, error, isLoading, refresh, setData } = useAsyncData(async () => {
    const res = await api.notifications();
    return unwrap(res);
  }, []);

  const notifications = data?.notifications ?? [];
  const unread = data?.unread ?? 0;

  const markRead = async (id: string) => {
    // Optimistic removal from the unread set keeps the click feeling instant;
    // a failure drops the local change and refetches the authoritative list.
    if (!data) return;
    setData({
      ...data,
      notifications: data.notifications.map((n) =>
        n.id === id ? { ...n, read_at: new Date().toISOString() } : n,
      ),
      unread: Math.max(0, data.unread - 1),
    });
    try {
      await api.markNotificationRead(id);
    } catch {
      refresh();
    }
  };

  const markAllRead = async () => {
    if (!data) return;
    const now = new Date().toISOString();
    setData({
      ...data,
      notifications: data.notifications.map((n) => ({ ...n, read_at: n.read_at ?? now })),
      unread: 0,
    });
    try {
      await api.markAllNotificationsRead();
    } catch {
      refresh();
    }
  };

  const visible = filter === "unread" ? notifications.filter((n) => !n.read_at) : notifications;

  return (
    <div className="max-w-3xl mx-auto px-4 py-6 space-y-6 animate-fade-in">
      <div className="flex items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Notifications</h1>
          <p className="text-sm text-gray-500">
            {unread > 0 ? `${unread} unread` : "You are all caught up"}
          </p>
        </div>
        <button
          onClick={() => void markAllRead()}
          disabled={unread === 0}
          className="text-sm text-lpu-primary hover:underline disabled:opacity-40 disabled:no-underline"
        >
          Mark all as read
        </button>
      </div>

      <div className="flex gap-1 bg-surface-raised border border-surface-border rounded-lg p-1 w-fit">
        {(["all", "unread"] as const).map((id) => (
          <button
            key={id}
            onClick={() => setFilter(id)}
            className={
              filter === id
                ? "px-4 py-1.5 rounded-md text-sm font-medium bg-white text-gray-900 shadow-sm"
                : "px-4 py-1.5 rounded-md text-sm font-medium text-gray-500 hover:text-gray-700"
            }
          >
            {id === "all" ? "All" : `Unread${unread > 0 ? ` (${unread})` : ""}`}
          </button>
        ))}
      </div>

      {error && (
        <div className="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
          {error}
          <button onClick={refresh} className="ml-2 underline font-medium">
            Retry
          </button>
        </div>
      )}

      {isLoading ? (
        <p className="text-sm text-gray-500">Loading…</p>
      ) : (
        <div className="bg-white rounded-xl border border-surface-border overflow-hidden">
          {visible.length === 0 ? (
            <p className="text-sm text-gray-500 p-6 text-center">
              {filter === "unread" ? "No unread notifications." : "No notifications yet."}
            </p>
          ) : (
            visible.map((n) => {
              const isUnread = !n.read_at;
              const body = (
                <>
                  <div className="w-10 h-10 rounded-full flex items-center justify-center flex-shrink-0 text-xl bg-surface-raised">
                    {ICONS[n.type] ?? "🔔"}
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex justify-between items-start mb-1 gap-2">
                      <h3
                        className={
                          isUnread ? "text-sm font-semibold text-gray-900" : "text-sm font-medium text-gray-700"
                        }
                      >
                        {n.title}
                      </h3>
                      <span className="text-xs text-gray-400 whitespace-nowrap">
                        {formatRelativeDate(new Date(n.created_at))}
                      </span>
                    </div>
                    {n.body && <p className="text-sm text-gray-600 mb-2">{n.body}</p>}
                    <Badge variant={ACCENT[n.type] ?? "default"}>{n.type.toUpperCase()}</Badge>
                  </div>
                </>
              );

              if (n.link) {
                return (
                  <Link
                    key={n.id}
                    href={n.link}
                    onClick={() => {
                      if (isUnread) void markRead(n.id);
                    }}
                    className={`p-4 flex gap-4 border-b border-surface-border last:border-0 transition-colors ${
                      isUnread ? "bg-lpu-light/30 hover:bg-lpu-light/50" : "hover:bg-surface-raised"
                    }`}
                  >
                    {body}
                  </Link>
                );
              }

              return isUnread ? (
                <button
                  key={n.id}
                  onClick={() => void markRead(n.id)}
                  className="w-full text-left p-4 flex gap-4 border-b border-surface-border last:border-0 bg-lpu-light/30 hover:bg-lpu-light/50 transition-colors"
                >
                  {body}
                </button>
              ) : (
                <div
                  key={n.id}
                  className="p-4 flex gap-4 border-b border-surface-border last:border-0"
                >
                  {body}
                </div>
              );
            })
          )}
        </div>
      )}
    </div>
  );
}