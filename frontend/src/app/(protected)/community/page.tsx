"use client";

import { useEffect, useState } from "react";
import { useSearchParams } from "next/navigation";
import { useAsyncData } from "@/hooks/useAsyncData";
import { api, unwrap, type Club, type Discussion, type DiscussionReply } from "@/lib/api";
import { cn, formatRelativeDate } from "@/lib/utils";
import { Badge } from "@/components/ui/Badge";

type Tab = "discussions" | "groups";

const TABS: { id: Tab; label: string }[] = [
  { id: "discussions", label: "Discussions" },
  { id: "groups", label: "Groups & Clubs" },
];

const CATEGORIES = [
  { value: "GENERAL", label: "General" },
  { value: "ANNOUNCEMENT", label: "Announcement" },
  { value: "QUESTION", label: "Question" },
  { value: "EVENT", label: "Event" },
  { value: "OPPORTUNITY", label: "Opportunity" },
  { value: "RESEARCH", label: "Research" },
];

export default function CommunityPage() {
  const [tab, setTab] = useState<Tab>("discussions");

  const { data, error, isLoading, refresh } = useAsyncData(async () => {
    const [threads, groups] = await Promise.all([api.discussions(), api.clubs()]);
    return {
      discussions: unwrap(threads).discussions,
      clubs: unwrap(groups).clubs,
    };
  }, []);

  const params = useSearchParams();
  const openDiscussionId = params.get("discussion");

  return (
    <div className="max-w-5xl mx-auto px-4 py-6 space-y-6 animate-fade-in">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Community</h1>
        <p className="text-sm text-gray-500 mt-0.5">
          Everything happening on this campus, in one place
        </p>
      </div>

      <div className="flex gap-1 bg-surface-raised border border-surface-border rounded-lg p-1 w-fit">
        {TABS.map(({ id, label }) => (
          <button
            key={id}
            onClick={() => setTab(id)}
            className={cn(
              "px-4 py-1.5 rounded-md text-sm font-medium transition-all",
              tab === id ? "bg-white text-gray-900 shadow-sm" : "text-gray-500 hover:text-gray-700",
            )}
          >
            {label}
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
      ) : !data ? null : tab === "discussions" ? (
        <DiscussionsTab discussions={data.discussions} openId={openDiscussionId} onPosted={refresh} />
      ) : (
        <GroupsTab clubs={data.clubs} onChanged={refresh} />
      )}
    </div>
  );
}

function DiscussionsTab({
  discussions,
  openId,
  onPosted,
}: {
  discussions: Discussion[];
  openId: string | null;
  onPosted: () => void;
}) {
  const [composerOpen, setComposerOpen] = useState(false);
  const [title, setTitle] = useState("");
  const [body, setBody] = useState("");
  const [category, setCategory] = useState("GENERAL");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const submit = async () => {
    if (!title.trim() || !body.trim()) return;
    setSubmitting(true);
    setError(null);
    try {
      await api.createDiscussion({ title: title.trim(), body: body.trim(), category });
      setTitle("");
      setBody("");
      setCategory("GENERAL");
      setComposerOpen(false);
      onPosted();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not post the discussion.");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <p className="text-sm text-gray-500">
          {discussions.length === 0
            ? "No discussions yet. Start the first one."
            : `${discussions.length} discussion${discussions.length === 1 ? "" : "s"}`}
        </p>
        <button
          onClick={() => setComposerOpen((v) => !v)}
          className="text-sm font-medium text-lpu-primary hover:underline"
        >
          {composerOpen ? "Cancel" : "New discussion"}
        </button>
      </div>

      {composerOpen && (
        <div className="space-y-3 rounded-lg border border-surface-border bg-white p-4">
          <select
            value={category}
            onChange={(e) => setCategory(e.target.value)}
            aria-label="Discussion category"
            className="w-full rounded-lg border border-surface-border px-3 py-2 text-sm"
          >
            {CATEGORIES.map((c) => (
              <option key={c.value} value={c.value}>
                {c.label}
              </option>
            ))}
          </select>
          <input
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            placeholder="What is this about?"
            aria-label="Discussion title"
            className="w-full rounded-lg border border-surface-border px-3 py-2 text-sm"
          />
          <textarea
            value={body}
            onChange={(e) => setBody(e.target.value)}
            placeholder="Share the details with your campus community…"
            aria-label="Discussion body"
            rows={4}
            className="w-full rounded-lg border border-surface-border px-3 py-2 text-sm resize-y"
          />
          {error && <p className="text-sm text-red-600">{error}</p>}
          <button
            onClick={() => void submit()}
            disabled={submitting || !title.trim() || !body.trim()}
            className="rounded-lg bg-lpu-primary px-4 py-2 text-sm font-medium text-white disabled:opacity-50"
          >
            {submitting ? "Posting…" : "Post"}
          </button>
        </div>
      )}

      {discussions.map((thread) => (
        <DiscussionCard key={thread.id} thread={thread} expanded={thread.id === openId} />
      ))}
    </div>
  );
}

function DiscussionCard({ thread, expanded }: { thread: Discussion; expanded: boolean }) {
  const [replies, setReplies] = useState<DiscussionReply[] | null>(null);
  const [reply, setReply] = useState("");
  const [posting, setPosting] = useState(false);

  const loadReplies = async () => {
    try {
      const res = await api.discussion(thread.id);
      setReplies(unwrap(res).replies);
    } catch {
      setReplies([]);
    }
  };

  // Replies are only fetched when a thread is opened, so the list page stays
  // cheap regardless of how many threads exist.
  useEffect(() => {
    if (!expanded) return;
    let cancelled = false;
    void (async () => {
      try {
        const res = await api.discussion(thread.id);
        if (!cancelled) setReplies(unwrap(res).replies);
      } catch {
        if (!cancelled) setReplies([]);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [expanded, thread.id]);

  const postReply = async () => {
    if (!reply.trim()) return;
    setPosting(true);
    try {
      await api.replyToDiscussion(thread.id, reply.trim());
      setReply("");
      await loadReplies();
    } finally {
      setPosting(false);
    }
  };

  return (
    <article className="rounded-lg border border-surface-border bg-white p-4 space-y-2">
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <h2 className="font-semibold text-gray-900">{thread.title}</h2>
          <p className="mt-1 text-sm text-gray-600 whitespace-pre-wrap">{thread.body}</p>
        </div>
        {thread.is_pinned && <Badge variant="info">Pinned</Badge>}
      </div>

      <div className="flex flex-wrap items-center gap-2 text-xs text-gray-500">
        <Badge variant="default">{thread.category}</Badge>
        <span>{thread.author_name || "Member"}</span>
        <span>·</span>
        <span>{formatRelativeDate(new Date(thread.last_activity_at))}</span>
        <span>·</span>
        <button
          onClick={() => void loadReplies()}
          className="hover:text-gray-700 hover:underline"
        >
          {thread.reply_count} repl{thread.reply_count === 1 ? "y" : "ies"}
        </button>
      </div>

      {expanded && (
        <div className="space-y-3 border-t border-surface-border pt-3">
          {replies === null && <p className="text-sm text-gray-500">Loading replies…</p>}

          {replies?.length === 0 && (
            <p className="text-sm text-gray-500">No replies yet.</p>
          )}

          {replies?.map((r) => (
            <div key={r.id} className="rounded-md bg-surface-raised px-3 py-2">
              <p className="text-sm text-gray-800 whitespace-pre-wrap">{r.body}</p>
              <p className="mt-1 text-xs text-gray-500">
                {r.author_name || "Member"} · {formatRelativeDate(new Date(r.created_at))}
              </p>
            </div>
          ))}

          <div className="flex gap-2">
            <input
              value={reply}
              onChange={(e) => setReply(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter" && !e.shiftKey) void postReply();
              }}
              placeholder="Write a reply…"
              aria-label="Reply text"
              className="flex-1 rounded-lg border border-surface-border px-3 py-2 text-sm"
            />
            <button
              onClick={() => void postReply()}
              disabled={posting || !reply.trim()}
              className="rounded-lg bg-lpu-primary px-3 py-2 text-sm font-medium text-white disabled:opacity-50"
            >
              Reply
            </button>
          </div>
        </div>
      )}
    </article>
  );
}

function GroupsTab({ clubs, onChanged }: { clubs: Club[]; onChanged: () => void }) {
  const [joining, setJoining] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  const toggle = async (club: Club) => {
    setJoining(club.id);
    setError(null);
    try {
      if (club.is_member) await api.leaveClub(club.id);
      else await api.joinClub(club.id);
      onChanged();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not join that group.");
    } finally {
      setJoining(null);
    }
  };

  return (
    <div className="space-y-3">
      <p className="text-sm text-gray-500">
        {clubs.length === 0
          ? "No groups have been set up yet."
          : `${clubs.length} group${clubs.length === 1 ? "" : "s"} on this campus`}
      </p>

      {error && <p className="text-sm text-red-600">{error}</p>}

      <div className="grid gap-3 sm:grid-cols-2">
        {clubs.map((club) => (
          <div
            key={club.id}
            className="flex flex-col rounded-lg border border-surface-border bg-white p-4 space-y-2"
          >
            <div className="flex items-start justify-between gap-2">
              <h2 className="font-semibold text-gray-900">{club.name}</h2>
              {club.category && <Badge variant="default">{club.category}</Badge>}
            </div>

            {club.description && (
              <p className="text-sm text-gray-600 line-clamp-3">{club.description}</p>
            )}

            <div className="mt-auto flex items-center justify-between pt-2">
              <span className="text-xs text-gray-500">
                {club.member_count ?? 0} member{club.member_count === 1 ? "" : "s"}
              </span>
              <button
                onClick={() => void toggle(club)}
                disabled={joining === club.id}
                className="rounded-lg border border-lpu-primary px-3 py-1.5 text-xs font-medium text-lpu-primary hover:bg-lpu-light disabled:opacity-50"
              >
                {joining === club.id
                  ? club.is_member
                    ? "Leaving…"
                    : "Joining…"
                  : club.is_member
                    ? "Leave"
                    : "Join"}
              </button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}