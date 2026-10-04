"use client";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { AIInsightCard } from "@/components/ai/AIComponents";
import { cn } from "@/lib/utils";
import { useState } from "react";

const COURSES = [
  { code: "CSE316", name: "Machine Learning", faculty: "Dr. Arun Sharma", attendance: 85, nextClass: "Mon 10:30 AM", credits: 4, grade: "A" },
  { code: "CSE303", name: "Operating Systems", faculty: "Dr. Meena Gill", attendance: 78, nextClass: "Tue 9:00 AM", credits: 4, grade: "B+" },
  { code: "CSE412", name: "Cloud Computing", faculty: "Dr. Vikas Singh", attendance: 91, nextClass: "Wed 11:00 AM", credits: 3, grade: "A+" },
  { code: "CSE499", name: "Capstone Project", faculty: "Dr. Ritu Kapoor", attendance: 100, nextClass: "Fri 2:00 PM", credits: 6, grade: "A" },
  { code: "HUM201", name: "Technical Writing", faculty: "Prof. Sonal Jain", attendance: 72, nextClass: "Thu 3:00 PM", credits: 2, grade: "B" },
];

const TIMETABLE = {
  Mon: [
    { time: "9:00-10:00", code: "CSE303", name: "Operating Systems", room: "B14-201" },
    { time: "10:30-11:30", code: "CSE316", name: "Machine Learning", room: "B32-Lab4" },
    { time: "2:00-3:00", code: "CSE412", name: "Cloud Computing", room: "B11-105" },
  ],
  Tue: [
    { time: "9:00-10:00", code: "CSE303", name: "Operating Systems", room: "B14-201" },
    { time: "11:00-1:00", code: "CSE499", name: "Capstone Project", room: "Innovation Lab" },
  ],
  Wed: [
    { time: "10:30-11:30", code: "CSE316", name: "Machine Learning", room: "B32-Lab4" },
    { time: "11:00-12:00", code: "CSE412", name: "Cloud Computing", room: "B11-105" },
  ],
  Thu: [
    { time: "3:00-4:00", code: "HUM201", name: "Technical Writing", room: "B5-102" },
  ],
  Fri: [
    { time: "2:00-5:00", code: "CSE499", name: "Capstone Project", room: "Innovation Lab" },
  ],
};

const DAYS = ["Mon", "Tue", "Wed", "Thu", "Fri"] as const;

export default function AcademicsPage() {
  const [tab, setTab] = useState<"overview" | "courses" | "timetable" | "attendance">("overview");

  return (
    <div className="max-w-5xl mx-auto px-4 py-6 space-y-6 animate-fade-in">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Academics</h1>
        <p className="text-sm text-gray-500 mt-0.5">B.Tech CSE · Semester 5 · 2026–27</p>
      </div>

      {/* Tabs */}
      <div className="flex gap-1 bg-surface-raised border border-surface-border rounded-lg p-1 w-fit">
        {(["overview", "courses", "timetable", "attendance"] as const).map(t => (
          <button
            key={t}
            onClick={() => setTab(t)}
            className={cn(
              "px-4 py-1.5 rounded-md text-sm font-medium transition-all",
              tab === t ? "bg-white text-gray-900 shadow-sm" : "text-gray-500 hover:text-gray-700"
            )}
          >
            {t.charAt(0).toUpperCase() + t.slice(1)}
          </button>
        ))}
      </div>

      {/* Overview tab */}
      {tab === "overview" && (
        <div className="space-y-4">
          <AIInsightCard
            title="This week's academic priority"
            content="Your Capstone Phase 1 deadline is tomorrow. Machine Learning attendance is strong (85%). OS attendance (78%) is approaching the warning zone — attend Thursday's class."
            sources={["Assignment Tracker", "Attendance Records"]}
            updatedAt="5 min ago"
          />
          <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
            {[
              { label: "SGPA (Current Est.)", value: "8.7", sub: "Semester 5" },
              { label: "Overall CGPA", value: "8.4", sub: "All Semesters" },
              { label: "Credits Completed", value: "112", sub: "of 160" },
              { label: "Avg Attendance", value: "85%", sub: "This Semester" },
            ].map((s, i) => (
              <div key={i} className="bg-white rounded-xl border border-surface-border p-4">
                <p className="text-xs text-gray-500 mb-1">{s.label}</p>
                <p className="text-2xl font-bold text-gray-900">{s.value}</p>
                <p className="text-xs text-gray-400 mt-0.5">{s.sub}</p>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Courses tab */}
      {tab === "courses" && (
        <div className="space-y-3">
          {COURSES.map(course => {
            const attStatus = course.attendance >= 80 ? "success" : course.attendance >= 75 ? "warning" : "error";
            return (
              <div key={course.code} className="bg-white rounded-xl border border-surface-border p-4 hover:shadow-card transition-shadow">
                <div className="flex items-start justify-between gap-3">
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 mb-0.5">
                      <span className="text-xs font-mono text-lpu-primary bg-lpu-light px-2 py-0.5 rounded">{course.code}</span>
                      <Badge variant={attStatus}>
                        {course.attendance}% attendance
                      </Badge>
                    </div>
                    <h3 className="font-semibold text-gray-900">{course.name}</h3>
                    <p className="text-xs text-gray-500 mt-0.5">{course.faculty} · {course.credits} credits</p>
                    <p className="text-xs text-gray-400 mt-1">Next: {course.nextClass}</p>
                  </div>
                  <div className="text-right flex-shrink-0">
                    <p className="text-xs text-gray-400">Est. Grade</p>
                    <p className="text-xl font-bold text-gray-900">{course.grade}</p>
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      )}

      {/* Timetable tab */}
      {tab === "timetable" && (
        <div className="bg-white rounded-xl border border-surface-border overflow-hidden">
          <div className="grid grid-cols-5 divide-x divide-surface-border">
            {DAYS.map(day => (
              <div key={day} className="min-w-0">
                <div className="p-3 bg-surface-raised border-b border-surface-border text-center">
                  <p className="text-xs font-semibold text-gray-600">{day}</p>
                </div>
                <div className="p-2 space-y-2 min-h-[200px]">
                  {(TIMETABLE[day] || []).map((cls, i) => (
                    <div key={i} className="bg-lpu-light border border-orange-100 rounded-lg p-2">
                      <p className="text-xs font-semibold text-lpu-dark leading-tight">{cls.code}</p>
                      <p className="text-xs text-gray-600 truncate">{cls.name}</p>
                      <p className="text-xs text-gray-400 mt-0.5">{cls.time}</p>
                      <p className="text-xs text-gray-400 truncate">{cls.room}</p>
                    </div>
                  ))}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Attendance tab */}
      {tab === "attendance" && (
        <div className="space-y-3">
          {COURSES.map(course => {
            const safe = course.attendance >= 80;
            const warn = course.attendance >= 75 && course.attendance < 80;
            return (
              <div key={course.code} className="bg-white rounded-xl border border-surface-border p-4">
                <div className="flex items-center justify-between mb-3">
                  <div>
                    <p className="font-medium text-gray-900">{course.name}</p>
                    <p className="text-xs text-gray-500">{course.code}</p>
                  </div>
                  <Badge variant={safe ? "success" : warn ? "warning" : "error"}>
                    {course.attendance}%
                  </Badge>
                </div>
                <div className="h-2 bg-gray-100 rounded-full overflow-hidden">
                  <div
                    className={cn("h-full rounded-full transition-all duration-700",
                      safe ? "bg-green-500" : warn ? "bg-amber-400" : "bg-red-500"
                    )}
                    style={{ width: `${course.attendance}%` }}
                  />
                </div>
                {!safe && (
                  <p className="mt-2 text-xs text-amber-700">
                    {warn
                      ? "⚠️ Attend next 3 classes to reach the safe zone."
                      : "🔴 Critical — contact your faculty advisor."}
                  </p>
                )}
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
