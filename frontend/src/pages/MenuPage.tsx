import { useEffect, useMemo, useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { Footer } from "../components/layout/Footer";
import { Header } from "../components/layout/Header";
import { MenuCard } from "../components/menu/MenuCard";
import { useMenu } from "../hooks/useMenu";
import { SeoHead } from "../components/seo/SeoHead";

export default function MenuPage() {
  const { data, loading, error, retry } = useMenu();
  const location = useLocation();
  const navigate = useNavigate();
  const [search, setSearch] = useState("");
  const [vegOnly, setVegOnly] = useState(false);
  const [categoryId, setCategoryId] = useState<number | null>(null);

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
      if (!term) return true;
      return `${item.name} ${item.description}`.toLowerCase().includes(term);
    }) ?? [];
  }, [data, search, vegOnly, categoryId]);

  return (
    <div>
      <SeoHead title="Menu" description="Browse the digital menu at Indian Restaurant & Sweets with categories for starters, mains, beverages, and sweets." />
      <Header />
      <main className="mx-auto max-w-7xl px-5 py-10 pb-24 lg:px-8">
        <div className="mb-8">
          <p className="text-sm font-bold uppercase tracking-[0.2em] text-orange-600">Digital menu</p>
          <h1 className="mt-2 text-4xl font-black">Something delicious for everyone.</h1>
        </div>
        <div className="sticky top-[72px] z-20 mb-8 rounded-3xl border border-black/5 bg-white/95 p-4 shadow-sm backdrop-blur">
          <div className="flex flex-col gap-3 md:flex-row">
            <input value={search} onChange={(event) => setSearch(event.target.value)} placeholder="Search dishes…" className="min-w-0 flex-1 rounded-2xl bg-slate-100 px-4 py-3 outline-none ring-orange-300 focus:ring-2" />
            <button onClick={() => setVegOnly((value) => !value)} className={`rounded-2xl px-5 py-3 text-sm font-bold ${vegOnly ? "bg-emerald-100 text-emerald-800" : "bg-slate-100"}`}>● Veg only</button>
          </div>
          <div className="mt-3 flex gap-2 overflow-x-auto pb-1">
            <button onClick={() => {
              setCategoryId(null);
              window.history.replaceState(null, "", "/menu");
            }} className={`shrink-0 rounded-full px-4 py-2 text-sm font-bold ${categoryId === null ? "bg-slate-950 text-white" : "bg-slate-100"}`}>All</button>
            {data?.categories.map((category) => <button key={category.id} onClick={() => {
              setCategoryId(category.id);
              window.history.replaceState(null, "", `/menu#${category.slug}`);
            }} className={`shrink-0 rounded-full px-4 py-2 text-sm font-bold ${categoryId === category.id ? "bg-slate-950 text-white" : "bg-slate-100"}`}>{category.name}</button>)}
          </div>
        </div>

        {loading ? <p>Loading menu…</p> : error ? (
          <div className="rounded-2xl bg-red-50 p-5 text-red-700">
            <p>{error}</p>
            <button type="button" onClick={retry} className="mt-3 rounded-full bg-red-700 px-4 py-2 text-sm font-semibold text-white">Try again</button>
          </div>
        ) : items.length === 0 ? <p className="rounded-2xl bg-slate-100 p-5 text-slate-600">No dishes match your filters.</p> : (
          <div className="grid gap-5 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
            {items.map((item) => <MenuCard key={item.id} item={item} />)}
          </div>
        )}
      </main>
      <Footer />
    </div>
  );
}
