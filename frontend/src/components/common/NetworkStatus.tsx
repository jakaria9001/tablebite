import { useEffect, useState } from "react";

export function NetworkStatus() {
  const [offline, setOffline] = useState(() => !navigator.onLine);

  useEffect(() => {
    const handleOffline = () => setOffline(true);
    const handleOnline = () => setOffline(false);
    window.addEventListener("offline", handleOffline);
    window.addEventListener("online", handleOnline);
    return () => {
      window.removeEventListener("offline", handleOffline);
      window.removeEventListener("online", handleOnline);
    };
  }, []);

  return offline ? (
    <div role="status" className="fixed inset-x-0 bottom-0 z-50 bg-slate-950 px-4 py-3 text-center text-sm font-semibold text-white shadow-lg">
      You are offline. Some actions may be unavailable until your connection returns.
    </div>
  ) : null;
}
