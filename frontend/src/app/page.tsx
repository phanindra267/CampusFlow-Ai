"use client";
import { useAuth } from "@/contexts/AuthContext";
import { useRouter } from "next/navigation";
import { useEffect } from "react";

export default function RootPage() {
  const { user } = useAuth();
  const router = useRouter();
  useEffect(() => {
    if (user) router.replace("/dashboard");
    else router.replace("/login");
  }, [user, router]);
  return null;
}
