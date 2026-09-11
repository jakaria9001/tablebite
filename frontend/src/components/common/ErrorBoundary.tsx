import { Component, type ErrorInfo, type ReactNode } from "react";

interface Props {
  children: ReactNode;
}

interface State {
  hasError: boolean;
}

export class ErrorBoundary extends Component<Props, State> {
  state: State = { hasError: false };

  static getDerivedStateFromError(): State {
    return { hasError: true };
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    console.error("Unhandled frontend error", { error, componentStack: info.componentStack });
  }

  render() {
    if (!this.state.hasError) {
      return this.props.children;
    }

    return (
      <main className="flex min-h-screen items-center justify-center bg-[#fffaf5] px-5 py-16 text-center text-slate-900">
        <div className="max-w-lg rounded-[2rem] border border-black/5 bg-white p-10 shadow-sm">
          <p className="text-sm font-semibold uppercase tracking-[0.3em] text-orange-600">Something went wrong</p>
          <h1 className="mt-3 text-3xl font-black">The page could not load.</h1>
          <p className="mt-4 text-slate-600">Please try again. Your data has not been changed.</p>
          <button type="button" onClick={() => window.location.reload()} className="mt-8 rounded-full bg-orange-600 px-5 py-3 text-sm font-semibold text-white hover:bg-orange-700">
            Reload page
          </button>
        </div>
      </main>
    );
  }
}
