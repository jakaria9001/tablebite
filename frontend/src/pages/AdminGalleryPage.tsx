import { useEffect, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { createGalleryImage, deleteGalleryImage, getAdminGalleryImages, reorderGalleryImages, updateGalleryImage, type GalleryImage, type GalleryImagePayload } from "../api/gallery";
import { Footer } from "../components/layout/Footer";
import { Header } from "../components/layout/Header";
import { ImageUploadField } from "../components/common/ImageUploadField";

const emptyForm = (): GalleryImagePayload => ({
  image_url: "",
  caption: "",
  is_active: true,
  is_featured: false,
});

export default function AdminGalleryPage() {
  const navigate = useNavigate();
  const [images, setImages] = useState<GalleryImage[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [draggingId, setDraggingId] = useState<number | null>(null);
  const [feedback, setFeedback] = useState<string | null>(null);
  const [form, setForm] = useState<GalleryImagePayload>(emptyForm());
  const [editingId, setEditingId] = useState<number | null>(null);

  const orderedImages = useMemo(() => [...images].sort((a, b) => a.display_order - b.display_order || a.id - b.id), [images]);

  const refresh = async () => {
    setLoading(true);
    try {
      const data = await getAdminGalleryImages();
      setImages(data.images);
    } catch (cause) {
      setFeedback(cause instanceof Error ? cause.message : "Unable to load gallery");
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
    if (!form.image_url.trim()) {
      setFeedback("Please provide an image URL.");
      return;
    }
    setSaving(true);
    setFeedback(null);
    try {
      if (editingId) {
        await updateGalleryImage(editingId, { ...form, id: editingId });
        setFeedback("Image updated.");
      } else {
        await createGalleryImage(form);
        setFeedback("Image created.");
      }
      resetForm();
      await refresh();
    } catch (cause) {
      setFeedback(cause instanceof Error ? cause.message : "Unable to save image");
    } finally {
      setSaving(false);
    }
  };

  const startEdit = (image: GalleryImage) => {
    setEditingId(image.id);
    setForm({
      id: image.id,
      image_url: image.image_url,
      caption: image.caption,
      is_active: image.is_active,
      is_featured: image.is_featured,
    });
  };

  const handleDelete = async (image: GalleryImage) => {
    try {
      await deleteGalleryImage(image.id);
      await refresh();
      setFeedback("Image deleted.");
    } catch (cause) {
      setFeedback(cause instanceof Error ? cause.message : "Unable to delete image");
    }
  };

  const handleDrop = async (targetId: number) => {
    if (draggingId === null || draggingId === targetId) return;
    const next = [...orderedImages];
    const fromIndex = next.findIndex((item) => item.id === draggingId);
    const toIndex = next.findIndex((item) => item.id === targetId);
    if (fromIndex < 0 || toIndex < 0) return;
    const [moved] = next.splice(fromIndex, 1);
    next.splice(toIndex, 0, moved);
    const reorderedIds = next.map((item) => item.id);
    try {
      await reorderGalleryImages(reorderedIds);
      await refresh();
      setFeedback("Gallery order updated.");
    } catch (cause) {
      setFeedback(cause instanceof Error ? cause.message : "Unable to reorder gallery");
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
              <p className="text-sm font-semibold uppercase tracking-[0.3em] text-orange-600">Gallery Management</p>
              <h1 className="text-3xl font-black tracking-tight">Upload, reorder, and spotlight your gallery images</h1>
            </div>
            <button type="button" onClick={() => navigate("/admin")} className="rounded-full border border-slate-200 px-4 py-2 text-sm font-semibold text-slate-900 transition hover:border-orange-300 hover:text-orange-600">Back to Dashboard</button>
          </div>

          <div className="grid gap-6 lg:grid-cols-[0.95fr_0.75fr]">
            <section className="rounded-[2rem] border border-black/5 bg-white p-6 shadow-sm">
              <div className="flex items-center justify-between">
                <h2 className="text-xl font-black">Gallery</h2>
                <p className="text-sm text-slate-600">Drag to reorder. The homepage uses a random selection of these images.</p>
              </div>
              {feedback ? <p className="mt-4 rounded-full bg-slate-100 px-4 py-2 text-sm text-slate-700">{feedback}</p> : null}
              {loading ? (
                <div className="mt-6 rounded-[1.5rem] border border-slate-200 bg-slate-50 p-6 text-sm text-slate-600">Loading gallery...</div>
              ) : (
                <div className="mt-6 grid gap-4 md:grid-cols-2">
                  {orderedImages.map((image) => (
                    <div key={image.id} draggable onDragStart={() => setDraggingId(image.id)} onDragOver={(event) => event.preventDefault()} onDrop={() => handleDrop(image.id)} className="overflow-hidden rounded-[1.5rem] border border-slate-200">
                      <img src={image.image_url} alt={image.caption || "Gallery image"} className="aspect-[4/3] w-full object-cover" />
                      <div className="p-4">
                        <div className="flex flex-wrap items-center gap-2">
                          <span className={`rounded-full px-3 py-1 text-xs font-semibold uppercase tracking-[0.3em] ${image.is_active ? "bg-emerald-100 text-emerald-700" : "bg-rose-100 text-rose-700"}`}>
                            {image.is_active ? "Active" : "Disabled"}
                          </span>
                          {image.is_featured ? <span className="rounded-full bg-orange-100 px-3 py-1 text-xs font-semibold uppercase tracking-[0.3em] text-orange-700">Featured</span> : null}
                        </div>
                        <p className="mt-3 font-black">{image.caption || "Untitled"}</p>
                        <div className="mt-4 flex flex-wrap gap-2">
                          <button type="button" onClick={() => startEdit(image)} className="rounded-full border border-slate-200 px-3 py-2 text-sm font-semibold text-slate-900 transition hover:border-orange-300 hover:text-orange-600">Edit</button>
                          <button type="button" onClick={() => handleDelete(image)} className="rounded-full border border-slate-200 px-3 py-2 text-sm font-semibold text-slate-900 transition hover:border-orange-300 hover:text-orange-600">Delete</button>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </section>

            <section className="rounded-[2rem] border border-black/5 bg-white p-6 shadow-sm">
              <div className="flex items-center justify-between">
                <h2 className="text-xl font-black">{editingId ? "Edit Image" : "Add Image"}</h2>
                {editingId ? <button type="button" onClick={resetForm} className="text-sm font-semibold text-slate-600">Cancel</button> : null}
              </div>
              <form className="mt-6 space-y-4" onSubmit={handleSubmit}>
                <ImageUploadField
                  value={form.image_url}
                  onChange={(value) => setForm((current) => ({ ...current, image_url: value }))}
                  label="Image URL"
                  placeholder="https://..."
                />
                <div>
                  <label className="text-sm font-semibold text-slate-700">Caption</label>
                  <input value={form.caption} onChange={(event) => setForm((current) => ({ ...current, caption: event.target.value }))} className="mt-2 w-full rounded-full border border-slate-200 px-4 py-3 text-sm outline-none" />
                </div>
                <label className="flex items-center gap-2 rounded-full border border-slate-200 px-4 py-3 text-sm font-semibold text-slate-700">
                  <input type="checkbox" checked={form.is_active} onChange={(event) => setForm((current) => ({ ...current, is_active: event.target.checked }))} />
                  Active
                </label>
                <label className="flex items-center gap-2 rounded-full border border-slate-200 px-4 py-3 text-sm font-semibold text-slate-700">
                  <input type="checkbox" checked={form.is_featured} onChange={(event) => setForm((current) => ({ ...current, is_featured: event.target.checked }))} />
                  Featured
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
