import { useState } from "react";
import { apiPost, formatPrice, type CartItem } from "../lib/api";

type Props = {
  cart: CartItem[];
  onCheckout: () => void;
  onChanged: () => void;
};

type Promo = { code?: string };

export default function Cart({ cart, onCheckout, onChanged: _onChanged }: Props) {
  const [promo, setPromo] = useState<Promo>({});
  const [promoError, setPromoError] = useState<string | null>(null);

  async function applyPromo() {
    setPromoError(null);
    // Bug 3 — when the user clicks before typing, promo.code is undefined.
    // .toUpperCase() on undefined throws TypeError. No try/catch on purpose:
    // when Traceway is added on demo day, the error boundary catches this
    // and the rrweb session replay attaches automatically.
    const normalized = promo.code!.toUpperCase();
    try {
      await apiPost("/api/promo", { code: normalized });
    } catch (e) {
      setPromoError(String(e));
    }
  }

  if (cart.length === 0) {
    return (
      <div>
        <h2>Your cart</h2>
        <div className="empty">Your cart is empty.</div>
      </div>
    );
  }

  const subtotal = cart.reduce((s, i) => s + i.product.priceCents * i.qty, 0);

  return (
    <div>
      <h2>Your cart</h2>
      {cart.map((item) => (
        <div className="row" key={item.product.id}>
          <div>
            <div style={{ fontWeight: 600 }}>{item.product.name}</div>
            <div className="qty">qty {item.qty}</div>
          </div>
          <div className="price">{formatPrice(item.product.priceCents * item.qty)}</div>
        </div>
      ))}

      <div className="promo-row">
        <input
          type="text"
          placeholder="Promo code"
          value={promo.code ?? ""}
          onChange={(e) =>
            setPromo((p) => ({ ...p, code: e.target.value === "" ? undefined : e.target.value }))
          }
        />
        <button className="ghost" onClick={applyPromo}>
          Apply
        </button>
      </div>
      {promoError && <div className="banner error" style={{ marginTop: 12 }}>{promoError}</div>}

      <div className="summary">
        <div className="sub-row"><span>Subtotal</span><span>{formatPrice(subtotal)}</span></div>
        <div className="sub-row"><span>Shipping</span><span>$0.00</span></div>
        <div className="total-row"><span>Total</span><span>{formatPrice(subtotal)}</span></div>
        <div style={{ marginTop: 18, textAlign: "right" }}>
          <button onClick={onCheckout}>Proceed to checkout →</button>
        </div>
      </div>
    </div>
  );
}
