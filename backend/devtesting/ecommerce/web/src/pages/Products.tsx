import { useEffect, useState } from "react";
import { apiGet, apiPost, formatPrice, type Product } from "../lib/api";

type Props = {
  productId: number | null;
  onSelect: (id: number) => void;
  onBack: () => void;
  onAdded: () => void;
  onGoToCart: () => void;
};

export default function Products({ productId, onSelect, onBack, onAdded, onGoToCart }: Props) {
  if (productId !== null) {
    return <ProductDetail id={productId} onBack={onBack} onAdded={onAdded} onGoToCart={onGoToCart} />;
  }
  return <ProductsList onSelect={onSelect} />;
}

function ProductsList({ onSelect }: { onSelect: (id: number) => void }) {
  const [products, setProducts] = useState<Product[] | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    apiGet<Product[]>("/api/products")
      .then(setProducts)
      .catch((e) => setError(String(e)));
  }, []);

  if (error) return <div className="banner error">Failed to load products: {error}</div>;
  if (!products) return <div className="empty"><span className="spinner" />Loading…</div>;

  return (
    <div>
      <h2>Shop</h2>
      <div className="grid">
        {products.map((p) => (
          <div
            key={p.id}
            className={`card${p.id === 42 ? " flagged" : ""}`}
            onClick={() => onSelect(p.id)}
          >
            <div className="name">{p.name}</div>
            <div className="desc">{p.description}</div>
            <div className="price">{formatPrice(p.priceCents)}</div>
          </div>
        ))}
      </div>
    </div>
  );
}

function ProductDetail({
  id,
  onBack,
  onAdded,
  onGoToCart,
}: {
  id: number;
  onBack: () => void;
  onAdded: () => void;
  onGoToCart: () => void;
}) {
  const [product, setProduct] = useState<Product | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [adding, setAdding] = useState(false);
  const [added, setAdded] = useState(false);

  useEffect(() => {
    setProduct(null);
    setError(null);
    setAdded(false);
    apiGet<Product>(`/api/products/${id}`)
      .then(setProduct)
      .catch((e) => setError(String(e)));
  }, [id]);

  async function add() {
    if (!product) return;
    setAdding(true);
    try {
      await apiPost("/api/cart", { productId: product.id, qty: 1 });
      setAdded(true);
      onAdded();
    } catch (e) {
      setError(String(e));
    } finally {
      setAdding(false);
    }
  }

  if (error) {
    return (
      <div>
        <button className="ghost" onClick={onBack}>← Back to shop</button>
        <div className="banner error" style={{ marginTop: 24 }}>
          Couldn&apos;t load this product. The server returned an error.
          <div style={{ fontSize: 12, marginTop: 6, opacity: 0.7 }}>{error}</div>
        </div>
      </div>
    );
  }

  if (!product) {
    return <div className="empty"><span className="spinner" />Loading…</div>;
  }

  return (
    <div>
      <button className="ghost" onClick={onBack}>← Back to shop</button>
      <div style={{ marginTop: 32, maxWidth: 560 }}>
        <h1 style={{ fontSize: 28 }}>{product.name}</h1>
        <p className="lead">{product.description}</p>
        <div style={{ fontSize: 22, fontWeight: 600, margin: "20px 0" }}>{formatPrice(product.priceCents)}</div>
        <div style={{ display: "flex", gap: 10 }}>
          <button onClick={add} disabled={adding}>
            {adding ? "Adding…" : "Add to cart"}
          </button>
          {added && (
            <button className="ghost" onClick={onGoToCart}>
              View cart →
            </button>
          )}
        </div>
      </div>
    </div>
  );
}
