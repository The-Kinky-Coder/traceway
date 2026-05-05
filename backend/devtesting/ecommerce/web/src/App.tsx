import { useEffect, useState } from "react";
import { apiGet, type CartItem } from "./lib/api";
import Home from "./pages/Home";
import Products from "./pages/Products";
import Cart from "./pages/Cart";
import Checkout from "./pages/Checkout";

type Page = "home" | "products" | "cart" | "checkout";

export default function App() {
  const [page, setPage] = useState<Page>("home");
  const [cart, setCart] = useState<CartItem[]>([]);
  const [productId, setProductId] = useState<number | null>(null);

  async function refreshCart() {
    try {
      const items = await apiGet<CartItem[]>("/api/cart");
      setCart(items ?? []);
    } catch {
      setCart([]);
    }
  }

  useEffect(() => {
    refreshCart();
  }, [page]);

  const cartCount = cart.reduce((s, i) => s + i.qty, 0);

  function navigate(p: Page) {
    setPage(p);
    setProductId(null);
  }

  function viewProduct(id: number) {
    setProductId(id);
    setPage("products");
  }

  return (
    <div className="app">
      <header className="nav">
        <div className="brand" onClick={() => navigate("home")} style={{ cursor: "pointer" }}>
          Loomstead
        </div>
        <nav>
          <a className={page === "home" ? "active" : ""} onClick={() => navigate("home")}>
            Home
          </a>
          <a className={page === "products" ? "active" : ""} onClick={() => navigate("products")}>
            Shop
          </a>
          <a className={page === "cart" ? "active" : ""} onClick={() => navigate("cart")}>
            Cart
          </a>
          <a className={page === "checkout" ? "active" : ""} onClick={() => navigate("checkout")}>
            Checkout
          </a>
        </nav>
        <div className="cart-pill">{cartCount > 0 ? `${cartCount} in cart` : ""}</div>
      </header>

      {page === "home" && <Home onShop={() => navigate("products")} />}
      {page === "products" && (
        <Products
          productId={productId}
          onSelect={viewProduct}
          onBack={() => setProductId(null)}
          onAdded={refreshCart}
          onGoToCart={() => navigate("cart")}
        />
      )}
      {page === "cart" && <Cart cart={cart} onCheckout={() => navigate("checkout")} onChanged={refreshCart} />}
      {page === "checkout" && <Checkout cart={cart} onComplete={refreshCart} onHome={() => navigate("home")} />}

      <footer className="foot">Loomstead — sample ecommerce demo · not a real store</footer>
    </div>
  );
}
