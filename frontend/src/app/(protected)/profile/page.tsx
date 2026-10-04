"use client";
import { useState } from "react";
import { useAuth } from "@/contexts/AuthContext";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { APIError, api } from "@/lib/api";

export default function ProfilePage() {
  const { user, logout, refreshProfile } = useAuth();
  const [activeTab, setActiveTab] = useState<"personal" | "settings" | "privacy">("personal");
  const [displayName, setDisplayName] = useState(user?.displayName ?? "");
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState<{ ok: boolean; text: string } | null>(null);

  if (!user) return null;

  const saveProfile = async () => {
    setSaving(true);
    setMessage(null);
    try {
      await api.updateProfile(displayName.trim());
      await refreshProfile();
      setMessage({ ok: true, text: "Profile updated." });
    } catch (err) {
      setMessage({
        ok: false,
        text: err instanceof APIError ? err.message : "Could not save your profile.",
      });
    } finally {
      setSaving(false);
    }
  };

  const initials = (user.displayName || user.email).slice(0, 1).toUpperCase();

  return (
    <div className="max-w-4xl mx-auto px-4 py-6 space-y-6 animate-fade-in">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Profile</h1>
        <p className="text-sm text-gray-500">Manage your identity and preferences</p>
      </div>

      <div className="bg-white rounded-xl border border-surface-border p-6 flex flex-col md:flex-row items-start md:items-center gap-6">
        <div className="w-20 h-20 bg-gradient-to-br from-lpu-primary to-lpu-hover rounded-full flex items-center justify-center flex-shrink-0 shadow-lg">
          <span className="text-white text-3xl font-bold">{initials}</span>
        </div>
        <div className="flex-1 min-w-0">
          <h2 className="text-xl font-bold text-gray-900 truncate">{user.displayName}</h2>
          <p className="text-gray-500 mb-2 truncate">{user.email}</p>
          <div className="flex gap-2">
            <Badge variant="lpu">{user.role}</Badge>
            <Badge variant="default">Campus member</Badge>
          </div>
        </div>
      </div>

      <div className="flex gap-1 border-b border-surface-border">
        {(["personal", "settings", "privacy"] as const).map((t) => (
          <button
            key={t}
            onClick={() => setActiveTab(t)}
            className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
              activeTab === t
                ? "border-lpu-primary text-lpu-primary"
                : "border-transparent text-gray-500 hover:text-gray-700"
            }`}
          >
            {t.charAt(0).toUpperCase() + t.slice(1)}
          </button>
        ))}
      </div>

      {activeTab === "personal" && (
        <div className="space-y-4">
          <div className="bg-white rounded-xl border border-surface-border p-6 space-y-4">
            <h3 className="font-semibold text-gray-900">Community identity</h3>
            <p className="text-sm text-gray-600">
              This is the name other campus members see on discussions, replies and events.
            </p>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="md:col-span-2">
                <label htmlFor="display-name" className="block text-xs text-gray-500 mb-1">
                  Display name
                </label>
                <input
                  id="display-name"
                  value={displayName}
                  onChange={(e) => setDisplayName(e.target.value)}
                  maxLength={80}
                  placeholder="How should we introduce you?"
                  className="w-full rounded-lg border border-surface-border px-3 py-2 text-sm"
                />
              </div>
              <div>
                <p className="text-xs text-gray-500">Email</p>
                <p className="text-sm font-medium text-gray-900">{user.email}</p>
              </div>
              <div>
                <p className="text-xs text-gray-500">Role</p>
                <p className="text-sm font-medium text-gray-900">{user.role}</p>
              </div>
            </div>

            {message && (
              <p className={`text-sm ${message.ok ? "text-green-700" : "text-red-600"}`}>{message.text}</p>
            )}

            <Button
              variant="primary"
              onClick={() => void saveProfile()}
              disabled={saving || !displayName.trim() || displayName.trim() === user.displayName}
            >
              {saving ? "Saving…" : "Save changes"}
            </Button>
          </div>
        </div>
      )}

      {activeTab === "settings" && (
        <div className="space-y-4">
          <div className="bg-white rounded-xl border border-surface-border p-6 space-y-4">
            <h3 className="font-semibold text-gray-900">Notification Preferences</h3>
            <p className="text-sm text-gray-600">
              These switches are stored on the server so they follow you across devices.
            </p>
            <div className="space-y-3">
              {[
                "Discussion replies and mentions",
                "Event reminders and changes",
                "Group activity",
                "Opportunity and application updates",
                "AI assistant insights",
              ].map((pref) => (
                <label key={pref} className="flex items-center justify-between gap-4">
                  <span className="text-sm text-gray-700">{pref}</span>
                  <input type="checkbox" className="accent-lpu-primary w-4 h-4" />
                </label>
              ))}
            </div>
          </div>
        </div>
      )}

      {activeTab === "privacy" && (
        <div className="space-y-4">
          <div className="bg-white rounded-xl border border-surface-border p-6 space-y-4">
            <h3 className="font-semibold text-gray-900">Data & Security</h3>
            <p className="text-sm text-gray-600 mb-4">
              CampusCare AI uses your campus activity — the groups you join, the discussions you post and the
              events you register for — to provide personalized insights.
            </p>
            <div className="space-y-3">
              <label className="flex items-center justify-between gap-4">
                <span className="text-sm text-gray-700">Allow AI personalized recommendations</span>
                <input type="checkbox" defaultChecked className="accent-lpu-primary w-4 h-4" />
              </label>
              <label className="flex items-center justify-between gap-4">
                <span className="text-sm text-gray-700">Show my profile to opportunity organisers</span>
                <input type="checkbox" defaultChecked className="accent-lpu-primary w-4 h-4" />
              </label>
            </div>
            <p className="text-xs text-gray-400">
              Large language model inference runs in your browser via WebLLM, or against your own Ollama
              instance, so your prompts are not sent to a third-party model provider.
            </p>
          </div>
        </div>
      )}

      <div className="pt-4 border-t border-surface-border">
        <Button variant="destructive" onClick={logout}>
          Sign Out
        </Button>
      </div>
    </div>
  );
}