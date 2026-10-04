import { type ClassValue, clsx } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function formatRelativeDate(date: Date): string {
  const now = new Date();
  const diff = date.getTime() - now.getTime();
  const absDiff = Math.abs(diff);

  if (absDiff < 60000) return "Just now";
  if (absDiff < 3600000) return `${Math.round(absDiff / 60000)}m`;
  if (absDiff < 86400000) {
    const h = Math.round(absDiff / 3600000);
    return diff > 0 ? `in ${h}h` : `${h}h ago`;
  }
  const d = Math.round(absDiff / 86400000);
  if (d === 1) return diff > 0 ? "Tomorrow" : "Yesterday";
  if (d <= 7) return diff > 0 ? `in ${d}d` : `${d}d ago`;
  return date.toLocaleDateString("en-IN", { day: "numeric", month: "short" });
}

export function getAttendanceStatus(percentage: number): "safe" | "warning" | "danger" {
  if (percentage >= 80) return "safe";
  if (percentage >= 75) return "warning";
  return "danger";
}
