"use client";
import { useMemo, useState } from "react";
import Link from "next/link";
import { useAsyncData } from "@/hooks/useAsyncData";
import { Badge, type BadgeVariant } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { AIInsightCard } from "@/components/ai/AIComponents";
import { APIError, api, unwrap, type CampusEvent } from "@/lib/api";
import { cn, formatDateTime, formatRelativeDate, isPast } from "@/lib/utils";

type Section = "events" | "clubs";

export default function DiscoverPage() {
  const [section, setSection] = useState<Section>("events");
  const [category, setCategory] = useState("all");
  const [registering, setRegistering] = useState<string | null>(null);
  const [actionError, setActionError] = useState<string | null>(null);

  const { data, error: loadError, isLoading, refresh, setData } = useAsyncData(async () => {
    const [eventRes, clubRes, myClubRes] = await Promise.all([api.events(), api.clubs(), api.myClubs()]);

    // The "closing soon" window is computed here, inside the fetch, so the
    // render pass stays pure and the deadline maths uses one consistent instant.
    const now = Date.now();
    const published = unwrap(eventRes)
      .events.filter((e) => e.status === "PUBLISHED")
      .sort((a, b) => +new Date(a.start_time) - +new Date(b.start_time));

    return {
      events: published,
      clubs: unwrap(clubRes).clubs,
      joined: new Set(unwrap(myClubRes).clubs.map((c) => c.id)),
      closingSoon: published.filter((e) => {
        const deadline = new Date(e.registration_deadline);
        return !isPast(deadline) && +deadline - now < 7 * 86400000;
      }),
    };
  }, []);

  // Derived slices are memoised so the category list and the deadline maths use
  // stable references instead of rebuilding arrays on every render.
  const events = useMemo(() => data?.events ?? [], [data]);
  const clubs = useMemo(() => data?.clubs ?? [], [data]);
  const closingSoon = useMemo(() => data?.closingSoon ?? [], [data]);

  const categories = useMemo(
    () => Array.from(new Set(events.map((e) => e.category).filter(Boolean))).sort(),
    [events],
  );

  const filtered = category === "all" ? events : events.filter((e) => e.category === category);

  // Membership changes are applied locally as well as on the server so the button
  // reflects the action immediately without a full refetch.
  const joined = useMemo(() => data?.joined ?? new Set<string>(), [data]);
  const setJoined = (next: Set<string>) => {
    if (data) setData({ ...data, joined: next });
  };

  const register = async (event: CampusEvent) => {
    setRegistering(event.id);
    setActionError(null);
    try {
      await api.registerForEvent(event.id);
      refresh();
    } catch (err) {
      setActionError(err instanceof APIError ? err.message : "Could not register for that event.");
    } finally {
      setRegistering(null);
    }
  };

  const toggleJoin = async (clubId: string, currentlyJoined: boolean) => {
    setActionError(null);
    try {
      if (currentlyJoined) {
        await api.leaveClub(clubId);
        setJoined(new Set([...joined].filter((id) => id !== clubId)));
      } else {
        await api.joinClub(clubId);
        setJoined(new Set([...joined, clubId]));
      }
    } catch (err) {
      setActionError(err instanceof APIError ? err.message : "Could not update that membership.");
    }
  };

  const error = actionError ?? loadError;

  return (
    <div className="max-w-5xl mx-auto px-4 py-6 space-y-6 animate-fade-in">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Discover</h1>
        <p className="text-sm text-gray-500">Events and groups happening on this campus</p>
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
        title="What is coming up"
        content={
isLoading
        ? "Loading the campus calendar…"
            : closingSoon.length > 0
              ? `${closingSoon.length} event${closingSoon.length === 1 ? "" : "s"} close registration within a week: ${closingSoon.map((e) => e.title).join(", ")}.`
              : `${events.length} event${events.length === 1 ? "" : "s"} published. Nothing closes within the next week, so you have time.`
        }
        sources={["Event Calendar"]}
        updatedAt="just now"
      />

      <div className="flex gap-1 bg-surface-raised border border-surface-border rounded-lg p-1 w-fit">
        {(["events", "clubs"] as const).map((s) => (
          <button
            key={s}
            onClick={() => setSection(s)}
            className={cn(
              "px-4 py-1.5 rounded-md text-sm font-medium transition-all",
              section === s ? "bg-white text-gray-900 shadow-sm" : "text-gray-500 hover:text-gray-700",
            )}
          >
            {s === "clubs" ? "Groups & Clubs" : "Events"}
          </button>
        ))}
      </div>

      {isLoading ? (
        <p className="text-sm text-gray-500">Loading…</p>
      ) : section === "events" ? (
        <div className="space-y-4">
          {categories.length > 0 && (
            <div className="flex items-center gap-2 flex-wrap">
              {["all", ...categories].map((c) => (
                <button
                  key={c}
                  onClick={() => setCategory(c)}
                  className={cn(
                    "px-3 py-1.5 rounded-full text-xs font-medium border transition-all",
                    category === c
                      ? "bg-lpu-primary text-white border-lpu-primary"
                      : "bg-white text-gray-600 border-surface-border hover:border-gray-300",
                  )}
                >
                  {c === "all" ? "All Types" : c}
                </button>
              ))}
            </div>
          )}

          {filtered.length === 0 ? (
            <p className="text-sm text-gray-500 py-6 text-center">No published events right now.</p>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {filtered.map((event) => {
                const registered = event.registration_status === "REGISTERED";
                const past = isPast(new Date(event.end_time));
                const deadlinePassed = isPast(new Date(event.registration_deadline));
                return (
                  <div
                    key={event.id}
                    className="bg-white rounded-xl border border-surface-border p-5 hover:shadow-card transition-shadow"
                  >
                    <div className="flex items-start justify-between gap-2 mb-3">
                      {event.category ? (
                        <Badge variant={categoryVariant(event.category)}>{event.category}</Badge>
                      ) : (
                        <Badge variant="default">EVENT</Badge>
                      )}
                      <div className="flex items-center gap-1.5 flex-shrink-0">
                        {event.is_online && <Badge variant="info">Online</Badge>}
                        {registered && <Badge variant="success">Registered</Badge>}
                        {past && <Badge variant="default">Ended</Badge>}
                      </div>
                    </div>

                    <h3 className="font-semibold text-gray-900 leading-tight">{event.title}</h3>
                    {event.description && (
                      <p className="text-sm text-gray-600 mt-2 leading-relaxed line-clamp-3">
                        {event.description}
                      </p>
                    )}

                    <div className="mt-3 flex flex-wrap items-center gap-3 text-xs text-gray-400">
                      <span>{formatDateTime(new Date(event.start_time))}</span>
                      {event.venue && <span>{event.venue}</span>}
                      <span>{event.capacity} spots</span>
                    </div>

                    {!isPast(new Date(event.registration_deadline)) && (
                      <p className="mt-2 text-xs text-amber-600">
                        Registration closes {formatRelativeDate(new Date(event.registration_deadline))}
                      </p>
                    )}

                    <div className="flex gap-2 mt-4">
                      <Button
                        variant={registered ? "secondary" : "primary"}
                        size="sm"
                        className="flex-1"
                        disabled={registered || past || deadlinePassed || registering === event.id}
                        onClick={() => void register(event)}
                      >
                        {registering === event.id
                          ? "Registering…"
                          : registered
                            ? "Registered"
                            : past || deadlinePassed
                              ? "Registration closed"
                              : "Register now"}
                      </Button>
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {clubs.length === 0 ? (
            <p className="text-sm text-gray-500 py-6 text-center">No groups have been set up yet.</p>
          ) : (
            clubs.map((club) => {
              const isJoined = joined.has(club.id);
              return (
                <div
                  key={club.id}
                  className="bg-white rounded-xl border border-surface-border p-5 hover:shadow-card transition-shadow"
                >
                  <div className="flex items-start gap-3 mb-3">
                    <div className="w-12 h-12 bg-lpu-light rounded-xl flex items-center justify-center flex-shrink-0">
                      <span className="text-xl">🏆</span>
                    </div>
                    <div className="flex-1 min-w-0">
                      <h3 className="font-semibold text-gray-900 leading-tight">{club.name}</h3>
                      <p className="text-xs text-gray-500">
                        {club.member_count ?? 0} member{club.member_count === 1 ? "" : "s"}
                        {club.category ? ` · ${club.category}` : ""}
                      </p>
                    </div>
                  </div>

                  {club.description && (
                    <p className="text-sm text-gray-600 mb-3 line-clamp-2">{club.description}</p>
                  )}

                  <div className="flex gap-2">
                    <Button
                      variant={isJoined ? "secondary" : "primary"}
                      size="sm"
                      className="flex-1"
                      onClick={() => void toggleJoin(club.id, isJoined)}
                    >
                      {isJoined ? "Leave group" : "Join group"}
                    </Button>
                    <Link href="/community">
                      <Button variant="ghost" size="sm">
                        Discussions
                      </Button>
                    </Link>
                  </div>
                </div>
              );
            })
          )}
        </div>
      )}
    </div>
  );
}

function categoryVariant(category: string): BadgeVariant {
  switch (category.toUpperCase()) {
    case "HACKATHON":
    case "COMPETITION":
      return "success";
    case "WORKSHOP":
      return "info";
    case "SEMINAR":
    case "TALKS":
      return "default";
    default:
      return "lpu";
  }
}