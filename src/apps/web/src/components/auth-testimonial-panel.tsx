"use client";

import { useEffect, useState } from "react";

const TESTIMONIALS = [
  {
    quote: "We went from brief to 12 ready-to-test variants before the morning standup ended.",
    author: "Media Buyer, performance agency",
  },
  {
    quote: "Qvora cut our concept-to-launch time from two weeks to under an hour.",
    author: "Creative Director, DTC agency",
  },
  {
    quote: "No briefs to write, no editors to brief — it just hands us a ready batch.",
    author: "Account Manager, growth agency",
  },
];

export function AuthTestimonialPanel() {
  const [index, setIndex] = useState(0);

  useEffect(() => {
    const id = setInterval(() => setIndex((i) => (i + 1) % TESTIMONIALS.length), 4000);
    return () => clearInterval(id);
  }, []);

  // TESTIMONIALS is a fixed non-empty literal array and `index` is always kept in
  // bounds via modulo in the interval above — safe to assert non-undefined here.
  const current = TESTIMONIALS[index] as (typeof TESTIMONIALS)[number];

  return (
    <div className="relative hidden flex-1 items-center justify-center overflow-hidden bg-[linear-gradient(135deg,rgba(123,47,255,0.25),rgba(10,10,15,0.98)_55%,rgba(46,156,255,0.15))] p-12 lg:flex">
      <div className="max-w-md space-y-4" key={index}>
        <p className="text-2xl font-medium leading-snug text-white/90">
          &ldquo;{current.quote}&rdquo;
        </p>
        <p className="text-sm text-white/50">— {current.author}</p>
      </div>
    </div>
  );
}
