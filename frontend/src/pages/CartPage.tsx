import { Link } from "react-router-dom";
import { CartPanel } from "../components/cart/CartPanel";
import { useCart } from "../context/CartContext";
import { Footer } from "../components/layout/Footer";
import { Header } from "../components/layout/Header";

export default function CartPage() {
  useCart();

  return (
    <div>
      <Header />
      <main className="mx-auto max-w-6xl px-5 py-10 pb-24 lg:px-8">
        <div className="mb-8">
          <p className="text-sm font-bold uppercase tracking-[0.2em] text-orange-600">Checkout</p>
          <h1 className="mt-2 text-4xl font-black">Your order</h1>
        </div>
        <CartPanel />
      </main>
      <div className="fixed inset-x-0 bottom-0 z-40 border-t border-white/10 bg-slate-950 px-3 py-3 shadow-[0_-10px_30px_rgba(0,0,0,0.35)] backdrop-blur md:hidden">
        <Link to="/menu" className="mx-auto flex w-fit items-center justify-center rounded-full border border-white/20 bg-white/20 px-5 py-3 text-sm font-black uppercase tracking-[0.2em] text-white shadow-lg" style={{ color: "#ffffff" }}>
          <span className="text-white">Go to Menu</span>
        </Link>
      </div>
      <Footer />
    </div>
  );
}
