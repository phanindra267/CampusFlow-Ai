import { cn } from "@/lib/utils";

// Skeleton loader — content-shape aware
export function Skeleton({ className }: { className?: string }) {
  return (
    <div className={cn("skeleton rounded-md", className)} />
  );
}

// Card skeleton — mimics a typical LPU card shape
export function CardSkeleton() {
  return (
    <div className="bg-white rounded-xl p-4 space-y-3 shadow-card border border-surface-border animate-fade-in">
      <div className="flex items-center gap-3">
        <Skeleton className="w-10 h-10 rounded-full flex-shrink-0" />
        <div className="flex-1 space-y-2">
          <Skeleton className="h-4 w-3/4" />
          <Skeleton className="h-3 w-1/2" />
        </div>
      </div>
      <Skeleton className="h-3 w-full" />
      <Skeleton className="h-3 w-5/6" />
      <div className="flex gap-2 pt-1">
        <Skeleton className="h-6 w-16 rounded-full" />
        <Skeleton className="h-6 w-20 rounded-full" />
      </div>
    </div>
  );
}

// Stat card skeleton
export function StatSkeleton() {
  return (
    <div className="bg-white rounded-xl p-4 space-y-2 shadow-card border border-surface-border">
      <Skeleton className="h-3 w-24" />
      <Skeleton className="h-8 w-16" />
      <Skeleton className="h-3 w-32" />
    </div>
  );
}
