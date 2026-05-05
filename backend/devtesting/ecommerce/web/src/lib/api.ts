import { captureException } from "@tracewayapp/frontend";

export async function apiGet<T>(path: string): Promise<T> {
  try {
    const r = await fetch(path, { headers: { Accept: "application/json" } });
    if (!r.ok) {
      const traceId = r.headers.get("traceway-trace-id") ?? undefined;
      throw Object.assign(new Error(`${path} → ${r.status}`), { traceId });
    }
    return await r.json();
  } catch (e) {
    const err = e instanceof Error ? e : new Error(String(e));
    captureException(err, { distributedTraceId: (err as { traceId?: string }).traceId });
    throw err;
  }
}

export async function apiPost<T>(path: string, body: unknown): Promise<T> {
  try {
    const r = await fetch(path, {
      method: "POST",
      headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
      },
      body: JSON.stringify(body),
    });
    if (!r.ok) {
      const traceId = r.headers.get("traceway-trace-id") ?? undefined;
      throw Object.assign(new Error(`${path} → ${r.status}`), { traceId });
    }
    return await r.json();
  } catch (e) {
    const err = e instanceof Error ? e : new Error(String(e));
    captureException(err, { distributedTraceId: (err as { traceId?: string }).traceId });
    throw err;
  }
}

export function formatPrice(cents: number): string {
  return `$${(cents / 100).toFixed(2)}`;
}

export type Product = {
  id: number;
  name: string;
  priceCents: number;
  description: string;
};

export type CartItem = {
  product: Product;
  qty: number;
};
