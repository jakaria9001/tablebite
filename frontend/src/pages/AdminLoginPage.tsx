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
    <div className="flex min-h-screen items-center justify-center bg-slate-950 px-4 py-10 text-white">
      <form onSubmit={onSubmit} className="w-full max-w-md rounded-[2rem] border border-white/10 bg-white/10 p-8 shadow-2xl backdrop-blur">
        <p className="text-sm font-semibold uppercase tracking-[0.3em] text-orange-300">Admin access</p>
        <h1 className="mt-3 text-3xl font-black">Secure sign in</h1>
        <p className="mt-3 text-sm text-slate-300">Use your administrator credentials to continue.</p>

        <label className="mt-6 block text-sm font-semibold text-slate-200">Email</label>
        <input
          value={email}
          onChange={(event) => setEmail(event.target.value)}
          className="mt-2 w-full rounded-2xl border border-white/10 bg-slate-900/70 px-4 py-3 outline-none ring-orange-400 focus:ring-2"
          placeholder="admin@example.com"
          type="email"
          autoComplete="username"
          autoCapitalize="none"
          autoCorrect="off"
          spellCheck={false}
        />

        <label className="mt-4 block text-sm font-semibold text-slate-200">Password</label>
        <input
          type="password"
          value={password}
          onChange={(event) => setPassword(event.target.value)}
          className="mt-2 w-full rounded-2xl border border-white/10 bg-slate-900/70 px-4 py-3 outline-none ring-orange-400 focus:ring-2"
          placeholder="••••••••"
          autoComplete="current-password"
          autoCapitalize="none"
          autoCorrect="off"
          spellCheck={false}
        />

        {error ? <p className="mt-4 rounded-2xl bg-red-500/15 px-3 py-2 text-sm text-red-200">{error}</p> : null}

        <button disabled={!canSubmit || loading} className="mt-6 w-full rounded-2xl bg-orange-500 px-4 py-3 font-black text-white disabled:cursor-not-allowed disabled:opacity-60">
          {loading ? "Signing in…" : "Sign in"}
        </button>
      </form>
    </div>
  );
}
