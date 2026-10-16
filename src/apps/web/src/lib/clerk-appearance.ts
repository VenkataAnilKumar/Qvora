// Shared Clerk `appearance` config to match Qvora's dark/volt brand across all
// Clerk-rendered UI (sign-in, sign-up, onboarding org creation, team management).
export const clerkAppearance = {
  variables: {
    colorPrimary: "#7B2FFF",
    colorBackground: "#0A0A0F",
    colorInputBackground: "#141420",
    colorText: "#F5F5F7",
    colorTextSecondary: "#A0A0AD",
    borderRadius: "0.75rem",
  },
  elements: {
    rootBox: "w-full",
    card: "w-full shadow-none border-0 bg-transparent",
    socialButtonsBlockButton: "bg-white text-[#0A0A0F] hover:bg-white/90",
    dividerLine: "bg-white/10",
    dividerText: "text-white/40",
    formButtonPrimary: "bg-[#7B2FFF] hover:bg-[#7B2FFF]/90",
    footerActionLink: "text-[#7B2FFF] hover:text-[#7B2FFF]/80",
  },
};
