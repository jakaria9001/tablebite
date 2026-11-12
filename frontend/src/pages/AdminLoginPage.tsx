import { useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { loginAdmin } from "../api/auth";

export default function AdminLoginPage() {
  const navigate = useNavigate();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const canSubmit = useMemo(() => email.trim().length > 0 && password.length > 0, [email, password]);

  async function onSubmit(event: React.FormEvent) {
    event.preventDefault();
    setLoading(true);
    setError("");

    try {
      await loginAdmin(email, password);
      navigate("/admin");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Login failed");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-[#fffaf5] px-4 py-10 text-slate-950">
      <div className="fixed inset-0 bg-slate-950/45 backdrop-blur-sm" />
      <form onSubmit={onSubmit} role="dialog" aria-modal="true" aria-labelledby="admin-login-title" className="relative z-10 w-full max-w-md rounded-[2rem] border border-slate-200 bg-white p-8 shadow-2xl">
        <div className="flex items-start justify-between gap-4">
          <div>
            <p className="text-sm font-semibold uppercase tracking-[0.3em] text-orange-600">Admin access</p>
            <h1 id="admin-login-title" className="mt-3 text-3xl font-black">Secure sign in</h1>
          </div>
          <button type="button" onClick={() => navigate("/")} className="rounded-full border border-slate-200 px-3 py-1 text-sm font-bold text-slate-500">Close</button>
        </div>
        <p className="mt-3 text-sm text-slate-600">Use your administrator credentials to continue.</p>

        <label className="mt-6 block text-sm font-semibold text-slate-700">Email</label>
        <input
          value={email}
          onChange={(event) => setEmail(event.target.value)}
          className="mt-2 w-full rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3 outline-none ring-orange-300 focus:ring-2"
          placeholder="admin@example.com"
          type="email"
          autoComplete="username"
          autoCapitalize="none"
          autoCorrect="off"
          spellCheck={false}
        />

        <label className="mt-4 block text-sm font-semibold text-slate-700">Password</label>
        <input
          type="password"
          value={password}
          onChange={(event) => setPassword(event.target.value)}
          className="mt-2 w-full rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3 outline-none ring-orange-300 focus:ring-2"
          placeholder="••••••••"
          autoComplete="current-password"
          autoCapitalize="none"
          autoCorrect="off"
          spellCheck={false}
        />

        {error ? <p className="mt-4 rounded-2xl bg-red-500/15 px-3 py-2 text-sm text-red-200">{error}</p> : null}

        <button disabled={!canSubmit || loading} className="mt-6 min-h-12 w-full rounded-2xl bg-slate-950 px-4 py-3 font-black text-white disabled:cursor-not-allowed disabled:opacity-60">
          {loading ? "Signing in…" : "Sign in"}
        </button>
      </form>
    </div>
  );
}
