"use client";
import { useAuth } from "@/contexts/AuthContext";
import { AIInsightCard } from "@/components/ai/AIComponents";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { cn, formatRelativeDate } from "@/lib/utils";
import Link from "next/link";

// ─── Mock Data ────────────────────────────────────────────────────────────────
const today = new Date();
const tomorrow = new Date(today); tomorrow.setDate(today.getDate() + 1);
const in3h = new Date(today); in3h.setHours(today.getHours() + 3);

const STUDENT_DATA = {
  nextClass: { course: "CSE316 — Machine Learning", time: "10:30 AM", room: "Block 32 · Lab 4", inMinutes: 45 },
  attendance: { percentage: 82, threshold: 75, trend: "stable" as const },
  deadlines: [
    { id: 1, title: "Capstone Phase 1 Report", course: "CSE499", due: tomorrow, urgent: true },
    { id: 2, title: "OS Assignment 3", course: "CSE303", due: in3h, urgent: true },
    { id: 3, title: "Cloud Computing Quiz", course: "CSE412", due: new Date(today.getTime() + 5 * 86400000), urgent: false },
  ],
  recommendations: [
    { id: 1, type: "event", title: "AWS Cloud Practitioner Workshop", org: "GDSC LPU", deadline: "2 days", match: 94 },
    { id: 2, type: "opportunity", title: "Microsoft SWE Intern — Summer 2027", org: "Microsoft", deadline: "10 days", match: 87 },
    { id: 3, type: "club", title: "Competitive Programming Club", org: "CP Club LPU", deadline: null, match: 91 },
  ],
  career: { readiness: 72, skills: ["Python", "React", "SQL", "ML"], missingFor: "SDE Roles: DSA, System Design" },
};

const FACULTY_DATA = {
  todayClasses: [
    { course: "CSE301 — Data Structures", time: "9:00 AM", room: "Block 14 · Room 201", students: 62 },
    { course: "CSE402 — Advanced Algorithms", time: "2:00 PM", room: "Block 14 · Room 105", students: 48 },
  ],
  pendingApprovals: 4,
  researchAlerts: 2,
};

// ─── Components ───────────────────────────────────────────────────────────────
function AttendanceMeter({ pct, threshold }: { pct: number; threshold: number }) {
  const status = pct >= 80 ? "safe" : pct >= threshold ? "warning" : "danger";
  const color = status === "safe" ? "bg-green-500" : status === "warning" ? "bg-amber-400" : "bg-red-500";
  const textColor = status === "safe" ? "text-green-700" : status === "warning" ? "text-amber-700" : "text-red-700";
  const bg = status === "safe" ? "bg-green-50" : status === "warning" ? "bg-amber-50" : "bg-red-50";

  return (
    <div className={cn("rounded-xl p-4 border", bg, status === "safe" ? "border-green-100" : status === "warning" ? "border-amber-100" : "border-red-100")}>
      <div className="flex items-center justify-between mb-2">
        <p className="text-sm font-medium text-gray-700">Attendance</p>
        <Badge variant={status === "safe" ? "success" : status === "warning" ? "warning" : "error"}>
          {status === "safe" ? "On Track" : status === "warning" ? "Warning" : "At Risk"}
        </Badge>
      </div>
      <div className="flex items-end gap-2">
        <span className={cn("text-3xl font-bold", textColor)}>{pct}%</span>
        <span className="text-sm text-gray-400 mb-1">/ {threshold}% required</span>
      </div>
      <div className="mt-3 h-2 bg-white rounded-full overflow-hidden border border-white/60">
        <div
          className={cn("h-full rounded-full transition-all duration-700", color)}
          style={{ width: `${pct}%` }}
        />
      </div>
      <p className="mt-2 text-xs text-gray-500">
        {status === "safe" ? "Great! You can miss 2 more classes safely." : "Attend your next 3 classes to be safe."}
      </p>
    </div>
  );
}

function DeadlineCard({ deadline }: { deadline: typeof STUDENT_DATA.deadlines[0] }) {
  return (
    <div className={cn(
      "flex items-start gap-3 p-3 rounded-lg border transition-colors",
      deadline.urgent ? "bg-red-50 border-red-100" : "bg-white border-surface-border hover:bg-surface-raised"
    )}>
      <div className={cn(
        "w-2 h-2 mt-1.5 rounded-full flex-shrink-0",
        deadline.urgent ? "bg-red-500" : "bg-gray-300"
      )} />
      <div className="flex-1 min-w-0">
        <p className="text-sm font-medium text-gray-900 truncate">{deadline.title}</p>
        <p className="text-xs text-gray-500">{deadline.course}</p>
      </div>
      <span className={cn(
        "text-xs font-medium flex-shrink-0 px-2 py-0.5 rounded-full",
        deadline.urgent ? "text-red-700 bg-red-100" : "text-gray-500 bg-gray-100"
      )}>
        {formatRelativeDate(deadline.due)}
      </span>
    </div>
  );
}

function RecommendationCard({ rec }: { rec: typeof STUDENT_DATA.recommendations[0] }) {
  const typeIcon = rec.type === "event" ? "🎉" : rec.type === "opportunity" ? "💼" : "🏆";
  const typeLabel = rec.type === "event" ? "Event" : rec.type === "opportunity" ? "Opportunity" : "Club";

  return (
    <div className="bg-white rounded-xl border border-surface-border p-4 hover:shadow-card transition-shadow group">
      <div className="flex items-start justify-between gap-2 mb-2">
        <div className="flex items-center gap-2">
          <span className="text-lg">{typeIcon}</span>
          <Badge variant={rec.type === "opportunity" ? "lpu" : "default"}>{typeLabel}</Badge>
        </div>
        <span className="text-xs font-semibold text-green-600 bg-green-50 px-2 py-0.5 rounded-full flex-shrink-0">
          {rec.match}% match
        </span>
      </div>
      <h3 className="text-sm font-semibold text-gray-900 leading-tight">{rec.title}</h3>
      <p className="text-xs text-gray-500 mt-0.5">{rec.org}</p>
      {rec.deadline && <p className="text-xs text-amber-600 mt-1.5">⏰ {rec.deadline} left</p>}
      <Button variant="secondary" size="sm" className="mt-3 w-full text-xs group-hover:border-lpu-primary group-hover:text-lpu-primary">
        View Details →
      </Button>
    </div>
  );
}

// ─── Student Dashboard ────────────────────────────────────────────────────────
function StudentHome() {
  const { user } = useAuth();
  const d = STUDENT_DATA;
  const hour = new Date().getHours();
  const greeting = hour < 12 ? "Good morning" : hour < 17 ? "Good afternoon" : "Good evening";

  return (
    <div className="max-w-6xl mx-auto px-4 py-6 space-y-6">
      {/* Header */}
      <div className="flex items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">
            {greeting}, {user?.name?.split(" ")[0]}. 👋
          </h1>
          <p className="text-sm text-gray-500 mt-0.5">{user?.program}</p>
        </div>
        <Link href="/ai">
          <Button variant="ai" size="sm" className="flex-shrink-0">
            ✨ Ask CampusCare
          </Button>
        </Link>
      </div>

      {/* AI Priority Insight */}
      <AIInsightCard
        title="Your day needs attention"
        content={`You have ${d.deadlines.filter(x => x.urgent).length} urgent deadlines and your next class starts in ${d.nextClass.inMinutes} minutes (${d.nextClass.course}).`}
        sources={["Academic Profile", "Timetable", "Assignment Tracker"]}
        updatedAt="2 min ago"
      />

      {/* "Your Day" + Attendance grid */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {/* Next class */}
        <div className="md:col-span-2 bg-white rounded-xl border border-surface-border p-4">
          <p className="text-xs font-semibold text-gray-400 uppercase tracking-wide mb-3">Next Class</p>
          <div className="flex items-center gap-4">
            <div className="w-12 h-12 bg-lpu-light rounded-xl flex items-center justify-center flex-shrink-0">
              <span className="text-2xl">📖</span>
            </div>
            <div>
              <p className="font-semibold text-gray-900">{d.nextClass.course}</p>
              <p className="text-sm text-gray-500">{d.nextClass.time} · {d.nextClass.room}</p>
            </div>
            <div className="ml-auto">
              <Badge variant="warning">In {d.nextClass.inMinutes}m</Badge>
            </div>
          </div>
        </div>

        {/* Attendance meter */}
        <AttendanceMeter pct={d.attendance.percentage} threshold={d.attendance.threshold} />
      </div>

      {/* Deadlines + Recommendations */}
      <div className="grid grid-cols-1 lg:grid-cols-5 gap-4">
        {/* Deadlines */}
        <div className="lg:col-span-2 bg-white rounded-xl border border-surface-border p-4">
          <div className="flex items-center justify-between mb-3">
            <p className="text-sm font-semibold text-gray-900">Deadlines</p>
            <Link href="/academics" className="text-xs text-lpu-primary hover:underline">See all</Link>
          </div>
          <div className="space-y-2">
            {d.deadlines.map(dl => <DeadlineCard key={dl.id} deadline={dl} />)}
          </div>
        </div>

        {/* For You */}
        <div className="lg:col-span-3">
          <div className="flex items-center justify-between mb-3">
            <p className="text-sm font-semibold text-gray-900">For You</p>
            <Link href="/discover" className="text-xs text-lpu-primary hover:underline">See all</Link>
          </div>
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
            {d.recommendations.map(rec => <RecommendationCard key={rec.id} rec={rec} />)}
          </div>
        </div>
      </div>

      {/* Career Snapshot */}
      <div className="bg-white rounded-xl border border-surface-border p-4">
        <div className="flex items-center justify-between mb-4">
          <p className="text-sm font-semibold text-gray-900">Career Readiness</p>
          <Link href="/career">
            <Button variant="secondary" size="sm">View Career Profile →</Button>
          </Link>
        </div>
        <div className="flex items-center gap-6">
          <div className="flex-shrink-0">
            <div className="relative w-20 h-20">
              <svg className="transform -rotate-90 w-20 h-20" viewBox="0 0 36 36">
                <circle cx="18" cy="18" r="15" fill="none" stroke="#F3F4F6" strokeWidth="3" />
                <circle cx="18" cy="18" r="15" fill="none" stroke="#F07C00" strokeWidth="3"
                  strokeDasharray={`${d.career.readiness * 94.2 / 100} 94.2`} strokeLinecap="round" />
              </svg>
              <span className="absolute inset-0 flex items-center justify-center text-xl font-bold text-gray-900">
                {d.career.readiness}%
              </span>
            </div>
          </div>
          <div className="flex-1 min-w-0">
            <div className="flex flex-wrap gap-1.5 mb-2">
              {d.career.skills.map(s => <Badge key={s} variant="success">{s}</Badge>)}
            </div>
            <p className="text-xs text-amber-700 bg-amber-50 px-3 py-2 rounded-lg mt-2">
              ⚠️ <strong>Gap detected:</strong> {d.career.missingFor}
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}

// ─── Faculty Dashboard ────────────────────────────────────────────────────────
function FacultyHome() {
  const { user } = useAuth();
  const d = FACULTY_DATA;

  return (
    <div className="max-w-5xl mx-auto px-4 py-6 space-y-6">
      <div className="flex items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Hello, {user?.name?.split(" ")[0]}. 👋</h1>
          <p className="text-sm text-gray-500">{user?.department}</p>
        </div>
        <Link href="/ai"><Button variant="ai" size="sm">✨ AI Copilot</Button></Link>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        {[
          { label: "Classes Today", value: d.todayClasses.length, sub: "On your schedule", icon: "📅" },
          { label: "Pending Approvals", value: d.pendingApprovals, sub: "Student requests", icon: "✅", urgent: true },
          { label: "Research Alerts", value: d.researchAlerts, sub: "New opportunities", icon: "🔬" },
        ].map((stat, i) => (
          <div key={i} className={cn(
            "bg-white rounded-xl border p-4",
            stat.urgent ? "border-amber-200 bg-amber-50" : "border-surface-border"
          )}>
            <div className="flex items-center gap-2 mb-2">
              <span className="text-xl">{stat.icon}</span>
              <p className="text-xs text-gray-500">{stat.label}</p>
            </div>
            <p className={cn("text-3xl font-bold", stat.urgent ? "text-amber-700" : "text-gray-900")}>{stat.value}</p>
            <p className="text-xs text-gray-400 mt-0.5">{stat.sub}</p>
          </div>
        ))}
      </div>

      <div className="bg-white rounded-xl border border-surface-border p-4">
        <p className="text-sm font-semibold text-gray-900 mb-3">Today's Classes</p>
        <div className="space-y-3">
          {d.todayClasses.map((cls, i) => (
            <div key={i} className="flex items-center gap-4 p-3 bg-surface-raised rounded-lg border border-surface-border">
              <div className="w-10 h-10 bg-lpu-light rounded-lg flex items-center justify-center flex-shrink-0">
                <span className="text-lg">📖</span>
              </div>
              <div className="flex-1">
                <p className="text-sm font-medium text-gray-900">{cls.course}</p>
                <p className="text-xs text-gray-500">{cls.time} · {cls.room} · {cls.students} students</p>
              </div>
              <Button variant="secondary" size="sm">View →</Button>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

// ─── Admin Dashboard ──────────────────────────────────────────────────────────
function AdminHome() {
  return (
    <div className="max-w-6xl mx-auto px-4 py-6 space-y-6">
      <div className="flex items-start justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Campus Overview</h1>
          <p className="text-sm text-gray-500">LPU — Operational Intelligence</p>
        </div>
        <Link href="/ai"><Button variant="ai" size="sm">✨ AI Copilot</Button></Link>
      </div>

      <AIInsightCard
        title="Placement Intelligence Alert"
        content="B.Tech CSE placements are tracking 12% behind last year. Knowledge Graph identifies 'Generative AI' as a skill gap across 3rd-year students. Recommend launching a targeted bootcamp before the next recruitment cycle."
        sources={["Placement Records", "Skill Database", "Employer Requirements"]}
        updatedAt="1h ago"
      />

      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        {[
          { label: "Active Students", value: "18,420", delta: "+234", positive: true },
          { label: "Events This Month", value: "47", delta: "+8", positive: true },
          { label: "Open Service Tickets", value: "128", delta: "+23", positive: false },
          { label: "Placement Rate", value: "68%", delta: "+3%", positive: true },
        ].map((s, i) => (
          <div key={i} className="bg-white rounded-xl border border-surface-border p-4">
            <p className="text-xs text-gray-500 mb-1">{s.label}</p>
            <p className="text-2xl font-bold text-gray-900">{s.value}</p>
            <p className={cn("text-xs mt-1 font-medium", s.positive ? "text-green-600" : "text-red-500")}>
              {s.delta} vs last month
            </p>
          </div>
        ))}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        <div className="bg-white rounded-xl border border-surface-border p-4">
          <p className="text-sm font-semibold text-gray-900 mb-4">Pending Approvals</p>
          {["IoT Lab Access — Riya Sharma", "Research Grant Application — Dr. Kapoor", "New Club Registration — Robotics Club"].map((item, i) => (
            <div key={i} className="flex items-center gap-3 py-2.5 border-b border-surface-border last:border-0">
              <span className="text-sm text-gray-700 flex-1">{item}</span>
              <Button variant="primary" size="sm">Review</Button>
            </div>
          ))}
        </div>
        <div className="bg-white rounded-xl border border-surface-border p-4">
          <p className="text-sm font-semibold text-gray-900 mb-4">System Health</p>
          {[
            { service: "API Server", status: "Operational" },
            { service: "Database", status: "Operational" },
            { service: "Notification Engine", status: "Degraded" },
            { service: "AI Pipeline", status: "Not Connected" },
          ].map((svc, i) => (
            <div key={i} className="flex items-center justify-between py-2.5 border-b border-surface-border last:border-0">
              <span className="text-sm text-gray-700">{svc.service}</span>
              <Badge variant={
                svc.status === "Operational" ? "success" :
                svc.status === "Degraded" ? "warning" : "error"
              }>{svc.status}</Badge>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

// ─── Route component ──────────────────────────────────────────────────────────
export default function DashboardPage() {
  const { user } = useAuth();
  if (!user) return null;
  if (user.role === "FACULTY" || user.role === "RESEARCHER") return <FacultyHome />;
  if (user.role === "ADMIN") return <AdminHome />;
  return <StudentHome />;
}
