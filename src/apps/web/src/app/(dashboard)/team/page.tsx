import { OrganizationProfile } from "@clerk/nextjs";

export const metadata = { title: "Team — Qvora" };

// Clerk Organizations already implements membership, invites, and role
// management end-to-end — this app uses Clerk orgs as the workspace model,
// so Team management binds directly to Clerk's prebuilt component rather
// than a custom tRPC/Go membership API (none exists, and none should be
// built just to re-implement what Clerk already owns).
export default function TeamPage() {
  return (
    <main className="mx-auto flex min-h-screen w-full max-w-5xl flex-1 flex-col gap-8 px-6 py-8 sm:px-8 lg:px-12">
      <div>
        <p className="text-xs uppercase tracking-[0.2em] text-white/40">Workspace</p>
        <h1 className="text-3xl font-semibold tracking-[-0.03em]">Team</h1>
      </div>

      <div className="rounded-2xl border border-white/8 bg-white/[0.03] p-2">
        <OrganizationProfile
          routing="hash"
          appearance={{
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
              cardBox: "w-full shadow-none border-0 bg-transparent",
            },
          }}
        />
      </div>
    </main>
  );
}
