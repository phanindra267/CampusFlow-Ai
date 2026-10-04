"use client";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { useAuth } from "@/contexts/AuthContext";
import { Button } from "@/components/ui/Button";
import { APIError } from "@/lib/api";
import { cn } from "@/lib/utils";

export default function LoginPage() {
  const { login, isLoading } = useAuth();
  const router = useRouter();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    try {
      await login(email.trim(), password);
      router.replace("/dashboard");
    } catch (err) {
      if (err instanceof APIError) {
        setError(err.details ? `${err.message} — ${err.details}` : err.message);
      } else {
        setError("Something went wrong. Please try again.");
      }
    }
  }

  return (
    <div className="min-h-screen bg-surface-raised flex">
      {/* Left — brand panel */}
      <div className="hidden lg:flex lg:w-1/2 bg-gray-900 flex-col justify-between p-12 relative overflow-hidden">
        <div
          className="absolute inset-0 opacity-10"
          style={{
            backgroundImage:
              "radial-gradient(circle at 20% 50%, #F07C00 0%, transparent 50%), radial-gradient(circle at 80% 20%, #6366F1 0%, transparent 40%)",
          }}
        />
        <div className="relative z-10">
          <div className="flex items-center gap-3 mb-12">
            <div className="w-10 h-10 bg-lpu-primary rounded-xl flex items-center justify-center">
              <span className="text-white font-bold text-lg">C</span>
            </div>
            <div>
              <p className="text-white font-semibold text-lg leading-none">CampusCare AI</p>
              <p className="text-gray-400 text-xs">One campus, one community</p>
            </div>
          </div>
          <h1 className="text-4xl font-bold text-white leading-tight">
            Your intelligent<br />campus companion.
          </h1>
          <p className="mt-4 text-gray-400 text-lg max-w-sm">
            Groups, events, opportunities, services and discussions — unified in one AI-powered experience
            built for this campus.
          </p>
        </div>

        <div className="relative z-10 space-y-4">
          {[
            { icon: "💬", text: "Discussions that stay with your groups" },
            { icon: "🎉", text: "Events and registration in one tap" },
            { icon: "💼", text: "Opportunities matched to your interests" },
            { icon: "🏛️", text: "Bookable campus services" },
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
          <div className="flex items-center gap-2 mb-8">
            <div className="w-8 h-8 bg-lpu-primary rounded-lg flex items-center justify-center">
              <span className="text-white font-bold text-sm">C</span>
            </div>
            <p className="font-semibold text-gray-900">CampusCare AI</p>
          </div>

          <h2 className="text-2xl font-bold text-gray-900 mb-1">Sign in</h2>
          <p className="text-sm text-gray-500 mb-8">Use the email address registered with your campus account</p>

          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label htmlFor="email" className="block text-sm font-medium text-gray-700 mb-1.5">
                Email
              </label>
              <input
                id="email"
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
                autoComplete="email"
                placeholder="you@campus.edu"
                className={cn(
                  "w-full px-3 py-2.5 rounded-lg border text-sm transition-colors",
                  "focus:outline-none focus:ring-2 focus:ring-lpu-primary focus:border-transparent",
                  "border-surface-border bg-white text-gray-900 placeholder:text-gray-400",
                )}
              />
            </div>

            <div>
              <div className="flex items-center justify-between mb-1.5">
                <label htmlFor="password" className="block text-sm font-medium text-gray-700">
                  Password
                </label>
              </div>
              <input
                id="password"
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
                autoComplete="current-password"
                placeholder="••••••••"
                className={cn(
                  "w-full px-3 py-2.5 rounded-lg border text-sm transition-colors",
                  "focus:outline-none focus:ring-2 focus:ring-lpu-primary focus:border-transparent",
                  "border-surface-border bg-white text-gray-900 placeholder:text-gray-400",
                )}
              />
              <p className="mt-1.5 text-xs text-gray-500">
                For password resets, contact your campus administrator.
              </p>
            </div>

            {error && (
              <div className="p-3 bg-red-50 border border-red-200 rounded-lg text-sm text-red-700">{error}</div>
            )}

            <Button
              type="submit"
              variant="primary"
              size="lg"
              isLoading={isLoading}
              disabled={!email.trim() || !password}
              className="w-full mt-2"
            >
              {isLoading ? "Signing in…" : "Sign in to CampusCare"}
            </Button>
          </form>

          <p className="mt-6 text-xs text-center text-gray-400">
            Sessions use a short-lived access token and a rotating refresh token. Signing out clears both from
            this device.
          </p>
        </div>
      </div>
    </div>
  );
}