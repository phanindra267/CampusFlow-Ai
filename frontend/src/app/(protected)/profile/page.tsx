"use client";
import { useAuth } from "@/contexts/AuthContext";
import { Button } from "@/components/ui/Button";
import { Badge } from "@/components/ui/Badge";
import { useState } from "react";

export default function ProfilePage() {
  const { user, logout } = useAuth();
  const [activeTab, setActiveTab] = useState<"personal" | "settings" | "privacy">("personal");

  if (!user) return null;

  return (
    <div className="max-w-4xl mx-auto px-4 py-6 space-y-6 animate-fade-in">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Profile</h1>
        <p className="text-sm text-gray-500">Manage your identity and preferences</p>
      </div>

      <div className="bg-white rounded-xl border border-surface-border p-6 flex flex-col md:flex-row items-start md:items-center gap-6">
        <div className="w-20 h-20 bg-gradient-to-br from-lpu-primary to-lpu-hover rounded-full flex items-center justify-center flex-shrink-0 shadow-lg">
          <span className="text-white text-3xl font-bold">{user.name[0]}</span>
        </div>
        <div className="flex-1">
          <h2 className="text-xl font-bold text-gray-900">{user.name}</h2>
          <p className="text-gray-500 mb-2">{user.email}</p>
          <div className="flex gap-2">
            <Badge variant="lpu">{user.role}</Badge>
            {user.program && <Badge variant="default">{user.program}</Badge>}
            {user.department && <Badge variant="default">{user.department}</Badge>}
          </div>
        </div>
      </div>

      {/* Tabs */}
      <div className="flex gap-1 border-b border-surface-border">
        {(["personal", "settings", "privacy"] as const).map(t => (
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
            <h3 className="font-semibold text-gray-900">Academic Details</h3>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <p className="text-xs text-gray-500">Registration Number</p>
                <p className="text-sm font-medium text-gray-900">12345678</p>
              </div>
              <div>
                <p className="text-xs text-gray-500">Batch</p>
                <p className="text-sm font-medium text-gray-900">2024-2028</p>
              </div>
            </div>
          </div>
        </div>
      )}

      {activeTab === "settings" && (
        <div className="space-y-4">
          <div className="bg-white rounded-xl border border-surface-border p-6 space-y-4">
            <h3 className="font-semibold text-gray-900">Notification Preferences</h3>
            <div className="space-y-3">
              {[
                "Academic Alerts (Attendance, Deadlines)",
                "Career & Placement Updates",
                "Campus Events & Clubs",
                "AI Assistant Insights"
              ].map(pref => (
                <div key={pref} className="flex items-center justify-between">
                  <span className="text-sm text-gray-700">{pref}</span>
                  <input type="checkbox" defaultChecked className="accent-lpu-primary w-4 h-4" />
                </div>
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
              CampusCare AI uses your academic and activity data to provide personalized insights.
            </p>
            <div className="space-y-3">
               <div className="flex items-center justify-between">
                  <span className="text-sm text-gray-700">Allow AI personalized recommendations</span>
                  <input type="checkbox" defaultChecked className="accent-lpu-primary w-4 h-4" />
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-sm text-gray-700">Make profile visible to recruiters</span>
                  <input type="checkbox" defaultChecked className="accent-lpu-primary w-4 h-4" />
                </div>
            </div>
          </div>
        </div>
      )}

      <div className="pt-4 border-t border-surface-border">
         <Button variant="destructive" onClick={logout}>Sign Out</Button>
      </div>
    </div>
  );
}
