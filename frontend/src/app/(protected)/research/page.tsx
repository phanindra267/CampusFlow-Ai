"use client";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { AIInsightCard } from "@/components/ai/AIComponents";

export default function ResearchPage() {
  const faculty = [
    { name: "Dr. Arun Sharma", dept: "CSE", area: "Machine Learning, NLP", publications: 28, projects: 4 },
    { name: "Dr. Meena Gill", dept: "CSE", area: "Operating Systems, Distributed Computing", publications: 19, projects: 2 },
    { name: "Dr. Ritu Kapoor", dept: "IT", area: "Cybersecurity, Blockchain", publications: 14, projects: 3 },
  ];
  const opportunities = [
    { title: "ML Research Assistant — NLP Lab", faculty: "Dr. Arun Sharma", match: 91, deadline: "Oct 15", skills: ["Python", "NLP", "PyTorch"] },
    { title: "Cybersecurity Research Project", faculty: "Dr. Ritu Kapoor", match: 72, deadline: "Nov 1", skills: ["Python", "Security"] },
  ];

  return (
    <div className="max-w-5xl mx-auto px-4 py-6 space-y-6 animate-fade-in">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Research</h1>
        <p className="text-sm text-gray-500">LPU Research Groups · Faculty · Opportunities</p>
      </div>

      <AIInsightCard
        title="Research match found"
        content="Based on your CSE316 grade (A) and Python proficiency (85%), you are a strong candidate for Dr. Arun Sharma's NLP Research Assistant position. Application closes Oct 15."
        sources={["Academic Profile", "Faculty Research Database"]}
        updatedAt="30 min ago"
      />

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div>
          <h2 className="text-sm font-semibold text-gray-900 mb-3">Open Research Opportunities</h2>
          <div className="space-y-3">
            {opportunities.map((opp, i) => (
              <div key={i} className="bg-white rounded-xl border border-surface-border p-4">
                <div className="flex items-start justify-between gap-2 mb-2">
                  <h3 className="font-medium text-gray-900">{opp.title}</h3>
                  <span className="text-xs font-semibold text-green-600 bg-green-50 px-2 py-0.5 rounded-full flex-shrink-0">{opp.match}%</span>
                </div>
                <p className="text-xs text-gray-500 mb-2">{opp.faculty}</p>
                <div className="flex flex-wrap gap-1.5 mb-3">
                  {opp.skills.map(s => <Badge key={s} variant="info">{s}</Badge>)}
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-xs text-amber-600">Deadline: {opp.deadline}</span>
                  <Button variant="primary" size="sm">Apply</Button>
                </div>
              </div>
            ))}
          </div>
        </div>

        <div>
          <h2 className="text-sm font-semibold text-gray-900 mb-3">LPU Faculty Researchers</h2>
          <div className="space-y-3">
            {faculty.map((f, i) => (
              <div key={i} className="bg-white rounded-xl border border-surface-border p-4">
                <div className="flex items-center gap-3 mb-2">
                  <div className="w-10 h-10 bg-lpu-light rounded-full flex items-center justify-center flex-shrink-0">
                    <span className="font-bold text-lpu-primary">{f.name[3]}</span>
                  </div>
                  <div>
                    <p className="font-medium text-gray-900">{f.name}</p>
                    <p className="text-xs text-gray-500">{f.dept}</p>
                  </div>
                </div>
                <p className="text-xs text-gray-600 mb-2">{f.area}</p>
                <div className="flex gap-3 text-xs text-gray-400">
                  <span>📄 {f.publications} publications</span>
                  <span>🔬 {f.projects} active projects</span>
                </div>
                <Button variant="secondary" size="sm" className="mt-3 w-full">View Profile</Button>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}
