import { SignIn } from "@clerk/nextjs";
import { AuthTestimonialPanel } from "@/components/auth-testimonial-panel";
import { clerkAppearance } from "@/lib/clerk-appearance";

export default function SignInPage() {
  return (
    <div className="flex min-h-screen bg-[var(--color-qvora-black)]">
      <div className="flex flex-1 flex-col items-center justify-center px-6 py-12">
        <div className="w-full max-w-sm space-y-6">
          <div className="text-center">
            <h1 className="text-2xl font-semibold tracking-[-0.02em] text-white">Welcome back</h1>
          </div>
          <SignIn appearance={clerkAppearance} />
        </div>
      </div>
      <AuthTestimonialPanel />
    </div>
  );
}
