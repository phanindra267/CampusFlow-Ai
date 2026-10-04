"use client";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { AIInsightCard } from "@/components/ai/AIComponents";
import { cn, formatRelativeDate } from "@/lib/utils";
import { useState } from "react";

const EVENTS = [
  {
    id: 1, type: "hackathon", title: "Smart India Hackathon 2027 — LPU Chapter",
    org: "GDSC LPU", date: new Date(Date.now() + 3 * 86400000), location: "Innovation Lab, Block 40",
    eligible: true, registered: false, tags: ["Coding", "Innovation"], match: 96,
    description: "Build solutions for real government problem statements. Teams of 6.",
  },
  {
    id: 2, type: "workshop", title: "AWS Cloud Practitioner Boot Camp",
    org: "AWS Campus Club", date: new Date(Date.now() + 5 * 86400000), location: "Block 32, Lab 3",
    eligible: true, registered: true, tags: ["Cloud", "AWS"], match: 89,
    description: "2-day intensive prep for the AWS Certified Cloud Practitioner exam.",
  },
  {
    id: 3, type: "seminar", title: "Industry Talk: AI in Healthcare",
    org: "School of CSE", date: new Date(Date.now() + 2 * 86400000), location: "Auditorium, Block 1",
    eligible: true, registered: false, tags: ["AI", "Healthcare"], match: 72,
    description: "Industry experts discuss real-world AI deployment challenges in healthcare.",
  },
  {
    id: 4, type: "competition", title: "National Coding Championship — Qualifier",
    org: "CP Club LPU", date: new Date(Date.now() + 8 * 86400000), location: "Online",
    eligible: true, registered: false, tags: ["Competitive", "Coding"], match: 82,
    description: "Top 100 qualify for the national finals. Registration closes in 3 days.",
  },
];

const CLUBS = [
  { id: 1, name: "Google Developer Student Club", members: 420, activity: "High", tags: ["Tech", "Google"], joined: true },
  { id: 2, name: "Competitive Programming Club", members: 280, activity: "High", tags: ["Algorithms", "Coding"], joined: false, match: 94 },
  { id: 3, name: "Robotics & Automation Society", members: 150, activity: "Medium", tags: ["Robotics", "IoT"], joined: false, match: 78 },
  { id: 4, name: "Entrepreneurship Cell (E-Cell)", members: 340, activity: "High", tags: ["Startup", "Business"], joined: false, match: 65 },
];

type FilterType = "all" | "hackathon" | "workshop" | "seminar" | "competition";

const typeColors: Record<string, string> = {
  hackathon:   "lpu",
  workshop:    "info",
  seminar:     "default",
  competition: "success",
};

export default function DiscoverPage() {
  const [section, setSection] = useState<"events" | "clubs" | "opportunities">("events");
  const [filter, setFilter] = useState<FilterType>("all");
  const [joinedIds, setJoinedIds] = useState<Set<number>>(new Set([1]));

  const filteredEvents = EVENTS.filter(e => filter === "all" || e.type === filter);

  return (
    <div className="max-w-5xl mx-auto px-4 py-6 space-y-6 animate-fade-in">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Discover</h1>
        <p className="text-sm text-gray-500">LPU events, clubs, and opportunities tailored for you</p>
      </div>

      <AIInsightCard
        title="3 high-match events this week"
        content="Smart India Hackathon (96% match) registration closes in 3 days. Your ML skills make you a strong candidate for the AI in Healthcare seminar. 2 events conflict — want me to help you prioritize?"
        sources={["Academic Profile", "Skill Tags", "Event Calendar"]}
        updatedAt="Just now"
      />

      {/* Section tabs */}
      <div className="flex gap-1 bg-surface-raised border border-surface-border rounded-lg p-1 w-fit">
        {(["events", "clubs", "opportunities"] as const).map(s => (
          <button key={s} onClick={() => setSection(s)}
            className={cn("px-4 py-1.5 rounded-md text-sm font-medium transition-all",
              section === s ? "bg-white text-gray-900 shadow-sm" : "text-gray-500 hover:text-gray-700"
            )}>
            {s.charAt(0).toUpperCase() + s.slice(1)}
          </button>
        ))}
      </div>

      {/* EVENTS */}
      {section === "events" && (
        <div className="space-y-4">
          {/* Filters */}
          <div className="flex items-center gap-2 flex-wrap">
            {(["all", "hackathon", "workshop", "seminar", "competition"] as FilterType[]).map(f => (
              <button key={f} onClick={() => setFilter(f)}
                className={cn("px-3 py-1.5 rounded-full text-xs font-medium border transition-all",
                  filter === f ? "bg-lpu-primary text-white border-lpu-primary" : "bg-white text-gray-600 border-surface-border hover:border-gray-300"
                )}>
                {f === "all" ? "All Types" : f.charAt(0).toUpperCase() + f.slice(1) + "s"}
              </button>
            ))}
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {filteredEvents.map(event => (
              <div key={event.id} className="bg-white rounded-xl border border-surface-border p-5 hover:shadow-card transition-shadow">
                <div className="flex items-start justify-between gap-2 mb-3">
                  <Badge variant={typeColors[event.type] as any}>{event.type}</Badge>
                  <div className="flex items-center gap-1.5 flex-shrink-0">
                    {event.match >= 85 && (
                      <span className="text-xs font-semibold text-green-600 bg-green-50 px-2 py-0.5 rounded-full">
                        {event.match}% match
                      </span>
                    )}
                    {event.registered && <Badge variant="success">✓ Registered</Badge>}
                  </div>
                </div>
                <h3 className="font-semibold text-gray-900 leading-tight">{event.title}</h3>
                <p className="text-xs text-gray-500 mt-1">{event.org}</p>
                <p className="text-sm text-gray-600 mt-2 leading-relaxed">{event.description}</p>
                <div className="mt-3 flex items-center gap-3 text-xs text-gray-400">
                  <span>📅 {formatRelativeDate(event.date)}</span>
                  <span>📍 {event.location}</span>
                </div>
                <div className="flex gap-2 mt-4">
                  {event.registered
                    ? <Button variant="secondary" size="sm" className="flex-1">View Details</Button>
                    : <Button variant="primary" size="sm" className="flex-1">Register Now</Button>
                  }
                  <Button variant="ghost" size="sm">Save</Button>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* CLUBS */}
      {section === "clubs" && (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {CLUBS.map(club => (
            <div key={club.id} className="bg-white rounded-xl border border-surface-border p-5 hover:shadow-card transition-shadow">
              <div className="flex items-start gap-3 mb-3">
                <div className="w-12 h-12 bg-lpu-light rounded-xl flex items-center justify-center flex-shrink-0">
                  <span className="text-xl">🏆</span>
                </div>
                <div className="flex-1 min-w-0">
                  <h3 className="font-semibold text-gray-900 leading-tight">{club.name}</h3>
                  <p className="text-xs text-gray-500">{club.members} members</p>
                </div>
                {club.match && (
                  <span className="text-xs font-semibold text-green-600 bg-green-50 px-2 py-0.5 rounded-full flex-shrink-0">
                    {club.match}% fit
                  </span>
                )}
              </div>
              <div className="flex gap-1.5 mb-3">
                {club.tags.map(t => <Badge key={t} variant="default">{t}</Badge>)}
                <Badge variant={club.activity === "High" ? "success" : "warning"}>{club.activity} activity</Badge>
              </div>
              <Button
                variant={joinedIds.has(club.id) ? "secondary" : "primary"}
                size="sm"
                className="w-full"
                onClick={() => setJoinedIds(prev => {
                  const next = new Set(prev);
                  joinedIds.has(club.id) ? next.delete(club.id) : next.add(club.id);
                  return next;
                })}
              >
                {joinedIds.has(club.id) ? "✓ Joined" : "Join Club"}
              </Button>
            </div>
          ))}
        </div>
      )}

      {/* OPPORTUNITIES */}
      {section === "opportunities" && (
        <div className="flex flex-col items-center justify-center py-16 text-center">
          <span className="text-4xl mb-4">💼</span>
          <h3 className="font-semibold text-gray-900 mb-1">Explore Career Opportunities</h3>
          <p className="text-sm text-gray-500 mb-4">View internships, jobs, and competitions matched to your profile.</p>
          <Button variant="primary" onClick={() => window.location.href = "/career"}>Go to Career Hub →</Button>
        </div>
      )}
    </div>
  );
}
