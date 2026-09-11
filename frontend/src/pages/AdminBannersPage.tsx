import { useEffect, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { createBanner, deleteBanner, getAdminBanners, reorderBanners, updateBanner, type Banner, type BannerPayload } from "../api/banners";
import { Footer } from "../components/layout/Footer";
import { Header } from "../components/layout/Header";
import { ImageUploadField } from "../components/common/ImageUploadField";

const emptyForm = (): BannerPayload => ({
  title: "",
  subtitle: "",
  image_url: "",
  cta_label: "Explore Menu",
  cta_url: "/menu",
  starts_at: "",
  ends_at: "",
  is_active: true,
});

export default function AdminBannersPage() {
  const navigate = useNavigate();
  const [banners, setBanners] = useState<Banner[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [draggingId, setDraggingId] = useState<number | null>(null);
  const [feedback, setFeedback] = useState<string | null>(null);
  const [form, setForm] = useState<BannerPayload>(emptyForm());
  const [editingId, setEditingId] = useState<number | null>(null);

  const orderedBanners = useMemo(() => [...banners].sort((a, b) => a.display_order - b.display_order || a.id - b.id), [banners]);

  const refresh = async () => {
    setLoading(true);
    try {
      const data = await getAdminBanners();
      setBanners(data.banners);
    } catch (cause) {
      setFeedback(cause instanceof Error ? cause.message : "Unable to load banners");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void refresh();
  }, []);

  const resetForm = () => {
    setForm(emptyForm());
    setEditingId(null);
  };

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault();
    if (!form.title.trim() || !form.subtitle.trim()) {
      setFeedback("Please provide a title and subtitle.");
      return;
    }

    setSaving(true);
    setFeedback(null);
    try {
      if (editingId) {
        await updateBanner(editingId, { ...form, id: editingId });
        setFeedback("Banner updated.");
      } else {
        await createBanner(form);
        setFeedback("Banner created.");
      }
      resetForm();
      await refresh();
    } catch (cause) {
      setFeedback(cause instanceof Error ? cause.message : "Unable to save banner");
    } finally {
      setSaving(false);
    }
  };

  const startEdit = (banner: Banner) => {
    setEditingId(banner.id);
    setForm({
      id: banner.id,
      title: banner.title,
      subtitle: banner.subtitle,
      image_url: banner.image_url ?? "",
      cta_label: banner.cta_label ?? "",
      cta_url: banner.cta_url ?? "",
      starts_at: banner.starts_at ?? "",
      ends_at: banner.ends_at ?? "",
      is_active: banner.is_active,
    });
  };

  const handleDelete = async (banner: Banner) => {
    try {
      await deleteBanner(banner.id);
      await refresh();
      setFeedback("Banner deleted.");
    } catch (cause) {
      setFeedback(cause instanceof Error ? cause.message : "Unable to delete banner");
    }
  };

  const handleDrop = async (targetId: number) => {
    if (draggingId === null || draggingId === targetId) return;
    const next = [...orderedBanners];
    const fromIndex = next.findIndex((item) => item.id === draggingId);
    const toIndex = next.findIndex((item) => item.id === targetId);
    if (fromIndex < 0 || toIndex < 0) return;
    const [moved] = next.splice(fromIndex, 1);
    next.splice(toIndex, 0, moved);
    const reorderedIds = next.map((item) => item.id);
    try {
      await reorderBanners(reorderedIds);
      await refresh();
      setFeedback("Banner order updated.");
    } catch (cause) {
      setFeedback(cause instanceof Error ? cause.message : "Unable to reorder banners");
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
              <p className="text-sm font-semibold uppercase tracking-[0.3em] text-orange-600">Promotional Banner Management</p>
              <h1 className="text-3xl font-black tracking-tight">Create and schedule promotional banners for the homepage</h1>
            </div>
            <button type="button" onClick={() => navigate("/admin")} className="rounded-full border border-slate-200 px-4 py-2 text-sm font-semibold text-slate-900 transition hover:border-orange-300 hover:text-orange-600">Back to Dashboard</button>
          </div>

          <div className="grid gap-6 lg:grid-cols-[0.95fr_0.75fr]">
            <section className="rounded-[2rem] border border-black/5 bg-white p-6 shadow-sm">
              <div className="flex items-center justify-between">
                <h2 className="text-xl font-black">Banners</h2>
                <p className="text-sm text-slate-600">Drag to reorder. Active banners are shown on the homepage.</p>
              </div>
              {feedback ? <p className="mt-4 rounded-full bg-slate-100 px-4 py-2 text-sm text-slate-700">{feedback}</p> : null}
              {loading ? (
                <div className="mt-6 rounded-[1.5rem] border border-slate-200 bg-slate-50 p-6 text-sm text-slate-600">Loading banners...</div>
              ) : (
                <div className="mt-6 grid gap-4">
                  {orderedBanners.map((banner) => (
                    <div key={banner.id} draggable onDragStart={() => setDraggingId(banner.id)} onDragOver={(event) => event.preventDefault()} onDrop={() => handleDrop(banner.id)} className="rounded-[1.5rem] border border-slate-200 p-4">
                      <div className="flex flex-wrap items-center gap-2">
                        <span className={`rounded-full px-3 py-1 text-xs font-semibold uppercase tracking-[0.3em] ${banner.is_active ? "bg-emerald-100 text-emerald-700" : "bg-rose-100 text-rose-700"}`}>
                          {banner.is_active ? "Active" : "Disabled"}
                        </span>
                      </div>
                      <p className="mt-3 font-black">{banner.title}</p>
                      <p className="mt-2 text-sm text-slate-600">{banner.subtitle}</p>
                      <div className="mt-4 flex flex-wrap gap-2">
                        <button type="button" onClick={() => startEdit(banner)} className="rounded-full border border-slate-200 px-3 py-2 text-sm font-semibold text-slate-900 transition hover:border-orange-300 hover:text-orange-600">Edit</button>
                        <button type="button" onClick={() => handleDelete(banner)} className="rounded-full border border-slate-200 px-3 py-2 text-sm font-semibold text-slate-900 transition hover:border-orange-300 hover:text-orange-600">Delete</button>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </section>

            <section className="rounded-[2rem] border border-black/5 bg-white p-6 shadow-sm">
              <div className="flex items-center justify-between">
                <h2 className="text-xl font-black">{editingId ? "Edit Banner" : "Add Banner"}</h2>
                {editingId ? <button type="button" onClick={resetForm} className="text-sm font-semibold text-slate-600">Cancel</button> : null}
              </div>
              <form className="mt-6 space-y-4" onSubmit={handleSubmit}>
                <div>
                  <label className="text-sm font-semibold text-slate-700">Title</label>
                  <input value={form.title} onChange={(event) => setForm((current) => ({ ...current, title: event.target.value }))} className="mt-2 w-full rounded-full border border-slate-200 px-4 py-3 text-sm outline-none" required />
                </div>
                <div>
                  <label className="text-sm font-semibold text-slate-700">Subtitle</label>
                  <textarea value={form.subtitle} onChange={(event) => setForm((current) => ({ ...current, subtitle: event.target.value }))} className="mt-2 min-h-[90px] w-full rounded-[1.25rem] border border-slate-200 px-4 py-3 text-sm outline-none" required />
                </div>
                <ImageUploadField
                  value={form.image_url}
                  onChange={(value) => setForm((current) => ({ ...current, image_url: value }))}
                  label="Image URL"
                  placeholder="https://..."
                />
                <div>
                  <label className="text-sm font-semibold text-slate-700">CTA text</label>
                  <input value={form.cta_label} onChange={(event) => setForm((current) => ({ ...current, cta_label: event.target.value }))} className="mt-2 w-full rounded-full border border-slate-200 px-4 py-3 text-sm outline-none" />
                </div>
                <div>
                  <label className="text-sm font-semibold text-slate-700">CTA URL</label>
                  <input value={form.cta_url} onChange={(event) => setForm((current) => ({ ...current, cta_url: event.target.value }))} className="mt-2 w-full rounded-full border border-slate-200 px-4 py-3 text-sm outline-none" />
                </div>
                <div className="grid gap-4 md:grid-cols-2">
                  <div>
                    <label className="text-sm font-semibold text-slate-700">Start date</label>
                    <input type="datetime-local" value={form.starts_at} onChange={(event) => setForm((current) => ({ ...current, starts_at: event.target.value }))} className="mt-2 w-full rounded-full border border-slate-200 px-4 py-3 text-sm outline-none" />
                  </div>
                  <div>
                    <label className="text-sm font-semibold text-slate-700">End date</label>
                    <input type="datetime-local" value={form.ends_at} onChange={(event) => setForm((current) => ({ ...current, ends_at: event.target.value }))} className="mt-2 w-full rounded-full border border-slate-200 px-4 py-3 text-sm outline-none" />
                  </div>
                </div>
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
