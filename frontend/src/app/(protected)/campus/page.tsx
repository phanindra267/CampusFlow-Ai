"use client";
import { useCallback, useState } from "react";
import { useAsyncData } from "@/hooks/useAsyncData";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { APIError, api, unwrap, type CampusResource, type ResourceBooking } from "@/lib/api";
import { cn, formatDateTime, formatRelativeDate } from "@/lib/utils";

const ICONS: Record<string, string> = {
  ROOM: "🏠",
  LAB: "🔬",
  STUDIO: "🎧",
  SPORTS: "🏀",
  TRANSPORT: "🚌",
  IT: "💻",
  HEALTH: "🏥",
  LIBRARY: "📚",
  OTHER: "🏛️",
};

export default function CampusPage() {
  const [view, setView] = useState<"services" | "bookings">("services");
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  const { data, error, isLoading, refresh, setData } = useAsyncData(async () => {
    const [serviceRes, bookingRes] = await Promise.all([api.services(), api.myBookings()]);
    return {
      resources: unwrap(serviceRes).resources,
      bookings: unwrap(bookingRes).bookings,
    };
  }, []);

  const resources = data?.resources ?? [];
  const bookings = data?.bookings ?? [];
  const selected = resources.find((r) => r.id === selectedId) ?? null;

  const reload = useCallback(async () => {
    setBusy(true);
    try {
      refresh();
    } finally {
      setBusy(false);
    }
  }, [refresh]);

  const cancelBooking = async (id: string) => {
    try {
      await api.cancelBooking(id);
      // Cancelling locally first keeps the list responsive; the server response
      // is authoritative and a refetch happens on the next load anyway.
      if (data) {
        setData({
          ...data,
          bookings: data.bookings.map((b) => (b.id === id ? { ...b, status: "CANCELLED" } : b)),
        });
      }
    } catch (err) {
      throw err instanceof APIError ? err : new Error("Could not cancel that booking.");
    }
  };

  const grouped = resources.reduce<Record<string, typeof resources>>((acc, resource) => {
    const key = resource.resource_type || "OTHER";
    (acc[key] ??= []).push(resource);
    return acc;
  }, {});

  return (
    <div className="max-w-4xl mx-auto px-4 py-6 space-y-6 animate-fade-in">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Campus Services</h1>
        <p className="text-sm text-gray-500">Browse shared resources and book the ones you need</p>
      </div>

      {error && (
        <div className="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
          {error}
          <button onClick={refresh} className="ml-2 underline font-medium">
            Retry
          </button>
        </div>
      )}

      <div className="flex gap-1 bg-surface-raised border border-surface-border rounded-lg p-1 w-fit">
        {(["services", "bookings"] as const).map((t) => (
          <button
            key={t}
            onClick={() => {
              setView(t);
              setSelectedId(null);
            }}
            className={cn(
              "px-4 py-1.5 rounded-md text-sm font-medium transition-all",
              view === t ? "bg-white text-gray-900 shadow-sm" : "text-gray-500 hover:text-gray-700",
            )}
          >
            {t === "services" ? "Services" : "My Bookings"}
          </button>
        ))}
      </div>

      {isLoading ? (
        <p className="text-sm text-gray-500">Loading…</p>
      ) : view === "services" ? (
        <div className="space-y-6">
          {resources.length === 0 ? (
            <p className="text-sm text-gray-500 py-6 text-center">No services have been published yet.</p>
          ) : (
            Object.entries(grouped).map(([type, items]) => (
              <section key={type}>
                <h2 className="text-sm font-semibold text-gray-900 mb-3">{titleCase(type)}</h2>
                <div className="grid grid-cols-2 md:grid-cols-3 gap-3">
                  {items.map((resource) => (
                    <button
                      key={resource.id}
                      onClick={() => resource.is_bookable && setSelectedId(resource.id)}
                      disabled={!resource.is_bookable}
                      className="bg-white rounded-xl border border-surface-border p-4 text-left hover:shadow-card hover:border-lpu-primary/30 transition-all group disabled:cursor-default disabled:opacity-70"
                    >
                      <span className="text-3xl block mb-2">{ICONS[type.toUpperCase()] ?? ICONS.OTHER}</span>
                      <p className="font-semibold text-gray-900 text-sm group-hover:text-lpu-primary transition-colors">
                        {resource.name}
                      </p>
                      {resource.description && (
                        <p className="text-xs text-gray-500 mt-0.5 line-clamp-2">{resource.description}</p>
                      )}
                      <div className="mt-2">
                        {resource.is_bookable ? (
                          <Badge variant="success">Bookable</Badge>
                        ) : (
                          <Badge variant="default">Browse only</Badge>
                        )}
                      </div>
                    </button>
                  ))}
                </div>
              </section>
            ))
          )}
        </div>
      ) : (
        <BookingsView bookings={bookings} busy={busy} onCancel={cancelBooking} />
      )}

      {selected && (
        <BookingDialog
          resource={selected}
          onClose={() => setSelectedId(null)}
          onBooked={() => {
            setSelectedId(null);
            void reload();
          }}
        />
      )}
    </div>
  );
}

function BookingsView({
  bookings,
  busy,
  onCancel,
}: {
  bookings: ResourceBooking[];
  busy: boolean;
  onCancel: (id: string) => Promise<void>;
}) {
  const [error, setError] = useState<string | null>(null);

  if (bookings.length === 0) {
    return <p className="text-sm text-gray-500 py-6 text-center">You have no bookings yet.</p>;
  }

  return (
    <div className="space-y-3">
      {error && <p className="text-sm text-red-600">{error}</p>}
      {bookings.map((booking) => (
        <div key={booking.id} className="bg-white rounded-xl border border-surface-border p-4">
          <div className="flex items-start justify-between gap-3">
            <div className="min-w-0">
              <p className="font-medium text-gray-900">{booking.resource_nm ?? "Booking"}</p>
              <p className="text-xs text-gray-500 mt-0.5">
                {formatDateTime(new Date(booking.start_time))} → {formatDateTime(new Date(booking.end_time))}
              </p>
              {booking.note && <p className="text-xs text-gray-500 mt-1">{booking.note}</p>}
            </div>
            <Badge variant={booking.status === "CANCELLED" ? "error" : "success"}>{booking.status}</Badge>
          </div>

          {booking.status !== "CANCELLED" && (
            <div className="mt-3 flex items-center justify-between">
              <span className="text-xs text-gray-400">
                Starts {formatRelativeDate(new Date(booking.start_time))}
              </span>
              <Button
                variant="secondary"
                size="sm"
                disabled={busy}
                onClick={async () => {
                  setError(null);
                  try {
                    await onCancel(booking.id);
                  } catch (err) {
                    setError(err instanceof Error ? err.message : "Could not cancel that booking.");
                  }
                }}
              >
                Cancel
              </Button>
            </div>
          )}
        </div>
      ))}
    </div>
  );
}

function BookingDialog({
  resource,
  onClose,
  onBooked,
}: {
  resource: CampusResource;
  onClose: () => void;
  onBooked: () => void;
}) {
  const [start, setStart] = useState(() => toLocalInput(nextHour()));
  const [end, setEnd] = useState(() => toLocalInput(new Date(nextHour().getTime() + 3600000)));
  const [note, setNote] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const submit = async () => {
    setSubmitting(true);
    setError(null);
    try {
      await api.bookService(resource.id, new Date(start).toISOString(), new Date(end).toISOString(), note);
      onBooked();
    } catch (err) {
      setError(
        err instanceof APIError
          ? `${err.message}${err.details ? ` — ${err.details}` : ""}`
          : "Could not create that booking.",
      );
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 bg-black/40 backdrop-blur-sm flex items-center justify-center px-4" onClick={onClose}>
      <div
        className="w-full max-w-md bg-white rounded-xl shadow-modal border border-surface-border p-5 animate-slide-up"
        onClick={(e) => e.stopPropagation()}
      >
        <h2 className="text-lg font-semibold text-gray-900">Book {resource.name}</h2>
        {resource.description && <p className="text-sm text-gray-500 mt-1">{resource.description}</p>}

        <div className="space-y-3 mt-4">
          <div>
            <label htmlFor="start" className="block text-xs text-gray-500 mb-1">
              From
            </label>
            <input
              id="start"
              type="datetime-local"
              value={start}
              onChange={(e) => setStart(e.target.value)}
              className="w-full rounded-lg border border-surface-border px-3 py-2 text-sm"
            />
          </div>
          <div>
            <label htmlFor="end" className="block text-xs text-gray-500 mb-1">
              To
            </label>
            <input
              id="end"
              type="datetime-local"
              value={end}
              onChange={(e) => setEnd(e.target.value)}
              className="w-full rounded-lg border border-surface-border px-3 py-2 text-sm"
            />
          </div>
          <div>
            <label htmlFor="note" className="block text-xs text-gray-500 mb-1">
              Purpose (optional)
            </label>
            <input
              id="note"
              value={note}
              onChange={(e) => setNote(e.target.value)}
              placeholder="What will you use it for?"
              className="w-full rounded-lg border border-surface-border px-3 py-2 text-sm"
            />
          </div>
        </div>

        {error && <p className="mt-3 text-sm text-red-600">{error}</p>}

        <div className="mt-5 flex gap-2">
          <Button variant="secondary" className="flex-1" onClick={onClose}>
            Cancel
          </Button>
          <Button variant="primary" className="flex-1" isLoading={submitting} onClick={() => void submit()}>
            Confirm booking
          </Button>
        </div>
      </div>
    </div>
  );
}

/** The next whole hour, used as the default booking slot. */
function nextHour(): Date {
  const d = new Date(Date.now() + 3600000);
  d.setMinutes(0, 0, 0);
  return d;
}

function toLocalInput(date: Date): string {
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`;
}

function titleCase(value: string): string {
  return value.charAt(0).toUpperCase() + value.slice(1).toLowerCase().replace(/_/g, " ");
}