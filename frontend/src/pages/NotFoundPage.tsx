import { SeoHead } from "../components/seo/SeoHead";

export default function NotFoundPage() {
  return (
    <div className="min-h-screen bg-[#fffaf5] px-5 py-16 text-slate-900">
      <SeoHead title="Page not found" description="The page you requested could not be found at Indian Restaurant & Sweets." noIndex />
      <div className="mx-auto max-w-2xl rounded-[2rem] border border-black/5 bg-white p-10 text-center shadow-sm">
        <p className="text-sm font-semibold uppercase tracking-[0.3em] text-orange-600">404</p>
        <h1 className="mt-3 text-3xl font-black">Page not found</h1>
        <p className="mt-4 text-lg leading-8 text-slate-600">The link you followed may be broken or the page may have moved.</p>
        <a href="/" className="mt-8 inline-flex rounded-full bg-orange-600 px-5 py-3 text-sm font-semibold text-white transition hover:bg-orange-700">
          Back to home
        </a>
      </div>
    </div>
  );
}
