"use client";
import { useState } from "react";
import Link from "next/link";
import { useAsyncData } from "@/hooks/useAsyncData";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { APIError, api, unwrap, type Opportunity } from "@/lib/api";
import { cn, formatRelativeDate, isPast } from "@/lib/utils";

type Tab = "opportunities" | "applications" | "prepare";

const PREPARATION_STARTERS = [
  "I have three months before a software internship interview. What should I learn first?",
  "Help me turn my project work into three strong resume bullets.",
  "Explain system design the way an interviewer would want to hear it.",
  "What questions should I ask at the end of a technical interview?",
  "How do I prepare for a data structures round in two weeks?",
];

export default function CareerPage() {
  const [tab, setTab] = useState<Tab>("opportunities");
  const [applying, setApplying] = useState<string | null>(null);
  const [withdrawing, setWithdrawing] = useState<string | null>(null);
  const [actionError, setActionError] = useState<string | null>(null);

  const { data, error: loadError, isLoading, refresh } = useAsyncData(async () => {
    const [opportunityRes, applicationRes] = await Promise.all([api.opportunities(), api.myApplications()]);
    return {
      opportunities: unwrap(opportunityRes).opportunities,
      applications: unwrap(applicationRes).applications,
    };
  }, []);

  const opportunities = data?.opportunities ?? [];
  const applications = data?.applications ?? [];

  const open = opportunities.filter(
    (o) => o.status === "OPEN" && !isPast(new Date(o.registration_deadline)),
  );
  const closed = opportunities.filter((o) => !open.includes(o));
  const activeApplications = applications.filter((a) => a.status !== "WITHDRAWN");
  const error = actionError ?? loadError;

  const apply = async (id: string) => {
    setApplying(id);
    setActionError(null);
    try {
      await api.applyToOpportunity(id, "");
      refresh();
    } catch (err) {
      setActionError(err instanceof APIError ? err.message : "Could not submit that application.");
    } finally {
      setApplying(null);
    }
  };

  const withdraw = async (id: string) => {
    setWithdrawing(id);
    setActionError(null);
    try {
      await api.withdrawApplication(id);
      refresh();
    } catch (err) {
      setActionError(err instanceof APIError ? err.message : "Could not withdraw that application.");
    } finally {
      setWithdrawing(null);
    }
  };

  return (
    <div className="max-w-5xl mx-auto px-4 py-6 space-y-6 animate-fade-in">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Career Hub</h1>
        <p className="text-sm text-gray-500">
          Opportunities posted to this campus and your applications to them
        </p>
      </div>

      {error && (
        <div className="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
          {error}
          <button onClick={refresh} className="ml-2 underline font-medium">
            Retry
          </button>
        </div>
      )}

      <div className="flex gap-1 bg-surface-raised border border-surface-border rounded-lg p-1 w-fit flex-wrap">
        {(["opportunities", "applications", "prepare"] as const).map((t) => (
          <button
            key={t}
            onClick={() => setTab(t)}
            className={cn(
              "px-4 py-1.5 rounded-md text-sm font-medium transition-all",
              tab === t ? "bg-white text-gray-900 shadow-sm" : "text-gray-500 hover:text-gray-700",
            )}
          >
            {t === "applications" ? `Applications (${activeApplications.length})` : t.charAt(0).toUpperCase() + t.slice(1)}
          </button>
        ))}
      </div>

      {isLoading ? (
        <p className="text-sm text-gray-500">Loading…</p>
      ) : tab === "opportunities" ? (
        <div className="space-y-6">
          <section>
            <h2 className="text-sm font-semibold text-gray-900 mb-3">
              Open now ({open.length})
            </h2>
            {open.length === 0 ? (
              <p className="text-sm text-gray-500 py-4">
                Nothing is open right now. Check back soon or start a discussion in your group.
              </p>
            ) : (
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                {open.map((opp) => (
                  <OpportunityCard
                    key={opp.id}
                    opportunity={opp}
                    applying={applying === opp.id}
                    applied={applications.some((a) => a.opportunity_id === opp.id && a.status !== "WITHDRAWN")}
                    onApply={() => void apply(opp.id)}
                  />
                ))}
              </div>
            )}
          </section>

          {closed.length > 0 && (
            <section>
              <h2 className="text-sm font-semibold text-gray-900 mb-3">Closed ({closed.length})</h2>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4 opacity-70">
                {closed.map((opp) => (
                  <OpportunityCard
                    key={opp.id}
                    opportunity={opp}
                    applying={applying === opp.id}
                    applied={applications.some((a) => a.opportunity_id === opp.id && a.status !== "WITHDRAWN")}
                    onApply={() => void apply(opp.id)}
                  />
                ))}
              </div>
            </section>
          )}
        </div>
      ) : tab === "applications" ? (
        <div className="space-y-3">
          {applications.length === 0 ? (
            <p className="text-sm text-gray-500 py-6 text-center">
              You have not applied to anything yet.
            </p>
          ) : (
            applications.map((application) => (
              <div
                key={application.id}
                className="bg-white rounded-xl border border-surface-border p-4 flex items-start justify-between gap-3"
              >
                <div className="min-w-0">
                  <p className="font-medium text-gray-900">{application.title || "Opportunity"}</p>
                  <p className="text-xs text-gray-500 mt-0.5">
                    Applied {formatRelativeDate(new Date(application.applied_at))}
                  </p>
                </div>
                <div className="flex items-center gap-2 flex-shrink-0">
                  <Badge
                    variant={
                      application.status === "ACCEPTED"
                        ? "success"
                        : application.status === "REJECTED" || application.status === "WITHDRAWN"
                          ? "error"
                          : "warning"
                    }
                  >
                    {application.status}
                  </Badge>
                  {application.status !== "WITHDRAWN" && (
                    <Button
                      variant="secondary"
                      size="sm"
                      disabled={withdrawing === application.id}
                      onClick={() => void withdraw(application.opportunity_id)}
                    >
                      {withdrawing === application.id ? "Withdrawing…" : "Withdraw"}
                    </Button>
                  )}
                </div>
              </div>
            ))
          )}
        </div>
      ) : (
        <div className="bg-white rounded-xl border border-surface-border p-6 space-y-4">
          <div>
            <h2 className="font-semibold text-gray-900">Prepare with the AI assistant</h2>
            <p className="text-sm text-gray-600 mt-1">
              The assistant runs entirely on your device using WebLLM on WebGPU, so your preparation
              notes are never sent to an external model provider. Pick a starting point:
            </p>
          </div>
          <div className="space-y-2">
            {PREPARATION_STARTERS.map((starter) => (
              <Link
                key={starter}
                href="/ai"
                className="block rounded-lg border border-surface-border px-4 py-3 text-sm text-gray-800 hover:bg-surface-raised transition-colors"
              >
                {starter}
              </Link>
            ))}
          </div>
          <p className="text-xs text-gray-400">
            CampusCare does not publish invented match scores or placement statistics. Readiness data on this
            page comes only from opportunities and applications stored on this campus server.
          </p>
        </div>
      )}
    </div>
  );
}

function OpportunityCard({
  opportunity,
  applied,
  applying,
  onApply,
}: {
  opportunity: Opportunity;
  applied: boolean;
  applying: boolean;
  onApply: () => void;
}) {
  const closed = isPast(new Date(opportunity.registration_deadline)) || opportunity.status !== "OPEN";

  return (
    <div className="bg-white rounded-xl border border-surface-border p-5 flex flex-col">
      <div className="flex items-start justify-between gap-2 mb-2">
        <h3 className="font-semibold text-gray-900 leading-tight">{opportunity.title}</h3>
        {opportunity.type && <Badge variant="lpu">{opportunity.type}</Badge>}
      </div>

      {opportunity.description && (
        <p className="text-sm text-gray-600 mb-3 line-clamp-3">{opportunity.description}</p>
      )}

      <div className="flex flex-wrap gap-1.5 mb-3">
        {opportunity.category && <Badge variant="info">{opportunity.category}</Badge>}
        {opportunity.delivery_mode && <Badge variant="default">{opportunity.delivery_mode}</Badge>}
        {opportunity.capacity > 0 && <Badge variant="default">{opportunity.capacity} seats</Badge>}
      </div>

      <div className="mt-auto">
        <p className="text-xs text-amber-600 mb-3">
          {closed
            ? "Applications closed"
            : `Closes ${formatRelativeDate(new Date(opportunity.registration_deadline))}`}
        </p>
        <Button
          variant={applied ? "secondary" : "primary"}
          size="sm"
          className="w-full"
          disabled={closed || applied || applying}
          onClick={onApply}
        >
          {applying ? "Applying…" : applied ? "Applied" : closed ? "Closed" : "Apply now"}
        </Button>
      </div>
    </div>
  );
}