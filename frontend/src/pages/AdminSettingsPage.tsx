import { useEffect, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { getAdminRestaurantSettings, updateAdminRestaurantSettings, type RestaurantSettings } from "../api/settings";
import { Footer } from "../components/layout/Footer";
import { Header } from "../components/layout/Header";

const EMPTY_SETTINGS: RestaurantSettings = {
  restaurant_name: "",
  tagline: "",
  logo: "",
  hero_image: "",
  phone: "",
  whatsapp: "",
  email: "",
  address: "",
  google_maps_url: "",
  hours_monday: "",
  hours_tuesday: "",
  hours_wednesday: "",
  hours_thursday: "",
  hours_friday: "",
  hours_saturday: "",
  hours_sunday: "",
  instagram: "",
  facebook: "",
  youtube: "",
  google_business_profile: "",
  seo_title: "",
  seo_description: "",
  google_review_url: "",
};

const sections = [
  {
    key: "restaurant",
    title: "Restaurant",
    fields: [
      { key: "restaurant_name", label: "Restaurant Name", type: "text" },
      { key: "tagline", label: "Tagline", type: "text" },
      { key: "logo", label: "Logo", type: "text" },
      { key: "hero_image", label: "Hero Image", type: "text" },
    ],
  },
  {
    key: "contact",
    title: "Contact",
    fields: [
      { key: "phone", label: "Phone", type: "text" },
      { key: "whatsapp", label: "WhatsApp", type: "text" },
      { key: "email", label: "Email", type: "text" },
      { key: "address", label: "Address", type: "textarea" },
      { key: "google_maps_url", label: "Google Maps URL", type: "text" },
    ],
  },
  {
    key: "hours",
    title: "Hours",
    fields: [
      { key: "hours_monday", label: "Monday", type: "text" },
      { key: "hours_tuesday", label: "Tuesday", type: "text" },
      { key: "hours_wednesday", label: "Wednesday", type: "text" },
      { key: "hours_thursday", label: "Thursday", type: "text" },
      { key: "hours_friday", label: "Friday", type: "text" },
      { key: "hours_saturday", label: "Saturday", type: "text" },
      { key: "hours_sunday", label: "Sunday", type: "text" },
    ],
  },
  {
    key: "social",
    title: "Social",
    fields: [
      { key: "instagram", label: "Instagram", type: "text" },
      { key: "facebook", label: "Facebook", type: "text" },
      { key: "youtube", label: "YouTube", type: "text" },
      { key: "google_business_profile", label: "Google Business Profile", type: "text" },
    ],
  },
  {
    key: "website",
    title: "Website",
    fields: [
      { key: "seo_title", label: "SEO Title", type: "text" },
      { key: "seo_description", label: "SEO Description", type: "textarea" },
      { key: "google_review_url", label: "Google Review URL", type: "text" },
    ],
  },
];

export default function AdminSettingsPage() {
  const navigate = useNavigate();
  const [settings, setSettings] = useState<RestaurantSettings>(EMPTY_SETTINGS);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [feedback, setFeedback] = useState<string | null>(null);

  useEffect(() => {
    const loadSettings = async () => {
      try {
        const data = await getAdminRestaurantSettings();
        setSettings({ ...EMPTY_SETTINGS, ...data });
      } catch (cause) {
        setFeedback(cause instanceof Error ? cause.message : "Unable to load settings");
      } finally {
        setLoading(false);
      }
    };

    void loadSettings();
  }, []);

  const saveSettings = async (event: React.FormEvent) => {
    event.preventDefault();
    setSaving(true);
    setFeedback(null);
    try {
      await updateAdminRestaurantSettings(settings);
      setFeedback("Settings updated.");
    } catch (cause) {
      setFeedback(cause instanceof Error ? cause.message : "Unable to save settings");
    } finally {
      setSaving(false);
    }
  };

  const formSummary = useMemo(() => {
    const name = settings.restaurant_name?.trim() || "Your Restaurant";
    const tagline = settings.tagline?.trim();
    return tagline ? `${name} • ${tagline}` : name;
  }, [settings]);

  return (
    <div className="min-h-screen bg-[#fffaf5] text-slate-900">
      <Header />
      <main className="px-5 py-10 sm:px-8 lg:px-10">
        <div className="mx-auto flex max-w-7xl flex-col gap-6">
          <div className="flex flex-wrap items-center justify-between gap-4 rounded-[2rem] border border-black/5 bg-white p-6 shadow-sm">
            <div>
              <p className="text-sm font-semibold uppercase tracking-[0.3em] text-orange-600">Restaurant Settings</p>
              <h1 className="text-3xl font-black tracking-tight">Manage your public identity and storefront details</h1>
            </div>
            <button type="button" onClick={() => navigate("/admin")} className="rounded-full border border-slate-200 px-4 py-2 text-sm font-semibold text-slate-900 transition hover:border-orange-300 hover:text-orange-600">Back to Dashboard</button>
          </div>

          <section className="rounded-[2rem] border border-black/5 bg-white p-6 shadow-sm">
            <div className="flex flex-wrap items-center justify-between gap-3">
              <div>
                <h2 className="text-xl font-black">Storefront settings</h2>
                <p className="mt-1 text-sm text-slate-600">{formSummary}</p>
              </div>
              {feedback ? <p className="rounded-full bg-slate-100 px-4 py-2 text-sm text-slate-700">{feedback}</p> : null}
            </div>

            {loading ? (
              <div className="mt-6 rounded-[1.5rem] border border-slate-200 bg-slate-50 p-6 text-sm text-slate-600">Loading settings...</div>
            ) : (
              <form className="mt-6 space-y-8" onSubmit={saveSettings}>
                {sections.map((section) => (
                  <div key={section.key} className="rounded-[1.5rem] border border-slate-200 p-5">
                    <h3 className="text-lg font-black">{section.title}</h3>
                    <div className="mt-4 grid gap-4 md:grid-cols-2">
                      {section.fields.map((field) => (
                        <div key={field.key} className={field.type === "textarea" ? "md:col-span-2" : ""}>
                          <label className="text-sm font-semibold text-slate-700">{field.label}</label>
                          {field.type === "textarea" ? (
                            <textarea
                              value={settings[field.key] ?? ""}
                              onChange={(event) => setSettings((current) => ({ ...current, [field.key]: event.target.value }))}
                              className="mt-2 min-h-24 w-full rounded-[1rem] border border-slate-200 px-4 py-3 text-sm outline-none"
                            />
                          ) : (
                            <input
                              value={settings[field.key] ?? ""}
                              onChange={(event) => setSettings((current) => ({ ...current, [field.key]: event.target.value }))}
                              className="mt-2 w-full rounded-full border border-slate-200 px-4 py-3 text-sm outline-none"
                            />
                          )}
                        </div>
                      ))}
                    </div>
                  </div>
                ))}

                <div className="flex justify-end">
                  <button type="submit" disabled={saving} className="rounded-full bg-orange-600 px-5 py-3 text-sm font-semibold text-white transition hover:bg-orange-700 disabled:opacity-60">
                    {saving ? "Saving..." : "Save Settings"}
                  </button>
                </div>
              </form>
            )}
          </section>
        </div>
      </main>
      <Footer />
    </div>
  );
}
