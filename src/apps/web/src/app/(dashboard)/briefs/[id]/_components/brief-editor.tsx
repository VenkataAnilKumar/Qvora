"use client";

import { useMemo, useState } from "react";
import { Badge, Button, Card, CardContent, CardHeader, CardTitle, Input } from "@qvora/ui";
import { trpc } from "@/lib/trpc/client";

type EditableAngle = {
  angle: string;
  headline: string;
  script: string;
  cta: string;
  voiceTone: string;
};

type BriefEditorProps = {
  briefId: string;
  productUrl: string;
  status: string;
  createdAt: string;
  updatedAt: string;
  angles: EditableAngle[];
  hooks: string[];
};

export function BriefEditor(props: BriefEditorProps) {
  const [angles, setAngles] = useState<EditableAngle[]>(props.angles);
  const [hooks, setHooks] = useState<string[]>(props.hooks);
  const [updatedAt, setUpdatedAt] = useState<string>(props.updatedAt);
  const [error, setError] = useState<string | null>(null);
  const [info, setInfo] = useState<string | null>(null);

  const updateMutation = trpc.briefs.updateContent.useMutation({
    onSuccess: (result) => {
      setError(null);
      setInfo("Brief edits saved.");
      setUpdatedAt(result.updatedAt);
    },
    onError: (mutationError) => {
      setError(mutationError.message);
    },
  });

  const regenerateAngleMutation = trpc.briefs.regenerateAngle.useMutation({
    onSuccess: (result) => {
      setAngles((prev) =>
        prev.map((item, idx) => (idx === result.angleIndex ? result.angle : item)),
      );
      setError(null);
      setInfo(
        result.under10Seconds
          ? "Angle regenerated in < 10s."
          : `Angle regenerated in ${(result.elapsedMs / 1000).toFixed(1)}s.`,
      );
    },
    onError: (mutationError) => {
      setError(mutationError.message);
    },
  });

  const regenerateHookMutation = trpc.briefs.regenerateHook.useMutation({
    onSuccess: (result) => {
      setHooks(result.hooks);
      setError(null);
      setInfo(
        result.under10Seconds
          ? "Hook regenerated in < 10s."
          : `Hook regenerated in ${(result.elapsedMs / 1000).toFixed(1)}s.`,
      );
    },
    onError: (mutationError) => {
      setError(mutationError.message);
    },
  });

  const isBusy =
    updateMutation.isPending ||
    regenerateAngleMutation.isPending ||
    regenerateHookMutation.isPending;

  const canSave = useMemo(() => {
    if (angles.length === 0 || hooks.length === 0) return false;
    return angles.every(
      (angle) =>
        angle.angle.trim().length > 0 &&
        angle.headline.trim().length > 0 &&
        angle.script.trim().length > 0 &&
        angle.cta.trim().length > 0,
    );
  }, [angles, hooks]);

  const save = () => {
    if (isBusy || !canSave) {
      return;
    }
    setInfo(null);
    setError(null);
    updateMutation.mutate({
      briefId: props.briefId,
      angles,
      hooks,
    });
  };

  return (
    <div className="space-y-6">
      <Card className="border-white/8 bg-white/[0.03]">
        <CardHeader className="gap-3">
          <div className="flex items-center justify-between gap-4">
            <CardTitle className="text-2xl tracking-[-0.02em]">Brief {props.briefId}</CardTitle>
            <Badge variant={props.status === "generated" ? "success" : "outline"}>
              {props.status}
            </Badge>
          </div>
          <p className="break-all text-sm text-white/60">{props.productUrl}</p>
        </CardHeader>
        <CardContent className="space-y-2 text-sm text-white/65">
          <p>Created: {new Date(props.createdAt).toLocaleString()}</p>
          <p>Updated: {new Date(updatedAt).toLocaleString()}</p>
        </CardContent>
      </Card>

      <Card className="border-white/8 bg-white/[0.03]">
        <CardHeader>
          <CardTitle className="text-xl">Angles</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {angles.map((angle, index) => (
            <div
              key={`${angle.angle}-${index}`}
              className="space-y-2 rounded-xl border border-white/10 p-4"
            >
              <div className="flex items-center justify-between gap-3">
                <p className="text-sm font-medium text-white">{angle.angle}</p>
                <Button
                  type="button"
                  variant="outline"
                  disabled={isBusy}
                  onClick={() =>
                    regenerateAngleMutation.mutate({ briefId: props.briefId, angleIndex: index })
                  }
                >
                  Regenerate angle
                </Button>
              </div>

              <Input
                value={angle.headline}
                onChange={(event) => {
                  const value = event.target.value;
                  setAngles((prev) =>
                    prev.map((item, idx) => (idx === index ? { ...item, headline: value } : item)),
                  );
                }}
                disabled={isBusy}
                placeholder="Headline"
              />

              <textarea
                className="min-h-28 w-full rounded-md border border-white/10 bg-black/20 px-3 py-2 text-sm text-white outline-none ring-0 transition focus:border-[var(--color-volt)]/50"
                value={angle.script}
                onChange={(event) => {
                  const value = event.target.value;
                  setAngles((prev) =>
                    prev.map((item, idx) => (idx === index ? { ...item, script: value } : item)),
                  );
                }}
                disabled={isBusy}
                placeholder="Script"
              />

              <div className="grid gap-2 md:grid-cols-2">
                <Input
                  value={angle.cta}
                  onChange={(event) => {
                    const value = event.target.value;
                    setAngles((prev) =>
                      prev.map((item, idx) => (idx === index ? { ...item, cta: value } : item)),
                    );
                  }}
                  disabled={isBusy}
                  placeholder="CTA"
                />
                <Input
                  value={angle.voiceTone}
                  onChange={(event) => {
                    const value = event.target.value;
                    setAngles((prev) =>
                      prev.map((item, idx) =>
                        idx === index ? { ...item, voiceTone: value } : item,
                      ),
                    );
                  }}
                  disabled={isBusy}
                  placeholder="Voice tone"
                />
              </div>
            </div>
          ))}
        </CardContent>
      </Card>

      <Card className="border-white/8 bg-white/[0.03]">
        <CardHeader>
          <CardTitle className="text-xl">Hooks</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          {hooks.map((hook, index) => (
            <div key={`${index}-${hook}`} className="flex items-center gap-3">
              <Input
                value={hook}
                onChange={(event) => {
                  const value = event.target.value;
                  setHooks((prev) => prev.map((item, idx) => (idx === index ? value : item)));
                }}
                disabled={isBusy}
                placeholder="Hook"
              />
              <Button
                type="button"
                variant="outline"
                disabled={isBusy}
                onClick={() =>
                  regenerateHookMutation.mutate({ briefId: props.briefId, hookIndex: index })
                }
              >
                Regenerate hook
              </Button>
            </div>
          ))}
        </CardContent>
      </Card>

      {error ? (
        <p className="rounded-xl border border-[var(--color-signal-red)]/30 bg-[var(--color-signal-red)]/10 px-3 py-2 text-sm text-[var(--color-signal-red)]">
          {error}
        </p>
      ) : null}

      {info ? <p className="text-sm text-[var(--color-convert-green)]">{info}</p> : null}

      <div className="flex justify-end">
        <Button type="button" disabled={isBusy || !canSave} onClick={save}>
          {updateMutation.isPending ? "Saving…" : "Save brief edits"}
        </Button>
      </div>
    </div>
  );
}
