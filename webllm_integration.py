import os

BASE_DIR = r"c:\Users\phani\Downloads\CampusCare AI\frontend\src\app\(protected)\ai"

def write_file(path, content):
    full_path = os.path.join(BASE_DIR, path)
    os.makedirs(os.path.dirname(full_path), exist_ok=True)
    with open(full_path, "w", encoding="utf-8") as f:
        f.write(content.strip())
    print(f"Written {path}")

write_file("page.tsx", """
"use client";
import { AIChatMessage, SuggestedActions } from "@/components/ai/AIComponents";
import { Button } from "@/components/ui/Button";
import { cn } from "@/lib/utils";
import { useState, useRef, useEffect } from "react";
import { useAuth } from "@/contexts/AuthContext";
import { fetchAPI } from "@/lib/api";
import { CreateMLCEngine, MLCEngine, InitProgressReport } from "@mlc-ai/web-llm";

const DEMO_SUGGESTIONS = [
  "Check my attendance status",
  "Find hackathons this month",
  "Explain my career skill gaps",
  "What internships suit my profile?",
];

interface Message {
  id: string;
  role: "user" | "assistant" | "system";
  content: string;
  citations?: Array<{ id: string; source: string }>;
}

export default function AIPage() {
  const { user } = useAuth();
  const [messages, setMessages] = useState<Message[]>([]);
  const [input, setInput] = useState("");
  const [isTyping, setIsTyping] = useState(false);
  
  // WebLLM State
  const [engine, setEngine] = useState<MLCEngine | null>(null);
  const [loadingMsg, setLoadingMsg] = useState("Initializing WebLLM...");
  const [isEngineReady, setIsEngineReady] = useState(false);
  
  const bottomRef = useRef<HTMLDivElement>(null);

  // Initialize WebLLM on mount
  useEffect(() => {
    async function initLLM() {
      try {
        const initProgressCallback = (initProgress: InitProgressReport) => {
          setLoadingMsg(`Loading AI Model: ${Math.round(initProgress.progress * 100)}%`);
        };
        // Using a highly efficient quantized model for browser compatibility
        const selectedModel = "Llama-3.2-1B-Instruct-q4f32_1-MLC";
        const newEngine = await CreateMLCEngine(selectedModel, { initProgressCallback });
        setEngine(newEngine);
        setIsEngineReady(true);
        setLoadingMsg("AI Model Ready");
      } catch (err) {
        console.error("WebLLM initialization failed", err);
        setLoadingMsg("Failed to initialize local AI. Please ensure WebGPU is supported.");
      }
    }
    initLLM();
  }, []);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages, isTyping, loadingMsg]);

  async function sendMessage(text: string) {
    if (!text.trim() || !engine) return;
    
    const userMsg: Message = { id: Date.now().toString(), role: "user", content: text };
    setMessages(prev => [...prev, userMsg]);
    setInput("");
    setIsTyping(true);

    try {
      // 1. RAG Retrieval via Go Backend
      let contextString = "";
      let citations: Array<{id: string, source: string}> = [];
      
      try {
        // Fallback or demo context handling since DB might not have real documents yet
        contextString = `You are CampusCare AI, an assistant for Lovely Professional University.
        The user is asking about: ${text}.
        Answer based on the assumption you have real LPU data.`;
        
        // Example of a real call if the backend endpoint existed:
        // const contextResp = await fetchAPI(`/ai/context?q=${encodeURIComponent(text)}`);
        // contextString = contextResp.data.context;
      } catch (e) {
        console.warn("Could not fetch RAG context, using fallback");
      }

      // 2. Format messages for WebLLM
      const chatHistory = messages.map(m => ({ role: m.role, content: m.content }));
      
      const systemPrompt = {
        role: "system" as const,
        content: `You are CampusCare AI, the intelligent campus assistant for Lovely Professional University (LPU). 
        Always be concise, helpful, and polite. 
        Context information retrieved for this query: ${contextString}`
      };

      const requestMessages = [systemPrompt, ...chatHistory, { role: "user" as const, content: text }];

      // 3. Generate Response
      const reply = await engine.chat.completions.create({
        messages: requestMessages,
        temperature: 0.7,
      });

      const aiMsg: Message = {
        id: (Date.now() + 1).toString(),
        role: "assistant",
        content: reply.choices[0].message.content || "I couldn't generate a response.",
        citations: citations.length > 0 ? citations : undefined,
      };

      setMessages(prev => [...prev, aiMsg]);
    } catch (err) {
      console.error(err);
      setMessages(prev => [...prev, {
        id: Date.now().toString(),
        role: "assistant",
        content: "Sorry, I encountered an error generating the response."
      }]);
    } finally {
      setIsTyping(false);
    }
  }

  function handleSuggestion(s: string) { sendMessage(s); }

  const hour = new Date().getHours();
  const timeGreeting = hour < 12 ? "morning" : hour < 17 ? "afternoon" : "evening";

  return (
    <div className="flex flex-col h-full max-h-screen">
      <div className="px-4 py-4 bg-white border-b border-surface-border flex-shrink-0 flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="w-9 h-9 bg-ai-light border border-ai-accent/20 rounded-full flex items-center justify-center">
            <svg className="w-5 h-5 text-ai-accent" viewBox="0 0 24 24" fill="currentColor">
              <path d="M12 2l2.4 7.2H22l-6.2 4.5 2.4 7.2L12 17l-6.2 3.9 2.4-7.2L2 9.2h7.6L12 2z" />
            </svg>
          </div>
          <div>
            <p className="font-semibold text-gray-900 text-sm">CampusCare AI</p>
            <p className="text-xs text-gray-500">Powered by WebLLM — running locally on your device</p>
          </div>
        </div>
        {!isEngineReady && (
          <div className="text-xs font-medium text-amber-600 bg-amber-50 px-3 py-1.5 rounded-full border border-amber-200 animate-pulse">
            {loadingMsg}
          </div>
        )}
        {isEngineReady && (
          <div className="text-xs font-medium text-green-600 bg-green-50 px-3 py-1.5 rounded-full border border-green-200">
            Engine Ready (WebGPU)
          </div>
        )}
      </div>

      <div className="flex-1 overflow-y-auto px-4 py-6">
        <div className="max-w-3xl mx-auto space-y-4">
          {messages.length === 0 && (
            <div className="text-center py-8 animate-fade-in">
              <div className="w-16 h-16 bg-ai-light rounded-2xl flex items-center justify-center mx-auto mb-4">
                <svg className="w-8 h-8 text-ai-accent" viewBox="0 0 24 24" fill="currentColor">
                  <path d="M12 2l2.4 7.2H22l-6.2 4.5 2.4 7.2L12 17l-6.2 3.9 2.4-7.2L2 9.2h7.6L12 2z" />
                </svg>
              </div>
              <h2 className="text-xl font-bold text-gray-900">Good {timeGreeting}, {user?.name?.split(" ")[0]}.</h2>
              <p className="text-sm text-gray-500 mt-1 mb-6">How can I help with your LPU day?</p>

              <SuggestedActions actions={DEMO_SUGGESTIONS} onSelect={handleSuggestion} />
            </div>
          )}

          {messages.map(msg => (
            msg.role !== "system" && (
              <AIChatMessage
                key={msg.id}
                role={msg.role as "user" | "assistant"}
                content={msg.content}
                citations={msg.citations}
              />
            )
          ))}

          {isTyping && (
            <div className="flex gap-3">
              <div className="w-8 h-8 rounded-full bg-ai-light flex items-center justify-center">
                <svg className="w-4 h-4 text-ai-accent animate-spin" viewBox="0 0 24 24" fill="currentColor">
                  <path d="M12 2l2.4 7.2H22l-6.2 4.5 2.4 7.2L12 17l-6.2 3.9 2.4-7.2L2 9.2h7.6L12 2z" />
                </svg>
              </div>
              <div className="bg-white px-4 py-3 rounded-2xl rounded-tl-sm shadow-card border border-surface-border text-sm text-gray-500">
                Generating response locally...
              </div>
            </div>
          )}

          <div ref={bottomRef} />
        </div>
      </div>

      <div className="px-4 py-4 bg-white border-t border-surface-border flex-shrink-0">
        <div className="max-w-3xl mx-auto">
          <div className="flex gap-2 items-end bg-surface-raised border border-surface-border rounded-xl p-2 focus-within:border-ai-accent focus-within:ring-2 focus-within:ring-ai-accent/20 transition-all">
            <textarea
              value={input}
              onChange={e => setInput(e.target.value)}
              onKeyDown={e => { if (e.key === "Enter" && !e.shiftKey) { e.preventDefault(); sendMessage(input); } }}
              placeholder={isEngineReady ? "Ask about LPU academics, opportunities..." : "Waiting for model to load..."}
              disabled={!isEngineReady || isTyping}
              rows={1}
              className="flex-1 resize-none bg-transparent text-sm text-gray-900 outline-none py-1 px-1 disabled:opacity-50"
            />
            <Button
              variant="ai"
              size="sm"
              onClick={() => sendMessage(input)}
              disabled={!input.trim() || isTyping || !isEngineReady}
            >
              <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2}>
                <line x1="22" y1="2" x2="11" y2="13" />
                <polygon points="22,2 15,22 11,13 2,9 22,2" />
              </svg>
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}
""")
