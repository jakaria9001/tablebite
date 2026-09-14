import { useEffect, useRef, useState } from "react";
import { Link, useLocation, useNavigate } from "react-router-dom";
import { getAdminMe, logoutAdmin } from "../../api/auth";
import { useCart } from "../../context/CartContext";
import { useRestaurantSettings } from "../../hooks/useRestaurantSettings";
import { FallbackImage } from "../common/FallbackImage";

export function Header() {
  const { itemCount } = useCart();
  const { settings } = useRestaurantSettings();
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
  const [adminVerified, setAdminVerified] = useState(false);
  const mobileMenuRef = useRef<HTMLDivElement>(null);
  const location = useLocation();
  const navigate = useNavigate();
  const restaurantName = settings.restaurant_name?.trim() || "Restaurant";
  const logo = settings.logo?.trim();

  useEffect(() => {
    let cancelled = false;

    void getAdminMe().then(() => {
      if (!cancelled) {
        setAdminVerified(true);
      }
    }).catch(() => {
      if (!cancelled) {
        setAdminVerified(false);
      }
    });

    return () => {
      cancelled = true;
    };
  }, []);

  useEffect(() => {
    if (!mobileMenuOpen) {
      return;
    }

    const handleOutsidePointer = (event: PointerEvent) => {
      if (mobileMenuRef.current && !mobileMenuRef.current.contains(event.target as Node)) {
        setMobileMenuOpen(false);
      }
    };

    document.addEventListener("pointerdown", handleOutsidePointer);
    return () => document.removeEventListener("pointerdown", handleOutsidePointer);
  }, [mobileMenuOpen]);

  async function handleLogout() {
    await logoutAdmin().catch(() => undefined);
    setAdminVerified(false);
    setMobileMenuOpen(false);
    navigate("/");
  }

  const showLogout = adminVerified && location.pathname === "/admin";
  const navItems = [
    { label: "HOME", to: "/" },
    { label: "MENU", to: "/menu" },
    { label: "ABOUT", href: "/#about" },
    { label: "GALLERY", href: "/#gallery" },
    { label: "CONTACT", href: "/#contact" },
    showLogout ? { label: "LOGOUT", action: handleLogout } : { label: "PROFILE", to: "/admin" },
  ];

  return (
    <header className="sticky top-0 z-30 border-b border-[var(--line)] bg-[var(--cream)]/95 backdrop-blur">
      <div className="mx-auto flex max-w-7xl items-center justify-between gap-4 px-5 py-4 lg:px-8">
        <Link to="/" className="flex items-center gap-3">
          <FallbackImage src={logo || "/indian_restaurant_logo.jpg"} alt={`${restaurantName} logo`} className="h-12 w-auto rounded-full object-contain" />
          <p className="font-serif text-lg font-black tracking-tight text-[var(--maroon-900)] sm:text-xl">{restaurantName}</p>
        </Link>

        <div className="ml-auto flex items-center gap-3">
          <nav className="hidden items-center gap-5 text-sm font-semibold md:flex">
            {navItems.map((item) =>
              "action" in item ? (
                <button key={item.label} type="button" onClick={item.action} className="hover:text-orange-600">
                  {item.label}
                </button>
              ) : item.to ? (
                <Link key={item.label} to={item.to} className="hover:text-orange-600">
                  {item.label}
                </Link>
              ) : (
                <a key={item.label} href={item.href} className="hover:text-orange-600">
                  {item.label}
                </a>
              )
            )}
          </nav>

          <div ref={mobileMenuRef} className="relative md:hidden">
            <button
              type="button"
              aria-label="Open menu"
              onClick={() => setMobileMenuOpen((open) => !open)}
              className="flex h-11 w-11 items-center justify-center rounded-full border border-slate-200 bg-white text-xl font-black shadow-sm"
            >
              ⋯
            </button>

            {mobileMenuOpen ? (
              <div className="absolute right-0 top-12 w-44 overflow-hidden rounded-2xl border border-slate-200 bg-white p-2 shadow-xl">
                {navItems.map((item) =>
                  "action" in item ? (
                    <button
                      key={item.label}
                      type="button"
                      onClick={item.action}
                      className="block w-full rounded-xl px-3 py-2 text-left text-sm font-semibold text-slate-700 hover:bg-orange-50 hover:text-orange-600"
                    >
                      {item.label}
                    </button>
                  ) : item.to ? (
                    <Link
                      key={item.label}
                      to={item.to}
                      onClick={() => setMobileMenuOpen(false)}
                      className="block rounded-xl px-3 py-2 text-sm font-semibold text-slate-700 hover:bg-orange-50 hover:text-orange-600"
                    >
                      {item.label}
                    </Link>
                  ) : (
                    <a
                      key={item.label}
                      href={item.href}
                      onClick={() => setMobileMenuOpen(false)}
                      className="block rounded-xl px-3 py-2 text-sm font-semibold text-slate-700 hover:bg-orange-50 hover:text-orange-600"
                    >
                      {item.label}
                    </a>
                  )
                )}
              </div>
            ) : null}
          </div>

          <Link to="/cart" aria-label={`Cart${itemCount > 0 ? `, ${itemCount} items` : ""}`} className="relative flex h-11 w-11 items-center justify-center rounded-full border border-[var(--line)] bg-white shadow-sm transition hover:border-[var(--gold-500)] hover:text-[var(--maroon-700)]">
            <img src="/shopping-cart.png" alt="Cart" className="h-5 w-5 object-contain" />
            {itemCount > 0 ? (
              <span className="absolute -right-1 -top-1 flex h-5 min-w-5 items-center justify-center rounded-full bg-orange-600 px-1 text-[10px] font-black text-white">
                {itemCount}
              </span>
            ) : null}
          </Link>
        </div>
      </div>
    </header>
  );
}
