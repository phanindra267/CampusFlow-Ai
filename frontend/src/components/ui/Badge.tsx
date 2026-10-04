import { cn } from "@/lib/utils";

export type BadgeVariant = "default" | "success" | "warning" | "error" | "ai" | "lpu" | "info";

interface BadgeProps {
  children: React.ReactNode;
  variant?: BadgeVariant;
  className?: string;
}

export function Badge({ children, variant = "default", className }: BadgeProps) {
  const variants = {
    default: "bg-gray-100 text-gray-700",
    success: "bg-green-50 text-green-700 ring-1 ring-green-200",
    warning: "bg-amber-50 text-amber-700 ring-1 ring-amber-200",
    error:   "bg-red-50 text-red-700 ring-1 ring-red-200",
    ai:      "bg-ai-light text-ai-accent ring-1 ring-indigo-200",
    lpu:     "bg-lpu-light text-lpu-dark ring-1 ring-orange-200",
    info:    "bg-blue-50 text-blue-700 ring-1 ring-blue-200",
  };

  return (
    <span className={cn(
      "inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium",
      variants[variant],
      className
    )}>
      {children}
    </span>
  );
}

// Status dot
interface StatusDotProps {
  status: "online" | "offline" | "away";
}
export function StatusDot({ status }: StatusDotProps) {
  const colors = { online: "bg-green-400", offline: "bg-gray-300", away: "bg-amber-400" };
  return <span className={cn("inline-block w-2 h-2 rounded-full", colors[status])} />;
}
