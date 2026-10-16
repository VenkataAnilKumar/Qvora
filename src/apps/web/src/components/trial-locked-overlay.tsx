"use client";

import { Button } from "@qvora/ui";
import { trpc } from "@/lib/trpc/client";

const PLANS = [
  { tier: "starter", price: "$99/mo", note: "Up to 3 variants/angle" },
  { tier: "growth", price: "$149/mo", note: "Up to 10 variants/angle" },
  { tier: "agency", price: "$399/mo", note: "Unlimited variants/angle" },
] as const;

export function TrialLockedOverlay() {
  const { data: workspace } = trpc.workspace.get.useQuery();
  const { data: usage } = trpc.workspace.getUsage.useQuery();
  const checkoutMutation = trpc.workspace.createCheckoutSession.useMutation({
    onSuccess: (result) => {
      window.location.href = result.url;
    },
  });

  if (!workspace) return null;

  // No literal "trial_expired" status exists in this schema — the closest real signals are
  // subscriptionStatus === "past_due", or still "trialing" with trialEndsAt in the past.
  const trialEndsAt = workspace.trialEndsAt ? new Date(workspace.trialEndsAt) : null;
  const trialExpired =
    workspace.subscriptionStatus === "past_due" ||
    (workspace.subscriptionStatus === "trialing" && !!trialEndsAt && trialEndsAt < new Date());

  if (!trialExpired) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-[var(--color-qvora-black)]/95 p-4 backdrop-blur-sm">
      <div className="w-full max-w-md space-y-6 rounded-2xl border border-white/10 bg-black/40 p-8 text-center">
        <div>
          <p className="text-2xl">⏳</p>
          <h1 className="mt-3 text-2xl font-semibold tracking-[-0.02em]">Your trial has ended.</h1>
          <p className="mt-2 text-sm text-white/60">
            Your {usage?.usedVariants ?? 0} ad sets are saved and ready.
          </p>
          <p className="mt-1 text-sm text-white/60">Upgrade to keep generating.</p>
        </div>

        <div className="space-y-3 text-left">
          {PLANS.map((plan, i) => (
            <div
              key={plan.tier}
              className={
                i === 0
                  ? "rounded-xl border border-[var(--color-volt)] bg-[var(--color-volt)]/5 p-4"
                  : "rounded-xl border border-white/8 bg-black/20 p-4"
              }
            >
              <div className="flex items-center justify-between">
                <p className="font-medium capitalize text-white">
                  {plan.tier} · {plan.price}
                </p>
                {i === 0 && (
                  <Button
                    size="sm"
                    disabled={checkoutMutation.isPending}
                    onClick={() =>
                      checkoutMutation.mutate({ planTier: "starter", returnPath: "/briefs" })
                    }
                  >
                    {checkoutMutation.isPending ? "Redirecting…" : "Upgrade to Starter →"}
                  </Button>
                )}
              </div>
              <p className="mt-1 text-xs text-white/50">{plan.note}</p>
            </div>
          ))}
        </div>

        <p className="text-xs text-white/40">ⓘ Assets deleted after 30 days if no plan selected.</p>
      </div>
    </div>
  );
}
