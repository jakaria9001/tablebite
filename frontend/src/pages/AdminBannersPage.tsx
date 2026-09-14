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

const FIXED_CTA_LABEL = "Explore Menu";
const FIXED_CTA_URL = "/menu";

function localDateTimeValue(date: Date) {
  const offset = date.getTimezoneOffset();
  return new Date(date.getTime() - offset * 60_000).toISOString().slice(0, 16);
}

function startOfTodayValue() {
  const today = new Date();
  return localDateTimeValue(new Date(today.getFullYear(), today.getMonth(), today.getDate()));
}

export default function AdminBannersPage() {
  const navigate = useNavigate();
  const [banners, setBanners] = useState<Banner[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [draggingId, setDraggingId] = useState<number | null>(null);
  const [feedback, setFeedback] = useState<string | null>(null);
  const [form, setForm] = useState<BannerPayload>(emptyForm());
  const [editingId, setEditingId] = useState<number | null>(null);
  const [originalStartAt, setOriginalStartAt] = useState("");
  const [formOpen, setFormOpen] = useState(false);
  const minimumStart = startOfTodayValue();

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
    setOriginalStartAt("");
    setFormOpen(false);
  };

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault();
    if (!form.title.trim() || !form.subtitle.trim()) {
      setFeedback("Please provide a title and subtitle.");
      return;
    }
    if (form.starts_at && form.starts_at < minimumStart && (!editingId || form.starts_at !== originalStartAt)) {
      setFeedback("Banner start date cannot be before today.");
      return;
    }
    if (form.starts_at && form.ends_at && form.ends_at <= form.starts_at) {
      setFeedback("Banner end date must be after the start date.");
      return;
    }

    setSaving(true);
    setFeedback(null);
    try {
      const payload = { ...form, cta_label: FIXED_CTA_LABEL, cta_url: FIXED_CTA_URL };
      if (editingId) {
        await updateBanner(editingId, { ...payload, id: editingId });
        setFeedback("Banner updated.");
      } else {
        await createBanner(payload);
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
      cta_label: FIXED_CTA_LABEL,
      cta_url: FIXED_CTA_URL,
      starts_at: banner.starts_at ?? "",
      ends_at: banner.ends_at ?? "",
      is_active: banner.is_active,
    });
    setOriginalStartAt(banner.starts_at ?? "");
    setFormOpen(true);
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

          <div>
            <section className="rounded-[2rem] border border-black/5 bg-white p-6 shadow-sm">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <div><h2 className="text-xl font-black">Banners</h2><p className="mt-1 text-sm text-slate-600">Drag to reorder. Active banners are shown on the homepage.</p></div>
                  <button type="button" onClick={() => { setForm(emptyForm()); setEditingId(null); setFormOpen(true); }} className="min-h-11 rounded-full bg-[var(--maroon-800)] px-5 py-2 text-sm font-bold text-white">Add banner</button>
                </div>
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

            {formOpen ? <div className="fixed inset-0 z-50 overflow-y-auto bg-[var(--maroon-950)]/70 px-4 py-6 backdrop-blur-sm sm:py-10" role="dialog" aria-modal="true" aria-labelledby="banner-form-title">
            <section className="mx-auto max-w-xl rounded-[2rem] border border-[var(--line)] bg-[var(--cream)] p-6 shadow-[var(--shadow-lifted)] sm:p-8">
              <div className="flex items-center justify-between">
                <h2 id="banner-form-title" className="text-xl font-black text-[var(--maroon-900)]">{editingId ? "Edit Banner" : "Add Banner"}</h2>
                <button type="button" onClick={resetForm} aria-label="Close banner form" className="grid h-11 w-11 place-items-center rounded-full border border-[var(--line)] bg-white text-xl">×</button>
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
                <div className="rounded-2xl border border-[var(--gold-200)] bg-[var(--cream-deep)] px-4 py-3 text-sm text-[var(--maroon-900)]">
                  The banner button is fixed to <strong>Explore Menu</strong> and opens <strong>menu directly</strong>.
                </div>
                <div className="grid gap-4 md:grid-cols-2">
                  <div>
                    <label className="text-sm font-semibold text-slate-700">Start date</label>
                    <input type="datetime-local" min={minimumStart} value={form.starts_at} onChange={(event) => setForm((current) => ({ ...current, starts_at: event.target.value }))} className="mt-2 w-full rounded-full border border-slate-200 px-4 py-3 text-sm outline-none" />
                  </div>
                  <div>
                    <label className="text-sm font-semibold text-slate-700">End date</label>
                    <input type="datetime-local" min={form.starts_at || minimumStart} value={form.ends_at} onChange={(event) => setForm((current) => ({ ...current, ends_at: event.target.value }))} className="mt-2 w-full rounded-full border border-slate-200 px-4 py-3 text-sm outline-none" />
                  </div>
                </div>
                <label className="flex items-center gap-2 rounded-full border border-slate-200 px-4 py-3 text-sm font-semibold text-slate-700">
                  <input type="checkbox" checked={form.is_active} onChange={(event) => setForm((current) => ({ ...current, is_active: event.target.checked }))} />
                  Active
                </label>
                <div className="flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
                <button type="button" onClick={resetForm} className="min-h-12 rounded-full border border-[var(--line)] bg-white px-5 py-3 text-sm font-bold">Cancel</button>
                <button type="submit" disabled={saving} className="min-h-12 rounded-full bg-[var(--maroon-800)] px-6 py-3 text-sm font-bold text-white transition hover:bg-[var(--maroon-700)] disabled:opacity-60">
                  {saving ? "Saving..." : editingId ? "Save Changes" : "Save"}
                </button>
                </div>
              </form>
            </section>
            </div> : null}
          </div>
        </div>
      </main>
      <Footer />
    </div>
  );
}
