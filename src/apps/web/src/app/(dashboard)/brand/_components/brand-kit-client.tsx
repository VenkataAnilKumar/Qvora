"use client";

import { useState } from "react";
import type { inferRouterOutputs } from "@trpc/server";
import { Button, Card, CardContent, CardHeader, CardTitle, Input } from "@qvora/ui";
import { trpc } from "@/lib/trpc/client";
import type { AppRouter } from "@/server/trpc/routers";

type BrandKit = inferRouterOutputs<AppRouter>["workspace"]["getBrandKit"];

const HEX_RE = /^#[0-9a-f]{6}$/i;

export function BrandKitClient({ initialBrandKit }: { initialBrandKit: BrandKit }) {
  const utils = trpc.useUtils();
  const { data: brandKit } = trpc.workspace.getBrandKit.useQuery(undefined, {
    initialData: initialBrandKit,
  });

  const [name, setName] = useState(brandKit?.name ?? "");
  const [primaryColor, setPrimaryColor] = useState(brandKit?.primaryColor ?? "#7B2FFF");
  const [secondaryColor, setSecondaryColor] = useState(brandKit?.secondaryColor ?? "");
  const [fontFamily, setFontFamily] = useState(brandKit?.fontFamily ?? "");
  const [watermarkEnabled, setWatermarkEnabled] = useState(brandKit?.watermarkEnabled ?? true);
  const [saved, setSaved] = useState(false);

  const upsert = trpc.workspace.upsertBrandKit.useMutation({
    onSuccess: () => {
      setSaved(true);
      utils.workspace.getBrandKit.invalidate();
      setTimeout(() => setSaved(false), 2500);
    },
  });

  const primaryValid = HEX_RE.test(primaryColor);
  const secondaryValid = secondaryColor === "" || HEX_RE.test(secondaryColor);
  const canSave = name.trim().length > 0 && primaryValid && secondaryValid && !upsert.isPending;

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!canSave) return;
    upsert.mutate({
      name: name.trim(),
      primaryColor,
      secondaryColor: secondaryColor === "" ? undefined : secondaryColor,
      fontFamily: fontFamily === "" ? undefined : fontFamily,
      watermarkEnabled,
    });
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      <Card className="border-white/8 bg-white/[0.03]">
        <CardHeader>
          <CardTitle className="text-xl">Identity</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
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

          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-2">
              <label htmlFor="primary-color" className="text-sm font-medium text-white/78">
                Primary color
              </label>
              <div className="flex items-center gap-2">
                <input
                  type="color"
                  value={primaryValid ? primaryColor : "#7B2FFF"}
                  onChange={(e) => setPrimaryColor(e.target.value)}
                  className="h-10 w-10 shrink-0 cursor-pointer rounded-lg border border-white/10 bg-transparent"
                />
                <Input
                  id="primary-color"
                  value={primaryColor}
                  onChange={(e) => setPrimaryColor(e.target.value)}
                  placeholder="#7B2FFF"
                />
              </div>
              {!primaryValid && (
                <p className="text-xs text-[var(--color-signal-red)]">
                  Must be a hex color, e.g. #7B2FFF
                </p>
              )}
            </div>

            <div className="space-y-2">
              <label htmlFor="secondary-color" className="text-sm font-medium text-white/78">
                Secondary color <span className="text-white/40">(optional)</span>
              </label>
              <div className="flex items-center gap-2">
                <input
                  type="color"
                  value={secondaryValid && secondaryColor ? secondaryColor : "#00E87A"}
                  onChange={(e) => setSecondaryColor(e.target.value)}
                  className="h-10 w-10 shrink-0 cursor-pointer rounded-lg border border-white/10 bg-transparent"
                />
                <Input
                  id="secondary-color"
                  value={secondaryColor}
                  onChange={(e) => setSecondaryColor(e.target.value)}
                  placeholder="#00E87A"
                />
              </div>
              {!secondaryValid && (
                <p className="text-xs text-[var(--color-signal-red)]">
                  Must be a hex color, e.g. #00E87A
                </p>
              )}
            </div>
          </div>

          <div className="space-y-2">
            <label htmlFor="font-family" className="text-sm font-medium text-white/78">
              Font family <span className="text-white/40">(optional)</span>
            </label>
            <Input
              id="font-family"
              value={fontFamily}
              onChange={(e) => setFontFamily(e.target.value)}
              placeholder="Space Grotesk"
            />
          </div>

          <label className="flex cursor-pointer items-center justify-between rounded-xl border border-white/8 bg-black/20 px-4 py-3">
            <div>
              <p className="text-sm font-medium text-white">Watermark</p>
              <p className="text-xs text-white/45">Burn the brand logo into generated videos.</p>
            </div>
            <input
              type="checkbox"
              checked={watermarkEnabled}
              onChange={(e) => setWatermarkEnabled(e.target.checked)}
              className="h-5 w-5 rounded border-white/20 accent-[var(--color-volt)]"
            />
          </label>
        </CardContent>
      </Card>

      <Card className="border-white/8 bg-white/[0.03]">
        <CardHeader>
          <CardTitle className="text-xl">Logo</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex cursor-not-allowed flex-col items-center gap-2 rounded-xl border border-dashed border-white/15 bg-black/20 px-5 py-8 text-center opacity-60">
            <p className="text-sm text-white/60">Drag and drop PNG/SVG, or click to browse</p>
            <p className="text-xs text-white/35">
              Logo upload coming soon — no upload endpoint wired up yet.
            </p>
          </div>
        </CardContent>
      </Card>

      {upsert.error && (
        <p className="rounded-xl border border-[var(--color-signal-red)]/30 bg-[var(--color-signal-red)]/8 px-4 py-3 text-sm text-[var(--color-signal-red)]">
          {upsert.error.message}
        </p>
      )}

      <div className="flex items-center gap-3 border-t border-white/8 pt-4">
        <Button type="submit" disabled={!canSave}>
          {upsert.isPending ? "Saving…" : "Save brand kit"}
        </Button>
        {saved && <span className="text-sm text-[var(--color-convert-green)]">✓ Saved</span>}
      </div>
    </form>
  );
}
