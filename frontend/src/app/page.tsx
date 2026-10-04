"use client";
import { useAuth } from "@/contexts/AuthContext";
import { useRouter } from "next/navigation";
import { useEffect } from "react";

export default function RootPage() {
  const { user, isReady } = useAuth();
  const router = useRouter();
  useEffect(() => {
    if (!isReady) return;
    if (user) router.replace("/dashboard");
    else router.replace("/login");
  }, [isReady, user, router]);
  return null;
}
