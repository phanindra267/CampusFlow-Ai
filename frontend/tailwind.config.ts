import type { Config } from "tailwindcss";

const config: Config = {
  content: ["./src/**/*.{js,ts,jsx,tsx,mdx}"],
  darkMode: "class",
  theme: {
    extend: {
      colors: {
        lpu: {
          primary: "#F07C00",
          hover:   "#CC6900",
          light:   "#FFF4E6",
          dark:    "#8B4700",
        },
        ai: {
          accent:  "#6366F1",
          light:   "#EEF2FF",
          dark:    "#4338CA",
        },
        surface: {
          base:    "#FFFFFF",
          raised:  "#F8F9FA",
          border:  "#E5E7EB",
        },
      },
      fontFamily: {
        sans: ["Inter", "Geist", "system-ui", "sans-serif"],
      },
      borderRadius: {
        sm:   "4px",
        md:   "8px",
        lg:   "12px",
        xl:   "16px",
        "2xl":"24px",
      },
      boxShadow: {
        card:   "0 1px 3px 0 rgb(0 0 0 / 0.06), 0 1px 2px -1px rgb(0 0 0 / 0.04)",
        modal:  "0 20px 25px -5px rgb(0 0 0 / 0.10), 0 8px 10px -6px rgb(0 0 0 / 0.06)",
        nav:    "0 1px 0 0 #E5E7EB",
      },
      animation: {
        "fade-in":   "fadeIn 150ms cubic-bezier(0.4, 0, 0.2, 1)",
        "slide-up":  "slideUp 200ms cubic-bezier(0.4, 0, 0.2, 1)",
        "shimmer":   "shimmer 1.5s infinite",
      },
      keyframes: {
        fadeIn:   { from: { opacity: "0", transform: "translateY(4px)" }, to: { opacity: "1", transform: "translateY(0)" } },
        slideUp:  { from: { opacity: "0", transform: "translateY(12px)" }, to: { opacity: "1", transform: "translateY(0)" } },
        shimmer:  { "0%": { backgroundPosition: "-200% 0" }, "100%": { backgroundPosition: "200% 0" } },
      },
    },
  },
  plugins: [],
};

export default config;
