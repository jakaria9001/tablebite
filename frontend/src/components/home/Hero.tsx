import { useNavigate } from "react-router-dom";
import { useRestaurantSettings } from "../../hooks/useRestaurantSettings";

export function Hero() {
  const navigate = useNavigate();
  const { settings } = useRestaurantSettings();
  const restaurantName = settings.restaurant_name?.trim() || "Restaurant";
  const tagline = settings.tagline?.trim() || "Indian food, starters, shakes & sweets in Badarpur.";
  const heroImage = settings.hero_image?.trim() || "/hero.png";
  const heroAlt = `${restaurantName} hero`;

  return (
    <section className="mx-auto max-w-7xl px-5 py-8 lg:px-8 lg:py-10">
      <div className="overflow-hidden rounded-[2.5rem] border border-orange-100 bg-[linear-gradient(135deg,#fff6eb,#ffe7d0)] shadow-[0_30px_70px_rgba(15,23,42,0.08)]">
        <div className="grid gap-10 px-6 py-10 lg:grid-cols-[1.05fr_.95fr] lg:px-10 lg:py-14">
          <div className="flex flex-col justify-center">
            <span className="mb-5 w-fit rounded-full bg-orange-100 px-4 py-2 text-sm font-bold text-orange-700">{tagline}</span>
            <h1 className="max-w-3xl text-4xl font-black tracking-tight text-slate-950 sm:text-5xl lg:text-6xl">
              {restaurantName}
            </h1>
            <p className="mt-6 max-w-xl text-lg leading-8 text-slate-600">
              Enjoy fresh Indian food, Chinese starters, coffee, shakes, and sweets with a simple dine-in QR ordering experience.
            </p>
            <div className="mt-8 flex flex-wrap gap-3">
              <button onClick={() => navigate("/menu")} className="rounded-full bg-orange-600 px-6 py-3 font-bold text-white transition hover:bg-orange-700">
                View Menu
              </button>
              <button onClick={() => navigate("/menu")} className="rounded-full border border-slate-300 bg-white px-6 py-3 font-bold text-slate-900 transition hover:border-orange-300 hover:text-orange-700">
                Order at Table
              </button>
            </div>
          </div>
          <div className="rounded-[2rem] border border-white/70 bg-white/70 p-4 shadow-inner">
            <div className="flex h-full min-h-[320px] items-center justify-center overflow-hidden rounded-[1.6rem] bg-slate-100">
              <img
                src={heroImage}
                alt={heroAlt}
                loading="lazy"
                className="h-full w-full object-cover"
              />
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
