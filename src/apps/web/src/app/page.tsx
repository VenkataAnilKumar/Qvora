import Link from "next/link";
import {
  Badge,
  Button,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@qvora/ui";

const pipeline = [
  {
    step: "01",
    title: "Paste a URL",
    description:
      "Qvora scrapes the product page, extracts claims, pricing, proof points, and brand cues.",
  },
  {
    step: "02",
    title: "Generate 10 ad variants",
    description:
      "Angles, scripts, voice, video, lip-sync, export. Built for vertical paid social from the start.",
  },
  {
    step: "03",
    title: "Learn what converts",
    description:
      "Performance signals feed the next generation cycle so each round gets sharper, not noisier.",
  },
] as const;

const highlights = [
  "Agency-first V1",
  "Next.js 15 + Go + Rust",
  "FAL.AI async pipeline",
  "Server-side tier enforcement",
] as const;

const personas = [
  {
    title: "Media Buyers",
    body: "Spin up more testing angles without waiting on manual creative production every cycle.",
  },
  {
    title: "Creative Directors",
    body: "Keep voice and visual direction tight while scaling variants per offer, audience, and platform.",
  },
  {
    title: "Account Managers",
    body: "Review outputs, approve exports, and keep clients aligned without generation permissions.",
  },
] as const;

export default function HomePage() {
  return (
    <main className="min-h-screen overflow-x-hidden bg-[radial-gradient(circle_at_top_left,rgba(46,156,255,0.16),transparent_24%),radial-gradient(circle_at_top_right,rgba(123,47,255,0.22),transparent_30%),linear-gradient(180deg,#0a0a0f_0%,#0d0d14_45%,#0a0a0f_100%)] text-[var(--color-qvora-white)]">
      <section className="relative isolate mx-auto flex min-h-screen w-full max-w-7xl flex-col px-6 pb-16 pt-8 sm:px-8 lg:px-12">
        <div className="absolute inset-x-0 top-0 -z-10 h-64 bg-[linear-gradient(90deg,rgba(123,47,255,0),rgba(123,47,255,0.22),rgba(0,232,122,0.10),rgba(123,47,255,0))] blur-3xl" />

        <header className="flex items-center justify-between border-b border-white/10 py-5">
          <div>
            <p className="font-[var(--font-family-display)] text-2xl font-semibold tracking-tight">
              Qvora
            </p>
            <p className="text-xs uppercase tracking-[0.28em] text-white/45">Born to Convert</p>
          </div>
          <div className="flex items-center gap-3">
            <Button asChild variant="ghost">
              <Link href="/sign-in">Sign in</Link>
            </Button>
            <Button asChild size="lg">
              <Link href="/sign-up">Start free trial</Link>
            </Button>
          </div>
        </header>

        <div className="grid flex-1 items-center gap-14 py-14 lg:grid-cols-[1.15fr_0.85fr] lg:py-20">
          <div className="space-y-8">
            <div className="flex flex-wrap gap-2">
              <Badge className="border border-[var(--color-volt)]/40 bg-[var(--color-volt)]/10 text-[var(--color-qvora-white)]">
                AI performance creative SaaS
              </Badge>
              <Badge variant="outline">Agency V1</Badge>
              <Badge variant="outline">9:16 ads</Badge>
            </div>

            <div className="space-y-6">
              <h1 className="max-w-4xl font-[var(--font-family-display)] text-5xl font-semibold leading-none tracking-[-0.04em] sm:text-6xl lg:text-7xl">
                Paste a product URL.
                <br />
                <span className="text-[var(--color-volt)]">Launch 10 video ads.</span>
                <br />
                Know which ones win.
              </h1>
              <p className="max-w-2xl text-lg leading-8 text-white/68 sm:text-xl">
                Qvora turns a landing page into short-form paid social creative, then closes the
                loop with performance-aware iteration so your next batch starts smarter than the
                last one.
              </p>
            </div>

            <div className="flex flex-col gap-4 sm:flex-row">
              <Button asChild size="lg" className="shadow-[var(--shadow-glow-volt)]">
                <Link href="/sign-up">Start 7-day free trial</Link>
              </Button>
              <Button asChild size="lg" variant="outline">
                <Link href="/dashboard">View dashboard shell</Link>
              </Button>
            </div>

            <div className="grid gap-3 sm:grid-cols-2">
              {highlights.map((item) => (
                <div
                  key={item}
                  className="rounded-2xl border border-white/8 bg-white/4 px-4 py-3 text-sm text-white/72 backdrop-blur-sm"
                >
                  {item}
                </div>
              ))}
            </div>
          </div>

          <div className="relative">
            <div className="absolute -left-6 top-12 h-24 w-24 rounded-full bg-[var(--color-convert-green)]/20 blur-2xl" />
            <div className="absolute -right-4 top-0 h-28 w-28 rounded-full bg-[var(--color-data-blue)]/20 blur-2xl" />
            <Card className="overflow-hidden border-white/12 bg-[linear-gradient(180deg,rgba(255,255,255,0.08),rgba(255,255,255,0.03))] shadow-[var(--shadow-card)] backdrop-blur-xl">
              <CardHeader className="border-b border-white/8 pb-5">
                <CardDescription className="text-white/55">Creative signal board</CardDescription>
                <CardTitle className="text-2xl">Spring launch / supplement stack</CardTitle>
              </CardHeader>
              <CardContent className="space-y-6 pt-6">
                <div className="rounded-2xl border border-[var(--color-volt)]/20 bg-[var(--color-volt)]/10 p-4">
                  <div className="flex items-center justify-between text-sm text-white/60">
                    <span>Generation output</span>
                    <span>10 variants</span>
                  </div>
                  <div className="mt-4 grid gap-3 sm:grid-cols-2">
                    <div className="rounded-xl bg-black/25 p-4">
                      <p className="text-xs uppercase tracking-[0.2em] text-white/45">Top hook</p>
                      <p className="mt-3 text-lg font-medium">
                        "Your 3 p.m. crash is costing you sales."
                      </p>
                    </div>
                    <div className="rounded-xl bg-black/25 p-4">
                      <p className="text-xs uppercase tracking-[0.2em] text-white/45">
                        Winning angle
                      </p>
                      <p className="mt-3 text-lg font-medium text-[var(--color-convert-green)]">
                        Problem → proof → CTA
                      </p>
                    </div>
                  </div>
                </div>

                <div className="grid gap-3 sm:grid-cols-3">
                  <MetricCard
                    label="CTR lift"
                    value="+18%"
                    accent="text-[var(--color-convert-green)]"
                  />
                  <MetricCard
                    label="Time to first draft"
                    value="14m"
                    accent="text-[var(--color-data-blue)]"
                  />
                  <MetricCard
                    label="Creative fatigue"
                    value="Low"
                    accent="text-[var(--color-volt)]"
                  />
                </div>

                <div className="rounded-2xl border border-white/8 bg-black/20 p-4">
                  <div className="mb-3 flex items-center justify-between text-sm text-white/55">
                    <span>Pipeline status</span>
                    <span>Live</span>
                  </div>
                  <div className="space-y-3">
                    <ProgressRow
                      label="Scrape product page"
                      widthLabel="100%"
                      widthClass="w-full"
                      tone="bg-[var(--color-data-blue)]"
                    />
                    <ProgressRow
                      label="Generate brief angles"
                      widthLabel="84%"
                      widthClass="w-[84%]"
                      tone="bg-[var(--color-volt)]"
                    />
                    <ProgressRow
                      label="Render video variants"
                      widthLabel="62%"
                      widthClass="w-[62%]"
                      tone="bg-[var(--color-convert-green)]"
                    />
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>
        </div>
      </section>

      <section className="mx-auto grid w-full max-w-7xl gap-6 px-6 pb-8 sm:px-8 lg:grid-cols-[0.9fr_1.1fr] lg:px-12">
        <Card className="border-white/8 bg-white/[0.03]">
          <CardHeader>
            <CardDescription>Who this is for</CardDescription>
            <CardTitle className="text-3xl">
              Built for agency operators who need output and signal.
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4 text-white/68">
            <p>
              Qvora is scoped for agencies first. Media buyers move faster, creative directors keep
              quality high, and account managers review without breaking production flow.
            </p>
            <p>
              DTC brand workflows stay out of V1 on purpose. The product is being shaped around
              agency velocity, approvals, and repeatable paid social testing.
            </p>
          </CardContent>
        </Card>

        <div className="grid gap-4 md:grid-cols-3">
          {personas.map((persona) => (
            <Card
              key={persona.title}
              className="border-white/8 bg-[linear-gradient(180deg,rgba(255,255,255,0.05),rgba(255,255,255,0.02))]"
            >
              <CardHeader>
                <CardTitle>{persona.title}</CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-sm leading-7 text-white/62">{persona.body}</p>
              </CardContent>
            </Card>
          ))}
        </div>
      </section>

      <section className="mx-auto w-full max-w-7xl px-6 py-12 sm:px-8 lg:px-12">
        <div className="mb-6 flex items-end justify-between gap-6">
          <div>
            <p className="text-sm uppercase tracking-[0.24em] text-white/45">How it works</p>
            <h2 className="mt-3 font-[var(--font-family-display)] text-4xl font-semibold tracking-[-0.03em]">
              One pipeline. Three decisions.
            </h2>
          </div>
          <p className="max-w-xl text-sm leading-7 text-white/58">
            The product flow stays intentionally narrow: ingest, generate, learn. Everything in V1
            serves that loop.
          </p>
        </div>
        <div className="grid gap-4 lg:grid-cols-3">
          {pipeline.map((item) => (
            <Card key={item.step} className="border-white/10 bg-white/[0.03]">
              <CardHeader>
                <div className="mb-3 text-sm font-medium uppercase tracking-[0.25em] text-[var(--color-data-blue)]">
                  {item.step}
                </div>
                <CardTitle className="text-2xl">{item.title}</CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-sm leading-7 text-white/64">{item.description}</p>
              </CardContent>
            </Card>
          ))}
        </div>
      </section>

      <section className="mx-auto w-full max-w-7xl px-6 pb-20 pt-6 sm:px-8 lg:px-12">
        <Card className="border-white/10 bg-[linear-gradient(135deg,rgba(123,47,255,0.18),rgba(10,10,15,0.98)_45%,rgba(0,232,122,0.10))]">
          <CardContent className="flex flex-col gap-8 p-8 lg:flex-row lg:items-center lg:justify-between lg:p-10">
            <div className="max-w-2xl">
              <p className="text-sm uppercase tracking-[0.24em] text-white/50">
                Start narrow. Learn fast.
              </p>
              <h2 className="mt-3 font-[var(--font-family-display)] text-4xl font-semibold tracking-[-0.03em]">
                Build your first conversion-focused video set before the brief gets stale.
              </h2>
              <p className="mt-4 text-base leading-8 text-white/66">
                Seven-day free trial. No card up front. Generation locks on day eight unless the
                workspace upgrades.
              </p>
            </div>
            <div className="flex flex-col gap-3 sm:flex-row lg:flex-col">
              <Button asChild size="lg" variant="success">
                <Link href="/sign-up">Create workspace</Link>
              </Button>
              <Button asChild size="lg" variant="outline">
                <Link href="/sign-in">Resume existing org</Link>
              </Button>
            </div>
          </CardContent>
        </Card>
      </section>
    </main>
  );
}

function MetricCard({
  label,
  value,
  accent,
}: {
  label: string;
  value: string;
  accent: string;
}) {
  return (
    <div className="rounded-2xl border border-white/8 bg-white/[0.04] p-4">
      <p className="text-xs uppercase tracking-[0.2em] text-white/45">{label}</p>
      <p className={`mt-3 text-3xl font-semibold ${accent}`}>{value}</p>
    </div>
  );
}

function ProgressRow({
  label,
  widthLabel,
  widthClass,
  tone,
}: {
  label: string;
  widthLabel: string;
  widthClass: string;
  tone: string;
}) {
  return (
    <div>
      <div className="mb-2 flex items-center justify-between text-sm text-white/60">
        <span>{label}</span>
        <span>{widthLabel}</span>
      </div>
      <div className="h-2 rounded-full bg-white/8">
        <div className={`h-2 rounded-full ${tone} ${widthClass}`} />
      </div>
    </div>
  );
}
