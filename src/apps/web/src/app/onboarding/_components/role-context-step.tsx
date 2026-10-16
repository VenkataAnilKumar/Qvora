"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Button, Input } from "@qvora/ui";

const ROLES = [
  "Media Buyer / Performance Marketer",
  "Creative Director",
  "Agency Owner / Account Manager",
  "In-house Brand Manager",
];

const SPEND_BANDS = ["<$10K", "$10K–$100K", "$100K–$500K", ">$500K"];

export function RoleContextStep() {
  const router = useRouter();
  const [role, setRole] = useState("");
  const [spend, setSpend] = useState("");
  const [company, setCompany] = useState("");

  const handleContinue = () => {
    const qs = new URLSearchParams({ step: "brand", role, spend, company });
    router.push(`/onboarding?${qs.toString()}`);
  };

  return (
    <div className="space-y-6">
      <div className="text-center">
        <h1 className="text-2xl font-semibold tracking-[-0.02em] text-white">
          Quick setup — takes 60 seconds.
        </h1>
      </div>

      <div className="space-y-2">
        <p className="text-sm font-medium text-white/78">What describes you best?</p>
        {ROLES.map((r) => (
          <label
            key={r}
            className="flex cursor-pointer items-center justify-between rounded-xl border border-white/8 bg-black/20 px-4 py-3 has-[:checked]:border-l-2 has-[:checked]:border-[var(--color-volt)] has-[:checked]:bg-[var(--color-volt)]/10"
          >
            <span className="text-sm text-white">{r}</span>
            <input
              type="radio"
              name="role"
              checked={role === r}
              onChange={() => setRole(r)}
              className="h-4 w-4 accent-[var(--color-volt)]"
            />
          </label>
        ))}
      </div>

      <div className="space-y-2">
        <p className="text-sm font-medium text-white/78">Monthly ad spend you manage</p>
        <div className="grid grid-cols-2 gap-2">
          {SPEND_BANDS.map((s) => (
            <label
              key={s}
              className="flex cursor-pointer items-center justify-center rounded-xl border border-white/8 bg-black/20 px-3 py-2 has-[:checked]:border-[var(--color-volt)]/50 has-[:checked]:bg-[var(--color-volt)]/10"
            >
              <span className="text-sm text-white">{s}</span>
              <input
                type="radio"
                name="spend"
                checked={spend === s}
                onChange={() => setSpend(s)}
                className="sr-only"
              />
            </label>
          ))}
        </div>
      </div>

      <div className="space-y-2">
        <label htmlFor="company" className="text-sm font-medium text-white/78">
          Company / Agency name
        </label>
        <Input
          id="company"
          value={company}
          onChange={(e) => setCompany(e.target.value)}
          placeholder="Acme Agency"
        />
      </div>

      <Button className="w-full" disabled={!role} onClick={handleContinue}>
        Continue →
      </Button>
    </div>
  );
}
