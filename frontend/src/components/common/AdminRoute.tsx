import { useEffect, useState } from "react";
import { Navigate } from "react-router-dom";
import { getAdminMe } from "../../api/auth";

export function AdminRoute({ children }: { children: React.ReactNode }) {
  const [status, setStatus] = useState<"loading" | "allowed" | "denied">("loading");

  useEffect(() => {
    let cancelled = false;
    void getAdminMe().then(() => {
      if (!cancelled) setStatus("allowed");
    }).catch(() => {
      if (!cancelled) setStatus("denied");
    });
    return () => {
      cancelled = true;
    };
  }, []);

  if (status === "loading") {
    return <div className="flex min-h-screen items-center justify-center bg-slate-950 text-white">Checking admin access…</div>;
  }

  if (status === "denied") {
    return <Navigate to="/admin/login" replace />;
  }

  return <>{children}</>;
}
