import { useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { createAdminCategory, reorderAdminCategories, updateAdminCategory } from "../api/adminCategories";
import { Footer } from "../components/layout/Footer";
import { Header } from "../components/layout/Header";
import { ImageUploadField } from "../components/common/ImageUploadField";
import { useAdminCategories } from "../hooks/useAdminCategories";
import type { AdminCategory, AdminCategoryPayload } from "../types/adminCategories";

const emptyForm = (): AdminCategoryPayload => ({
  name: "",
  slug: "",
  image_url: "",
  is_active: true,
});

export default function AdminCategoriesPage() {
  const navigate = useNavigate();
  const { categories, loading, error, refresh } = useAdminCategories();
  const [form, setForm] = useState<AdminCategoryPayload>(emptyForm());
  const [editingId, setEditingId] = useState<number | null>(null);
  const [saving, setSaving] = useState(false);
  const [draggingId, setDraggingId] = useState<number | null>(null);
  const [feedback, setFeedback] = useState<string | null>(null);

  const orderedCategories = useMemo(() => [...categories].sort((a, b) => a.display_order - b.display_order || a.id - b.id), [categories]);

  const resetForm = () => {
    setForm(emptyForm());
    setEditingId(null);
  };

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault();
    if (!form.name.trim()) {
      setFeedback("Please enter a category name.");
      return;
    }

    setSaving(true);
    setFeedback(null);
    try {
      if (editingId) {
        await updateAdminCategory(editingId, { ...form, id: editingId });
        setFeedback("Category updated.");
      } else {
        await createAdminCategory(form);
        setFeedback("Category created.");
      }
      resetForm();
      await refresh();
    } catch (cause) {
      setFeedback(cause instanceof Error ? cause.message : "Unable to save category");
    } finally {
      setSaving(false);
    }
  };

  const startEdit = (category: AdminCategory) => {
    setEditingId(category.id);
    setForm({
      id: category.id,
      name: category.name,
      slug: category.slug,
      image_url: category.image_url ?? "",
      is_active: category.is_active,
    });
  };

  const toggleActive = async (category: AdminCategory) => {
    try {
      await updateAdminCategory(category.id, {
        id: category.id,
        name: category.name,
        slug: category.slug,
        image_url: category.image_url ?? "",
        is_active: !category.is_active,
      });
      await refresh();
    } catch (cause) {
      setFeedback(cause instanceof Error ? cause.message : "Unable to update category");
    }
  };

  const handleDrop = async (targetId: number) => {
    if (draggingId === null || draggingId === targetId) return;
    const next = [...orderedCategories];
    const fromIndex = next.findIndex((item) => item.id === draggingId);
    const toIndex = next.findIndex((item) => item.id === targetId);
    if (fromIndex < 0 || toIndex < 0) return;
    const [moved] = next.splice(fromIndex, 1);
    next.splice(toIndex, 0, moved);
    const reorderedIds = next.map((item) => item.id);
    try {
      await reorderAdminCategories(reorderedIds);
      await refresh();
      setFeedback("Category order updated.");
    } catch (cause) {
      setFeedback(cause instanceof Error ? cause.message : "Unable to reorder categories");
    }
    setDraggingId(null);
  };

  return (
    <div className="min-h-screen bg-[#fffaf5] text-slate-900">
      <Header />
      <main className="px-5 py-10 sm:px-8 lg:px-10">
        <div className="mx-auto flex max-w-7xl flex-col gap-6">
          <div className="flex flex-wrap items-center justify-between gap-4 rounded-[2rem] border border-black/5 bg-white p-6 shadow-sm">
            <div>
              <p className="text-sm font-semibold uppercase tracking-[0.3em] text-orange-600">Category Management</p>
              <h1 className="text-3xl font-black tracking-tight">Organize category order and visibility</h1>
            </div>
            <button type="button" onClick={() => navigate("/admin")} className="rounded-full border border-slate-200 px-4 py-2 text-sm font-semibold text-slate-900 transition hover:border-orange-300 hover:text-orange-600">Back to Dashboard</button>
          </div>

          <div className="grid gap-6 lg:grid-cols-[0.95fr_0.75fr]">
            <section className="rounded-[2rem] border border-black/5 bg-white p-6 shadow-sm">
              <div className="flex items-center justify-between">
                <h2 className="text-xl font-black">Categories</h2>
                <p className="text-sm text-slate-600">Drag to reorder. The homepage carousel uses this order.</p>
              </div>
              {feedback ? <p className="mt-4 rounded-full bg-slate-100 px-4 py-2 text-sm text-slate-700">{feedback}</p> : null}
              {loading ? (
                <div className="mt-6 rounded-[1.5rem] border border-slate-200 bg-slate-50 p-6 text-sm text-slate-600">Loading categories...</div>
              ) : error ? (
                <div className="mt-6 rounded-[1.5rem] border border-rose-200 bg-rose-50 p-6 text-sm text-rose-700">{error}</div>
              ) : (
                <div className="mt-6 space-y-3">
                  {orderedCategories.map((category) => (
                    <div key={category.id} draggable onDragStart={() => setDraggingId(category.id)} onDragOver={(event) => event.preventDefault()} onDrop={() => handleDrop(category.id)} className="flex flex-col gap-3 rounded-[1.5rem] border border-slate-200 p-4 md:flex-row md:items-center md:justify-between">
                      <div className="flex items-center gap-3">
                        <div className="flex h-10 w-10 items-center justify-center rounded-full bg-slate-100 text-sm font-black text-slate-700">{String(category.display_order).padStart(2, "0")}</div>
                        <div>
                          <p className="font-black">{category.name}</p>
                          <p className="text-sm text-slate-600">{category.slug}</p>
                        </div>
                      </div>
                      <div className="flex flex-wrap items-center gap-2">
                        <span className={`rounded-full px-3 py-1 text-xs font-semibold uppercase tracking-[0.3em] ${category.is_active ? "bg-emerald-100 text-emerald-700" : "bg-rose-100 text-rose-700"}`}>
                          {category.is_active ? "Active" : "Disabled"}
                        </span>
                        <button type="button" onClick={() => startEdit(category)} className="rounded-full border border-slate-200 px-3 py-2 text-sm font-semibold text-slate-900 transition hover:border-orange-300 hover:text-orange-600">Edit</button>
                        <button type="button" onClick={() => toggleActive(category)} className="rounded-full border border-slate-200 px-3 py-2 text-sm font-semibold text-slate-900 transition hover:border-orange-300 hover:text-orange-600">{category.is_active ? "Disable" : "Activate"}</button>
                        <button type="button" onClick={() => setDraggingId(category.id)} className="rounded-full border border-slate-200 px-3 py-2 text-sm font-semibold text-slate-900 transition hover:border-orange-300 hover:text-orange-600">Reorder</button>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </section>

            <section className="rounded-[2rem] border border-black/5 bg-white p-6 shadow-sm">
              <div className="flex items-center justify-between">
                <h2 className="text-xl font-black">{editingId ? "Edit Category" : "Add Category"}</h2>
                {editingId ? <button type="button" onClick={resetForm} className="text-sm font-semibold text-slate-600">Cancel</button> : null}
              </div>

              <form className="mt-6 space-y-4" onSubmit={handleSubmit}>
                <div>
                  <label className="text-sm font-semibold text-slate-700">Name</label>
                  <input value={form.name} onChange={(event) => setForm((current) => ({ ...current, name: event.target.value }))} className="mt-2 w-full rounded-full border border-slate-200 px-4 py-3 text-sm outline-none" required />
                </div>
                <div>
                  <label className="text-sm font-semibold text-slate-700">Slug</label>
                  <input value={form.slug} onChange={(event) => setForm((current) => ({ ...current, slug: event.target.value }))} className="mt-2 w-full rounded-full border border-slate-200 px-4 py-3 text-sm outline-none" required />
                </div>
                <ImageUploadField
                  value={form.image_url ?? ""}
                  onChange={(value) => setForm((current) => ({ ...current, image_url: value }))}
                  label="Image URL"
                  placeholder="https://..."
                />
                <label className="flex items-center gap-2 rounded-full border border-slate-200 px-4 py-3 text-sm font-semibold text-slate-700">
                  <input type="checkbox" checked={form.is_active} onChange={(event) => setForm((current) => ({ ...current, is_active: event.target.checked }))} />
                  Active
                </label>
                <button type="submit" disabled={saving} className="w-full rounded-full bg-orange-600 px-4 py-3 text-sm font-semibold text-white transition hover:bg-orange-700 disabled:opacity-60">
                  {saving ? "Saving..." : editingId ? "Save Changes" : "Save"}
                </button>
              </form>
            </section>
          </div>
        </div>
      </main>
      <Footer />
    </div>
  );
}
