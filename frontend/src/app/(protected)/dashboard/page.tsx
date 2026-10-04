"use client";
import Link from "next/link";
import { useAuth } from "@/contexts/AuthContext";
import { useAsyncData } from "@/hooks/useAsyncData";
import { AIInsightCard } from "@/components/ai/AIComponents";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { api, fetchAPI, unwrap } from "@/lib/api";
import { cn, formatDateTime, formatRelativeDate } from "@/lib/utils";

export default function DashboardPage() {
  const { user } = useAuth();
  if (!user) return null;
  if (user.role === "ADMIN" || user.role === "SUPER_ADMIN") return <AdminHome />;
  return <MemberHome />;
}

function greeting(): string {
  const hour = new Date().getHours();
  if (hour < 12) return "Good morning";
  if (hour < 17) return "Good afternoon";
  return "Good evening";
}

// ─── Member home ───────────────────────────────────────────────────────────────

function MemberHome() {
  const { user } = useAuth();

  const { data, error, isLoading, refresh } = useAsyncData(async () => {
    const [eventRes, discussionRes, clubRes, opportunityRes, notificationRes] = await Promise.all([
      api.events(),
      api.discussions({ limit: 5 }),
      api.myClubs(),
      api.opportunities(),
      api.notifications(),
    ]);

    return {
      events: unwrap(eventRes).events,
      discussions: unwrap(discussionRes).discussions,
      groups: unwrap(clubRes).clubs,
      opportunities: unwrap(opportunityRes).opportunities,
      notifications: unwrap(notificationRes),
    };
  }, []);

  const events = data?.events ?? [];
  const discussions = data?.discussions ?? [];
  const groups = data?.groups ?? [];
  const opportunities = data?.opportunities ?? [];
  const notifications = data?.notifications;

  const upcoming = events
    .filter((e) => e.status === "PUBLISHED" && new Date(e.start_time) >= new Date())
    .sort((a, b) => +new Date(a.start_time) - +new Date(b.start_time));

  const next = upcoming[0];
  const openOpportunities = opportunities.filter((o) => o.status === "OPEN");

  const insight = isLoading
    ? "Loading your campus context…"
    : next
      ? `Your next event is "${next.title}" on ${formatDateTime(new Date(next.start_time))}${next.venue ? ` at ${next.venue}` : ""}. You are in ${groups.length} group${groups.length === 1 ? "" : "s"} and have ${openOpportunities.length} open opportunit${openOpportunities.length === 1 ? "y" : "ies"}.`
      : `You are in ${groups.length} group${groups.length === 1 ? "" : "s"} with ${openOpportunities.length} open opportunit${openOpportunities.length === 1 ? "y" : "ies"}. No upcoming events right now.`;

  return (
    <div className="max-w-6xl mx-auto px-4 py-6 space-y-6">
      <div className="flex items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">
            {greeting()}, {user?.displayName?.split(" ")[0]}.
          </h1>
          <p className="text-sm text-gray-500 mt-0.5">{user?.email}</p>
        </div>
        <Link href="/ai" className="flex-shrink-0">
          <Button variant="ai" size="sm">
            ✨ Ask CampusCare
          </Button>
        </Link>
      </div>

      {error && (
        <div className="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
          {error}
          <button onClick={refresh} className="ml-2 underline font-medium">
            Retry
          </button>
        </div>
      )}

      <AIInsightCard
        title="What is happening on campus"
        content={insight}
        sources={["Events", "Groups", "Opportunities", "Notifications"]}
        updatedAt={isLoading ? "loading…" : "just now"}
      />

      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <StatTile
          label="Upcoming Events"
          value={upcoming.length}
          sub={next ? `Next: ${formatRelativeDate(new Date(next.start_time))}` : "Nothing scheduled"}
          icon="🎉"
          href="/discover"
        />
        <StatTile
          label="My Groups"
          value={groups.length}
          sub={groups.length ? groups.map((g) => g.name).slice(0, 2).join(", ") : "Join a group to start collaborating"}
          icon="👥"
          href="/community"
        />
        <StatTile
          label="Unread Notifications"
          value={notifications?.unread ?? 0}
          sub={notifications?.notifications?.[0]?.title ?? "You are all caught up"}
          icon="🔔"
          href="/notifications"
          highlight={(notifications?.unread ?? 0) > 0}
        />
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-5 gap-4">
        <div className="lg:col-span-3 bg-white rounded-xl border border-surface-border p-4">
          <div className="flex items-center justify-between mb-3">
            <p className="text-sm font-semibold text-gray-900">Latest discussions</p>
            <Link href="/community" className="text-xs text-lpu-primary hover:underline">
              Open community
            </Link>
          </div>

          {discussions.length === 0 ? (
            <EmptyHint text="No discussions yet. Start one from the Community page." />
          ) : (
            <div className="space-y-2">
              {discussions.map((thread) => (
                <Link
                  key={thread.id}
                  href={`/community?discussion=${thread.id}`}
                  className="block rounded-lg border border-surface-border p-3 hover:bg-surface-raised transition-colors"
                >
                  <div className="flex items-start justify-between gap-2">
                    <p className="text-sm font-medium text-gray-900">{thread.title}</p>
                    <Badge variant="default">{thread.category}</Badge>
                  </div>
                  <p className="mt-1 text-xs text-gray-500">
                    {thread.author_name || "Member"} · {thread.reply_count} repl
                    {thread.reply_count === 1 ? "y" : "ies"} ·{" "}
                    {formatRelativeDate(new Date(thread.last_activity_at))}
                  </p>
                </Link>
              ))}
            </div>
          )}
        </div>

        <div className="lg:col-span-2 bg-white rounded-xl border border-surface-border p-4 space-y-4">
          <div>
            <div className="flex items-center justify-between mb-3">
              <p className="text-sm font-semibold text-gray-900">Next up</p>
              <Link href="/discover" className="text-xs text-lpu-primary hover:underline">
                All events
              </Link>
            </div>

            {next ? (
              <div className="rounded-lg bg-lpu-light p-4">
                <p className="font-semibold text-gray-900 text-sm">{next.title}</p>
                <p className="text-xs text-gray-600 mt-1">{formatDateTime(new Date(next.start_time))}</p>
                <p className="text-xs text-gray-500">
                  {next.is_online ? "Online" : next.venue} · {next.capacity} spots
                </p>
                <Link href="/discover" className="mt-3 inline-block">
                  <Button variant="primary" size="sm">
                    View event
                  </Button>
                </Link>
              </div>
            ) : (
              <EmptyHint text="No upcoming events right now." />
            )}
          </div>

          <div>
            <div className="flex items-center justify-between mb-3">
              <p className="text-sm font-semibold text-gray-900">Open opportunities</p>
              <Link href="/career" className="text-xs text-lpu-primary hover:underline">
                See all
              </Link>
            </div>

            {openOpportunities.length === 0 ? (
              <EmptyHint text="Nothing open at the moment." />
            ) : (
              <div className="space-y-2">
                {openOpportunities.slice(0, 3).map((opp) => (
                  <div key={opp.id} className="rounded-lg border border-surface-border p-3">
                    <p className="text-sm font-medium text-gray-900">{opp.title}</p>
                    <p className="text-xs text-gray-500">
                      {opp.type} · closes {formatRelativeDate(new Date(opp.registration_deadline))}
                    </p>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}

// ─── Admin home ────────────────────────────────────────────────────────────────

type Analytics = {
  window_days: number;
  total_events?: number;
  previous_events?: number;
  trend?: string;
  by_category?: { category: string; count: number }[];
};

function AdminHome() {
  const { user } = useAuth();

  const { data, error } = useAsyncData(async () => {
    const [analyticsRes, notificationRes, clubRes] = await Promise.all([
      fetchAPI<{ data: Analytics }>("/institutional/analytics?days=30"),
      api.notifications(),
      api.clubs(),
    ]);
    return {
      analytics: analyticsRes.data,
      notifications: unwrap(notificationRes).notifications,
      clubs: unwrap(clubRes).clubs,
    };
  }, []);

  const analytics = data?.analytics ?? null;
  const notifications = data?.notifications ?? [];
  const clubs = data?.clubs ?? [];

  const total = analytics?.total_events ?? 0;
  const previous = analytics?.previous_events ?? 0;
  const delta = total - previous;

  return (
    <div className="max-w-6xl mx-auto px-4 py-6 space-y-6">
      <div className="flex items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Campus Overview</h1>
          <p className="text-sm text-gray-500 mt-0.5">
            Operational view for {user?.email}
          </p>
        </div>
        <Link href="/ai" className="flex-shrink-0">
          <Button variant="ai" size="sm">
            ✨ AI Copilot
          </Button>
        </Link>
      </div>

      {error && (
        <div className="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">{error}</div>
      )}

      <AIInsightCard
        title="Engagement summary"
        content={
          analytics
            ? `${total} event${total === 1 ? "" : "s"} ran in the last ${analytics.window_days} days (${delta >= 0 ? "+" : ""}${delta} vs the previous period), across ${analytics.by_category?.length ?? 0} categor${analytics.by_category?.length === 1 ? "y" : "ies"}. ${clubs.length} group${clubs.length === 1 ? "" : "s"} are active on this campus.`
            : "Loading engagement metrics…"
        }
        sources={["Event Analytics", "Groups"]}
        updatedAt="just now"
      />

      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        {[
          { label: `Events (${analytics?.window_days ?? 30}d)`, value: String(total), delta },
          { label: "Previous Period", value: String(previous), delta: 0 },
          { label: "Active Groups", value: String(clubs.length), delta: 0 },
          { label: "Notifications", value: String(notifications.length), delta: 0 },
        ].map((s) => (
          <div key={s.label} className="bg-white rounded-xl border border-surface-border p-4">
            <p className="text-xs text-gray-500 mb-1">{s.label}</p>
            <p className="text-2xl font-bold text-gray-900">{s.value}</p>
            {s.delta !== 0 && (
              <p className={cn("text-xs mt-1 font-medium", s.delta > 0 ? "text-green-600" : "text-red-500")}>
                {s.delta > 0 ? "+" : ""}
                {s.delta} vs previous
              </p>
            )}
          </div>
        ))}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        <div className="bg-white rounded-xl border border-surface-border p-4">
          <p className="text-sm font-semibold text-gray-900 mb-4">Groups by category</p>
          {clubs.length === 0 ? (
            <EmptyHint text="No groups registered yet." />
          ) : (
            <div className="space-y-2">
              {clubs.slice(0, 8).map((club) => (
                <div
                  key={club.id}
                  className="flex items-center justify-between py-2.5 border-b border-surface-border last:border-0"
                >
                  <div className="min-w-0">
                    <span className="text-sm text-gray-700 truncate">{club.name}</span>
                    <p className="text-xs text-gray-400">
                      {club.category || "Uncategorised"} · {club.status}
                    </p>
                  </div>
                  <Badge variant={club.verification_status === "APPROVED" ? "success" : "warning"}>
                    {club.member_count ?? 0} members
                  </Badge>
                </div>
              ))}
            </div>
          )}
        </div>

        <div className="bg-white rounded-xl border border-surface-border p-4">
          <p className="text-sm font-semibold text-gray-900 mb-4">Recent notifications</p>
          {notifications.length === 0 ? (
            <EmptyHint text="No notifications recorded." />
          ) : (
            <div className="space-y-2">
              {notifications.slice(0, 8).map((n) => (
                <div key={n.id} className="border-b border-surface-border last:border-0 py-2.5">
                  <p className="text-sm text-gray-800">{n.title}</p>
                  <p className="text-xs text-gray-500">{formatRelativeDate(new Date(n.created_at))}</p>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

// ─── Shared bits ───────────────────────────────────────────────────────────────

function StatTile({
  label,
  value,
  sub,
  icon,
  href,
  highlight,
}: {
  label: string;
  value: number;
  sub: string;
  icon: string;
  href: string;
  highlight?: boolean;
}) {
  return (
    <Link
      href={href}
      className={cn(
        "rounded-xl border p-4 transition-colors",
        highlight ? "border-amber-200 bg-amber-50" : "border-surface-border bg-white hover:bg-surface-raised",
      )}
    >
      <div className="flex items-center gap-2 mb-2">
        <span className="text-xl">{icon}</span>
        <p className="text-xs text-gray-500">{label}</p>
      </div>
      <p className="text-3xl font-bold text-gray-900">{value}</p>
      <p className="text-xs text-gray-400 mt-0.5 truncate">{sub}</p>
    </Link>
  );
}

function EmptyHint({ text }: { text: string }) {
  return <p className="text-sm text-gray-500 py-4 text-center">{text}</p>;
}