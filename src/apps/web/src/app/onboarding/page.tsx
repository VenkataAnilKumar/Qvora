import { redirect } from "next/navigation";
import { auth } from "@clerk/nextjs/server";
import { CreateOrganization } from "@clerk/nextjs";
import { clerkAppearance } from "@/lib/clerk-appearance";
import { RoleContextStep } from "./_components/role-context-step";
import { BrandSetupStep } from "./_components/brand-setup-step";

type SearchParams = {
  step?: string;
  role?: string;
  spend?: string;
  company?: string;
};

export default async function OnboardingPage({
  searchParams,
}: {
  searchParams: Promise<SearchParams>;
}) {
  const { userId } = await auth();
  if (!userId) {
    redirect("/sign-in");
  }

  const params = await searchParams;
  const step = params.step ?? "org";

  return (
    <div className="flex min-h-screen items-center justify-center bg-[var(--color-qvora-black)] px-6 py-12">
      <div className="w-full max-w-md space-y-6">
        {step !== "org" && (
          <div className="flex justify-center gap-2">
            <span
              className={`h-1.5 w-8 rounded-full ${
                step === "context" ? "bg-[var(--color-volt)]" : "bg-[var(--color-volt)]/40"
              }`}
            />
            <span
              className={`h-1.5 w-8 rounded-full ${
                step === "brand" ? "bg-[var(--color-volt)]" : "bg-white/10"
              }`}
            />
          </div>
        )}

        {step === "org" && (
          <>
            <div className="text-center">
              <h1 className="text-2xl font-semibold tracking-[-0.02em] text-white">
                Create your workspace
              </h1>
              <p className="mt-1 text-sm text-white/50">
                One workspace per agency or team — you can invite teammates later.
              </p>
            </div>
            <CreateOrganization
              afterCreateOrganizationUrl="/onboarding?step=context"
              appearance={clerkAppearance}
            />
          </>
        )}

        {step === "context" && <RoleContextStep />}

        {step === "brand" && (
          <BrandSetupStep
            role={params.role ?? ""}
            spend={params.spend ?? ""}
            company={params.company ?? ""}
          />
        )}
      </div>
    </div>
  );
}
