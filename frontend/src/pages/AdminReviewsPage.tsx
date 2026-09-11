import { useEffect, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { createReview, deleteReview, getAdminReviews, reorderReviews, updateReview, type Review, type ReviewPayload } from "../api/reviews";
import { Footer } from "../components/layout/Footer";
import { Header } from "../components/layout/Header";

const emptyForm = (): ReviewPayload => ({
  author_name: "",
  rating: 5,
  review_text: "",
  source: "manual",
  source_url: "",
  is_featured: true,
  is_active: true,
});

export default function AdminReviewsPage() {
  const navigate = useNavigate();
  const [reviews, setReviews] = useState<Review[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [draggingId, setDraggingId] = useState<number | null>(null);
  const [feedback, setFeedback] = useState<string | null>(null);
  const [form, setForm] = useState<ReviewPayload>(emptyForm());
  const [editingId, setEditingId] = useState<number | null>(null);

  const orderedReviews = useMemo(() => [...reviews].sort((a, b) => a.display_order - b.display_order || a.id - b.id), [reviews]);

  const refresh = async () => {
    setLoading(true);
    try {
      const data = await getAdminReviews();
      setReviews(data.reviews);
    } catch (cause) {
      setFeedback(cause instanceof Error ? cause.message : "Unable to load reviews");
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
    if (!form.author_name.trim() || !form.review_text.trim()) {
      setFeedback("Please provide a guest name and review text.");
      return;
    }

    setSaving(true);
    setFeedback(null);
    try {
      if (editingId) {
        await updateReview(editingId, { ...form, id: editingId });
        setFeedback("Review updated.");
      } else {
        await createReview(form);
        setFeedback("Review created.");
      }
      resetForm();
      await refresh();
    } catch (cause) {
      setFeedback(cause instanceof Error ? cause.message : "Unable to save review");
    } finally {
      setSaving(false);
    }
  };

  const startEdit = (review: Review) => {
    setEditingId(review.id);
    setForm({
      id: review.id,
      author_name: review.author_name,
      rating: review.rating,
      review_text: review.review_text,
      source: review.source,
      source_url: review.source_url,
      is_featured: review.is_featured,
      is_active: review.is_active,
    });
  };

  const handleDelete = async (review: Review) => {
    try {
      await deleteReview(review.id);
      await refresh();
      setFeedback("Review deleted.");
    } catch (cause) {
      setFeedback(cause instanceof Error ? cause.message : "Unable to delete review");
    }
  };

  const handleDrop = async (targetId: number) => {
    if (draggingId === null || draggingId === targetId) return;
    const next = [...orderedReviews];
    const fromIndex = next.findIndex((item) => item.id === draggingId);
    const toIndex = next.findIndex((item) => item.id === targetId);
    if (fromIndex < 0 || toIndex < 0) return;
    const [moved] = next.splice(fromIndex, 1);
    next.splice(toIndex, 0, moved);
    const reorderedIds = next.map((item) => item.id);
    try {
      await reorderReviews(reorderedIds);
      await refresh();
      setFeedback("Review order updated.");
    } catch (cause) {
      setFeedback(cause instanceof Error ? cause.message : "Unable to reorder reviews");
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
              <p className="text-sm font-semibold uppercase tracking-[0.3em] text-orange-600">Review Management</p>
              <h1 className="text-3xl font-black tracking-tight">Showcase guest feedback and highlight your best reviews</h1>
            </div>
            <button type="button" onClick={() => navigate("/admin")} className="rounded-full border border-slate-200 px-4 py-2 text-sm font-semibold text-slate-900 transition hover:border-orange-300 hover:text-orange-600">Back to Dashboard</button>
          </div>

          <div className="grid gap-6 lg:grid-cols-[0.95fr_0.75fr]">
            <section className="rounded-[2rem] border border-black/5 bg-white p-6 shadow-sm">
              <div className="flex items-center justify-between">
                <h2 className="text-xl font-black">Reviews</h2>
                <p className="text-sm text-slate-600">Drag to reorder. Only featured reviews appear on the homepage.</p>
              </div>
              {feedback ? <p className="mt-4 rounded-full bg-slate-100 px-4 py-2 text-sm text-slate-700">{feedback}</p> : null}
              {loading ? (
                <div className="mt-6 rounded-[1.5rem] border border-slate-200 bg-slate-50 p-6 text-sm text-slate-600">Loading reviews...</div>
              ) : (
                <div className="mt-6 grid gap-4">
                  {orderedReviews.map((review) => (
                    <div key={review.id} draggable onDragStart={() => setDraggingId(review.id)} onDragOver={(event) => event.preventDefault()} onDrop={() => handleDrop(review.id)} className="rounded-[1.5rem] border border-slate-200 p-4">
                      <div className="flex flex-wrap items-center gap-2">
                        <span className={`rounded-full px-3 py-1 text-xs font-semibold uppercase tracking-[0.3em] ${review.is_active ? "bg-emerald-100 text-emerald-700" : "bg-rose-100 text-rose-700"}`}>
                          {review.is_active ? "Active" : "Disabled"}
                        </span>
                        {review.is_featured ? <span className="rounded-full bg-orange-100 px-3 py-1 text-xs font-semibold uppercase tracking-[0.3em] text-orange-700">Featured</span> : null}
                      </div>
                      <p className="mt-3 font-black">{review.author_name}</p>
                      <p className="mt-2 text-sm text-slate-600">{review.review_text}</p>
                      <div className="mt-3 flex flex-wrap items-center gap-3 text-sm text-slate-600">
                        <span>Rating: {"★".repeat(review.rating)}</span>
                        <span>Source: {review.source}</span>
                      </div>
                      <div className="mt-4 flex flex-wrap gap-2">
                        <button type="button" onClick={() => startEdit(review)} className="rounded-full border border-slate-200 px-3 py-2 text-sm font-semibold text-slate-900 transition hover:border-orange-300 hover:text-orange-600">Edit</button>
                        <button type="button" onClick={() => handleDelete(review)} className="rounded-full border border-slate-200 px-3 py-2 text-sm font-semibold text-slate-900 transition hover:border-orange-300 hover:text-orange-600">Delete</button>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </section>

            <section className="rounded-[2rem] border border-black/5 bg-white p-6 shadow-sm">
              <div className="flex items-center justify-between">
                <h2 className="text-xl font-black">{editingId ? "Edit Review" : "Add Review"}</h2>
                {editingId ? <button type="button" onClick={resetForm} className="text-sm font-semibold text-slate-600">Cancel</button> : null}
              </div>
              <form className="mt-6 space-y-4" onSubmit={handleSubmit}>
                <div>
                  <label className="text-sm font-semibold text-slate-700">Guest Name</label>
                  <input value={form.author_name} onChange={(event) => setForm((current) => ({ ...current, author_name: event.target.value }))} className="mt-2 w-full rounded-full border border-slate-200 px-4 py-3 text-sm outline-none" required />
                </div>
                <div>
                  <label className="text-sm font-semibold text-slate-700">Review Text</label>
                  <textarea value={form.review_text} onChange={(event) => setForm((current) => ({ ...current, review_text: event.target.value }))} className="mt-2 min-h-[100px] w-full rounded-[1.25rem] border border-slate-200 px-4 py-3 text-sm outline-none" required />
                </div>
                <div>
                  <label className="text-sm font-semibold text-slate-700">Rating</label>
                  <select value={form.rating} onChange={(event) => setForm((current) => ({ ...current, rating: Number(event.target.value) }))} className="mt-2 w-full rounded-full border border-slate-200 px-4 py-3 text-sm outline-none">
                    {[1, 2, 3, 4, 5].map((value) => <option key={value} value={value}>{value} Star{value > 1 ? "s" : ""}</option>)}
                  </select>
                </div>
                <div>
                  <label className="text-sm font-semibold text-slate-700">Source</label>
                  <input value={form.source} onChange={(event) => setForm((current) => ({ ...current, source: event.target.value }))} className="mt-2 w-full rounded-full border border-slate-200 px-4 py-3 text-sm outline-none" />
                </div>
                <div>
                  <label className="text-sm font-semibold text-slate-700">Source URL</label>
                  <input value={form.source_url} onChange={(event) => setForm((current) => ({ ...current, source_url: event.target.value }))} className="mt-2 w-full rounded-full border border-slate-200 px-4 py-3 text-sm outline-none" />
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
