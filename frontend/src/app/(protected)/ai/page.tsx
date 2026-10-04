"use client";

import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { AIChatMessage } from "@/components/ai/AIComponents";
import { Button } from "@/components/ui/Button";
import { useAuth } from "@/contexts/AuthContext";
import { useLLM } from "@/hooks/useLLM";
import type { ChatMessage } from "@/lib/llm";
import { cn } from "@/lib/utils";

const SUGGESTIONS = [
  "What groups and clubs are running on this campus right now?",
  "How should I prepare for a technical internship interview?",
  "Explain the difference between a project and an internship on a resume.",
  "How can I get more out of my first hackathon?",
  "Help me draft a two-week plan to learn distributed systems.",
];

const SYSTEM_PROMPT = [
  "You are CampusCare AI, the campus assistant for the CampusCare community platform.",
  "Answer questions about campus life: clubs and groups, events, opportunities, internships,",
  "open source and research collaboration, campus services, and career development.",
  "Be concise and practical, and prefer concrete next steps.",
  "You do not have access to live records, so never invent an event, deadline, seat count,",
  "organiser name or placement statistic.",
  "If you need those details, say which CampusCare page the member should check.",
].join(" ");

type Message = {
  id: string;
  role: "user" | "assistant";
  content: string;
};

const WELCOME =
  "Hi! I'm CampusCare AI. Ask me about groups, events, opportunities, internships or career preparation.";

let messageCounter = 0;
function nextId(): string {
  messageCounter += 1;
  return `${Date.now().toString(36)}-${messageCounter}`;
}

export default function AIPage() {
  const { user } = useAuth();
  const { status, providerName, model, progress, error, isStreaming, initialize, generate, stop } =
    useLLM();

  const [messages, setMessages] = useState<Message[]>([]);
  const [input, setInput] = useState("");
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

  const sendMessage = useCallback(
    async (raw: string) => {
      const text = raw.trim();
      if (!text || isStreaming) return;

      const ready = (await initialize()) === true;
      if (!ready) return;

      const history = messagesRef.current;
      const userMessage: Message = { id: nextId(), role: "user", content: text };
      const assistantId = nextId();

      setMessages([...history, userMessage, { id: assistantId, role: "assistant", content: "" }]);
      setInput("");

      const transcript: ChatMessage[] = [
        { role: "system", content: buildSystemPrompt(user?.displayName) },
        ...history
          .filter((message) => message.content.trim().length > 0)
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
    [generate, initialize, isStreaming, user?.displayName],
  );

  const busy = isStreaming || status === "loading";

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
                ? `${providerName === "ollama" ? "Ollama" : "WebLLM"} · ${model}`
                : "Campus Intelligence · on-device language model"}
            </p>
          </div>
          <div className="ml-auto">
            <ProviderBadge status={status} />
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
              <p className="text-sm text-gray-500 mt-1">{WELCOME}</p>

              {status === "idle" && (
                <div className="mt-6 flex flex-col items-center gap-3">
                  <Button variant="ai" size="sm" onClick={() => void initialize()}>
                    Load the language model
                  </Button>
                  <p className="text-xs text-gray-400 max-w-md">
                    The first run downloads the model. Set NEXT_PUBLIC_LLM_PROVIDER=ollama to use a local
                    Ollama server instead.
                  </p>
                </div>
              )}

              <div className="mt-6">
                <p className="text-xs text-gray-400 mb-3 uppercase tracking-wide font-medium">
                  Try asking
                </p>
                <div className="flex flex-wrap gap-2 justify-center">
                  {SUGGESTIONS.map((suggestion) => (
                    <button
                      key={suggestion}
                      onClick={() => void sendMessage(suggestion)}
                      disabled={busy}
                      className={cn(
                        "text-sm text-gray-700 bg-white border border-surface-border",
                        "hover:border-ai-accent hover:text-ai-accent px-3 py-1.5 rounded-full",
                        "transition-all duration-150 shadow-sm disabled:opacity-50 disabled:cursor-not-allowed",
                      )}
                    >
                      {suggestion}
                    </button>
                  ))}
                </div>
              </div>
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
              <AIChatMessage
                key={message.id}
                role={message.role}
                content={message.content}
                isStreaming={message.role === "assistant" && isStreaming}
              />
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
              placeholder="Ask about events, groups, opportunities, or career."
              rows={1}
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
            Responses are generated on {providerName === "ollama" ? "your local Ollama server" : "this device"}.
            Always confirm event details and deadlines on the CampusCare pages themselves.
          </p>
        </div>
      </div>
    </div>
  );
}

function buildSystemPrompt(name: string | undefined): string {
  if (!name) return SYSTEM_PROMPT;
  return `${SYSTEM_PROMPT}\nThe campus member you are talking to is ${name}.`;
}

function ProviderBadge({ status }: { status: string }) {
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
      Model not loaded
    </span>
  );
}
