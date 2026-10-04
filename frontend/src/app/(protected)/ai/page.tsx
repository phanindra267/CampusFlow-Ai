"use client";
import { AIChatMessage, SuggestedActions } from "@/components/ai/AIComponents";
import { Button } from "@/components/ui/Button";
import { cn } from "@/lib/utils";
import { useState, useRef, useEffect } from "react";
import { useAuth } from "@/contexts/AuthContext";

const DEMO_SUGGESTIONS = [
  "Check my attendance status",
  "Find hackathons this month",
  "Explain my career skill gaps",
  "What internships suit my profile?",
  "Summarize my academic progress",
  "Find LPU research opportunities",
];

const CANNED_RESPONSES: Record<string, { text: string; citations?: Array<{id:string; source:string}> }> = {
  "check my attendance status": {
    text: "Your overall attendance this semester is 85%. Here's the breakdown:\n\n• CSE316 Machine Learning: 85% ✅\n• CSE303 Operating Systems: 78% ⚠️ (Warning — attend next 3 classes)\n• CSE412 Cloud Computing: 91% ✅\n• CSE499 Capstone: 100% ✅\n• HUM201 Technical Writing: 72% 🔴 (Critical)\n\nYou need to prioritize attending HUM201 immediately to avoid debarment from the exam.",
    citations: [
      { id: "1", source: "Attendance Records — Semester 5" },
      { id: "2", source: "LPU Academic Regulations 2026" },
    ],
  },
  "find hackathons this month": {
    text: "I found 2 hackathons this month that match your profile:\n\n**1. Smart India Hackathon 2027 — LPU Chapter** (96% match)\n• Team of 6 · Registration closes Oct 7\n• Location: Innovation Lab, Block 40\n• Your ML skills are a strong fit for the AI track\n\n**2. National Coding Championship Qualifier** (82% match)\n• Online · Registration closes Oct 12\n• Focuses on competitive programming\n\nShould I help you register for SIH 2027?",
    citations: [{ id: "1", source: "LPU Event Calendar — October 2026" }],
  },
  "explain my career skill gaps": {
    text: "Based on your profile and LPU placement data, here are your priority gaps for SDE roles:\n\n**Critical Gaps:**\n• System Design — 20% proficiency · Required by 87% of top recruiters\n• DSA (Advanced) — 60% · Need 200+ LeetCode problems, focus on Trees/Graphs\n\n**Strengths to highlight:**\n• Python (85%) — Strong evidence from CSE316 + 3 projects\n• SQL (68%) — Verified through DBMS coursework\n\n**Recommended next steps:**\n1. Take CSE Fundamentals of System Design (LPU elective, Sem 6)\n2. Practice 15 problems/week on CP Club platform",
    citations: [
      { id: "1", source: "Skill Profile — LPU Career Portal" },
      { id: "2", source: "Placement Analytics 2025–26" },
    ],
  },
};

function getBotResponse(message: string) {
  const key = message.toLowerCase().trim();
  const match = Object.keys(CANNED_RESPONSES).find(k => key.includes(k));
  if (match) return CANNED_RESPONSES[match];
  return {
    text: "I understand you're asking about \"" + message + "\". To give you an accurate, grounded answer I need to be connected to the live LPU Knowledge Base (coming soon). For now, I can answer questions about attendance, hackathons, skill gaps, and internships.",
    citations: [],
  };
}

interface Message {
  id: string;
  role: "user" | "assistant";
  content: string;
  citations?: Array<{ id: string; source: string }>;
}

export default function AIPage() {
  const { user } = useAuth();
  const [messages, setMessages] = useState<Message[]>([]);
  const [input, setInput] = useState("");
  const [isTyping, setIsTyping] = useState(false);
  const [started, setStarted] = useState(false);
  const bottomRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages, isTyping]);

  async function sendMessage(text: string) {
    if (!text.trim()) return;
    setStarted(true);
    const userMsg: Message = { id: Date.now().toString(), role: "user", content: text };
    setMessages(prev => [...prev, userMsg]);
    setInput("");
    setIsTyping(true);

    // Simulate AI response delay
    await new Promise(r => setTimeout(r, 1200));
    const resp = getBotResponse(text);
    const aiMsg: Message = {
      id: (Date.now() + 1).toString(),
      role: "assistant",
      content: resp.text,
      citations: resp.citations,
    };
    setIsTyping(false);
    setMessages(prev => [...prev, aiMsg]);
  }

  function handleSuggestion(s: string) { sendMessage(s); }

  const hour = new Date().getHours();
  const timeGreeting = hour < 12 ? "morning" : hour < 17 ? "afternoon" : "evening";

  return (
    <div className="flex flex-col h-full max-h-screen">
      {/* Header */}
      <div className="px-4 py-4 bg-white border-b border-surface-border flex-shrink-0">
        <div className="max-w-3xl mx-auto flex items-center gap-3">
          <div className="w-9 h-9 bg-ai-light border border-ai-accent/20 rounded-full flex items-center justify-center flex-shrink-0">
            <svg className="w-5 h-5 text-ai-accent" viewBox="0 0 24 24" fill="currentColor">
              <path d="M12 2l2.4 7.2H22l-6.2 4.5 2.4 7.2L12 17l-6.2 3.9 2.4-7.2L2 9.2h7.6L12 2z" />
            </svg>
          </div>
          <div>
            <p className="font-semibold text-gray-900 text-sm">CampusCare AI</p>
            <p className="text-xs text-gray-400">LPU Campus Intelligence · Knowledge-grounded responses</p>
          </div>
          <div className="ml-auto">
            <span className="text-xs bg-amber-50 text-amber-600 border border-amber-200 px-2 py-1 rounded-full">
              Demo mode — limited to sample queries
            </span>
          </div>
        </div>
      </div>

      {/* Messages */}
      <div className="flex-1 overflow-y-auto px-4 py-6">
        <div className="max-w-3xl mx-auto space-y-4">
          {!started && (
            <div className="text-center py-8 animate-fade-in">
              <div className="w-16 h-16 bg-ai-light rounded-2xl flex items-center justify-center mx-auto mb-4">
                <svg className="w-8 h-8 text-ai-accent" viewBox="0 0 24 24" fill="currentColor">
                  <path d="M12 2l2.4 7.2H22l-6.2 4.5 2.4 7.2L12 17l-6.2 3.9 2.4-7.2L2 9.2h7.6L12 2z" />
                </svg>
              </div>
              <h2 className="text-xl font-bold text-gray-900">Good {timeGreeting}, {user?.name?.split(" ")[0]}.</h2>
              <p className="text-sm text-gray-500 mt-1">How can I help with your LPU day?</p>

              <div className="mt-6">
                <p className="text-xs text-gray-400 mb-3 uppercase tracking-wide font-medium">Try asking</p>
                <SuggestedActions actions={DEMO_SUGGESTIONS} onSelect={handleSuggestion} />
              </div>
            </div>
          )}

          {messages.map(msg => (
            <AIChatMessage
              key={msg.id}
              role={msg.role}
              content={msg.content}
              citations={msg.citations}
            />
          ))}

          {isTyping && (
            <div className="flex gap-3">
              <div className="flex-shrink-0 w-8 h-8 rounded-full bg-ai-light border border-ai-accent/20 flex items-center justify-center">
                <svg className="w-4 h-4 text-ai-accent" viewBox="0 0 24 24" fill="currentColor">
                  <path d="M12 2l2.4 7.2H22l-6.2 4.5 2.4 7.2L12 17l-6.2 3.9 2.4-7.2L2 9.2h7.6L12 2z" />
                </svg>
              </div>
              <div className="bg-white rounded-2xl rounded-tl-sm px-4 py-3 shadow-card border border-surface-border">
                <div className="flex gap-1 items-center h-5">
                  {[0, 1, 2].map(i => (
                    <div
                      key={i}
                      className="w-2 h-2 bg-gray-300 rounded-full animate-bounce"
                      style={{ animationDelay: `${i * 0.15}s` }}
                    />
                  ))}
                </div>
              </div>
            </div>
          )}

          {/* Suggested follow-ups after last AI message */}
          {started && !isTyping && messages.length > 0 && (
            <div className="pl-11">
              <SuggestedActions
                actions={["Check my attendance", "Find internships for me", "Explain my skill gaps"]}
                onSelect={handleSuggestion}
              />
            </div>
          )}

          <div ref={bottomRef} />
        </div>
      </div>

      {/* Input */}
      <div className="px-4 py-4 bg-white border-t border-surface-border flex-shrink-0">
        <div className="max-w-3xl mx-auto">
          <div className="flex gap-2 items-end bg-surface-raised border border-surface-border rounded-xl p-2 focus-within:border-ai-accent focus-within:ring-2 focus-within:ring-ai-accent/20 transition-all">
            <textarea
              value={input}
              onChange={e => setInput(e.target.value)}
              onKeyDown={e => { if (e.key === "Enter" && !e.shiftKey) { e.preventDefault(); sendMessage(input); } }}
              placeholder="Ask about your LPU academics, events, opportunities..."
              rows={1}
              className="flex-1 resize-none bg-transparent text-sm text-gray-900 placeholder:text-gray-400 outline-none max-h-32 py-1 px-1"
            />
            <Button
              variant="ai"
              size="sm"
              onClick={() => sendMessage(input)}
              disabled={!input.trim() || isTyping}
              className="flex-shrink-0"
            >
              <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2}>
                <line x1="22" y1="2" x2="11" y2="13" />
                <polygon points="22,2 15,22 11,13 2,9 22,2" />
              </svg>
            </Button>
          </div>
          <p className="text-xs text-center text-gray-400 mt-2">
            Responses are grounded in LPU campus data. Always verify critical academic decisions with your faculty.
          </p>
        </div>
      </div>
    </div>
  );
}
