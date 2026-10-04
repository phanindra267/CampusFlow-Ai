"use client";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { AIInsightCard } from "@/components/ai/AIComponents";
import { cn } from "@/lib/utils";
import { useState } from "react";

const SKILLS = [
  { name: "Python", level: 85, evidence: "CSE316 Grade A · 3 Projects", verified: true, relevant: true },
  { name: "React.js", level: 72, evidence: "2 Projects · GDSC Workshop", verified: false, relevant: true },
  { name: "SQL", level: 68, evidence: "DBMS Grade B+ · 1 Project", verified: true, relevant: true },
  { name: "Machine Learning", level: 55, evidence: "CSE316 Ongoing", verified: false, relevant: true },
  { name: "System Design", level: 20, evidence: "Not started", verified: false, relevant: true, gap: true },
  { name: "DSA", level: 60, evidence: "CP Club · LeetCode 150 problems", verified: false, relevant: true, gap: true },
];

const INTERNSHIPS = [
  {
    id: 1, company: "Microsoft India", role: "Software Engineering Intern",
    location: "Hyderabad", duration: "3 months", stipend: "₹60,000/mo",
    deadline: "Oct 20", eligible: true, skills: ["Python", "DSA", "System Design"],
    matchSkills: ["Python"], missingSkills: ["System Design"], match: 74,
  },
  {
    id: 2, company: "Google (via GSoC)", role: "Open Source Contributor — AI/ML",
    location: "Remote", duration: "3 months", stipend: "₹1,20,000",
    deadline: "Nov 5", eligible: true, skills: ["Python", "ML", "Git"],
    matchSkills: ["Python", "Machine Learning"], missingSkills: [], match: 91,
  },
  {
    id: 3, company: "Flipkart", role: "Data Science Intern",
    location: "Bangalore", duration: "6 months", stipend: "₹45,000/mo",
    deadline: "Oct 28", eligible: true, skills: ["Python", "SQL", "ML"],
    matchSkills: ["Python", "SQL", "Machine Learning"], missingSkills: [], match: 88,
  },
];

export default function CareerPage() {
  const [tab, setTab] = useState<"overview" | "internships" | "skills" | "preparation">("overview");
  const [saved, setSaved] = useState<Set<number>>(new Set());

  const readiness = 72;

  return (
    <div className="max-w-5xl mx-auto px-4 py-6 space-y-6 animate-fade-in">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Career Hub</h1>
        <p className="text-sm text-gray-500">LPU Placement Cell · Personalized career intelligence</p>
      </div>

      {/* Tab nav */}
      <div className="flex gap-1 bg-surface-raised border border-surface-border rounded-lg p-1 w-fit flex-wrap">
        {(["overview", "internships", "skills", "preparation"] as const).map(t => (
          <button key={t} onClick={() => setTab(t)}
            className={cn("px-4 py-1.5 rounded-md text-sm font-medium transition-all",
              tab === t ? "bg-white text-gray-900 shadow-sm" : "text-gray-500 hover:text-gray-700"
            )}>
            {t.charAt(0).toUpperCase() + t.slice(1)}
          </button>
        ))}
      </div>

      {/* OVERVIEW */}
      {tab === "overview" && (
        <div className="space-y-4">
          <AIInsightCard
            title="Your SDE readiness: 72%"
            content="You match well on Python and ML. Your biggest gaps for SDE roles are System Design and advanced DSA. Starting these now gives you 6 months before the next LPU placement season."
            sources={["Skill Profile", "Placement Data 2025–26", "JD Analysis"]}
            updatedAt="1h ago"
          />
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {/* Readiness ring */}
            <div className="bg-white rounded-xl border border-surface-border p-6 flex items-center gap-6">
              <div className="relative w-24 h-24 flex-shrink-0">
                <svg className="transform -rotate-90 w-24 h-24" viewBox="0 0 36 36">
                  <circle cx="18" cy="18" r="15" fill="none" stroke="#F3F4F6" strokeWidth="3" />
                  <circle cx="18" cy="18" r="15" fill="none" stroke="#F07C00" strokeWidth="3"
                    strokeDasharray={`${readiness * 94.2 / 100} 94.2`} strokeLinecap="round" />
                </svg>
                <span className="absolute inset-0 flex items-center justify-center text-2xl font-bold text-gray-900">{readiness}%</span>
              </div>
              <div>
                <p className="font-semibold text-gray-900">Career Readiness</p>
                <p className="text-sm text-gray-500 mt-0.5">For SDE roles at product companies</p>
                <div className="mt-2 space-y-1">
                  {[["Skills", "68%"], ["Projects", "80%"], ["Experience", "60%"]].map(([k, v]) => (
                    <div key={k} className="flex items-center gap-2">
                      <span className="text-xs text-gray-500 w-16">{k}</span>
                      <div className="flex-1 h-1.5 bg-gray-100 rounded-full overflow-hidden">
                        <div className="h-full bg-lpu-primary rounded-full" style={{ width: v }} />
                      </div>
                      <span className="text-xs text-gray-400 w-8">{v}</span>
                    </div>
                  ))}
                </div>
              </div>
            </div>

            {/* Top gaps */}
            <div className="bg-white rounded-xl border border-surface-border p-4">
              <p className="text-sm font-semibold text-gray-900 mb-3">Priority Skill Gaps</p>
              {SKILLS.filter(s => s.gap).map(s => (
                <div key={s.name} className="mb-3">
                  <div className="flex items-center justify-between mb-1">
                    <span className="text-sm text-gray-700">{s.name}</span>
                    <span className="text-xs text-gray-500">{s.level}%</span>
                  </div>
                  <div className="h-2 bg-gray-100 rounded-full overflow-hidden">
                    <div className="h-full bg-amber-400 rounded-full" style={{ width: `${s.level}%` }} />
                  </div>
                  <p className="text-xs text-gray-400 mt-0.5">{s.evidence}</p>
                </div>
              ))}
              <Button variant="primary" size="sm" className="w-full mt-2" onClick={() => setTab("skills")}>
                View all skills →
              </Button>
            </div>
          </div>
        </div>
      )}

      {/* INTERNSHIPS */}
      {tab === "internships" && (
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <p className="text-sm text-gray-500">{INTERNSHIPS.length} opportunities matched to your profile</p>
            <Button variant="secondary" size="sm">Filters</Button>
          </div>
          {INTERNSHIPS.map(i => (
            <div key={i.id} className="bg-white rounded-xl border border-surface-border p-5">
              <div className="flex items-start gap-4">
                <div className="w-12 h-12 bg-gray-100 rounded-xl flex items-center justify-center flex-shrink-0">
                  <span className="text-xl">🏢</span>
                </div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-start justify-between gap-2">
                    <div>
                      <h3 className="font-semibold text-gray-900">{i.role}</h3>
                      <p className="text-sm text-gray-500">{i.company} · {i.location}</p>
                    </div>
                    <span className="text-sm font-semibold text-green-600 bg-green-50 px-2.5 py-1 rounded-full flex-shrink-0">
                      {i.match}% match
                    </span>
                  </div>

                  <div className="flex flex-wrap gap-2 mt-3">
                    <Badge variant="info">⏰ Deadline: {i.deadline}</Badge>
                    <Badge variant="default">💰 {i.stipend}</Badge>
                    <Badge variant="default">📅 {i.duration}</Badge>
                  </div>

                  {/* Match explanation — Trust UX */}
                  <div className="mt-3 p-3 bg-surface-raised rounded-lg border border-surface-border">
                    <p className="text-xs font-medium text-gray-700 mb-1.5">Why this matches you</p>
                    <div className="flex flex-wrap gap-1.5">
                      {i.matchSkills.map(s => (
                        <span key={s} className="text-xs font-medium text-green-700 bg-green-50 px-2 py-0.5 rounded-full">✓ {s}</span>
                      ))}
                      {i.missingSkills.map(s => (
                        <span key={s} className="text-xs font-medium text-amber-700 bg-amber-50 px-2 py-0.5 rounded-full">⚠ {s} (Gap)</span>
                      ))}
                    </div>
                  </div>

                  <div className="flex gap-2 mt-3">
                    <Button variant="primary" size="sm">Apply Now</Button>
                    <Button
                      variant="secondary"
                      size="sm"
                      onClick={() => setSaved(prev => { const n = new Set(prev); saved.has(i.id) ? n.delete(i.id) : n.add(i.id); return n; })}
                    >
                      {saved.has(i.id) ? "✓ Saved" : "Save"}
                    </Button>
                  </div>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* SKILLS */}
      {tab === "skills" && (
        <div className="space-y-3">
          {SKILLS.map(skill => (
            <div key={skill.name} className={cn(
              "bg-white rounded-xl border p-4",
              skill.gap ? "border-amber-100" : "border-surface-border"
            )}>
              <div className="flex items-center justify-between mb-2">
                <div className="flex items-center gap-2">
                  <span className="font-medium text-gray-900">{skill.name}</span>
                  {skill.verified && <Badge variant="success">Verified</Badge>}
                  {skill.gap && <Badge variant="warning">Gap</Badge>}
                </div>
                <span className="text-sm font-semibold text-gray-700">{skill.level}%</span>
              </div>
              <div className="h-2 bg-gray-100 rounded-full overflow-hidden mb-2">
                <div
                  className={cn("h-full rounded-full transition-all duration-700",
                    skill.gap ? "bg-amber-400" : skill.level >= 75 ? "bg-green-500" : "bg-lpu-primary"
                  )}
                  style={{ width: `${skill.level}%` }}
                />
              </div>
              <p className="text-xs text-gray-400">{skill.evidence}</p>
              {skill.gap && (
                <Button variant="secondary" size="sm" className="mt-2 text-xs">
                  Find learning resources →
                </Button>
              )}
            </div>
          ))}
        </div>
      )}

      {/* PREPARATION */}
      {tab === "preparation" && (
        <div className="space-y-4">
          {[
            { title: "Resume Review", desc: "AI-assisted resume analysis for SDE roles", icon: "📄", action: "Analyze Resume" },
            { title: "Mock Interview", desc: "Practice DSA and system design with AI", icon: "🎯", action: "Start Practice" },
            { title: "Company Research", desc: "Deep-dive on companies visiting LPU campus", icon: "🏢", action: "View Companies" },
            { title: "Placement Timeline", desc: "LPU campus placement schedule for 2026-27", icon: "📅", action: "View Timeline" },
          ].map((item, i) => (
            <div key={i} className="bg-white rounded-xl border border-surface-border p-4 flex items-center gap-4 hover:shadow-card transition-shadow">
              <div className="w-12 h-12 bg-surface-raised rounded-xl flex items-center justify-center flex-shrink-0 text-2xl">
                {item.icon}
              </div>
              <div className="flex-1">
                <p className="font-medium text-gray-900">{item.title}</p>
                <p className="text-sm text-gray-500 mt-0.5">{item.desc}</p>
              </div>
              <Button variant="secondary" size="sm" className="flex-shrink-0">{item.action}</Button>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
