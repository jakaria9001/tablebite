import { useNavigate } from "react-router-dom";
import { Footer } from "../components/layout/Footer";
import { Header } from "../components/layout/Header";
import { useAdminDashboardSummary } from "../hooks/useAdminDashboardSummary";
import { provisionWaiter } from "../api/staffOrders";
import { useState } from "react";

export default function AdminDashboardPage() {
  const navigate = useNavigate();
  const { data, loading, error } = useAdminDashboardSummary();
  const [waiterPhone, setWaiterPhone] = useState("");
  const [waiterName, setWaiterName] = useState("");
  const [waiterMessage, setWaiterMessage] = useState("");
  const [provisioning, setProvisioning] = useState(false);

  async function addWaiter() {
    setProvisioning(true);
    setWaiterMessage("");
    try {
      await provisionWaiter(waiterPhone, waiterName);
      setWaiterMessage("Waiter access enabled. They can now sign in with phone OTP.");
      setWaiterPhone("");
      setWaiterName("");
    } catch (cause) {
      setWaiterMessage(cause instanceof Error ? cause.message : "Unable to add waiter.");
    } finally {
      setProvisioning(false);
    }
  }

  return (
    <div className="min-h-screen bg-[#fffaf5] text-slate-900">
      <Header />
      <main className="px-5 py-10 sm:px-8 lg:px-10">
        <div className="mx-auto flex max-w-5xl flex-col rounded-[2rem] border border-black/5 bg-white p-8 shadow-sm md:p-10">
          <div className="border-b border-slate-200 pb-6">
            <h1 className="text-3xl font-black tracking-tight">Admin Dashboard</h1>
          </div>

          {loading ? (
            <div className="mt-8 rounded-[1.5rem] border border-slate-200 bg-slate-50 p-6 text-sm text-slate-600">
              Loading dashboard summary...
            </div>
          ) : error ? (
            <div className="mt-8 rounded-[1.5rem] border border-rose-200 bg-rose-50 p-6 text-sm text-rose-700">
              {error}
            </div>
          ) : data ? (
            <>
              <div className="mt-8 grid gap-6 md:grid-cols-3">
                <div className="rounded-[1.5rem] border border-slate-200 bg-slate-50 p-6">
                  <p className="text-sm font-semibold uppercase tracking-[0.3em] text-orange-600">Menu Items</p>
                  <p className="mt-3 text-4xl font-black">{data.menuItems}</p>
                </div>
                <div className="rounded-[1.5rem] border border-slate-200 bg-slate-50 p-6">
                  <p className="text-sm font-semibold uppercase tracking-[0.3em] text-orange-600">Categories</p>
                  <p className="mt-3 text-4xl font-black">{data.categories}</p>
                </div>
                <div className="rounded-[1.5rem] border border-slate-200 bg-slate-50 p-6">
                  <p className="text-sm font-semibold uppercase tracking-[0.3em] text-orange-600">Tables</p>
                  <p className="mt-3 text-4xl font-black">{data.tables}</p>
                </div>
              </div>

              <div className="mt-8 grid gap-6 md:grid-cols-3">
                <div className="rounded-[1.5rem] border border-slate-200 bg-slate-50 p-6">
                  <p className="text-sm font-semibold uppercase tracking-[0.3em] text-emerald-600">Available Items</p>
                  <p className="mt-3 text-4xl font-black">{data.availableItems}</p>
                </div>
                <div className="rounded-[1.5rem] border border-slate-200 bg-slate-50 p-6">
                  <p className="text-sm font-semibold uppercase tracking-[0.3em] text-rose-600">Sold Out Items</p>
                  <p className="mt-3 text-4xl font-black">{data.soldOutItems}</p>
                </div>
                <div className="rounded-[1.5rem] border border-slate-200 bg-slate-50 p-6">
                  <p className="text-sm font-semibold uppercase tracking-[0.3em] text-sky-600">Featured</p>
                  <p className="mt-3 text-4xl font-black">{data.featuredItems}</p>
                </div>
              </div>
            </>
          ) : null}

          <div className="mt-8 rounded-[1.5rem] border border-slate-200 bg-slate-50 p-6">
            <h2 className="text-xl font-black">Quick Actions</h2>
            <div className="mt-5 grid gap-3 sm:grid-cols-2">
              {[
                "Manage Menu",
                "Kitchen Availability",
                "Manage Categories",
                "Manage Tables",
                "Manage Gallery",
                "Manage Reviews",
                "Manage Banners",
                "Restaurant Settings",
              ].map((action) => (
                <button
                  key={action}
                  type="button"
                  onClick={() => {
                    if (action === "Manage Menu") {
                      navigate("/admin/menu");
                    }
                    if (action === "Kitchen Availability") {
                      navigate("/admin/kitchen-availability");
                    }
                    if (action === "Manage Categories") {
                      navigate("/admin/categories");
                    }
                    if (action === "Manage Tables") {
                      navigate("/admin/tables");
                    }
                    if (action === "Manage Gallery") {
                      navigate("/admin/gallery");
                    }
                    if (action === "Manage Reviews") {
                      navigate("/admin/reviews");
                    }
                    if (action === "Manage Banners") {
                      navigate("/admin/banners");
                    }
                    if (action === "Restaurant Settings") {
                      navigate("/admin/settings");
                    }
                  }}
                  className="rounded-full border border-slate-200 bg-white px-4 py-3 text-left text-sm font-semibold text-slate-900 transition hover:border-orange-300 hover:text-orange-600"
                >
                  {action}
                </button>
              ))}
            </div>
            <div className="mt-8 rounded-[1.5rem] border border-slate-200 bg-slate-50 p-6">
              <h2 className="text-xl font-black">Add waiter access</h2>
              <p className="mt-2 text-sm text-slate-600">The customer must have completed OTP verification first.</p>
              <div className="mt-5 grid gap-3 md:grid-cols-3">
                <input value={waiterName} onChange={(event) => setWaiterName(event.target.value)} placeholder="Waiter name" className="rounded-2xl border border-slate-200 bg-white px-4 py-3 outline-none focus:ring-2 focus:ring-orange-300" />
                <input value={waiterPhone} onChange={(event) => setWaiterPhone(event.target.value)} placeholder="9876543210" inputMode="tel" className="rounded-2xl border border-slate-200 bg-white px-4 py-3 outline-none focus:ring-2 focus:ring-orange-300" />
                <button type="button" disabled={provisioning || !waiterName.trim() || !waiterPhone.trim()} onClick={() => void addWaiter()} className="rounded-2xl bg-slate-950 px-4 py-3 font-black text-white disabled:opacity-50">{provisioning ? "Adding..." : "Add waiter"}</button>
              </div>
              {waiterMessage ? <p className="mt-3 text-sm font-semibold text-slate-700">{waiterMessage}</p> : null}
              <button type="button" onClick={() => navigate("/waiter/login")} className="mt-4 text-sm font-bold text-orange-700">Open waiter login</button>
            </div>
          </div>
        </div>
      </main>
      <Footer />
    </div>
  );
}
