"use client";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { cn } from "@/lib/utils";
import { useState } from "react";

const SERVICES = [
  { id: 1, icon: "🏠", name: "Hostel", desc: "Room allocation, mess, maintenance", category: "accommodation" },
  { id: 2, icon: "🚌", name: "Transport", desc: "Bus schedules, routes, pass renewal", category: "transport" },
  { id: 3, icon: "💻", name: "IT Support", desc: "Wi-Fi, lab access, accounts", category: "it" },
  { id: 4, icon: "🏥", name: "Health Centre", desc: "Appointments, medical support", category: "health" },
  { id: 5, icon: "📚", name: "Library", desc: "Books, resources, digital access", category: "academic" },
  { id: 6, icon: "🏦", name: "Accounts & Fees", desc: "Fee payment, receipts, scholarships", category: "admin" },
];

const TICKETS = [
  { id: "TK-1041", title: "AC not working in Room BH4-312", service: "Hostel", status: "IN_PROGRESS", created: "2 days ago",
    timeline: [
      { event: "Ticket Submitted", time: "Oct 2, 10:30 AM", done: true },
      { event: "Assigned to Maintenance Team", time: "Oct 2, 11:00 AM", done: true },
      { event: "Technician Scheduled", time: "Oct 4, 2:00 PM", done: false },
      { event: "Resolved", time: "Pending", done: false },
    ]
  },
  { id: "TK-1038", title: "Wi-Fi connectivity in Block 32 Lab", service: "IT Support", status: "RESOLVED", created: "5 days ago",
    timeline: [
      { event: "Ticket Submitted", time: "Sep 29", done: true },
      { event: "IT Team Notified", time: "Sep 29", done: true },
      { event: "Resolved", time: "Oct 1", done: true },
    ]
  },
];

export default function CampusPage() {
  const [view, setView] = useState<"services" | "tickets" | "create">("services");
  const [selectedTicket, setSelectedTicket] = useState<typeof TICKETS[0] | null>(null);
  const [form, setForm] = useState({ service: "", title: "", description: "" });
  const [submitted, setSubmitted] = useState(false);

  return (
    <div className="max-w-4xl mx-auto px-4 py-6 space-y-6 animate-fade-in">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Campus Services</h1>
          <p className="text-sm text-gray-500">Get help, report issues, track requests</p>
        </div>
        <Button variant="primary" size="sm" onClick={() => { setView("create"); setSubmitted(false); }}>
          + New Request
        </Button>
      </div>

      {/* Tab nav */}
      <div className="flex gap-1 bg-surface-raised border border-surface-border rounded-lg p-1 w-fit">
        {(["services", "tickets"] as const).map(t => (
          <button key={t} onClick={() => { setView(t); setSelectedTicket(null); }}
            className={cn("px-4 py-1.5 rounded-md text-sm font-medium transition-all",
              view === t ? "bg-white text-gray-900 shadow-sm" : "text-gray-500 hover:text-gray-700"
            )}>
            {t === "services" ? "Services" : "My Tickets"}
          </button>
        ))}
      </div>

      {/* SERVICES DIRECTORY */}
      {view === "services" && (
        <div className="grid grid-cols-2 md:grid-cols-3 gap-3">
          {SERVICES.map(svc => (
            <button
              key={svc.id}
              onClick={() => { setForm(f => ({...f, service: svc.name})); setView("create"); setSubmitted(false); }}
              className="bg-white rounded-xl border border-surface-border p-4 text-left hover:shadow-card hover:border-lpu-primary/30 transition-all group"
            >
              <span className="text-3xl block mb-2">{svc.icon}</span>
              <p className="font-semibold text-gray-900 text-sm group-hover:text-lpu-primary transition-colors">{svc.name}</p>
              <p className="text-xs text-gray-500 mt-0.5">{svc.desc}</p>
            </button>
          ))}
        </div>
      )}

      {/* TICKETS LIST */}
      {view === "tickets" && !selectedTicket && (
        <div className="space-y-3">
          {TICKETS.map(ticket => (
            <button
              key={ticket.id}
              onClick={() => setSelectedTicket(ticket)}
              className="w-full text-left bg-white rounded-xl border border-surface-border p-4 hover:shadow-card transition-shadow"
            >
              <div className="flex items-start justify-between gap-2">
                <div>
                  <div className="flex items-center gap-2 mb-1">
                    <span className="font-mono text-xs text-gray-400">{ticket.id}</span>
                    <Badge variant={ticket.status === "RESOLVED" ? "success" : ticket.status === "IN_PROGRESS" ? "warning" : "default"}>
                      {ticket.status === "IN_PROGRESS" ? "In Progress" : ticket.status}
                    </Badge>
                  </div>
                  <p className="font-medium text-gray-900">{ticket.title}</p>
                  <p className="text-xs text-gray-500 mt-0.5">{ticket.service} · {ticket.created}</p>
                </div>
                <svg className="w-4 h-4 text-gray-300 flex-shrink-0 mt-1" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2}>
                  <path d="M9 18l6-6-6-6" />
                </svg>
              </div>
            </button>
          ))}
        </div>
      )}

      {/* TICKET DETAIL */}
      {view === "tickets" && selectedTicket && (
        <div className="space-y-4">
          <button onClick={() => setSelectedTicket(null)} className="flex items-center gap-1 text-sm text-gray-500 hover:text-gray-700">
            ← Back to tickets
          </button>
          <div className="bg-white rounded-xl border border-surface-border p-5">
            <div className="flex items-start justify-between gap-2 mb-4">
              <div>
                <span className="font-mono text-xs text-gray-400">{selectedTicket.id}</span>
                <h2 className="font-semibold text-gray-900 mt-1">{selectedTicket.title}</h2>
                <p className="text-sm text-gray-500">{selectedTicket.service}</p>
              </div>
              <Badge variant={selectedTicket.status === "RESOLVED" ? "success" : "warning"}>
                {selectedTicket.status === "IN_PROGRESS" ? "In Progress" : selectedTicket.status}
              </Badge>
            </div>

            {/* Timeline */}
            <p className="text-xs font-semibold text-gray-400 uppercase tracking-wide mb-3">Progress</p>
            <div className="space-y-0">
              {selectedTicket.timeline.map((step, i) => (
                <div key={i} className="flex gap-4">
                  <div className="flex flex-col items-center">
                    <div className={cn(
                      "w-8 h-8 rounded-full flex items-center justify-center flex-shrink-0 text-sm",
                      step.done ? "bg-green-500 text-white" : "bg-gray-100 text-gray-400 border-2 border-dashed border-gray-200"
                    )}>
                      {step.done ? "✓" : i + 1}
                    </div>
                    {i < selectedTicket.timeline.length - 1 && (
                      <div className={cn("w-0.5 h-8", step.done ? "bg-green-200" : "bg-gray-100")} />
                    )}
                  </div>
                  <div className="pb-6">
                    <p className={cn("text-sm font-medium", step.done ? "text-gray-900" : "text-gray-400")}>{step.event}</p>
                    <p className="text-xs text-gray-400">{step.time}</p>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      )}

      {/* CREATE REQUEST FORM */}
      {view === "create" && !submitted && (
        <div className="bg-white rounded-xl border border-surface-border p-6 space-y-5">
          <h2 className="font-semibold text-gray-900">New Service Request</h2>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1.5">Service Category</label>
            <select
              value={form.service}
              onChange={e => setForm(f => ({...f, service: e.target.value}))}
              className="w-full px-3 py-2.5 rounded-lg border border-surface-border text-sm bg-white focus:outline-none focus:ring-2 focus:ring-lpu-primary"
            >
              <option value="">Select a service...</option>
              {SERVICES.map(s => <option key={s.id} value={s.name}>{s.name}</option>)}
            </select>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1.5">Summary</label>
            <input
              type="text"
              value={form.title}
              onChange={e => setForm(f => ({...f, title: e.target.value}))}
              placeholder="Briefly describe the issue..."
              className="w-full px-3 py-2.5 rounded-lg border border-surface-border text-sm bg-white focus:outline-none focus:ring-2 focus:ring-lpu-primary"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1.5">Description</label>
            <textarea
              value={form.description}
              onChange={e => setForm(f => ({...f, description: e.target.value}))}
              rows={4}
              placeholder="Provide details — location, when it started, what you've tried..."
              className="w-full px-3 py-2.5 rounded-lg border border-surface-border text-sm bg-white resize-none focus:outline-none focus:ring-2 focus:ring-lpu-primary"
            />
          </div>

          <div className="flex gap-3 pt-2">
            <Button
              variant="primary"
              onClick={() => { if (form.service && form.title) setSubmitted(true); }}
              disabled={!form.service || !form.title}
            >
              Submit Request
            </Button>
            <Button variant="ghost" onClick={() => setView("services")}>Cancel</Button>
          </div>
        </div>
      )}

      {/* SUCCESS STATE */}
      {view === "create" && submitted && (
        <div className="bg-white rounded-xl border border-green-100 p-8 text-center animate-fade-in">
          <div className="w-16 h-16 bg-green-50 rounded-full flex items-center justify-center mx-auto mb-4">
            <span className="text-3xl">✅</span>
          </div>
          <h2 className="text-lg font-semibold text-gray-900">Request Submitted!</h2>
          <p className="text-sm text-gray-500 mt-1 mb-4">Your ticket has been created. The {form.service} team will respond within 24 hours.</p>
          <div className="bg-surface-raised rounded-lg p-3 inline-block">
            <p className="text-xs text-gray-500">Ticket Number</p>
            <p className="font-mono font-bold text-gray-900">TK-{Math.floor(1000 + Math.random() * 9000)}</p>
          </div>
          <div className="mt-6 flex gap-3 justify-center">
            <Button variant="primary" onClick={() => setView("tickets")}>Track My Tickets</Button>
            <Button variant="ghost" onClick={() => { setView("services"); setForm({ service:"", title:"", description:"" }); }}>Back to Services</Button>
          </div>
        </div>
      )}
    </div>
  );
}
