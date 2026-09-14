import { useNavigate } from "react-router-dom";
import { useRestaurantSettings } from "../../hooks/useRestaurantSettings";
import { FallbackImage } from "../common/FallbackImage";

export function Hero() {
  const navigate = useNavigate();
  const { settings } = useRestaurantSettings();
  const restaurantName = settings.restaurant_name?.trim() || "Restaurant";
  const tagline = settings.tagline?.trim() || "Indian food, starters, shakes & sweets in Badarpur.";
  const heroImage = settings.hero_image?.trim() || "/gallery-1.jpg";
  const heroAlt = `${restaurantName} hero`;

  return (
    <section className="mx-auto max-w-7xl px-4 py-5 sm:px-5 lg:px-8 lg:py-8">
      <div className="overflow-hidden rounded-[2rem] border border-[var(--line)] bg-[var(--maroon-950)] shadow-[var(--shadow-lifted)]">
        <div className="grid lg:grid-cols-[0.9fr_1.1fr]">
          <div className="relative z-10 flex flex-col justify-center px-6 py-10 text-white sm:px-10 sm:py-14 lg:px-12">
            <span className="mb-5 w-fit rounded-full border border-[var(--gold-500)]/50 bg-[var(--gold-500)]/15 px-4 py-2 text-xs font-bold uppercase tracking-[0.18em] text-[var(--gold-200)]">{tagline}</span>
            <h1 className="max-w-3xl font-serif text-4xl font-black leading-[1.05] tracking-tight sm:text-5xl lg:text-6xl">
              {restaurantName}
            </h1>
            <p className="mt-5 max-w-xl text-base leading-7 text-white/75 sm:text-lg">
              Enjoy fresh Indian food, Chinese starters, coffee, shakes, and sweets with a simple dine-in QR ordering experience.
            </p>
            <div className="mt-8 flex flex-wrap gap-3">
              <button onClick={() => navigate("/menu")} className="min-h-12 rounded-full bg-[var(--gold-500)] px-6 py-3 font-bold text-[var(--maroon-950)] transition duration-200 hover:bg-[var(--gold-200)] active:scale-[0.98]">
                View Menu
              </button>
              <button onClick={() => navigate("/menu")} className="min-h-12 rounded-full border border-white/30 bg-white/10 px-6 py-3 font-bold text-white transition duration-200 hover:bg-white/20 active:scale-[0.98]">
                Order at Table
              </button>
            </div>
          </div>
          <div className="relative order-first min-h-[250px] overflow-hidden sm:min-h-[380px] lg:order-none lg:min-h-[500px]">
            <FallbackImage src={heroImage} alt={heroAlt} loading="eager" fetchPriority="high" sizes="(max-width: 1024px) 100vw, 55vw" className="absolute inset-0 h-full w-full object-cover object-center" />
            <div className="absolute inset-0 bg-gradient-to-r from-[var(--maroon-950)]/50 via-transparent to-transparent lg:from-[var(--maroon-950)]/20" />
            <div className="absolute bottom-5 right-5 rounded-full border border-white/25 bg-black/25 px-3 py-1.5 text-xs font-semibold text-white/85 backdrop-blur-sm">Fresh from our kitchen</div>
          </div>
        </div>
      </div>
    </section>
  );
}
