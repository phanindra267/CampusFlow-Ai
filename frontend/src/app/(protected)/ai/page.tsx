"use client";

import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { AIChatMessage } from "@/components/ai/AIComponents";
import { Button } from "@/components/ui/Button";
import { useAuth } from "@/contexts/AuthContext";
import { useLLM } from "@/hooks/useLLM";
import { api } from "@/lib/api";
import { route, type Intent, type Route } from "@/lib/ai/router";
import type { ChatMessage } from "@/lib/llm";
import { cn } from "@/lib/utils";

/**
 * Suggestions are split by intent so the first thing a member tries exercises
 * the fast path rather than the model.
 */
const STRUCTURED_SUGGESTIONS = [
  "Show upcoming events",
  "Find coding clubs",
  "Find internships",
  "My notifications",
];

const GENERATIVE_SUGGESTIONS = [
  "Based on my interests, what campus opportunities should I consider?",
  "Which events are most relevant to someone interested in cloud computing?",
  "Summarise these opportunities and explain which ones suit me best",
];

const SYSTEM_PROMPT = [
  "You are CampusCare AI, the campus assistant for the CampusCare community platform.",
  "Campus records are retrieved for you and included below the question.",
  "Answer using only those records. When they do not contain the answer, say plainly that",
  "the information is not available in CampusCare and name the page the member should check.",
  "Never invent events, deadlines, seat counts, organisers, rankings, placements, marks or",
  "attendance figures. Being concise and specific is better than being comprehensive.",
].join(" ");

type Message = {
  id: string;
  role: "user" | "assistant";
  content: string;
  /** Present on assistant messages that came from retrieval, not generation. */
  route?: Route;
  /** Records backing a retrieval answer. */
  records?: RetrievedRecord[];
};

type RetrievedRecord = {
  entity_type: string;
  entity_id: string;
  title: string;
  url: string;
  retrieved_by: string;
};

let messageCounter = 0;
function nextId(): string {
  messageCounter += 1;
  return `${Date.now().toString(36)}-${messageCounter}`;
}

type StructuredPlan = {
  /** Human name of the data source, used in the reply. */
  source: string;
  /** Where to look when nothing matched. */
  hint: string;
  /** Fetches the records for this intent. */
  load: () => Promise<RetrievedRecord[]>;
};

/**
 * structuredPlan maps a STRUCTURED intent onto the existing REST client. Each
 * entry reuses the same endpoint the corresponding page uses, so authorisation
 * and data shape stay in one place rather than being reimplemented here.
 */
function structuredPlan(intent: Intent, query: string): StructuredPlan {
  const q = query.trim() || undefined;

  switch (intent) {
    case "EVENT_SEARCH":
    case "REGISTRATION_QUERY":
      return {
        source: "events list",
        hint: "Try the Discover page to browse everything that is scheduled.",
        load: async () => {
          const { data } = await api.events({ q });
          return data.events.slice(0, RESULT_LIMIT).map((event) => ({
            entity_type: "event",
            entity_id: event.id,
            title: `${event.title} — ${formatWhen(event.start_time)}`,
            url: `/discover?event=${event.id}`,
            retrieved_by: "events",
          }));
        },
      };

    case "CLUB_SEARCH":
      return {
        source: "clubs list",
        hint: "Try the Community page to browse every club.",
        load: async () => {
          const { data } = await api.clubs({ q });
          return data.clubs.slice(0, RESULT_LIMIT).map((club) => ({
            entity_type: "club",
            entity_id: club.id,
            title: `${club.name}${club.member_count ? ` · ${club.member_count} members` : ""}`,
            url: `/community?club=${club.id}`,
            retrieved_by: "clubs",
          }));
        },
      };

    case "OPPORTUNITY_SEARCH":
    case "APPLICATION_QUERY":
      return {
        source: "opportunities list",
        hint: "Try the Career page to browse every opportunity.",
        load: async () => {
          const { data } = await api.opportunities({ q });
          return data.opportunities.slice(0, RESULT_LIMIT).map((opportunity) => ({
            entity_type: "opportunity",
            entity_id: opportunity.id,
            title: `${opportunity.title} — ${opportunity.type.replace(/_/g, " ").toLowerCase()}`,
            url: `/career?opportunity=${opportunity.id}`,
            retrieved_by: "opportunities",
          }));
        },
      };

    case "RESOURCE_BOOKING":
      return {
        source: "campus services list",
        hint: "Try the Campus page to see every bookable resource.",
        load: async () => {
          const { data } = await api.services({ q });
          return data.resources.slice(0, RESULT_LIMIT).map((resource) => ({
            entity_type: "resource",
            entity_id: resource.id,
            title: `${resource.name} — ${resource.location}${
              resource.is_bookable ? " · bookable" : ""
            }`,
            url: `/campus?resource=${resource.id}`,
            retrieved_by: "services",
          }));
        },
      };

    case "NOTIFICATION_QUERY":
      return {
        source: "notifications",
        hint: "Open the Notifications page for the full history.",
        load: async () => {
          const { data } = await api.notifications();
          return data.notifications.slice(0, RESULT_LIMIT).map((notification) => ({
            entity_type: "notification",
            entity_id: notification.id,
            title: `${notification.title}${
              notification.read_at ? "" : " · unread"
            } — ${formatWhen(notification.created_at)}`,
            url: "/notifications",
            retrieved_by: "notifications",
          }));
        },
      };

    default:
      // Intents without a dedicated list endpoint fall through to hybrid
      // retrieval, which still answers without a model.
      return {
        source: "campus records",
        hint: "Try different wording.",
        load: async () => {
          const { data } = await api.aiContext(query);
          return data.records.map((record) => ({
            entity_type: record.entity_type,
            entity_id: record.entity_id,
            title: record.title,
            url: record.url,
            retrieved_by: record.retrieved_by,
          }));
        },
      };
  }
}

const RESULT_LIMIT = 8;

const dateFormatter = new Intl.DateTimeFormat("en-IN", {
  dateStyle: "medium",
  timeStyle: "short",
});

function formatWhen(value: string): string {
  const parsed = new Date(value);
  return Number.isNaN(parsed.getTime()) ? value : dateFormatter.format(parsed);
}

export default function AIPage() {
  const { user } = useAuth();
  const {
    status,
    model,
    progress,
    error,
    capabilities,
    isStreaming,
    initialize,
    generate,
    stop,
  } = useLLM();

  const [messages, setMessages] = useState<Message[]>([]);
  const [input, setInput] = useState("");
  const [isRetrieving, setIsRetrieving] = useState(false);
  const bottomRef = useRef<HTMLDivElement>(null);
  const messagesRef = useRef<Message[]>([]);

  useEffect(() => {
    messagesRef.current = messages;
  }, [messages]);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages, isStreaming]);

  const greeting = useMemo(() => {
    const hour = new Date().getHours();
    return hour < 12 ? "Good morning" : hour < 17 ? "Good afternoon" : "Good evening";
  }, []);

  const hasStarted = messages.length > 0;

  const appendMessage = useCallback((message: Message) => {
    setMessages((current) => [...current, message]);
  }, []);

  /**
   * answerWithStructured calls the existing REST endpoint for the intent, so a
   * structured request is served by the same handlers, authorisation and audit
   * trail as the rest of the application. No model is involved.
   */
  const answerWithStructured = useCallback(
    async (text: string, currentRoute: Route) => {
      const plan = structuredPlan(currentRoute.intent, currentRoute.searchQuery);

      setIsRetrieving(true);
      try {
        const records = await plan.load();

        if (records.length === 0) {
          appendMessage({
            id: nextId(),
            role: "assistant",
            content: `Nothing in CampusCare matched that yet. ${plan.hint}`,
            route: currentRoute,
          });
          return;
        }

        appendMessage({
          id: nextId(),
          role: "assistant",
          content: `Here ${records.length === 1 ? "is the record" : `are the ${records.length} records`} from the ${plan.source}:`,
          route: currentRoute,
          records,
        });
      } catch {
        appendMessage({
          id: nextId(),
          role: "assistant",
          content: `I could not reach the ${plan.source} right now. Please try again shortly.`,
          route: currentRoute,
        });
      } finally {
        setIsRetrieving(false);
      }
    },
    [appendMessage],
  );

  /**
   * answerWithSearch handles free-text queries through hybrid retrieval. Still
   * no model: the answer is a factual list of campus records.
   */
  const answerWithSearch = useCallback(
    async (text: string, currentRoute: Route) => {
      setIsRetrieving(true);
      try {
        const envelope = await api.aiContext(currentRoute.searchQuery || text);
        const data = envelope.data;

        if (!data.found || data.records.length === 0) {
          appendMessage({
            id: nextId(),
            role: "assistant",
            content:
              "I could not find anything in CampusCare matching that. Try different wording, or check the relevant page directly.",
            route: currentRoute,
          });
          return;
        }

        appendMessage({
          id: nextId(),
          role: "assistant",
          content: `Found ${data.count} matching ${data.count === 1 ? "record" : "records"}${
            data.vector_hits > 0 ? " using keyword and semantic matching" : ""
          }:`,
          route: currentRoute,
          records: data.records.map((record) => ({
            entity_type: record.entity_type,
            entity_id: record.entity_id,
            title: record.title,
            url: record.url,
            retrieved_by: record.retrieved_by,
          })),
        });
      } catch {
        appendMessage({
          id: nextId(),
          role: "assistant",
          content: "Campus search is unavailable right now. Please try again shortly.",
          route: currentRoute,
        });
      } finally {
        setIsRetrieving(false);
      }
    },
    [appendMessage],
  );

  /** answerWithGeneration retrieves context first, then streams a WebLLM reply. */
  const answerWithGeneration = useCallback(
    async (text: string, currentRoute: Route) => {
      setIsRetrieving(true);

      let context = "";
      let found = false;
      try {
        const envelope = await api.aiContext(currentRoute.searchQuery || text);
        context = envelope.data.context;
        found = envelope.data.found;
      } catch {
        // Retrieval failure must not block generation; the prompt below tells
        // the model it has no records to work from.
        context = "";
      } finally {
        setIsRetrieving(false);
      }

      const ready = (await initialize()) === true;
      if (!ready) return;

      const history = messagesRef.current;
      const assistantId = nextId();

      appendMessage({ id: nextId(), role: "user", content: text });
      appendMessage({ id: assistantId, role: "assistant", content: "" });

      const transcript: ChatMessage[] = [
        { role: "system", content: buildSystemPrompt(user?.displayName, context, found) },
        ...history
          .filter((message) => message.role === "user" || message.content.trim().length > 0)
          .filter((message) => message.id !== assistantId)
          .map((message) => ({ role: message.role, content: message.content })),
        { role: "user", content: text },
      ];

      try {
        await generate(transcript, (delta) => {
          setMessages((current) =>
            current.map((message) =>
              message.id === assistantId
                ? { ...message, content: message.content + delta }
                : message,
            ),
          );
        });
      } catch {
        setMessages((current) =>
          current.map((message) =>
            message.id === assistantId && message.content.trim().length === 0
              ? { ...message, content: "I could not complete that response. Please try again." }
              : message,
          ),
        );
      }
    },
    [appendMessage, generate, initialize, user?.displayName],
  );

  const sendMessage = useCallback(
    async (raw: string) => {
      const text = raw.trim();
      if (!text || isStreaming || isRetrieving) return;

      const currentRoute = route(text);

      if (currentRoute.kind === "GENERATIVE") {
        await answerWithGeneration(text, currentRoute);
        return;
      }

      appendMessage({ id: nextId(), role: "user", content: text });

      if (currentRoute.kind === "STRUCTURED") {
        await answerWithStructured(text, currentRoute);
        return;
      }

      await answerWithSearch(text, currentRoute);
    },
    [
      answerWithGeneration,
      answerWithSearch,
      answerWithStructured,
      appendMessage,
      isRetrieving,
      isStreaming,
    ],
  );

  const busy = isStreaming || isRetrieving || status === "loading";

  return (
    <div className="flex flex-col h-full max-h-screen">
      <div className="px-4 py-4 bg-white border-b border-surface-border flex-shrink-0">
        <div className="max-w-3xl mx-auto flex items-center gap-3">
          <div className="w-9 h-9 bg-ai-light border border-ai-accent/20 rounded-full flex items-center justify-center flex-shrink-0">
            <svg className="w-5 h-5 text-ai-accent" viewBox="0 0 24 24" fill="currentColor">
              <path d="M12 2l2.4 7.2H22l-6.2 4.5 2.4 7.2L12 17l-6.2 3.9 2.4-7.2L2 9.2h7.6L12 2z" />
            </svg>
          </div>
          <div>
            <p className="font-semibold text-gray-900 text-sm">CampusCare AI</p>
            <p className="text-xs text-gray-400">
              {status === "ready" && model
                ? `WebLLM · ${model}`
                : "Campus search · on-device generation when needed"}
            </p>
          </div>
          <div className="ml-auto">
            <EngineBadge status={status} />
          </div>
        </div>
      </div>

      {status === "loading" && progress && (
        <div className="px-4 py-2 bg-ai-light border-b border-ai-accent/20 flex-shrink-0">
          <div className="max-w-3xl mx-auto">
            <p className="text-xs text-ai-accent truncate">{progress.text}</p>
            {progress.progress !== null && (
              <div className="h-1 bg-white rounded-full mt-1 overflow-hidden">
                <div
                  className="h-full bg-ai-accent transition-all duration-300"
                  style={{ width: `${Math.round(progress.progress * 100)}%` }}
                />
              </div>
            )}
          </div>
        </div>
      )}

      {status === "error" && error && (
        <div className="px-4 py-2 bg-red-50 border-b border-red-200 flex-shrink-0">
          <div className="max-w-3xl mx-auto flex items-center gap-3">
            <p className="text-xs text-red-700 flex-1">{error}</p>
            <button
              onClick={() => void initialize()}
              className="text-xs font-medium text-red-700 underline"
            >
              Retry
            </button>
          </div>
        </div>
      )}

      {!capabilities.canGenerate && capabilities.reason && (
        <div className="px-4 py-2 bg-amber-50 border-b border-amber-200 flex-shrink-0">
          <div className="max-w-3xl mx-auto">
            <p className="text-xs text-amber-700">{capabilities.reason}</p>
          </div>
        </div>
      )}

      <div className="flex-1 overflow-y-auto px-4 py-6">
        <div className="max-w-3xl mx-auto space-y-4">
          {!hasStarted && (
            <div className="text-center py-8 animate-fade-in">
              <div className="w-16 h-16 bg-ai-light rounded-2xl flex items-center justify-center mx-auto mb-4">
                <svg className="w-8 h-8 text-ai-accent" viewBox="0 0 24 24" fill="currentColor">
                  <path d="M12 2l2.4 7.2H22l-6.2 4.5 2.4 7.2L12 17l-6.2 3.9 2.4-7.2L2 9.2h7.6L12 2z" />
                </svg>
              </div>
              <h2 className="text-xl font-bold text-gray-900">
                {greeting}
                {user?.displayName ? `, ${user.displayName.split(" ")[0]}.` : "."}
              </h2>
              <p className="text-sm text-gray-500 mt-1">
                Search events, clubs, opportunities and services instantly. Generative
                answers load an on-device model only when a question needs one.
              </p>

              <div className="mt-6">
                <p className="text-xs text-gray-400 mb-3 uppercase tracking-wide font-medium">
                  Search campus
                </p>
                <div className="flex flex-wrap gap-2 justify-center">
                  {STRUCTURED_SUGGESTIONS.map((suggestion) => (
                    <SuggestionChip
                      key={suggestion}
                      label={suggestion}
                      onSelect={() => void sendMessage(suggestion)}
                      disabled={busy}
                    />
                  ))}
                </div>
              </div>

              <div className="mt-5">
                <p className="text-xs text-gray-400 mb-3 uppercase tracking-wide font-medium">
                  Ask for advice
                </p>
                <div className="flex flex-wrap gap-2 justify-center">
                  {GENERATIVE_SUGGESTIONS.map((suggestion) => (
                    <SuggestionChip
                      key={suggestion}
                      label={suggestion}
                      onSelect={() => void sendMessage(suggestion)}
                      disabled={busy}
                    />
                  ))}
                </div>
              </div>

              {status === "idle" && capabilities.canGenerate && (
                <div className="mt-6 flex flex-col items-center gap-2">
                  <Button variant="ai" size="sm" onClick={() => void initialize()}>
                    Load the on-device model
                  </Button>
                  <p className="text-xs text-gray-400 max-w-md">
                    Optional. Only needed for questions that ask for advice, comparison or a
                    summary. The first run downloads the model.
                  </p>
                </div>
              )}
            </div>
          )}

          {messages.map((message) =>
            message.role === "assistant" && message.content.length === 0 ? (
              <div key={message.id} className="flex gap-3">
                <div className="flex-shrink-0 w-8 h-8 rounded-full bg-ai-light border border-ai-accent/20 flex items-center justify-center">
                  <svg className="w-4 h-4 text-ai-accent" viewBox="0 0 24 24" fill="currentColor">
                    <path d="M12 2l2.4 7.2H22l-6.2 4.5 2.4 7.2L12 17l-6.2 3.9 2.4-7.2L2 9.2h7.6L12 2z" />
                  </svg>
                </div>
                <div className="bg-white rounded-2xl rounded-tl-sm px-4 py-3 shadow-card border border-surface-border">
                  <span
                    className="ai-shimmer inline-block h-4 w-40 rounded"
                    aria-label="AI is thinking..."
                  />
                </div>
              </div>
            ) : (
              <MessageBlock key={message.id} message={message} isStreaming={isStreaming} />
            ),
          )}

          <div ref={bottomRef} />
        </div>
      </div>

      <div className="px-4 py-4 bg-white border-t border-surface-border flex-shrink-0">
        <div className="max-w-3xl mx-auto">
          <div className="flex gap-2 items-end bg-surface-raised border border-surface-border rounded-xl p-2 focus-within:border-ai-accent focus-within:ring-2 focus-within:ring-ai-accent/20 transition-all">
            <textarea
              value={input}
              onChange={(event) => setInput(event.target.value)}
              onKeyDown={(event) => {
                if (event.key === "Enter" && !event.shiftKey) {
                  event.preventDefault();
                  void sendMessage(input);
                }
              }}
              placeholder="Search events, clubs, opportunities, or ask for advice."
              rows={1}
              aria-label="Message CampusCare AI"
              className="flex-1 resize-none bg-transparent text-sm text-gray-900 placeholder:text-gray-400 outline-none max-h-32 py-1 px-1"
            />
            {isStreaming ? (
              <Button variant="ghost" size="sm" onClick={stop} className="flex-shrink-0">
                Stop
              </Button>
            ) : (
              <Button
                variant="ai"
                size="sm"
                onClick={() => void sendMessage(input)}
                disabled={!input.trim() || busy}
                className="flex-shrink-0"
                aria-label="Send message"
              >
                <svg
                  className="w-4 h-4"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth={2}
                >
                  <line x1="22" y1="2" x2="11" y2="13" />
                  <polygon points="22,2 15,22 11,13 2,9 22,2" />
                </svg>
              </Button>
            )}
          </div>
          <p className="text-xs text-center text-gray-400 mt-2">
            Campus data is retrieved from CampusCare through its API. Generated text runs on
            this device via WebGPU and can be wrong — confirm details on the CampusCare
            pages.
          </p>
        </div>
      </div>
    </div>
  );
}

function MessageBlock({ message, isStreaming }: { message: Message; isStreaming: boolean }) {
  return (
    <div className="space-y-2">
      <AIChatMessage
        role={message.role}
        content={message.content}
        isStreaming={message.role === "assistant" && isStreaming}
      />
      {message.records && message.records.length > 0 && (
        <RecordList records={message.records} />
      )}
    </div>
  );
}

/** RecordList renders retrieved campus facts as links, not prose. */
function RecordList({ records }: { records: RetrievedRecord[] }) {
  return (
    <ul className="ml-11 space-y-1">
      {records.map((record) => (
        <li key={`${record.entity_type}-${record.entity_id}`}>
          <a
            href={record.url}
            className="flex items-center gap-2 text-sm text-gray-700 hover:text-ai-accent transition-colors"
          >
            <span className="w-1.5 h-1.5 rounded-full bg-ai-accent flex-shrink-0" />
            <span className="truncate">{record.title}</span>
            <span className="text-xs text-gray-400 flex-shrink-0">{record.entity_type}</span>
          </a>
        </li>
      ))}
    </ul>
  );
}

function SuggestionChip({
  label,
  onSelect,
  disabled,
}: {
  label: string;
  onSelect: () => void;
  disabled: boolean;
}) {
  return (
    <button
      onClick={onSelect}
      disabled={disabled}
      className={cn(
        "text-sm text-gray-700 bg-white border border-surface-border",
        "hover:border-ai-accent hover:text-ai-accent px-3 py-1.5 rounded-full",
        "transition-all duration-150 shadow-sm disabled:opacity-50 disabled:cursor-not-allowed",
      )}
    >
      {label}
    </button>
  );
}

function buildSystemPrompt(
  name: string | undefined,
  context: string,
  found: boolean,
): string {
  const parts = [SYSTEM_PROMPT];

  if (name) {
    parts.push(`The campus member you are talking to is ${name}.`);
  }

  if (found && context) {
    parts.push(context);
  } else {
    parts.push(
      "No campus records matched this question. Tell the member the information is not " +
        "available in CampusCare and suggest where to look instead of guessing.",
    );
  }

  return parts.join("\n\n");
}

function EngineBadge({ status }: { status: string }) {
  if (status === "ready") {
    return (
      <span className="text-xs bg-green-50 text-green-700 border border-green-200 px-2 py-1 rounded-full">
        Model ready
      </span>
    );
  }
  if (status === "loading") {
    return (
      <span className="text-xs bg-ai-light text-ai-accent border border-ai-accent/20 px-2 py-1 rounded-full">
        Loading model…
      </span>
    );
  }
  if (status === "error") {
    return (
      <span className="text-xs bg-red-50 text-red-700 border border-red-200 px-2 py-1 rounded-full">
        Model unavailable
      </span>
    );
  }
  return (
    <span className="text-xs bg-amber-50 text-amber-600 border border-amber-200 px-2 py-1 rounded-full">
      Search only
    </span>
  );
}