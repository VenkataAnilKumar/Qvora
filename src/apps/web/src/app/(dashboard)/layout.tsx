import { redirect } from "next/navigation";
import { auth } from "@clerk/nextjs/server";

export default async function DashboardLayout({ children }: { children: React.ReactNode }) {
  const { userId, orgId } = await auth();

  if (!userId) {
    redirect("/sign-in");
  }

  if (!orgId) {
    // User is authenticated but hasn't selected/created a workspace org yet
    redirect("/onboarding");
  }

  return <div className="flex min-h-screen flex-col bg-[var(--color-surface-0)]">{children}</div>;
}
