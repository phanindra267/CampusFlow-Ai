"use client";
import Link from "next/link";
import { useAsyncData } from "@/hooks/useAsyncData";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { AIInsightCard } from "@/components/ai/AIComponents";
import { api, unwrap } from "@/lib/api";
import { formatRelativeDate } from "@/lib/utils";

export default function ResearchPage() {
  const { data, error, isLoading, refresh } = useAsyncData(async () => {
    const [clubRes, opportunityRes] = await Promise.all([api.clubs(), api.opportunities()]);
    const allClubs = unwrap(clubRes).clubs;
    const allOpportunities = unwrap(opportunityRes).opportunities;

    // Research groups are the club records tagged as research; when the campus has
    // none yet the rest of the community stays visible as adjacent spaces.
    const research = allClubs.filter((c) => /research|study|lab/i.test(`${c.category} ${c.name}`));
    return {
      groups: research.length > 0 ? research : allClubs,
      projects: allOpportunities.filter((o) => o.status === "OPEN").slice(0, 6),
    };
  }, []);

  return (
    <div className="max-w-5xl mx-auto px-4 py-6 space-y-6 animate-fade-in">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Research & Collaboration</h1>
        <p className="text-sm text-gray-500">Open projects, study groups and the people working on them</p>
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
        title="Where to start"
        content={
          isLoading
            ? "Loading the research landscape…"
            : `${data?.groups.length ?? 0} group${data?.groups.length === 1 ? "" : "s"} are open for collaboration and ${data?.projects.length ?? 0} project${data?.projects.length === 1 ? "" : "s"} are accepting members. Join a group to see its discussions.`
        }
        sources={["Groups", "Opportunities"]}
        updatedAt="just now"
      />

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div>
          <div className="flex items-center justify-between mb-3">
            <h2 className="text-sm font-semibold text-gray-900">Open projects</h2>
            <Link href="/career" className="text-xs text-lpu-primary hover:underline">
              All opportunities
            </Link>
          </div>

          {isLoading ? (
            <p className="text-sm text-gray-500">Loading…</p>
          ) : (data?.projects.length ?? 0) === 0 ? (
            <p className="text-sm text-gray-500 py-4">No open projects right now.</p>
          ) : (
            <div className="space-y-3">
              {(data?.projects ?? []).map((opp) => (
                <div key={opp.id} className="bg-white rounded-xl border border-surface-border p-4">
                  <div className="flex items-start justify-between gap-2 mb-2">
                    <h3 className="font-medium text-gray-900">{opp.title}</h3>
                    {opp.category && <Badge variant="info">{opp.category}</Badge>}
                  </div>
                  {opp.description && (
                    <p className="text-xs text-gray-500 mb-2 line-clamp-2">{opp.description}</p>
                  )}
                  <div className="flex items-center justify-between gap-3">
                    <span className="text-xs text-amber-600">
                      Closes {formatRelativeDate(new Date(opp.registration_deadline))}
                    </span>
                    <Link href="/career">
                      <Button variant="primary" size="sm">
                        Apply
                      </Button>
                    </Link>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        <div>
          <div className="flex items-center justify-between mb-3">
            <h2 className="text-sm font-semibold text-gray-900">Research & study groups</h2>
            <Link href="/community" className="text-xs text-lpu-primary hover:underline">
              Browse all groups
            </Link>
          </div>

          {isLoading ? (
            <p className="text-sm text-gray-500">Loading…</p>
          ) : (data?.groups.length ?? 0) === 0 ? (
            <p className="text-sm text-gray-500 py-4">No research groups have been set up yet.</p>
          ) : (
            <div className="space-y-3">
              {(data?.groups ?? []).map((group) => (
                <div key={group.id} className="bg-white rounded-xl border border-surface-border p-4">
                  <div className="flex items-center gap-3 mb-2">
                    <div className="w-10 h-10 bg-lpu-light rounded-full flex items-center justify-center flex-shrink-0">
                      <span className="font-bold text-lpu-primary">
                        {(group.name || "G").slice(0, 1).toUpperCase()}
                      </span>
                    </div>
                    <div className="min-w-0">
                      <p className="font-medium text-gray-900 truncate">{group.name}</p>
                      <p className="text-xs text-gray-500">{group.category || "General"}</p>
                    </div>
                  </div>
                  {group.description && (
                    <p className="text-xs text-gray-600 mb-2 line-clamp-2">{group.description}</p>
                  )}
                  <div className="flex items-center justify-between gap-3">
                    <span className="text-xs text-gray-400">
                      {group.member_count ?? 0} member{group.member_count === 1 ? "" : "s"}
                    </span>
                    <Link href="/community">
                      <Button variant="secondary" size="sm">
                        Join group
                      </Button>
                    </Link>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}