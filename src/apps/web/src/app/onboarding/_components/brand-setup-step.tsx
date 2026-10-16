"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Button, Input } from "@qvora/ui";
import { trpc } from "@/lib/trpc/client";

const HEX_RE = /^#[0-9a-f]{6}$/i;

export function BrandSetupStep({
  role,
  spend,
  company,
}: {
  role: string;
  spend: string;
  company: string;
}) {
  const router = useRouter();
  const [name, setName] = useState(company || "");
  const [primaryColor, setPrimaryColor] = useState("#7B2FFF");

  const upsert = trpc.workspace.upsertBrandKit.useMutation({
    onSuccess: () => router.push("/briefs"),
  });

  // role/spend aren't persisted anywhere — no backend field exists for them yet.
  // They're carried through the URL only so a future analytics/onboarding hook can pick them up.
  void role;
  void spend;

  const canSubmit = name.trim().length > 0 && HEX_RE.test(primaryColor) && !upsert.isPending;

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!canSubmit) return;
    upsert.mutate({ name: name.trim(), primaryColor, watermarkEnabled: true });
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      <div className="text-center">
        <h1 className="text-2xl font-semibold tracking-[-0.02em] text-white">
          Set up your first brand
        </h1>
        <p className="mt-1 text-sm text-white/50">
          You can add more detail in Brand Settings later.
        </p>
      </div>

      <div className="space-y-2">
        <label htmlFor="brand-name" className="text-sm font-medium text-white/78">
          Brand name
        </label>
        <Input
          id="brand-name"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="Acme Co"
          required
        />
      </div>

      <div className="space-y-2">
        <label htmlFor="primary-color" className="text-sm font-medium text-white/78">
          Primary color
        </label>
        <div className="flex items-center gap-2">
          <input
            type="color"
            value={HEX_RE.test(primaryColor) ? primaryColor : "#7B2FFF"}
            onChange={(e) => setPrimaryColor(e.target.value)}
            className="h-10 w-10 shrink-0 cursor-pointer rounded-lg border border-white/10 bg-transparent"
          />
          <Input
            id="primary-color"
            value={primaryColor}
            onChange={(e) => setPrimaryColor(e.target.value)}
          />
        </div>
      </div>

      <div className="space-y-2">
        <p className="text-sm font-medium text-white/78">Logo</p>
        <div className="flex flex-col items-center gap-1 rounded-xl border border-dashed border-white/15 bg-black/20 px-5 py-6 text-center opacity-60">
          <p className="text-sm text-white/60">Drag and drop PNG/SVG</p>
        </div>
        <p className="text-right">
          <button
            type="button"
            onClick={() => router.push("/briefs")}
            className="text-sm text-white/40 hover:text-white/60"
          >
            Skip for now →
          </button>
        </p>
      </div>

      {upsert.error && (
        <p className="rounded-xl border border-[var(--color-signal-red)]/30 bg-[var(--color-signal-red)]/8 px-4 py-3 text-sm text-[var(--color-signal-red)]">
          {upsert.error.message}
        </p>
      )}

      <Button type="submit" className="w-full" disabled={!canSubmit}>
        {upsert.isPending ? "Creating…" : "Create brand & continue →"}
      </Button>

      <p className="text-center text-xs text-white/35">
        ⓘ You can add fonts, colours, and intro/outro in Brand Settings anytime.
      </p>
    </form>
  );
}
