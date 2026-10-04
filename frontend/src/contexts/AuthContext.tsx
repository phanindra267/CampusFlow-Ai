"use client";
import React, { createContext, useContext, useState, useEffect } from "react";
import { fetchAPI } from "@/lib/api";

type User = {
  id: string;
  name: string;
  email: string;
  role: string;
  program?: string;
  department?: string;
};

type AuthContextType = {
  user: User | null;
  login: (email: string, password: string) => Promise<void>;
  logout: () => void;
  isLoading: boolean;
};

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const token = localStorage.getItem("campuscare_token");
    const storedUser = localStorage.getItem("campuscare_user");
    if (token && storedUser) {
      setUser(JSON.parse(storedUser));
    }
    setIsLoading(false);
  }, []);

  const login = async (email: string, password: string) => {
    setIsLoading(true);
    try {
      // If demo accounts, simulate for presentation, otherwise hit real backend
      if (password === "demo") {
        await new Promise(r => setTimeout(r, 800));
        let demoUser: User = { id: "student_1", name: "Rohan Sharma", email, role: "STUDENT", program: "B.Tech CSE" };
        if (email.includes("faculty")) demoUser = { id: "fac_1", name: "Dr. Arun Sharma", email, role: "FACULTY", department: "School of CSE" };
        if (email.includes("admin")) demoUser = { id: "admin_1", name: "Admin Setup", email, role: "ADMIN" };
        
        setUser(demoUser);
        localStorage.setItem("campuscare_user", JSON.stringify(demoUser));
        localStorage.setItem("campuscare_token", "demo-token");
        return;
      }

      // Real API Call
      const res = await fetchAPI("/auth/login", {
        method: "POST",
        body: JSON.stringify({ email, password }),
      });
      
      const realUser: User = {
        id: res.data.user.id,
        name: res.data.user.email.split("@")[0], // Fallback name
        email: res.data.user.email,
        role: res.data.user.role,
      };

      setUser(realUser);
      localStorage.setItem("campuscare_user", JSON.stringify(realUser));
      localStorage.setItem("campuscare_token", res.data.token);
    } catch (err: any) {
      throw new Error(err.message || "Invalid credentials");
    } finally {
      setIsLoading(false);
    }
  };

  const logout = () => {
    setUser(null);
    localStorage.removeItem("campuscare_token");
    localStorage.removeItem("campuscare_user");
  };

  return (
    <AuthContext.Provider value={{ user, login, logout, isLoading }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) throw new Error("useAuth must be used within an AuthProvider");
  return context;
}