import { useState } from "react";
import { apiPost, formatPrice, type CartItem } from "../lib/api";

type Props = {
  cart: CartItem[];
  onComplete: () => void;
  onHome: () => void;
};

type CheckoutResponse = { orderId: string; durationMs: number };

export default function Checkout({ cart, onComplete, onHome }: Props) {
  const [placing, setPlacing] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [order, setOrder] = useState<CheckoutResponse | null>(null);

  if (cart.length === 0 && !order) {
    return (
      <div>
        <h2>Checkout</h2>
        <div className="empty">Your cart is empty. Add something first.</div>
      </div>
    );
  }

  if (order) {
    return (
      <div>
        <h2>Order placed</h2>
        <p className="lead">
          Order <code>{order.orderId}</code> confirmed. Took {(order.durationMs / 1000).toFixed(2)}s.
        </p>
        <button onClick={onHome}>Back to home</button>
      </div>
    );
  }

  const total = cart.reduce((s, i) => s + i.product.priceCents * i.qty, 0);

  async function placeOrder() {
    setPlacing(true);
    setError(null);
    try {
      const resp = await apiPost<CheckoutResponse>("/api/checkout", {});
      setOrder(resp);
      onComplete();
    } catch (e) {
      setError(String(e));
    } finally {
      setPlacing(false);
    }
  }

  return (
    <div>
      <h2>Checkout</h2>
      {placing && (
        <div className="banner">
          <span className="spinner" />
          Placing your order — this can take a few seconds…
        </div>
      )}
      {error && <div className="banner error">{error}</div>}

      <div className="summary">
        {cart.map((item) => (
          <div className="sub-row" key={item.product.id}>
            <span>
              {item.product.name} <span style={{ color: "var(--ink-3)" }}>× {item.qty}</span>
            </span>
            <span>{formatPrice(item.product.priceCents * item.qty)}</span>
          </div>
        ))}
        <div className="total-row">
          <span>Total</span>
          <span>{formatPrice(total)}</span>
        </div>
        <div style={{ marginTop: 24, textAlign: "right" }}>
          <button onClick={placeOrder} disabled={placing}>
            {placing ? "Placing…" : "Place order"}
          </button>
        </div>
      </div>
    </div>
  );
}
