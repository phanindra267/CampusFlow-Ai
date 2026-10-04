"use client";
import { Badge } from "@/components/ui/Badge";

const NOTIFICATIONS = [
  { id: 1, type: "academic", title: "Assignment Due Soon", message: "Capstone Phase 1 Report is due tomorrow at 11:59 PM.", time: "2 hours ago", unread: true },
  { id: 2, type: "career", title: "New Internship Match", message: "A new role at Microsoft matches 85% of your skill profile.", time: "5 hours ago", unread: true },
  { id: 3, type: "campus", title: "Service Ticket Updated", message: "Your ticket TK-1041 has been assigned to a technician.", time: "1 day ago", unread: false },
  { id: 4, type: "system", title: "System Maintenance", message: "CampusCare will be down for maintenance from 2 AM to 4 AM on Sunday.", time: "2 days ago", unread: false },
];

export default function NotificationsPage() {
  return (
    <div className="max-w-3xl mx-auto px-4 py-6 space-y-6 animate-fade-in">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Notifications</h1>
          <p className="text-sm text-gray-500">Stay updated on your LPU activities</p>
        </div>
        <button className="text-sm text-lpu-primary hover:underline">Mark all as read</button>
      </div>

      <div className="bg-white rounded-xl border border-surface-border overflow-hidden">
        {NOTIFICATIONS.map((n, i) => (
          <div key={n.id} className={`p-4 flex gap-4 border-b border-surface-border last:border-0 hover:bg-surface-raised transition-colors ${n.unread ? "bg-lpu-light/30" : ""}`}>
             <div className="w-10 h-10 rounded-full flex items-center justify-center flex-shrink-0 text-xl bg-surface-raised">
               {n.type === "academic" ? "📚" : n.type === "career" ? "💼" : n.type === "campus" ? "🏛️" : "⚙️"}
             </div>
             <div className="flex-1 min-w-0">
               <div className="flex justify-between items-start mb-1">
                 <h3 className={`text-sm ${n.unread ? "font-semibold text-gray-900" : "font-medium text-gray-700"}`}>
                   {n.title}
                 </h3>
                 <span className="text-xs text-gray-400 whitespace-nowrap ml-2">{n.time}</span>
               </div>
               <p className="text-sm text-gray-600 mb-2">{n.message}</p>
               <Badge variant={n.type === "academic" ? "warning" : n.type === "career" ? "success" : "default"}>
                 {n.type.toUpperCase()}
               </Badge>
             </div>
          </div>
        ))}
      </div>
    </div>
  );
}
