import { useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { createAdminMenuItem, deleteAdminMenuItem, updateAdminMenuItem } from "../api/adminMenu";
import { Footer } from "../components/layout/Footer";
import { Header } from "../components/layout/Header";
import { ImageUploadField } from "../components/common/ImageUploadField";
import { useAdminMenu } from "../hooks/useAdminMenu";
import { markAdminDataUpdated } from "../hooks/useAdminDashboardSummary";
import type { AdminMenuItem, AdminMenuItemPayload } from "../types/adminMenu";

const emptyForm = (): AdminMenuItemPayload => ({
  name: "",
  description: "",
  category_id: 0,
  price: null,
  price_half: null,
  price_full: null,
  is_veg: true,
  is_available: true,
  is_featured: false,
  is_bestseller: false,
  is_top10: false,
  sort_order: 0,
  image_url: "",
});

export default function AdminMenuPage() {
  const navigate = useNavigate();
  const { data, loading, error, refresh } = useAdminMenu();
  const [query, setQuery] = useState("");
  const [selectedCategory, setSelectedCategory] = useState("all");
  const [form, setForm] = useState<AdminMenuItemPayload>(emptyForm());
  const [editingId, setEditingId] = useState<number | null>(null);
  const [formOpen, setFormOpen] = useState(false);
  const [saving, setSaving] = useState(false);
  const [busyId, setBusyId] = useState<number | null>(null);
  const [feedback, setFeedback] = useState<string | null>(null);
  const [page, setPage] = useState(1);
  const pageSize = 8;

  const filteredItems = useMemo(() => {
    const items = data?.items ?? [];
    const needle = query.trim().toLowerCase();
    return items.filter((item) => {
      const matchesQuery = !needle || [item.name, item.description].some((value) => value.toLowerCase().includes(needle));
      const matchesCategory = selectedCategory === "all" || item.category_id.toString() === selectedCategory;
      return matchesQuery && matchesCategory;
    });
  }, [data, query, selectedCategory]);

  const pageCount = Math.max(1, Math.ceil(filteredItems.length / pageSize));
  const safePage = Math.min(page, pageCount);
  const pagedItems = useMemo(() => {
    const start = (safePage - 1) * pageSize;
    return filteredItems.slice(start, start + pageSize);
  }, [filteredItems, safePage]);

  const resetForm = () => {
    setForm(emptyForm());
    setEditingId(null);
    setFormOpen(false);
  };

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault();
    if (!form.name.trim()) {
      setFeedback("Please enter a menu item name.");
      return;
    }

    setSaving(true);
    setFeedback(null);

    try {
      if (editingId) {
        await updateAdminMenuItem(editingId, form);
        setFeedback("Menu item updated.");
      } else {
        await createAdminMenuItem(form);
        setFeedback("Menu item created.");
      }
      resetForm();
      await refresh();
      markAdminDataUpdated();
    } catch (cause) {
      setFeedback(cause instanceof Error ? cause.message : "Unable to save menu item");
    } finally {
      setSaving(false);
    }
  };

  const startEdit = (item: AdminMenuItem) => {
    setEditingId(item.id);
    setForm({
      id: item.id,
      name: item.name,
      description: item.description,
      category_id: item.category_id,
      price: item.price,
      price_half: item.price_half,
      price_full: item.price_full,
      is_veg: item.is_veg,
      is_available: item.is_available,
      is_featured: item.is_featured,
      is_bestseller: item.is_bestseller,
      is_top10: item.is_top10,
      sort_order: item.sort_order,
      image_url: item.image_url ?? "",
    });
    setFormOpen(true);
  };

  const toggleStatus = async (item: AdminMenuItem, key: "is_available" | "is_featured" | "is_bestseller" | "is_top10") => {
    setBusyId(item.id);
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
        is_available: key === "is_available" ? !item.is_available : item.is_available,
        is_featured: key === "is_featured" ? !item.is_featured : item.is_featured,
        is_bestseller: key === "is_bestseller" ? !item.is_bestseller : item.is_bestseller,
        is_top10: key === "is_top10" ? !item.is_top10 : item.is_top10,
        sort_order: item.sort_order,
        image_url: item.image_url ?? "",
      });
      await refresh();
      markAdminDataUpdated();
    } catch (cause) {
      setFeedback(cause instanceof Error ? cause.message : "Unable to update item");
    } finally {
      setBusyId(null);
    }
  };

  const handleDelete = async (item: AdminMenuItem) => {
    if (!window.confirm(`Delete ${item.name}?`)) return;
    setBusyId(item.id);
    try {
      await deleteAdminMenuItem(item.id);
      await refresh();
      markAdminDataUpdated();
      setFeedback("Menu item deleted.");
    } catch (cause) {
      setFeedback(cause instanceof Error ? cause.message : "Unable to delete item");
    } finally {
      setBusyId(null);
    }
  };

  const goToPage = (nextPage: number) => {
    setPage(Math.min(Math.max(1, nextPage), pageCount));
  };

  return (
    <div className="min-h-screen bg-[#fffaf5] text-slate-900">
      <Header />
      <main className="px-5 py-10 sm:px-8 lg:px-10">
        <div className="mx-auto flex max-w-7xl flex-col gap-6">
          <div className="flex flex-wrap items-center justify-between gap-4 rounded-[2rem] border border-black/5 bg-white p-6 shadow-sm">
            <div>
              <p className="text-sm font-semibold uppercase tracking-[0.3em] text-orange-600">Menu Management</p>
              <h1 className="text-3xl font-black tracking-tight">Manage your restaurant menu</h1>
            </div>
            <button
              type="button"
              onClick={() => navigate("/admin")}
              className="rounded-full border border-slate-200 px-4 py-2 text-sm font-semibold text-slate-900 transition hover:border-orange-300 hover:text-orange-600"
            >
              Back to Dashboard
            </button>
          </div>

          <section className="rounded-[2rem] border border-black/5 bg-white p-6 shadow-sm">
              <div className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
                <div className="flex-1">
                  <label className="text-sm font-semibold text-slate-700">Search</label>
                  <input
                    value={query}
                    onChange={(event) => setQuery(event.target.value)}
                    placeholder="Search menu items"
                    className="mt-2 w-full rounded-full border border-slate-200 px-4 py-3 text-sm outline-none ring-0"
                  />
                </div>
                <div className="flex min-w-[220px] flex-col gap-2 sm:flex-row sm:items-end">
                  <div className="flex-1">
                  <label className="text-sm font-semibold text-slate-700">Filter by category</label>
                  <select
                    value={selectedCategory}
                    onChange={(event) => setSelectedCategory(event.target.value)}
                    className="mt-2 w-full rounded-full border border-slate-200 px-4 py-3 text-sm outline-none"
                  >
                    <option value="all">All categories</option>
                    {(data?.categories ?? []).map((category) => (
                      <option key={category.id} value={category.id}>
                        {category.name}
                      </option>
                    ))}
                  </select>
                  </div>
                  <button type="button" onClick={() => { setForm(emptyForm()); setEditingId(null); setFormOpen(true); }} className="min-h-11 rounded-full bg-[var(--maroon-800)] px-5 py-3 text-sm font-bold text-white transition hover:bg-[var(--maroon-700)] active:scale-[0.98]">
                    Add Item
                  </button>
                </div>
              </div>

              {feedback ? <p className="mt-4 rounded-full bg-slate-100 px-4 py-2 text-sm text-slate-700">{feedback}</p> : null}

              {loading ? (
                <div className="mt-6 rounded-[1.5rem] border border-slate-200 bg-slate-50 p-6 text-sm text-slate-600">Loading menu...</div>
              ) : error ? (
                <div className="mt-6 rounded-[1.5rem] border border-rose-200 bg-rose-50 p-6 text-sm text-rose-700">{error}</div>
              ) : (
                <>
                  <div className="mt-6 flex flex-wrap items-center justify-between gap-3 rounded-full border border-slate-200 bg-slate-50 px-4 py-3 text-sm text-slate-600">
                    <span>
                      Showing {filteredItems.length === 0 ? 0 : (safePage - 1) * pageSize + 1}-{Math.min(safePage * pageSize, filteredItems.length)} of {filteredItems.length} items
                    </span>
                    <div className="flex items-center gap-2">
                      <button type="button" onClick={() => goToPage(safePage - 1)} disabled={safePage === 1} className="rounded-full border border-slate-200 px-3 py-1.5 font-semibold text-slate-700 disabled:cursor-not-allowed disabled:opacity-50">Prev</button>
                      <span className="rounded-full bg-white px-3 py-1.5 font-semibold text-slate-700">{safePage}/{pageCount}</span>
                      <button type="button" onClick={() => goToPage(safePage + 1)} disabled={safePage === pageCount} className="rounded-full border border-slate-200 px-3 py-1.5 font-semibold text-slate-700 disabled:cursor-not-allowed disabled:opacity-50">Next</button>
                    </div>
                  </div>
                  <div className="mt-4 space-y-4">
                    {pagedItems.map((item) => (
                    <div key={item.id} className="rounded-[1.5rem] border border-slate-200 p-5">
                      <div className="flex flex-col gap-4 md:flex-row md:items-start md:justify-between">
                        <div>
                          <div className="flex flex-wrap items-center gap-2">
                            <h2 className="text-lg font-black">{item.name}</h2>
                            <span className="rounded-full bg-slate-100 px-3 py-1 text-xs font-semibold uppercase tracking-[0.3em] text-slate-700">
                              {item.is_veg ? "Veg" : "Non-Veg"}
                            </span>
                            {!item.is_available ? <span className="rounded-full bg-rose-100 px-3 py-1 text-xs font-semibold uppercase tracking-[0.3em] text-rose-700">Sold Out</span> : null}
                            {item.is_featured ? <span className="rounded-full bg-amber-100 px-3 py-1 text-xs font-semibold uppercase tracking-[0.3em] text-amber-700">Featured</span> : null}
                            {item.is_bestseller ? <span className="rounded-full bg-emerald-100 px-3 py-1 text-xs font-semibold uppercase tracking-[0.3em] text-emerald-700">Bestseller</span> : null}
                            {item.is_top10 ? <span className="rounded-full bg-sky-100 px-3 py-1 text-xs font-semibold uppercase tracking-[0.3em] text-sky-700">Top 10</span> : null}
                          </div>
                          <p className="mt-2 text-sm text-slate-600">{item.description || "No description"}</p>
                          <p className="mt-2 text-sm font-semibold text-slate-900">
                            {item.price ? `₹${item.price}` : ""}
                            {item.price_half ? ` • Half ₹${item.price_half}` : ""}
                            {item.price_full ? ` • Full ₹${item.price_full}` : ""}
                          </p>
                        </div>
                        <div className="flex flex-wrap gap-2">
                          <button type="button" onClick={() => startEdit(item)} className="rounded-full border border-slate-200 px-3 py-2 text-sm font-semibold text-slate-900 transition hover:border-orange-300 hover:text-orange-600">Edit</button>
                          <button type="button" onClick={() => toggleStatus(item, "is_available")} disabled={busyId === item.id} className="rounded-full border border-slate-200 px-3 py-2 text-sm font-semibold text-slate-900 transition hover:border-orange-300 hover:text-orange-600 disabled:opacity-60">
                            {item.is_available ? "Mark Sold Out" : "Mark Available"}
                          </button>
                          <button type="button" onClick={() => toggleStatus(item, "is_featured")} disabled={busyId === item.id} className="rounded-full border border-slate-200 px-3 py-2 text-sm font-semibold text-slate-900 transition hover:border-orange-300 hover:text-orange-600 disabled:opacity-60">
                            {item.is_featured ? "Remove Featured" : "Featured"}
                          </button>
                          <button type="button" onClick={() => toggleStatus(item, "is_bestseller")} disabled={busyId === item.id} className="rounded-full border border-slate-200 px-3 py-2 text-sm font-semibold text-slate-900 transition hover:border-orange-300 hover:text-orange-600 disabled:opacity-60">
                            {item.is_bestseller ? "Remove Bestseller" : "Bestseller"}
                          </button>
                          <button type="button" onClick={() => toggleStatus(item, "is_top10")} disabled={busyId === item.id} className="rounded-full border border-slate-200 px-3 py-2 text-sm font-semibold text-slate-900 transition hover:border-orange-300 hover:text-orange-600 disabled:opacity-60">
                            {item.is_top10 ? "Remove Top 10" : "Top 10"}
                          </button>
                          <button type="button" onClick={() => handleDelete(item)} disabled={busyId === item.id} className="rounded-full border border-rose-200 px-3 py-2 text-sm font-semibold text-rose-700 transition hover:border-rose-300 hover:bg-rose-50 disabled:opacity-60">
                            Delete
                          </button>
                        </div>
                      </div>
                    </div>
                    ))}
                  </div>
                </>
              )}
          </section>

          {formOpen ? <div className="fixed inset-0 z-50 overflow-y-auto bg-[var(--maroon-950)]/70 px-4 py-6 backdrop-blur-sm sm:py-10" role="dialog" aria-modal="true" aria-labelledby="menu-form-title">
            <section className="mx-auto max-w-2xl rounded-[2rem] border border-[var(--line)] bg-[var(--cream)] p-6 shadow-[var(--shadow-lifted)] sm:p-8">
              <div className="flex items-center justify-between gap-4">
                <div>
                  <p className="text-xs font-bold uppercase tracking-[0.2em] text-[var(--gold-600)]">Menu management</p>
                  <h2 id="menu-form-title" className="mt-1 text-2xl font-black text-[var(--maroon-900)]">{editingId ? "Edit item" : "Add item"}</h2>
                </div>
                <button type="button" onClick={resetForm} aria-label="Close item form" className="grid h-11 w-11 place-items-center rounded-full border border-[var(--line)] bg-white text-xl text-[var(--maroon-900)] transition hover:border-[var(--gold-500)]" >×</button>
              </div>
              <form className="mt-6 space-y-4" onSubmit={handleSubmit}>
                <div>
                  <label className="text-sm font-semibold text-slate-700">Name</label>
                  <input value={form.name} onChange={(event) => setForm((current) => ({ ...current, name: event.target.value }))} className="mt-2 w-full rounded-full border border-slate-200 px-4 py-3 text-sm outline-none" required />
                </div>
                <div>
                  <label className="text-sm font-semibold text-slate-700">Description</label>
                  <textarea value={form.description} onChange={(event) => setForm((current) => ({ ...current, description: event.target.value }))} className="mt-2 min-h-[100px] w-full rounded-[1.25rem] border border-slate-200 px-4 py-3 text-sm outline-none" />
                </div>
                <div>
                  <label className="text-sm font-semibold text-slate-700">Category</label>
                  <select value={form.category_id} onChange={(event) => setForm((current) => ({ ...current, category_id: Number(event.target.value) }))} className="mt-2 w-full rounded-full border border-slate-200 px-4 py-3 text-sm outline-none" required>
                    <option value={0}>Select category</option>
                    {(data?.categories ?? []).map((category) => (
                      <option key={category.id} value={category.id}>{category.name}</option>
                    ))}
                  </select>
                </div>
                <div>
                  <label className="text-sm font-semibold text-slate-700">Veg / Non-Veg</label>
                  <select value={form.is_veg ? "veg" : "non-veg"} onChange={(event) => setForm((current) => ({ ...current, is_veg: event.target.value === "veg" }))} className="mt-2 w-full rounded-full border border-slate-200 px-4 py-3 text-sm outline-none">
                    <option value="veg">Veg</option>
                    <option value="non-veg">Non-Veg</option>
                  </select>
                </div>

                <div className="grid gap-4 md:grid-cols-3">
                  <div>
                    <label className="text-sm font-semibold text-slate-700">Regular</label>
                    <input type="number" step="0.01" value={form.price ?? ""} onChange={(event) => setForm((current) => ({ ...current, price: event.target.value ? Number(event.target.value) : null }))} className="mt-2 w-full rounded-full border border-slate-200 px-4 py-3 text-sm outline-none" />
                  </div>
                  <div>
                    <label className="text-sm font-semibold text-slate-700">Half</label>
                    <input type="number" step="0.01" value={form.price_half ?? ""} onChange={(event) => setForm((current) => ({ ...current, price_half: event.target.value ? Number(event.target.value) : null }))} className="mt-2 w-full rounded-full border border-slate-200 px-4 py-3 text-sm outline-none" />
                  </div>
                  <div>
                    <label className="text-sm font-semibold text-slate-700">Full</label>
                    <input type="number" step="0.01" value={form.price_full ?? ""} onChange={(event) => setForm((current) => ({ ...current, price_full: event.target.value ? Number(event.target.value) : null }))} className="mt-2 w-full rounded-full border border-slate-200 px-4 py-3 text-sm outline-none" />
                  </div>
                </div>

                <ImageUploadField
                  value={form.image_url ?? ""}
                  onChange={(value) => setForm((current) => ({ ...current, image_url: value }))}
                  label="Image URL"
                  placeholder="https://..."
                />

                <div className="grid gap-4 sm:grid-cols-2">
                  <label className="flex items-center gap-2 rounded-full border border-slate-200 px-4 py-3 text-sm font-semibold text-slate-700">
                    <input type="checkbox" checked={form.is_available} onChange={(event) => setForm((current) => ({ ...current, is_available: event.target.checked }))} />
                    Available
                  </label>
                  <label className="flex items-center gap-2 rounded-full border border-slate-200 px-4 py-3 text-sm font-semibold text-slate-700">
                    <input type="checkbox" checked={form.is_featured} onChange={(event) => setForm((current) => ({ ...current, is_featured: event.target.checked }))} />
                    Featured
                  </label>
                  <label className="flex items-center gap-2 rounded-full border border-slate-200 px-4 py-3 text-sm font-semibold text-slate-700">
                    <input type="checkbox" checked={form.is_bestseller} onChange={(event) => setForm((current) => ({ ...current, is_bestseller: event.target.checked }))} />
                    Bestseller
                  </label>
                  <label className="flex items-center gap-2 rounded-full border border-slate-200 px-4 py-3 text-sm font-semibold text-slate-700">
                    <input type="checkbox" checked={form.is_top10} onChange={(event) => setForm((current) => ({ ...current, is_top10: event.target.checked }))} />
                    Top 10
                  </label>
                </div>

                <div>
                  <label className="text-sm font-semibold text-slate-700">Display Order</label>
                  <input type="number" value={form.sort_order} onChange={(event) => setForm((current) => ({ ...current, sort_order: Number(event.target.value) }))} className="mt-2 w-full rounded-full border border-slate-200 px-4 py-3 text-sm outline-none" />
                </div>

                <div className="flex flex-col-reverse gap-2 pt-2 sm:flex-row sm:justify-end">
                  <button type="button" onClick={resetForm} className="min-h-12 rounded-full border border-[var(--line)] bg-white px-5 py-3 text-sm font-bold text-[var(--ink)]">Cancel</button>
                  <button type="submit" disabled={saving} className="min-h-12 rounded-full bg-[var(--maroon-800)] px-6 py-3 text-sm font-bold text-white transition hover:bg-[var(--maroon-700)] disabled:cursor-wait disabled:opacity-60">
                    {saving ? "Saving..." : editingId ? "Save changes" : "Add item"}
                  </button>
                </div>
              </form>
            </section>
          </div> : null}
        </div>
      </main>
      <Footer />
    </div>
  );
}
