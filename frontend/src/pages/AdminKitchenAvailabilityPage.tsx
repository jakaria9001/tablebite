import { useEffect, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { updateAdminMenuItem } from "../api/adminMenu";
import { Footer } from "../components/layout/Footer";
import { Header } from "../components/layout/Header";
import { useAdminMenu } from "../hooks/useAdminMenu";
import type { AdminMenuItem } from "../types/adminMenu";

export default function AdminKitchenAvailabilityPage() {
  const navigate = useNavigate();
  const { data, loading, error, refresh } = useAdminMenu();
  const [busyId, setBusyId] = useState<number | null>(null);
  const [feedback, setFeedback] = useState<string | null>(null);
  const [query, setQuery] = useState("");
  const [statusFilter, setStatusFilter] = useState<"all" | "available" | "soldout">("all");
  const [page, setPage] = useState(1);
  const pageSize = 10;

  const items = useMemo(() => {
    return [...(data?.items ?? [])].sort((left, right) => left.name.localeCompare(right.name));
  }, [data?.items]);

  const filteredItems = useMemo(() => {
    const needle = query.trim().toLowerCase();
    return items.filter((item) => {
      const matchesQuery = !needle || item.name.toLowerCase().includes(needle);
      const matchesStatus =
        statusFilter === "all" || (statusFilter === "available" && item.is_available) || (statusFilter === "soldout" && !item.is_available);
      return matchesQuery && matchesStatus;
    });
  }, [items, query, statusFilter]);

  const totalPages = Math.max(1, Math.ceil(filteredItems.length / pageSize));
  const visibleItems = useMemo(() => {
    const start = (page - 1) * pageSize;
    return filteredItems.slice(start, start + pageSize);
  }, [filteredItems, page]);

  useEffect(() => {
    setPage(1);
  }, [query, statusFilter]);

  const availableCount = items.filter((item) => item.is_available).length;
  const soldOutCount = items.length - availableCount;

  const toggleAvailability = async (item: AdminMenuItem) => {
    setBusyId(item.id);
    setFeedback(null);

    try {
      await updateAdminMenuItem(item.id, {
        id: item.id,
        name: item.name,
        description: item.description,
        category_id: item.category_id,
        price: item.price,
        price_half: item.price_half,
        price_full: item.price_full,
        is_veg: item.is_veg,
        is_available: !item.is_available,
        is_featured: item.is_featured,
        is_bestseller: item.is_bestseller,
        is_top10: item.is_top10,
        sort_order: item.sort_order,
        image_url: item.image_url ?? "",
      });
      await refresh();
    } catch (cause) {
      setFeedback(cause instanceof Error ? cause.message : "Unable to update availability");
    } finally {
      setBusyId(null);
    }
  };

  return (
    <div className="min-h-screen bg-[#fffaf5] text-slate-900">
      <Header />
      <main className="px-5 py-10 sm:px-8 lg:px-10">
        <div className="mx-auto flex max-w-5xl flex-col gap-6">
          <div className="flex flex-wrap items-center justify-between gap-4 rounded-[2rem] border border-black/5 bg-white p-6 shadow-sm">
            <div>
              <p className="text-sm font-semibold uppercase tracking-[0.3em] text-orange-600">Kitchen Mode</p>
              <h1 className="text-3xl font-black tracking-tight">Availability</h1>
              <p className="mt-2 text-sm text-slate-600">Switch dishes between available and sold out in a single tap.</p>
            </div>
            <button
              type="button"
              onClick={() => navigate("/admin")}
              className="rounded-full border border-slate-200 px-4 py-2 text-sm font-semibold text-slate-900 transition hover:border-orange-300 hover:text-orange-600"
            >
              Back to Dashboard
            </button>
          </div>

          <div className="grid gap-4 sm:grid-cols-2">
            <div className="rounded-[1.5rem] border border-emerald-200 bg-emerald-50 p-5">
              <p className="text-sm font-semibold uppercase tracking-[0.3em] text-emerald-700">Available</p>
              <p className="mt-3 text-3xl font-black text-emerald-800">{availableCount}</p>
            </div>
            <div className="rounded-[1.5rem] border border-rose-200 bg-rose-50 p-5">
              <p className="text-sm font-semibold uppercase tracking-[0.3em] text-rose-700">Sold Out</p>
              <p className="mt-3 text-3xl font-black text-rose-800">{soldOutCount}</p>
            </div>
          </div>

          {feedback ? <div className="rounded-full bg-slate-100 px-4 py-2 text-sm text-slate-700">{feedback}</div> : null}

          <div className="rounded-[2rem] border border-black/5 bg-white p-6 shadow-sm">
            <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
              <div className="flex-1">
                <label className="text-sm font-semibold text-slate-700">Search</label>
                <input
                  value={query}
                  onChange={(event) => setQuery(event.target.value)}
                  placeholder="Search dish"
                  className="mt-2 w-full rounded-full border border-slate-200 px-4 py-3 text-sm outline-none"
                />
              </div>
              <div className="min-w-[220px]">
                <label className="text-sm font-semibold text-slate-700">Filter</label>
                <select
                  value={statusFilter}
                  onChange={(event) => setStatusFilter(event.target.value as "all" | "available" | "soldout")}
                  className="mt-2 w-full rounded-full border border-slate-200 px-4 py-3 text-sm outline-none"
                >
                  <option value="all">All items</option>
                  <option value="available">Available</option>
                  <option value="soldout">Sold out</option>
                </select>
              </div>
            </div>

            {loading ? (
              <div className="mt-6 rounded-[1.5rem] border border-slate-200 bg-slate-50 p-6 text-sm text-slate-600">Loading availability list...</div>
            ) : error ? (
              <div className="mt-6 rounded-[1.5rem] border border-rose-200 bg-rose-50 p-6 text-sm text-rose-700">{error}</div>
            ) : (
              <div className="mt-6 space-y-3">
                {visibleItems.map((item) => (
                  <div key={item.id} className="flex flex-col gap-3 rounded-[1.25rem] border border-slate-200 p-4 sm:flex-row sm:items-center sm:justify-between">
                    <div className="flex-1">
                      <div className="flex items-center gap-2">
                        <span className={`h-3 w-3 rounded-full ${item.is_available ? "bg-emerald-500" : "bg-rose-500"}`} />
                        <h2 className="text-lg font-black">{item.name}</h2>
                      </div>
                      <p className="mt-1 text-sm text-slate-600">{item.description || "No description"}</p>
                    </div>
                    <button
                      type="button"
                      onClick={() => toggleAvailability(item)}
                      disabled={busyId === item.id}
                      className={`rounded-full px-4 py-2 text-sm font-bold uppercase tracking-[0.2em] transition ${item.is_available ? "bg-emerald-600 text-white hover:bg-emerald-700" : "bg-rose-600 text-white hover:bg-rose-700"} disabled:opacity-70`}
                    >
                      {busyId === item.id ? "Updating..." : item.is_available ? "Available" : "Sold Out"}
                    </button>
                  </div>
                ))}

                {filteredItems.length > pageSize ? (
                  <div className="flex flex-wrap items-center justify-between gap-3 rounded-[1rem] border border-slate-200 bg-slate-50 px-4 py-3 text-sm text-slate-700">
                    <p>
                      Showing {Math.min((page - 1) * pageSize + 1, filteredItems.length)}-{Math.min(page * pageSize, filteredItems.length)} of {filteredItems.length}
                    </p>
                    <div className="flex items-center gap-2">
                      <button
                        type="button"
                        onClick={() => setPage((current) => Math.max(1, current - 1))}
                        disabled={page === 1}
                        className="rounded-full border border-slate-200 px-3 py-2 font-semibold disabled:cursor-not-allowed disabled:opacity-60"
                      >
                        Previous
                      </button>
                      <span className="font-semibold">Page {page} of {totalPages}</span>
                      <button
                        type="button"
                        onClick={() => setPage((current) => Math.min(totalPages, current + 1))}
                        disabled={page === totalPages}
                        className="rounded-full border border-slate-200 px-3 py-2 font-semibold disabled:cursor-not-allowed disabled:opacity-60"
                      >
                        Next
                      </button>
                    </div>
                  </div>
                ) : null}
              </div>
            )}
          </div>
        </div>
      </main>
      <Footer />
    </div>
  );
}
