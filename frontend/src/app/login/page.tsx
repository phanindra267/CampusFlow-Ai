"use client";
import { useAuth } from "@/contexts/AuthContext";
import { Button } from "@/components/ui/Button";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { cn } from "@/lib/utils";

export default function LoginPage() {
  const { login, isLoading } = useAuth();
  const router = useRouter();
  const [email, setEmail] = useState("student@lpu.in");
  const [password, setPassword] = useState("demo");
  const [error, setError] = useState("");

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    try {
      await login(email, password);
      router.replace("/dashboard");
    } catch (err: any) {
      setError(err.message);
    }
  }

  return (
    <div className="min-h-screen bg-surface-raised flex">
      {/* Left — brand panel */}
      <div className="hidden lg:flex lg:w-1/2 bg-gray-900 flex-col justify-between p-12 relative overflow-hidden">
        {/* Pattern */}
        <div className="absolute inset-0 opacity-10"
          style={{ backgroundImage: "radial-gradient(circle at 20% 50%, #F07C00 0%, transparent 50%), radial-gradient(circle at 80% 20%, #6366F1 0%, transparent 40%)" }}
        />
        <div className="relative z-10">
          <div className="flex items-center gap-3 mb-12">
            <div className="w-10 h-10 bg-lpu-primary rounded-xl flex items-center justify-center">
              <span className="text-white font-bold text-lg">C</span>
            </div>
            <div>
              <p className="text-white font-semibold text-lg leading-none">CampusCare AI</p>
              <p className="text-gray-400 text-xs">Lovely Professional University</p>
            </div>
          </div>
          <h1 className="text-4xl font-bold text-white leading-tight">
            Your intelligent<br />campus companion.
          </h1>
          <p className="mt-4 text-gray-400 text-lg max-w-sm">
            Academics, opportunities, events, research, and career — unified in one AI-powered experience built for LPU.
          </p>
        </div>

        {/* Feature highlights */}
        <div className="relative z-10 space-y-4">
          {[
            { icon: "✨", text: "AI-powered academic insights" },
            { icon: "🎯", text: "Personalized opportunity matching" },
            { icon: "🏛️", text: "Smart campus services" },
            { icon: "📊", text: "Career readiness intelligence" },
          ].map((f, i) => (
            <div key={i} className="flex items-center gap-3 text-gray-300">
              <span className="text-xl">{f.icon}</span>
              <span className="text-sm">{f.text}</span>
            </div>
          ))}
        </div>
      </div>

      {/* Right — login form */}
      <div className="flex-1 flex items-center justify-center px-6 py-12">
        <div className="w-full max-w-md animate-fade-in">
          {/* Mobile logo */}
          <div className="flex items-center gap-2 mb-8 lg:hidden">
            <div className="w-8 h-8 bg-lpu-primary rounded-lg flex items-center justify-center">
              <span className="text-white font-bold text-sm">C</span>
            </div>
            <p className="font-semibold text-gray-900">CampusCare AI — LPU</p>
          </div>

          <h2 className="text-2xl font-bold text-gray-900 mb-1">Sign in</h2>
          <p className="text-sm text-gray-500 mb-8">Use your LPU portal credentials</p>

          {/* Demo hint */}
          <div className="mb-6 p-3 bg-ai-light border border-ai-accent/20 rounded-lg text-sm text-gray-700">
            <p className="font-medium text-ai-accent mb-1">✨ Demo Mode</p>
            <p>Try: <code className="bg-white px-1 rounded">student@lpu.in</code>, <code className="bg-white px-1 rounded">faculty@lpu.in</code>, or <code className="bg-white px-1 rounded">admin@lpu.in</code></p>
          </div>

          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1.5">
                LPU Email
              </label>
              <input
                type="email"
                value={email}
                onChange={e => setEmail(e.target.value)}
                required
                placeholder="rollno@lpu.in"
                className={cn(
                  "w-full px-3 py-2.5 rounded-lg border text-sm transition-colors",
                  "focus:outline-none focus:ring-2 focus:ring-lpu-primary focus:border-transparent",
                  "border-surface-border bg-white text-gray-900 placeholder:text-gray-400"
                )}
              />
            </div>

            <div>
              <div className="flex items-center justify-between mb-1.5">
                <label className="block text-sm font-medium text-gray-700">Password</label>
                <button type="button" className="text-xs text-lpu-primary hover:underline">Forgot password?</button>
              </div>
              <input
                type="password"
                value={password}
                onChange={e => setPassword(e.target.value)}
                required
                placeholder="••••••••"
                className={cn(
                  "w-full px-3 py-2.5 rounded-lg border text-sm transition-colors",
                  "focus:outline-none focus:ring-2 focus:ring-lpu-primary focus:border-transparent",
                  "border-surface-border bg-white text-gray-900 placeholder:text-gray-400"
                )}
              />
            </div>

            {error && (
              <div className="p-3 bg-red-50 border border-red-200 rounded-lg text-sm text-red-700">
                {error}
              </div>
            )}

            <Button type="submit" variant="primary" size="lg" isLoading={isLoading} className="w-full mt-2">
              {isLoading ? "Signing in..." : "Sign in to CampusCare"}
            </Button>
          </form>

          <p className="mt-6 text-xs text-center text-gray-400">
            By signing in, you agree to LPU's data usage and privacy policies.
          </p>
        </div>
      </div>
    </div>
  );
}
