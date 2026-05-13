import type { Config } from "tailwindcss";

export default {
  content: [
    "./app/**/*.{js,ts,jsx,tsx,mdx}",
    "./components/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  theme: {
    extend: {
      colors: {
        primary: {
          DEFAULT: "#FF6B4A",
          hover: "#fa5e3a",
        },
        surface: "#FFFFFF",
        background: "#FDFBF7",
        text: "#2D3142",
        muted: "#8C93A8",
        border: "#E2E8F0",
      },
      fontFamily: {
        heading: ["var(--font-heading)", "sans-serif"],
        body: ["var(--font-body)", "sans-serif"],
      },
      borderRadius: {
        card: "24px",
      },
      boxShadow: {
        soft: "0 12px 24px rgba(45, 49, 66, 0.05)",
        hover: "0 20px 40px rgba(45, 49, 66, 0.08)",
      },
    },
  },
  plugins: [],
} satisfies Config;
