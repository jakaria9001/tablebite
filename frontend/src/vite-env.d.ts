interface ImportMetaEnv {
    readonly VITE_API_BASE_URL: string;
    readonly VITE_FRONTEND_URL?: string;
    readonly VITE_WAITER_POLL_INTERVAL_MS?: string;
}

interface ImportMeta {
    readonly env: ImportMetaEnv;
}