"use client";
import { cn } from "@/lib/utils";
import { useAuth } from "@/contexts/AuthContext";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { useState } from "react";

const memberNav = [
  { href: "/dashboard", label: "Home", icon: HomeIcon },
  { href: "/community", label: "Community", icon: UsersIcon },
  { href: "/discover", label: "Discover", icon: CompassIcon },
  { href: "/career", label: "Career", icon: BriefcaseIcon },
  { href: "/research", label: "Research", icon: FlaskIcon },
  { href: "/campus", label: "Campus", icon: BuildingIcon },
  { href: "/ai", label: "AI Assistant", icon: SparkleIcon, accent: true },
];

const adminNav = [
  { href: "/dashboard", label: "Overview", icon: HomeIcon },
  { href: "/admin/operations", label: "Operations", icon: BuildingIcon },
  { href: "/admin/analytics", label: "Analytics", icon: ChartIcon },
  { href: "/admin/approvals", label: "Approvals", icon: CheckIcon },
  { href: "/admin/users", label: "Users", icon: UsersIcon },
  { href: "/ai", label: "AI Copilot", icon: SparkleIcon, accent: true },
];

interface SidebarProps {
  className?: string;
}

export function Sidebar({ className }: SidebarProps) {
  const { user, logout } = useAuth();
  const pathname = usePathname();
  const [collapsed, setCollapsed] = useState(false);

  const navItems = user?.role === "ADMIN" || user?.role === "SUPER_ADMIN" ? adminNav : memberNav;

  return (
    <aside className={cn(
      "flex flex-col bg-white border-r border-surface-border transition-all duration-200",
      collapsed ? "w-16" : "w-60",
      className
    )}>
      {/* Logo */}
      <div className="flex items-center gap-3 px-4 h-16 border-b border-surface-border flex-shrink-0">
        <div className="w-8 h-8 bg-lpu-primary rounded-lg flex items-center justify-center flex-shrink-0">
          <span className="text-white font-bold text-sm">C</span>
        </div>
        {!collapsed && (
          <div className="min-w-0">
            <p className="font-semibold text-gray-900 text-sm truncate">CampusCare AI</p>
            <p className="text-xs text-gray-400 truncate">Campus Community</p>
          </div>
        )}
        <button
          onClick={() => setCollapsed(!collapsed)}
          className="ml-auto text-gray-400 hover:text-gray-600 flex-shrink-0"
          aria-label={collapsed ? "Expand sidebar" : "Collapse sidebar"}
        >
          <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2}>
            {collapsed
              ? <path d="M9 18l6-6-6-6" />
              : <path d="M15 18l-6-6 6-6" />
            }
          </svg>
        </button>
      </div>

      {/* User pill */}
      {!collapsed && user && (
        <div className="mx-3 my-3 p-2.5 bg-surface-raised rounded-lg border border-surface-border">
          <p className="text-xs font-semibold text-gray-800 truncate">{user.displayName}</p>
          <p className="text-xs text-gray-400 truncate">{user.email}</p>
          <span className="mt-1 inline-block text-xs font-medium text-lpu-primary bg-lpu-light px-2 py-0.5 rounded-full">
            {user.role}
          </span>
        </div>
      )}

      {/* Nav */}
      <nav className="flex-1 px-2 py-2 space-y-0.5 overflow-y-auto">
        {navItems.map(({ href, label, icon: Icon, accent }) => {
          const active = pathname === href || pathname.startsWith(href + "/");
          return (
            <Link
              key={href}
              href={href}
              className={cn(
                "flex items-center gap-3 px-3 py-2 rounded-lg text-sm font-medium transition-all duration-100 group",
                active
                  ? accent
                    ? "bg-ai-light text-ai-accent"
                    : "bg-lpu-light text-lpu-primary"
                  : "text-gray-600 hover:bg-surface-raised hover:text-gray-900",
              )}
              title={collapsed ? label : undefined}
            >
              <Icon className={cn(
                "w-4 h-4 flex-shrink-0",
                active
                  ? accent ? "text-ai-accent" : "text-lpu-primary"
                  : "text-gray-400 group-hover:text-gray-600"
              )} />
              {!collapsed && <span className="truncate">{label}</span>}
              {!collapsed && accent && !active && (
                <span className="ml-auto text-xs bg-ai-light text-ai-accent px-1.5 py-0.5 rounded-full">AI</span>
              )}
            </Link>
          );
        })}
      </nav>

      {/* Footer */}
      <div className="px-2 py-3 border-t border-surface-border">
        <Link href="/profile" className={cn(
          "flex items-center gap-3 px-3 py-2 rounded-lg text-sm text-gray-600 hover:bg-surface-raised",
          collapsed && "justify-center"
        )}>
          <div className="w-7 h-7 bg-gradient-to-br from-lpu-primary to-lpu-hover rounded-full flex items-center justify-center flex-shrink-0">
            <span className="text-white text-xs font-bold">{user?.displayName?.[0] ?? "?"}</span>
          </div>
          {!collapsed && <span className="truncate text-xs">{user?.displayName}</span>}
        </Link>
        <button
          onClick={logout}
          className={cn(
            "w-full flex items-center gap-3 px-3 py-2 rounded-lg text-sm text-gray-500 hover:bg-red-50 hover:text-red-600 transition-colors mt-0.5",
            collapsed && "justify-center"
          )}
        >
          <LogOutIcon className="w-4 h-4 flex-shrink-0" />
          {!collapsed && <span>Sign out</span>}
        </button>
      </div>
    </aside>
  );
}

export function TopBar() {
  const [searchOpen, setSearchOpen] = useState(false);

  return (
    <header className="h-14 bg-white border-b border-surface-border flex items-center gap-4 px-4 flex-shrink-0 sticky top-0 z-10">
      {/* Search */}
      <button
        onClick={() => setSearchOpen(true)}
        className="flex items-center gap-2 text-sm text-gray-400 bg-surface-raised border border-surface-border rounded-lg px-3 py-1.5 hover:border-gray-300 transition-colors flex-1 max-w-sm"
      >
        <SearchIcon className="w-4 h-4 flex-shrink-0" />
        <span className="truncate">Search campus…</span>
        <kbd className="ml-auto text-xs bg-white border border-surface-border rounded px-1.5 py-0.5 hidden sm:inline-flex items-center gap-1 flex-shrink-0">
          <span className="text-gray-400">⌘K</span>
        </kbd>
      </button>

      {/* Command palette modal */}
      {searchOpen && (
        <CommandPalette onClose={() => setSearchOpen(false)} />
      )}

      <div className="ml-auto flex items-center gap-2">
        {/* AI Button */}
        <Link
          href="/ai"
          className="flex items-center gap-1.5 text-sm font-medium text-ai-accent bg-ai-light hover:bg-indigo-100 px-3 py-1.5 rounded-lg transition-colors"
        >
          <svg className="w-4 h-4" viewBox="0 0 24 24" fill="currentColor">
            <path d="M12 2l2.4 7.2H22l-6.2 4.5 2.4 7.2L12 17l-6.2 3.9 2.4-7.2L2 9.2h7.6L12 2z" />
          </svg>
          <span className="hidden sm:inline">Ask CampusCare</span>
        </Link>

        {/* Notifications */}
        <Link href="/notifications" className="relative p-2 text-gray-500 hover:text-gray-800 hover:bg-surface-raised rounded-lg transition-colors">
          <BellIcon className="w-5 h-5" />
          <span className="absolute top-1.5 right-1.5 w-2 h-2 bg-lpu-primary rounded-full ring-2 ring-white" />
        </Link>

        {/* Profile */}
        <Link href="/profile" className="w-8 h-8 bg-gradient-to-br from-lpu-primary to-lpu-hover rounded-full flex items-center justify-center flex-shrink-0">
          <span className="text-white text-xs font-bold">R</span>
        </Link>
      </div>
    </header>
  );
}

function CommandPalette({ onClose }: { onClose: () => void }) {
  const [query, setQuery] = useState("");
  const suggestions = [
    { label: "Community Discussions", href: "/community", icon: "💬" },
    { label: "Upcoming Events", href: "/discover", icon: "🎉" },
    { label: "Find Opportunities", href: "/career", icon: "💼" },
    { label: "Ask the AI Assistant", href: "/ai", icon: "✨" },
    { label: "Campus Services", href: "/campus", icon: "🏛️" },
    { label: "Research & Groups", href: "/research", icon: "🔬" },
  ].filter(s => !query || s.label.toLowerCase().includes(query.toLowerCase()));

  return (
    <div className="fixed inset-0 z-50 bg-black/40 backdrop-blur-sm flex items-start justify-center pt-20 px-4" onClick={onClose}>
      <div className="w-full max-w-lg bg-white rounded-xl shadow-modal border border-surface-border overflow-hidden animate-slide-up" onClick={e => e.stopPropagation()}>
        <div className="flex items-center gap-3 px-4 py-3 border-b border-surface-border">
          <SearchIcon className="w-4 h-4 text-gray-400 flex-shrink-0" />
          <input
            autoFocus
            value={query}
            onChange={e => setQuery(e.target.value)}
            placeholder="Search campus — events, groups, opportunities, services..."
            className="flex-1 text-sm outline-none bg-transparent"
          />
          <kbd className="text-xs text-gray-400 border border-surface-border rounded px-1.5 py-0.5 flex-shrink-0">Esc</kbd>
        </div>
        <ul className="py-2 max-h-80 overflow-y-auto">
          {suggestions.map((item, i) => (
            <li key={i}>
              <Link
                href={item.href}
                onClick={onClose}
                className="flex items-center gap-3 px-4 py-2.5 hover:bg-surface-raised transition-colors"
              >
                <span className="text-lg flex-shrink-0">{item.icon}</span>
                <span className="text-sm text-gray-800">{item.label}</span>
              </Link>
            </li>
          ))}
          {suggestions.length === 0 && (
            <li className="px-4 py-6 text-center text-sm text-gray-400">No results for &ldquo;{query}&rdquo;</li>
          )}
        </ul>
      </div>
    </div>
  );
}

// Mobile bottom nav
export function MobileNav() {
  const pathname = usePathname();
  const items = [
    { href: "/dashboard", label: "Home", icon: HomeIcon },
    { href: "/community", label: "Community", icon: UsersIcon },
    { href: "/ai", label: "AI", icon: SparkleIcon, accent: true },
    { href: "/discover", label: "Discover", icon: CompassIcon },
    { href: "/profile", label: "Profile", icon: UserIcon },
  ];

  return (
    <nav className="fixed bottom-0 left-0 right-0 bg-white border-t border-surface-border flex lg:hidden z-20">
      {items.map(({ href, label, icon: Icon, accent }) => {
        const active = pathname === href;
        return (
          <Link
            key={href}
            href={href}
            className={cn(
              "flex-1 flex flex-col items-center justify-center py-2 gap-0.5 text-xs transition-colors",
              active
                ? accent ? "text-ai-accent" : "text-lpu-primary"
                : "text-gray-400"
            )}
          >
            {accent && active
              ? <div className="w-10 h-10 bg-ai-accent rounded-full flex items-center justify-center -mt-5 shadow-lg">
                  <Icon className="w-5 h-5 text-white" />
                </div>
              : <Icon className="w-5 h-5" />
            }
            {!(accent && active) && <span>{label}</span>}
          </Link>
        );
      })}
    </nav>
  );
}

// Icon components
function HomeIcon({ className }: { className?: string }) {
  return <svg className={className} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2}><path d="M3 9l9-7 9 7v11a2 2 0 01-2 2H5a2 2 0 01-2-2z"/><polyline points="9,22 9,12 15,12 15,22"/></svg>;
}
function CompassIcon({ className }: { className?: string }) {
  return <svg className={className} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2}><circle cx="12" cy="12" r="10"/><polygon points="16.24,7.76 14.12,14.12 7.76,16.24 9.88,9.88 16.24,7.76"/></svg>;
}
function BriefcaseIcon({ className }: { className?: string }) {
  return <svg className={className} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2}><rect x="2" y="7" width="20" height="14" rx="2"/><path d="M16 7V5a2 2 0 00-2-2h-4a2 2 0 00-2 2v2"/><line x1="12" y1="12" x2="12" y2="12"/></svg>;
}
function FlaskIcon({ className }: { className?: string }) {
  return <svg className={className} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2}><path d="M9 3h6M9 3v8L5.5 17A2 2 0 007.3 20h9.4a2 2 0 001.8-3L15 11V3"/></svg>;
}
function BuildingIcon({ className }: { className?: string }) {
  return <svg className={className} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2}><rect x="3" y="3" width="18" height="18" rx="1"/><path d="M3 9h18M9 21V9M3 15h6"/></svg>;
}
function SparkleIcon({ className }: { className?: string }) {
  return <svg className={className} viewBox="0 0 24 24" fill="currentColor"><path d="M12 2l2.4 7.2H22l-6.2 4.5 2.4 7.2L12 17l-6.2 3.9 2.4-7.2L2 9.2h7.6L12 2z"/></svg>;
}
function ChartIcon({ className }: { className?: string }) {
  return <svg className={className} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2}><line x1="18" y1="20" x2="18" y2="10"/><line x1="12" y1="20" x2="12" y2="4"/><line x1="6" y1="20" x2="6" y2="14"/></svg>;
}
function CheckIcon({ className }: { className?: string }) {
  return <svg className={className} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2}><polyline points="20,6 9,17 4,12"/></svg>;
}
function UsersIcon({ className }: { className?: string }) {
  return <svg className={className} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2}><path d="M17 21v-2a4 4 0 00-4-4H5a4 4 0 00-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 00-3-3.87"/><path d="M16 3.13a4 4 0 010 7.75"/></svg>;
}
function UserIcon({ className }: { className?: string }) {
  return <svg className={className} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2}><path d="M20 21v-2a4 4 0 00-4-4H8a4 4 0 00-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>;
}
function SearchIcon({ className }: { className?: string }) {
  return <svg className={className} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2}><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>;
}
function BellIcon({ className }: { className?: string }) {
  return <svg className={className} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2}><path d="M18 8A6 6 0 006 8c0 7-3 9-3 9h18s-3-2-3-9"/><path d="M13.73 21a2 2 0 01-3.46 0"/></svg>;
}
function LogOutIcon({ className }: { className?: string }) {
  return <svg className={className} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2}><path d="M9 21H5a2 2 0 01-2-2V5a2 2 0 012-2h4"/><polyline points="16,17 21,12 16,7"/><line x1="21" y1="12" x2="9" y2="12"/></svg>;
}
