"use client";
import { useAuth } from "@/contexts/AuthContext";
import { Sidebar, TopBar, MobileNav } from "@/components/layout/Navigation";
import { useRouter } from "next/navigation";
import { useEffect, type ReactNode } from "react";

export default function ProtectedLayout({ children }: { children: ReactNode }) {
  const { user, isReady } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (isReady && !user) router.replace("/login");
  }, [isReady, user, router]);

  if (!isReady || !user) return null;

  return (
    <div className="flex h-screen bg-surface-raised overflow-hidden">
      {/* Desktop sidebar */}
      <Sidebar className="hidden lg:flex flex-shrink-0" />

      {/* Main area */}
      <div className="flex-1 flex flex-col min-w-0 overflow-hidden">
        <TopBar />
        <main className="flex-1 overflow-y-auto pb-20 lg:pb-0 page-transition">
          {children}
        </main>
      </div>

      {/* Mobile bottom nav */}
      <MobileNav />
    </div>
  );
}
