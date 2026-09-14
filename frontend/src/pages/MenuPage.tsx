import { useEffect, useMemo, useRef, useState } from "react";
import { Link, useLocation, useNavigate } from "react-router-dom";
import { Footer } from "../components/layout/Footer";
import { Header } from "../components/layout/Header";
import { MenuCard } from "../components/menu/MenuCard";
import { useMenu } from "../hooks/useMenu";
import { SeoHead } from "../components/seo/SeoHead";
import { useCart } from "../context/CartContext";

export default function MenuPage() {
  const { data, loading, error, retry } = useMenu();
  const location = useLocation();
  const navigate = useNavigate();
  const [search, setSearch] = useState("");
  const [vegOnly, setVegOnly] = useState(false);
  const [nonVegOnly, setNonVegOnly] = useState(false);
  const [recommendedOnly, setRecommendedOnly] = useState(false);
  const [priceFilter, setPriceFilter] = useState("all");
  const [categoryId, setCategoryId] = useState<number | null>(null);
  const [filterOpen, setFilterOpen] = useState(false);
  const filterMenuRef = useRef<HTMLDivElement>(null);
  const [isOnline, setIsOnline] = useState(() => typeof navigator === "undefined" || navigator.onLine);
  const [recentIds, setRecentIds] = useState<number[]>([]);
  const [loadingMessage, setLoadingMessage] = useState("Warming the kitchen");
  const { itemCount, subtotal } = useCart();

  useEffect(() => {
    try {
      const stored = JSON.parse(window.localStorage.getItem("tablebite-recent-items") ?? "[]") as unknown;
      if (Array.isArray(stored)) setRecentIds(stored.filter((id): id is number => typeof id === "number"));
    } catch {
      setRecentIds([]);
    }
  }, []);

  useEffect(() => {
    const handleOnline = () => setIsOnline(true);
    const handleOffline = () => setIsOnline(false);
    window.addEventListener("online", handleOnline);
    window.addEventListener("offline", handleOffline);
    return () => {
      window.removeEventListener("online", handleOnline);
      window.removeEventListener("offline", handleOffline);
    };
  }, []);

  useEffect(() => {
    if (!loading) {
      return;
    }

    const previousOverflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";

    return () => {
      document.body.style.overflow = previousOverflow;
    };
  }, [loading]);

  useEffect(() => {
    if (!loading) {
      return;
    }

    const messages = ["Warming the kitchen", "Checking today's availability", "Putting the favourites on the table"];
    let messageIndex = 0;
    const timer = window.setInterval(() => {
      messageIndex = (messageIndex + 1) % messages.length;
      setLoadingMessage(messages[messageIndex]);
    }, 1800);

    return () => window.clearInterval(timer);
  }, [loading]);

  useEffect(() => {
    if (!filterOpen) {
      return;
    }

    const handleOutsidePointer = (event: PointerEvent) => {
      if (filterMenuRef.current && !filterMenuRef.current.contains(event.target as Node)) {
        setFilterOpen(false);
      }
    };

    document.addEventListener("pointerdown", handleOutsidePointer);
    return () => document.removeEventListener("pointerdown", handleOutsidePointer);
  }, [filterOpen]);

  useEffect(() => {
    const slug = location.hash.replace(/^#/, "");
    if (!slug) {
      setCategoryId(null);
      return;
    }

    if (slug === "about" || slug === "gallery" || slug === "contact") {
      navigate(`/${slug}`, { replace: true });
      return;
    }

    const matchedCategory = data?.categories.find((category) => category.slug === slug);
    setCategoryId(matchedCategory?.id ?? null);
  }, [data?.categories, location.hash, navigate]);

  const items = useMemo(() => {
    const term = search.trim().toLowerCase();
    return data?.items.filter((item) => {
      if (categoryId !== null && item.category_id !== categoryId) return false;
      if (vegOnly && !item.is_veg) return false;
      if (nonVegOnly && item.is_veg) return false;
      if (recommendedOnly && !item.is_featured && !item.is_bestseller && !item.is_top10) return false;
      const price = item.price ?? item.price_full ?? 0;
      if (priceFilter === "under-200" && price >= 200) return false;
      if (priceFilter === "200-400" && (price < 200 || price > 400)) return false;
      if (priceFilter === "over-400" && price <= 400) return false;
      if (!term) return true;
      return `${item.name} ${item.description}`.toLowerCase().includes(term);
    }) ?? [];
  }, [data, search, vegOnly, nonVegOnly, recommendedOnly, priceFilter, categoryId]);

  const recentItems = useMemo(() => {
    if (!data?.items) return [];
    return recentIds.map((id) => data.items.find((item) => item.id === id)).filter((item): item is NonNullable<typeof item> => Boolean(item)).slice(0, 4);
  }, [data, recentIds]);

  const hasActiveFilters = Boolean(search.trim()) || vegOnly || nonVegOnly || recommendedOnly || priceFilter !== "all" || categoryId !== null;
  const activeFilterCount = [vegOnly, nonVegOnly, recommendedOnly, priceFilter !== "all"].filter(Boolean).length;

  return (
    <div>
      {loading ? (
        <div className="fixed inset-0 z-[60] grid place-items-center bg-[var(--maroon-950)]/90 px-6 backdrop-blur-sm" role="dialog" aria-modal="true" aria-labelledby="menu-loading-title">
          <div className="w-full max-w-xs text-center text-white">
            <div className="brand-pulse mx-auto grid h-16 w-16 place-items-center rounded-full border border-[var(--gold-500)]/50 bg-[var(--gold-500)]/15 text-2xl text-[var(--gold-200)]" aria-hidden="true">✦</div>
            <h2 id="menu-loading-title" className="mt-6 font-serif text-2xl font-black">Setting the table</h2>
            <p className="mt-2 text-sm leading-6 text-white/70">{loadingMessage}</p>
            <div className="mx-auto mt-6 h-1.5 w-40 overflow-hidden rounded-full bg-white/15" aria-hidden="true">
              <div className="h-full w-1/3 animate-[loading-slide_1.4s_ease-in-out_infinite] rounded-full bg-[var(--gold-500)]" />
            </div>
            <p className="sr-only" aria-live="polite">Loading menu items</p>
          </div>
        </div>
      ) : null}
      <SeoHead title="Menu" description="Browse the digital menu at Indian Restaurant & Sweets with categories for starters, mains, beverages, and sweets." />
      <Header />
      <main className="mx-auto max-w-7xl px-5 py-10 pb-24 lg:px-8">
        {!isOnline ? <div role="status" className="mb-5 rounded-2xl border border-[var(--gold-200)] bg-[var(--cream-deep)] px-4 py-3 text-sm font-semibold text-[var(--maroon-900)]">You are offline. Your cart is safe on this device; menu updates will resume when you reconnect.</div> : null}
        <div className="mb-6">
          <p className="text-xs font-bold uppercase tracking-[0.2em] text-[var(--gold-600)]">Digital menu</p>
          <h1 className="mt-2 max-w-xl font-serif text-3xl font-black leading-tight text-[var(--maroon-900)] sm:text-4xl">Something delicious for everyone.</h1>
        </div>
        <div className="sticky top-[72px] z-20 mb-8 border-y border-[var(--line)] bg-[var(--cream)]/95 py-2.5 backdrop-blur sm:rounded-[1.75rem] sm:border sm:p-3 sm:shadow-[var(--shadow-soft)]">
          <div className="flex gap-2.5">
            <div ref={filterMenuRef} className="relative shrink-0">
              <button type="button" onClick={() => setFilterOpen((value) => !value)} aria-expanded={filterOpen} aria-controls="menu-filters" className={`flex min-h-11 items-center gap-2 rounded-full border px-4 text-sm font-bold transition active:scale-[0.98] ${filterOpen || activeFilterCount > 0 ? "border-[var(--maroon-800)] bg-[var(--maroon-800)] text-white" : "border-[var(--gold-600)] bg-white text-[var(--ink)]"}`}>
                Filter by <span aria-hidden="true" className="text-base leading-none">☷</span>{activeFilterCount > 0 ? <span className="grid h-5 min-w-5 place-items-center rounded-full bg-[var(--gold-500)] px-1 text-xs text-[var(--maroon-950)]">{activeFilterCount}</span> : null}
              </button>
              {filterOpen ? <div id="menu-filters" className="absolute left-0 top-14 z-30 w-64 rounded-2xl border border-[var(--line)] bg-white p-3 shadow-[var(--shadow-lifted)]">
                <p className="px-2 pb-2 text-xs font-bold uppercase tracking-[0.15em] text-[var(--muted)]">Filter menu</p>
                <button type="button" onClick={() => { setVegOnly((value) => !value); setNonVegOnly(false); }} className={`flex min-h-11 w-full items-center justify-between rounded-xl px-3 text-left text-sm font-semibold ${vegOnly ? "bg-emerald-50 text-emerald-900" : "text-[var(--ink)] hover:bg-[var(--cream)]"}`}><span>Pure veg</span><span>{vegOnly ? "✓" : ""}</span></button>
                <button type="button" onClick={() => { setNonVegOnly((value) => !value); setVegOnly(false); }} className={`flex min-h-11 w-full items-center justify-between rounded-xl px-3 text-left text-sm font-semibold ${nonVegOnly ? "bg-[var(--cream-deep)] text-[var(--maroon-900)]" : "text-[var(--ink)] hover:bg-[var(--cream)]"}`}><span>Non-veg</span><span>{nonVegOnly ? "✓" : ""}</span></button>
                <button type="button" onClick={() => setRecommendedOnly((value) => !value)} className={`flex min-h-11 w-full items-center justify-between rounded-xl px-3 text-left text-sm font-semibold ${recommendedOnly ? "bg-[var(--cream-deep)] text-[var(--maroon-900)]" : "text-[var(--ink)] hover:bg-[var(--cream)]"}`}><span>Recommended</span><span>{recommendedOnly ? "✓" : ""}</span></button>
                <label className="mt-2 block border-t border-[var(--line)] px-2 pt-3 text-xs font-bold uppercase tracking-[0.15em] text-[var(--muted)]" htmlFor="price-filter">Price</label>
                <select id="price-filter" value={priceFilter} onChange={(event) => setPriceFilter(event.target.value)} className="mt-2 min-h-10 w-full rounded-xl border border-[var(--line)] bg-white px-3 text-sm font-semibold text-[var(--ink)] outline-none focus:ring-2 focus:ring-[var(--gold-500)]">
                  <option value="all">Any price</option>
                  <option value="under-200">Under ₹200</option>
                  <option value="200-400">₹200–₹400</option>
                  <option value="over-400">Over ₹400</option>
                </select>
                <button type="button" onClick={() => { setVegOnly(false); setNonVegOnly(false); setRecommendedOnly(false); setPriceFilter("all"); setFilterOpen(false); }} className="mt-3 w-full rounded-xl border border-[var(--line)] px-3 py-2 text-sm font-semibold text-[var(--muted)]">Clear filters</button>
              </div> : null}
            </div>
            <input value={search} onChange={(event) => setSearch(event.target.value)} placeholder="Search dishes" className="min-h-11 min-w-0 flex-1 rounded-full border border-[var(--line)] bg-white px-4 text-[var(--ink)] shadow-sm outline-none ring-[var(--gold-500)] placeholder:text-[var(--muted)] focus:ring-2" aria-label="Search dishes" />
          </div>
          <div className="mt-2.5 flex gap-2 overflow-x-auto pb-0.5 [scrollbar-width:none]">
            <button type="button" onClick={() => {
              setCategoryId(null);
              window.history.replaceState(null, "", "/menu");
            }} className={`min-h-10 shrink-0 rounded-full px-4 py-2 text-sm font-bold transition active:scale-[0.98] ${categoryId === null ? "bg-[var(--maroon-950)] text-white" : "bg-white text-[var(--ink)] shadow-sm"}`}>All</button>
            {data?.categories.map((category) => <button type="button" key={category.id} onClick={() => {
              setCategoryId(category.id);
              window.history.replaceState(null, "", `/menu#${category.slug}`);
            }} className={`min-h-10 shrink-0 rounded-full px-4 py-2 text-sm font-bold transition active:scale-[0.98] ${categoryId === category.id ? "bg-[var(--maroon-950)] text-white" : "bg-white text-[var(--ink)] shadow-sm"}`}>{category.name}</button>)}
          </div>
        </div>

        {loading ? <div className="grid gap-5 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">{Array.from({ length: 8 }, (_, index) => <div key={index} className="h-[390px] animate-pulse rounded-3xl bg-[var(--cream-deep)]" />)}</div> : error ? (
          <div className="rounded-2xl bg-red-50 p-5 text-red-700">
            <p>{error}</p>
            <button type="button" onClick={retry} className="mt-3 rounded-full bg-red-700 px-4 py-2 text-sm font-semibold text-white">Try again</button>
          </div>
        ) : items.length === 0 ? <p className="rounded-2xl bg-slate-100 p-5 text-slate-600">No dishes match your filters.</p> : (
          <>
            {!hasActiveFilters && recentItems.length > 0 ? <section className="mb-10" aria-labelledby="recent-heading">
              <div className="mb-5 flex flex-col gap-2 sm:flex-row sm:items-end sm:justify-between">
                <div><p className="text-xs font-bold uppercase tracking-[0.18em] text-[var(--gold-600)]">Welcome back</p><h2 id="recent-heading" className="mt-1 max-w-lg font-serif text-2xl font-black leading-tight text-[var(--maroon-900)] sm:text-3xl">Pick up where you left off</h2></div>
                <span className="w-fit rounded-full bg-[var(--cream-deep)] px-3 py-1.5 text-xs font-semibold text-[var(--muted)]">Saved on this device</span>
              </div>
              <div className="grid gap-5 sm:grid-cols-2 lg:grid-cols-4">{recentItems.map((item) => <MenuCard key={`recent-${item.id}`} item={item} />)}</div>
            </section> : null}
            <div className="grid gap-5 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
              {items.map((item) => <MenuCard key={item.id} item={item} />)}
            </div>
          </>
        )}
      </main>
      {itemCount > 0 ? (
        <Link to="/cart" className="fixed inset-x-4 bottom-4 z-30 flex min-h-14 items-center justify-between rounded-2xl bg-[var(--maroon-950)] px-5 py-3 text-white shadow-[var(--shadow-lifted)] transition hover:bg-[var(--maroon-800)] active:scale-[0.99] sm:left-auto sm:right-6 sm:w-80">
          <span className="font-black text-[var(--gold-200)]">{itemCount} item{itemCount === 1 ? "" : "s"} in your cart</span>
          <span className="font-black text-[var(--gold-200)]">₹{subtotal} · View</span>
        </Link>
      ) : null}
      <Footer />
    </div>
  );
}
