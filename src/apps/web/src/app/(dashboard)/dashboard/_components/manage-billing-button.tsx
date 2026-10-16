"use client";

import { Button } from "@qvora/ui";
import { trpc } from "@/lib/trpc/client";

export function ManageBillingButton() {
  const portalMutation = trpc.workspace.createPortalSession.useMutation({
    onSuccess: (result) => {
      window.location.href = result.url;
    },
  });

  return (
    <div className="flex flex-col items-start gap-1">
      <Button
        type="button"
        variant="outline"
        size="sm"
        disabled={portalMutation.isPending}
        onClick={() => portalMutation.mutate({ returnPath: "/dashboard" })}
      >
        {portalMutation.isPending ? "Opening…" : "Manage billing"}
      </Button>
      {portalMutation.error && (
        <p className="text-xs text-white/40">
          {portalMutation.error.message.includes("no_stripe_customer_yet")
            ? "Available after your first checkout."
            : portalMutation.error.message}
        </p>
      )}
    </div>
  );
}
