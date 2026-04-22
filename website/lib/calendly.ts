// Single source of truth for the "Book a demo / Book a call" destination.
// Set NEXT_PUBLIC_CALENDLY_URL in your deploy env (or .env.local for dev) to
// route every demo CTA to Calendly with the user's name + email pre-filled.
// When unset, CTAs fall back to the internal /contact form.

const RAW_CALENDLY = process.env.NEXT_PUBLIC_CALENDLY_URL;

export function getCalendlyUrl(prefill?: { email?: string; name?: string }): string {
  if (!RAW_CALENDLY) return "/contact";
  try {
    const url = new URL(RAW_CALENDLY);
    if (prefill?.email) url.searchParams.set("email", prefill.email);
    if (prefill?.name) url.searchParams.set("name", prefill.name);
    return url.toString();
  } catch {
    return "/contact";
  }
}

export const hasCalendly = !!RAW_CALENDLY;
