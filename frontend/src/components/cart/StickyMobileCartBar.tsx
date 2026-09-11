import { Link, useLocation } from "react-router-dom";
import { useCart } from "../../context/CartContext";

export function StickyMobileCartBar() {
  const { items, itemCount, subtotal } = useCart();
  const location = useLocation();

  if (items.length === 0 || location.pathname === "/cart") {
    return null;
  }

  return (
    <div className="fixed inset-x-0 bottom-0 z-40 border-t border-white/10 bg-slate-950 px-3 py-3 shadow-[0_-10px_30px_rgba(0,0,0,0.35)] backdrop-blur md:hidden">
      <Link to="/cart" className="flex items-center justify-between rounded-full border border-white/10 bg-white/10 px-4 py-3 text-white shadow-lg">
        <div className="text-left">
          <p className="text-sm font-black text-white">{itemCount} item{itemCount > 1 ? "s" : ""}</p>
          <p className="text-xs text-white/70">₹{subtotal}</p>
        </div>
        <span className="text-sm font-black uppercase tracking-[0.2em] text-white">View cart</span>
      </Link>
    </div>
  );
}
