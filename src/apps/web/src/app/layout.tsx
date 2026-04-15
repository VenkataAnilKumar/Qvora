import type { Metadata } from "next";
import { Inter } from "next/font/google";
import { ClerkProvider } from "@clerk/nextjs";
import { TRPCProvider } from "@/lib/trpc/client";
import "./globals.css";

const inter = Inter({
  subsets: ["latin"],
  variable: "--font-inter",
  display: "swap",
});

export const metadata: Metadata = {
  title: {
    default: "Qvora — Born to Convert",
    template: "%s | Qvora",
  },
  description:
    "AI-powered performance creative SaaS. Paste a URL. Get 10 video ads. Know which ones win — automatically.",
  metadataBase: new URL(process.env.NEXT_PUBLIC_APP_URL ?? "https://app.qvora.com"),
  openGraph: {
    type: "website",
    siteName: "Qvora",
  },
  robots: {
    index: false, // Private app — not for public indexing
  },
};

const clerkPublishableKey = process.env.NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY;

export default function RootLayout({ children }: { children: React.ReactNode }) {
  if (!clerkPublishableKey) {
    return (
      <html lang="en" className={inter.variable}>
        <body>
          <TRPCProvider>{children}</TRPCProvider>
        </body>
      </html>
    );
  }

  return (
    <ClerkProvider publishableKey={clerkPublishableKey}>
      <html lang="en" className={inter.variable}>
        <body>
          <TRPCProvider>{children}</TRPCProvider>
        </body>
      </html>
    </ClerkProvider>
  );
}
