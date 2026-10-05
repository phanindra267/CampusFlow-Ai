"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { useAuth } from "@/contexts/AuthContext";
import { APIError, fetchAPI, type Club } from "@/lib/api";
import { formatRelativeDate } from "@/lib/utils";
import { useAsyncData } from "@/hooks/useAsyncData";

type AdminEnvelope<T> = { data: T };

function AdminOnly({ children }: { children: React.ReactNode }) {
  const { user, isReady } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (isReady && user && user.role !== "ADMIN" && user.role !== "SUPER_ADMIN") {
      router.replace("/dashboard");
    }
  }, [isReady, user, router]);

  if (!isReady || !user || (user.role !== "ADMIN" && user.role !== "SUPER_ADMIN")) {
    return null;
  }
  return <>{children}</>;
}

function AdminHeading({ title, description }: { title: string; description: string }) {
  return (
    <header>
      <h1 className="text-2xl font-bold text-gray-900">{title}</h1>
      <p className="mt-1 text-sm text-gray-500">{description}</p>
    </header>
  );
}

function AdminError({ error, retry }: { error: string | null; retry: () => void }) {
  if (!error) return null;
  return (
    <div role="alert" className="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
      {error}
      <button onClick={retry} className="ml-2 font-medium underline">Retry</button>
    </div>
  );
}

type ServiceRequest = {
  id: string;
  reference: string;
  requester_name?: string;
  service_name?: string;
  kind: string;
  subject: string;
  description: string;
  status: string;
  priority: string;
  created_at: string;
  resolution?: string;
};

export function AdminOperationsPage() {
  const { data, error, isLoading, refresh } = useAsyncData(async () => {
    const response = await fetchAPI<AdminEnvelope<{ requests: ServiceRequest[] }>>(
      "/admin/service-requests?limit=100",
    );
    return response.data.requests;
  }, []);
  const [busyId, setBusyId] = useState<string | null>(null);
  const [resolution, setResolution] = useState<Record<string, string>>({});
  const [actionError, setActionError] = useState<string | null>(null);

  async function updateRequest(request: ServiceRequest, status: "IN_PROGRESS" | "RESOLVED") {
    const note = status === "RESOLVED" ? (resolution[request.id] ?? "").trim() : "Work started.";
    if (status === "RESOLVED" && !note) {
      setActionError("Enter a resolution before resolving a request.");
      return;
    }
    setBusyId(request.id);
    setActionError(null);
    try {
      await fetchAPI(`/admin/service-requests/${request.id}`, {
        method: "PATCH",
        body: JSON.stringify({
          status,
          note,
          ...(status === "RESOLVED" ? { resolution: note } : {}),
        }),
      });
      refresh();
    } catch (err) {
      setActionError(err instanceof APIError ? err.message : "Could not update the service request.");
    } finally {
      setBusyId(null);
    }
  }

  return (
    <AdminOnly>
      <div className="mx-auto max-w-6xl space-y-6 px-4 py-6">
        <AdminHeading title="Operations" description="Review and update campus service requests." />
        <AdminError error={error} retry={refresh} />
        {actionError && <p role="alert" className="text-sm text-red-600">{actionError}</p>}
        {isLoading ? <p className="text-sm text-gray-500">Loading service queue…</p> : (
          <div className="space-y-3">
            {(data ?? []).length === 0 ? (
              <p className="rounded-xl border border-surface-border bg-white p-6 text-center text-sm text-gray-500">
                There are no service requests in the queue.
              </p>
            ) : (data ?? []).map((request) => (
              <article key={request.id} className="rounded-xl border border-surface-border bg-white p-4">
                <div className="flex flex-wrap items-start justify-between gap-3">
                  <div>
                    <p className="font-semibold text-gray-900">{request.subject}</p>
                    <p className="mt-1 text-xs text-gray-500">
                      {request.reference} · {request.requester_name || "Campus member"} · {request.service_name || request.kind}
                    </p>
                  </div>
                  <div className="flex gap-2">
                    <Badge variant={request.priority === "URGENT" || request.priority === "HIGH" ? "warning" : "default"}>
                      {request.priority}
                    </Badge>
                    <Badge variant={request.status === "RESOLVED" ? "success" : "info"}>{request.status}</Badge>
                  </div>
                </div>
                <p className="mt-3 whitespace-pre-wrap text-sm text-gray-600">{request.description}</p>
                <p className="mt-2 text-xs text-gray-400">{formatRelativeDate(new Date(request.created_at))}</p>
                {(request.status === "OPEN" || request.status === "IN_PROGRESS") && (
                  <div className="mt-4 flex flex-col gap-2 sm:flex-row">
                    {request.status === "OPEN" && (
                      <Button size="sm" variant="secondary" disabled={busyId === request.id}
                        onClick={() => void updateRequest(request, "IN_PROGRESS")}>
                        Start work
                      </Button>
                    )}
                    <input
                      value={resolution[request.id] ?? ""}
                      onChange={(event) => setResolution({ ...resolution, [request.id]: event.target.value })}
                      aria-label={`Resolution for ${request.reference}`}
                      placeholder="Resolution required to close"
                      className="min-w-0 flex-1 rounded-lg border border-surface-border px-3 py-2 text-sm"
                    />
                    <Button size="sm" disabled={busyId === request.id}
                      onClick={() => void updateRequest(request, "RESOLVED")}>
                      Resolve
                    </Button>
                  </div>
                )}
                {request.resolution && <p className="mt-3 text-sm text-gray-600">Resolution: {request.resolution}</p>}
              </article>
            ))}
          </div>
        )}
      </div>
    </AdminOnly>
  );
}

type Analytics = {
  window_days: number;
  total_events?: number;
  previous_events?: number;
  trend?: string;
  by_category?: { category: string; count: number }[];
};

export function AdminAnalyticsPage() {
  const { data, error, isLoading, refresh } = useAsyncData(async () => {
    const response = await fetchAPI<AdminEnvelope<Analytics>>("/institutional/analytics?days=30");
    return response.data;
  }, []);
  const total = data?.total_events ?? 0;
  const previous = data?.previous_events ?? 0;

  return (
    <AdminOnly>
      <div className="mx-auto max-w-6xl space-y-6 px-4 py-6">
        <AdminHeading title="Analytics" description="Campus event engagement over the last 30 days." />
        <AdminError error={error} retry={refresh} />
        {isLoading ? <p className="text-sm text-gray-500">Loading analytics…</p> : (
          <>
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
              {[
                ["Events in period", total],
                ["Previous period", previous],
                ["Change", `${total - previous >= 0 ? "+" : ""}${total - previous}`],
              ].map(([label, value]) => (
                <div key={label} className="rounded-xl border border-surface-border bg-white p-5">
                  <p className="text-sm text-gray-500">{label}</p>
                  <p className="mt-2 text-3xl font-bold text-gray-900">{value}</p>
                </div>
              ))}
            </div>
            <section className="rounded-xl border border-surface-border bg-white p-5">
              <h2 className="font-semibold text-gray-900">Events by category</h2>
              {(data?.by_category ?? []).length === 0 ? (
                <p className="mt-4 text-sm text-gray-500">No event analytics are available yet.</p>
              ) : (
                <ul className="mt-3 divide-y divide-surface-border">
                  {(data?.by_category ?? []).map((category) => (
                    <li key={category.category} className="flex justify-between py-3 text-sm">
                      <span className="text-gray-700">{category.category}</span>
                      <span className="font-medium text-gray-900">{category.count}</span>
                    </li>
                  ))}
                </ul>
              )}
            </section>
          </>
        )}
      </div>
    </AdminOnly>
  );
}

export function AdminApprovalsPage() {
  const { data, error, isLoading, refresh, setData } = useAsyncData(async () => {
    const response = await fetchAPI<AdminEnvelope<{ clubs: Club[] }>>("/admin/approvals/clubs");
    return response.data.clubs;
  }, []);
  const [busyId, setBusyId] = useState<string | null>(null);
  const [actionError, setActionError] = useState<string | null>(null);

  async function decide(club: Club, status: "APPROVED" | "REJECTED") {
    setBusyId(club.id);
    setActionError(null);
    try {
      await fetchAPI(`/admin/approvals/clubs/${club.id}`, {
        method: "PATCH",
        body: JSON.stringify({ status }),
      });
      setData((clubs) => clubs?.filter((item) => item.id !== club.id) ?? []);
    } catch (err) {
      setActionError(err instanceof APIError ? err.message : "Could not update this approval.");
    } finally {
      setBusyId(null);
    }
  }

  return (
    <AdminOnly>
      <div className="mx-auto max-w-6xl space-y-6 px-4 py-6">
        <AdminHeading title="Approvals" description="Review campus groups awaiting verification." />
        <AdminError error={error} retry={refresh} />
        {actionError && <p role="alert" className="text-sm text-red-600">{actionError}</p>}
        {isLoading ? <p className="text-sm text-gray-500">Loading pending approvals…</p> : (
          <div className="space-y-3">
            {(data ?? []).length === 0 ? (
              <p className="rounded-xl border border-surface-border bg-white p-6 text-center text-sm text-gray-500">
                No groups are waiting for approval.
              </p>
            ) : (data ?? []).map((club) => (
              <article key={club.id} className="rounded-xl border border-surface-border bg-white p-4">
                <div className="flex flex-wrap items-start justify-between gap-3">
                  <div>
                    <h2 className="font-semibold text-gray-900">{club.name}</h2>
                    <p className="mt-1 text-xs text-gray-500">
                      {club.category}
                      {club.created_at ? ` · Submitted ${formatRelativeDate(new Date(club.created_at))}` : ""}
                    </p>
                  </div>
                  <Badge variant="warning">{club.verification_status}</Badge>
                </div>
                <p className="mt-3 text-sm text-gray-600">{club.description}</p>
                <div className="mt-4 flex gap-2">
                  <Button size="sm" disabled={busyId === club.id} onClick={() => void decide(club, "APPROVED")}>
                    Approve
                  </Button>
                  <Button size="sm" variant="secondary" disabled={busyId === club.id}
                    onClick={() => void decide(club, "REJECTED")}>
                    Reject
                  </Button>
                </div>
              </article>
            ))}
          </div>
        )}
      </div>
    </AdminOnly>
  );
}

type AdminUser = {
  id: string;
  email: string;
  display_name: string;
  role: string;
  status: string;
  created_at: string;
};

export function AdminUsersPage() {
  const { data, error, isLoading, refresh } = useAsyncData(async () => {
    const response = await fetchAPI<AdminEnvelope<{ users: AdminUser[] }>>("/admin/users?limit=100");
    return response.data.users;
  }, []);

  return (
    <AdminOnly>
      <div className="mx-auto max-w-6xl space-y-6 px-4 py-6">
        <AdminHeading title="Users" description="Registered campus accounts and their access roles." />
        <AdminError error={error} retry={refresh} />
        {isLoading ? <p className="text-sm text-gray-500">Loading users…</p> : (
          <div className="overflow-x-auto rounded-xl border border-surface-border bg-white">
            <table className="w-full text-left text-sm">
              <thead className="bg-surface-raised text-xs uppercase text-gray-500">
                <tr>
                  <th className="px-4 py-3">Name</th>
                  <th className="px-4 py-3">Email</th>
                  <th className="px-4 py-3">Role</th>
                  <th className="px-4 py-3">Status</th>
                  <th className="px-4 py-3">Joined</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-surface-border">
                {(data ?? []).map((user) => (
                  <tr key={user.id}>
                    <td className="px-4 py-3 font-medium text-gray-900">{user.display_name}</td>
                    <td className="px-4 py-3 text-gray-600">{user.email}</td>
                    <td className="px-4 py-3"><Badge variant="info">{user.role}</Badge></td>
                    <td className="px-4 py-3">{user.status}</td>
                    <td className="px-4 py-3 text-gray-500">{formatRelativeDate(new Date(user.created_at))}</td>
                  </tr>
                ))}
                {(data ?? []).length === 0 && (
                  <tr><td colSpan={5} className="px-4 py-8 text-center text-gray-500">No registered accounts were found.</td></tr>
                )}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </AdminOnly>
  );
}
