"use client";
import { cn } from "@/lib/utils";
import { useState } from "react";

// AI Insight Block — Indigo accent, clearly AI-labelled
interface AIInsightProps {
  title?: string;
  content: string;
  sources?: string[];
  updatedAt?: string;
  className?: string;
  onDismiss?: () => void;
}

export function AIInsightCard({ title, content, sources, updatedAt, className, onDismiss }: AIInsightProps) {
  return (
    <div className={cn(
      "relative rounded-xl border-l-4 border-ai-accent bg-ai-light p-4 space-y-2 animate-fade-in",
      className
    )}>
      <div className="flex items-start justify-between gap-2">
        <div className="flex items-center gap-1.5">
          <AISparkle />
          <span className="text-xs font-semibold text-ai-accent uppercase tracking-wide">CampusCare AI</span>
        </div>
        {onDismiss && (
          <button onClick={onDismiss} className="text-gray-400 hover:text-gray-600 transition-colors" aria-label="Dismiss AI insight">
            <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2}>
              <path d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        )}
      </div>
      {title && <p className="text-sm font-semibold text-gray-800">{title}</p>}
      <p className="text-sm text-gray-700 leading-relaxed">{content}</p>
      {(sources || updatedAt) && (
        <div className="flex flex-wrap items-center gap-2 pt-1 text-xs text-gray-500">
          {sources?.map((s, i) => (
            <span key={i} className="flex items-center gap-1">
              <span className="w-1.5 h-1.5 rounded-full bg-ai-accent inline-block" />
              {s}
            </span>
          ))}
          {updatedAt && <span className="ml-auto">Updated {updatedAt}</span>}
        </div>
      )}
    </div>
  );
}

// AI Chat Message
interface AIChatMessageProps {
  role: "user" | "assistant";
  content: string;
  citations?: Array<{ id: string; source: string }>;
  isStreaming?: boolean;
}

export function AIChatMessage({ role, content, citations, isStreaming }: AIChatMessageProps) {
  const [expandedCitation, setExpandedCitation] = useState<string | null>(null);

  if (role === "user") {
    return (
      <div className="flex justify-end">
        <div className="max-w-[80%] bg-lpu-primary text-white rounded-2xl rounded-tr-sm px-4 py-3 text-sm">
          {content}
        </div>
      </div>
    );
  }

  return (
    <div className="flex gap-3 max-w-[90%]">
      <div className="flex-shrink-0 w-8 h-8 rounded-full bg-ai-light border border-ai-accent/20 flex items-center justify-center">
        <AISparkle className="w-4 h-4" />
      </div>
      <div className="space-y-2">
        <div className="bg-white rounded-2xl rounded-tl-sm px-4 py-3 text-sm text-gray-800 shadow-card border border-surface-border">
          {isStreaming ? (
            <span className="ai-shimmer inline-block h-4 w-40 rounded" aria-label="AI is thinking..." />
          ) : (
            <p className="leading-relaxed">{content}</p>
          )}
        </div>
        {citations && citations.length > 0 && (
          <div className="flex flex-wrap gap-1.5 px-1">
            {citations.map((c) => (
              <button
                key={c.id}
                onClick={() => setExpandedCitation(expandedCitation === c.id ? null : c.id)}
                className="text-xs text-ai-accent bg-ai-light px-2 py-0.5 rounded-full hover:bg-indigo-100 transition-colors"
              >
                [{c.id}] {c.source}
              </button>
            ))}
          </div>
        )}
        {expandedCitation && (
          <div className="bg-ai-light border border-ai-accent/20 rounded-lg p-3 text-xs text-gray-700 animate-fade-in">
            <p className="font-medium text-ai-accent mb-1">Source Details</p>
            <p>{citations?.find(c => c.id === expandedCitation)?.source}</p>
            <p className="text-gray-400 mt-1">LPU Knowledge Base — This response is grounded in verified campus data.</p>
          </div>
        )}
      </div>
    </div>
  );
}

// Suggested actions for AI
interface SuggestedActionsProps {
  actions: string[];
  onSelect: (action: string) => void;
}

export function SuggestedActions({ actions, onSelect }: SuggestedActionsProps) {
  return (
    <div className="flex flex-wrap gap-2">
      {actions.map((action, i) => (
        <button
          key={i}
          onClick={() => onSelect(action)}
          className="text-sm text-gray-700 bg-white border border-surface-border hover:border-ai-accent hover:text-ai-accent px-3 py-1.5 rounded-full transition-all duration-150 shadow-sm"
        >
          {action}
        </button>
      ))}
    </div>
  );
}

// Sparkle icon for AI
function AISparkle({ className }: { className?: string }) {
  return (
    <svg className={cn("w-3.5 h-3.5 text-ai-accent", className)} viewBox="0 0 24 24" fill="currentColor">
      <path d="M12 2l2.4 7.2H22l-6.2 4.5 2.4 7.2L12 17l-6.2 3.9 2.4-7.2L2 9.2h7.6L12 2z" />
    </svg>
  );
}
