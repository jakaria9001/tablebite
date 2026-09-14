import { useMemo, useRef, useState } from "react";
import { QRCodeCanvas, QRCodeSVG } from "qrcode.react";
import { useNavigate } from "react-router-dom";
import { createAdminTable, deleteAdminTable, reorderAdminTables, updateAdminTable } from "../api/tables";
import { Footer } from "../components/layout/Footer";
import { Header } from "../components/layout/Header";
import { useAdminTables } from "../hooks/useAdminTables";
import { markAdminDataUpdated } from "../hooks/useAdminDashboardSummary";
import type { AdminTable, AdminTablePayload } from "../api/tables";

const emptyForm = (): AdminTablePayload => ({
  table_number: "",
  display_name: "",
  qr_token: "",
  is_active: true,
  display_order: 0,
});

function buildQrToken(tableNumber: string, displayName: string) {
  const safeTable = (tableNumber || "table").trim().toLowerCase().replace(/[^a-z0-9]+/g, "-");
  const safeName = (displayName || "table").trim().toLowerCase().replace(/[^a-z0-9]+/g, "-");
  return `tbl-${safeTable}-${safeName}`.replace(/-+/g, "-").replace(/^-|-$/g, "");
}

const FRONTEND_URL = (import.meta.env.VITE_FRONTEND_URL ?? window.location.origin).replace(/\/$/, "");

function buildTableUrl(token: string, tableNumber: string) {
  return `${FRONTEND_URL}/table/${tableNumber}/${token}`;
}

function renderQrPreview(url: string) {
  const normalized = url || FRONTEND_URL;

  return (
    <div className="rounded-2xl border border-slate-200 bg-slate-50 p-3">
      <QRCodeSVG value={normalized} size={160} level="H" includeMargin bgColor="#ffffff" fgColor="#0f172a" />
    </div>
  );
}

export default function AdminTablesPage() {
  const navigate = useNavigate();
  const { tables, loading, error, refresh } = useAdminTables();
  const [form, setForm] = useState<AdminTablePayload>(emptyForm());
  const [editingId, setEditingId] = useState<number | null>(null);
  const [formOpen, setFormOpen] = useState(false);
  const [saving, setSaving] = useState(false);
  const [draggingId, setDraggingId] = useState<number | null>(null);
  const [feedback, setFeedback] = useState<string | null>(null);
  const [qrModal, setQrModal] = useState<{ open: boolean; token: string; title: string }>({ open: false, token: "", title: "" });
  const [shareStatus, setShareStatus] = useState<string | null>(null);
  const qrDownloadRef = useRef<HTMLCanvasElement | null>(null);

  const orderedTables = useMemo(() => [...tables].sort((a, b) => a.display_order - b.display_order || a.id - b.id), [tables]);
  const currentQrToken = buildQrToken(form.table_number, form.display_name);
  const currentQrUrl = buildTableUrl(currentQrToken, form.table_number || "0");

  const resetForm = () => {
    setForm(emptyForm());
    setEditingId(null);
    setFormOpen(false);
  };

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault();
    if (!form.table_number.trim()) {
      setFeedback("Please enter a table number.");
      return;
    }

    const generatedToken = buildQrToken(form.table_number, form.display_name);
    const payload = { ...form, qr_token: generatedToken };

    setSaving(true);
    setFeedback(null);
    try {
      if (editingId) {
        await updateAdminTable(editingId, { ...payload, id: editingId });
        setFeedback("Table updated.");
      } else {
        await createAdminTable(payload);
        setFeedback("Table created.");
      }
      resetForm();
      await refresh();
      markAdminDataUpdated();
    } catch (cause) {
      setFeedback(cause instanceof Error ? cause.message : "Unable to save table");
    } finally {
      setSaving(false);
    }
  };

  const startEdit = (table: AdminTable) => {
    setEditingId(table.id);
    setForm({
      id: table.id,
      table_number: table.table_number,
      display_name: table.display_name ?? "",
      qr_token: table.qr_token ?? "",
      is_active: table.is_active,
      display_order: table.display_order,
    });
    setFormOpen(true);
  };

  const toggleActive = async (table: AdminTable) => {
    try {
      await updateAdminTable(table.id, {
        id: table.id,
        table_number: table.table_number,
        display_name: table.display_name ?? "",
        qr_token: table.qr_token ?? "",
        is_active: !table.is_active,
        display_order: table.display_order,
      });
      await refresh();
      markAdminDataUpdated();
    } catch (cause) {
      setFeedback(cause instanceof Error ? cause.message : "Unable to update table");
    }
  };

  const openQrModal = (token: string, title: string) => {
    setQrModal({ open: true, token, title });
    setShareStatus(null);
  };

  const handleGenerateQr = (source?: AdminTable | null) => {
    if (source) {
      const token = source.qr_token || buildQrToken(source.table_number, source.display_name);
      openQrModal(buildTableUrl(token, source.table_number), source.display_name || `Table ${source.table_number}`);
      return;
    }

    const token = currentQrToken;
    openQrModal(buildTableUrl(token, form.table_number || "0"), form.display_name || `Table ${form.table_number}`);
  };

  const handleDownloadQr = () => {
    const canvas = qrDownloadRef.current;
    if (!canvas) {
      return;
    }

    const link = document.createElement("a");
    link.download = `${(qrModal.title || "table-qr").trim().toLowerCase().replace(/[^a-z0-9]+/g, "-").replace(/^-|-$/g, "") || "table-qr"}.png`;
    link.href = canvas.toDataURL("image/png");
    link.click();
  };

  const handleShareQr = async () => {
    const shareTarget = qrModal.token || FRONTEND_URL;
    const shareTitle = qrModal.title || "Table QR Code";

    try {
      if (typeof navigator !== "undefined" && typeof navigator.share === "function") {
        await navigator.share({
          title: shareTitle,
          text: `Scan this QR code to open the ordering page for ${shareTitle}.`,
          url: shareTarget,
        });
        setShareStatus("QR link shared.");
        return;
      }

      if (typeof navigator !== "undefined" && navigator.clipboard?.writeText) {
        await navigator.clipboard.writeText(shareTarget);
        setShareStatus("QR link copied to clipboard.");
        return;
      }

      setShareStatus("Sharing is not available in this browser.");
    } catch (cause) {
      if (cause instanceof Error && cause.name === "AbortError") {
        setShareStatus("Sharing cancelled.");
        return;
      }

      if (typeof navigator !== "undefined" && navigator.clipboard?.writeText) {
        try {
          await navigator.clipboard.writeText(shareTarget);
          setShareStatus("QR link copied to clipboard.");
          return;
        } catch {
          // Ignore clipboard fallback failures and fall through.
        }
      }

      setShareStatus("Unable to share this QR link.");
    }
  };

  const handleDrop = async (targetId: number) => {
    if (draggingId === null || draggingId === targetId) return;
    const next = [...orderedTables];
    const fromIndex = next.findIndex((item) => item.id === draggingId);
    const toIndex = next.findIndex((item) => item.id === targetId);
    if (fromIndex < 0 || toIndex < 0) return;
    const [moved] = next.splice(fromIndex, 1);
    next.splice(toIndex, 0, moved);
    const reorderedIds = next.map((item) => item.id);
    try {
      await reorderAdminTables(reorderedIds);
      await refresh();
      markAdminDataUpdated();
      setFeedback("Table order updated.");
    } catch (cause) {
      setFeedback(cause instanceof Error ? cause.message : "Unable to reorder tables");
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
              <p className="text-sm font-semibold uppercase tracking-[0.3em] text-orange-600">Table Management</p>
              <h1 className="text-3xl font-black tracking-tight">Create and manage dining tables</h1>
            </div>
            <button type="button" onClick={() => navigate("/admin")} className="rounded-full border border-slate-200 px-4 py-2 text-sm font-semibold text-slate-900 transition hover:border-orange-300 hover:text-orange-600">Back to Dashboard</button>
          </div>

          <div>
            <section className="rounded-[2rem] border border-black/5 bg-white p-6 shadow-sm">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <div><h2 className="text-xl font-black">Tables</h2><p className="mt-1 text-sm text-slate-600">Drag to reorder. Each table gets a numbered label and QR token.</p></div>
                  <button type="button" onClick={() => { setForm(emptyForm()); setEditingId(null); setFormOpen(true); }} className="min-h-11 rounded-full bg-[var(--maroon-800)] px-5 py-2 text-sm font-bold text-white">Add table</button>
                </div>
              </div>
              {feedback ? <p className="mt-4 rounded-full bg-slate-100 px-4 py-2 text-sm text-slate-700">{feedback}</p> : null}
              {loading ? (
                <div className="mt-6 rounded-[1.5rem] border border-slate-200 bg-slate-50 p-6 text-sm text-slate-600">Loading tables...</div>
              ) : error ? (
                <div className="mt-6 rounded-[1.5rem] border border-rose-200 bg-rose-50 p-6 text-sm text-rose-700">{error}</div>
              ) : (
                <div className="mt-6 space-y-3">
                  {orderedTables.map((table) => (
                    <div key={table.id} draggable onDragStart={() => setDraggingId(table.id)} onDragOver={(event) => event.preventDefault()} onDrop={() => handleDrop(table.id)} className="flex flex-col gap-3 rounded-[1.5rem] border border-slate-200 p-4 md:flex-row md:items-center md:justify-between">
                      <div className="flex items-center gap-3">
                        <div className="flex h-10 w-10 items-center justify-center rounded-full bg-slate-100 text-sm font-black text-slate-700">{table.table_number}</div>
                        <div>
                          <p className="font-black">{table.display_name || `Table ${table.table_number}`}</p>
                          <p className="text-sm text-slate-600">{table.qr_token}</p>
                        </div>
                      </div>
                      <div className="flex flex-wrap items-center gap-2">
                        <span className={`rounded-full px-3 py-1 text-xs font-semibold uppercase tracking-[0.3em] ${table.is_active ? "bg-emerald-100 text-emerald-700" : "bg-rose-100 text-rose-700"}`}>
                          {table.is_active ? "Active" : "Disabled"}
                        </span>
                        <button type="button" onClick={() => startEdit(table)} className="rounded-full border border-slate-200 px-3 py-2 text-sm font-semibold text-slate-900 transition hover:border-orange-300 hover:text-orange-600">Edit</button>
                        <button type="button" onClick={() => handleGenerateQr(table)} className="rounded-full border border-slate-200 px-3 py-2 text-sm font-semibold text-slate-900 transition hover:border-orange-300 hover:text-orange-600">Generate QR</button>
                        <button type="button" onClick={() => toggleActive(table)} className="rounded-full border border-slate-200 px-3 py-2 text-sm font-semibold text-slate-900 transition hover:border-orange-300 hover:text-orange-600">{table.is_active ? "Disable" : "Activate"}</button>
                        <button type="button" onClick={async () => { if (window.confirm("Delete this table?")) { await deleteAdminTable(table.id); await refresh(); markAdminDataUpdated(); setFeedback("Table deleted."); } }} className="rounded-full border border-slate-200 px-3 py-2 text-sm font-semibold text-slate-900 transition hover:border-rose-300 hover:text-rose-600">Delete</button>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </section>

            {formOpen ? <div className="fixed inset-0 z-50 overflow-y-auto bg-[var(--maroon-950)]/70 px-4 py-6 backdrop-blur-sm sm:py-10" role="dialog" aria-modal="true" aria-labelledby="table-form-title">
            <section className="mx-auto max-w-xl rounded-[2rem] border border-[var(--line)] bg-[var(--cream)] p-6 shadow-[var(--shadow-lifted)] sm:p-8">
              <div className="flex items-center justify-between">
                <h2 id="table-form-title" className="text-xl font-black text-[var(--maroon-900)]">{editingId ? "Edit Table" : "Add Table"}</h2>
                <button type="button" onClick={resetForm} aria-label="Close table form" className="grid h-11 w-11 place-items-center rounded-full border border-[var(--line)] bg-white text-xl">×</button>
              </div>

              <form className="mt-6 space-y-4" onSubmit={handleSubmit}>
                <div>
                  <label className="text-sm font-semibold text-slate-700">Table Number</label>
                  <input value={form.table_number} onChange={(event) => setForm((current) => ({ ...current, table_number: event.target.value }))} className="mt-2 w-full rounded-full border border-slate-200 px-4 py-3 text-sm outline-none" placeholder="01" required />
                </div>
                <div>
                  <label className="text-sm font-semibold text-slate-700">Name</label>
                  <input value={form.display_name} onChange={(event) => setForm((current) => ({ ...current, display_name: event.target.value }))} className="mt-2 w-full rounded-full border border-slate-200 px-4 py-3 text-sm outline-none" placeholder="Window Table" />
                </div>
                <label className="flex items-center gap-2 rounded-full border border-slate-200 px-4 py-3 text-sm font-semibold text-slate-700">
                  <input type="checkbox" checked={form.is_active} onChange={(event) => setForm((current) => ({ ...current, is_active: event.target.checked }))} />
                  Active
                </label>
                <div>
                  <label className="text-sm font-semibold text-slate-700">QR Token</label>
                  <div className="mt-2 rounded-[1rem] border border-slate-200 bg-slate-50 px-4 py-3 text-sm font-medium text-slate-700">
                    {currentQrToken}
                  </div>
                  <div className="mt-3 flex flex-wrap items-center gap-2">
                    <button type="button" onClick={() => handleGenerateQr()} className="rounded-full bg-orange-600 px-4 py-2 text-sm font-semibold text-white transition hover:bg-orange-700">Generate QR</button>
                    {editingId ? <span className="text-xs text-slate-500">Existing QR codes can be reopened below.</span> : null}
                  </div>
                  <p className="mt-2 text-xs text-slate-500">This is generated automatically from the table number and name.</p>
                </div>
                <div>
                  <label className="text-sm font-semibold text-slate-700">QR Code</label>
                  <div className="mt-3 flex items-center gap-4 rounded-[1.25rem] border border-slate-200 bg-slate-50 p-4">
                    {renderQrPreview(currentQrUrl)}
                    <div className="text-sm text-slate-600">
                      <p className="font-semibold text-slate-800">Preview</p>
                      <p className="mt-1">This preview updates from the QR token and table number.</p>
                    </div>
                  </div>
                </div>
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
      {qrModal.open ? (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/70 px-4">
          <div className="w-full max-w-md rounded-[2rem] border border-slate-200 bg-white p-6 shadow-2xl">
            <div className="flex items-start justify-between gap-4">
              <div>
                <p className="text-sm font-semibold uppercase tracking-[0.3em] text-orange-600">QR Code</p>
                <h3 className="mt-2 text-xl font-black text-slate-900">{qrModal.title}</h3>
              </div>
              <div className="flex items-center gap-2">
                <button type="button" onClick={handleShareQr} className="rounded-full border border-slate-200 px-3 py-2 text-sm font-semibold text-slate-700 transition hover:border-orange-300 hover:text-orange-600">Share</button>
                <button type="button" onClick={handleDownloadQr} className="rounded-full bg-orange-600 px-3 py-2 text-sm font-semibold text-white transition hover:bg-orange-700">Download</button>
                <button type="button" onClick={() => { setQrModal({ open: false, token: "", title: "" }); setShareStatus(null); }} className="rounded-full border border-slate-200 px-3 py-2 text-sm font-semibold text-slate-700">Close</button>
              </div>
            </div>
            <div className="mt-6 flex flex-col items-center justify-center rounded-[1.5rem] border border-slate-200 bg-slate-50 p-6">
              <div className="sr-only">
                <QRCodeCanvas ref={qrDownloadRef} value={qrModal.token || FRONTEND_URL} size={1024} level="H" includeMargin bgColor="#ffffff" fgColor="#0f172a" />
              </div>
              {renderQrPreview(qrModal.token)}
              <p className="mt-4 break-all text-center text-sm font-medium text-slate-600">{qrModal.token}</p>
              {shareStatus ? <p className="mt-3 text-center text-sm font-semibold text-orange-600">{shareStatus}</p> : null}
            </div>
          </div>
        </div>
      ) : null}
      <Footer />
    </div>
  );
}
