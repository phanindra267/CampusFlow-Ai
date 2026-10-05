import { cn } from "@/lib/utils";
import { ButtonHTMLAttributes, forwardRef } from "react";

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: "primary" | "secondary" | "ghost" | "destructive" | "ai";
  size?: "sm" | "md" | "lg";
  isLoading?: boolean;
}

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  ({ variant = "primary", size = "md", isLoading, children, className, disabled, ...props }, ref) => {
    const base = "inline-flex items-center justify-center gap-2 font-medium rounded-md transition-all duration-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-lpu-primary focus-visible:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed";

    const variants = {
      primary:     "bg-gradient-to-r from-lpu-primary to-lpu-hover text-white hover:saturate-110 active:scale-[0.98] shadow-md shadow-orange-900/10 hover:shadow-lg hover:shadow-orange-900/20",
      secondary:   "bg-surface-base border border-surface-border text-gray-700 hover:bg-surface-raised",
      ghost:       "text-gray-600 hover:bg-surface-raised hover:text-gray-900",
      destructive: "bg-red-500 text-white hover:bg-red-600",
      ai:          "bg-gradient-to-r from-ai-accent to-ai-dark text-white hover:saturate-110 shadow-md shadow-indigo-900/10 hover:shadow-lg hover:shadow-indigo-900/20",
    };

    const sizes = {
      sm: "px-3 py-1.5 text-sm",
      md: "px-4 py-2 text-sm",
      lg: "px-6 py-3 text-base",
    };

    return (
      <button
        ref={ref}
        disabled={disabled || isLoading}
        className={cn(base, variants[variant], sizes[size], className)}
        {...props}
      >
        {isLoading && (
          <svg className="animate-spin h-4 w-4" viewBox="0 0 24 24" fill="none">
            <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"/>
            <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/>
          </svg>
        )}
        {children}
      </button>
    );
  }
);

Button.displayName = "Button";
