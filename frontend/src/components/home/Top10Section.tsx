import type { MenuItem } from "../../types/menu";
import { SectionHeading } from "../common/SectionHeading";
import { MenuBadge } from "../menu/MenuBadge";

interface Top10SectionProps {
  items: MenuItem[];
  starters: MenuItem[];
}

export function Top10Section({ items, starters }: Top10SectionProps) {
  return (
    <section className="mx-auto max-w-7xl px-5 py-8 lg:px-8">
      <div className="grid gap-6 lg:grid-cols-[1.05fr_0.95fr]">
        <div className="rounded-[2rem] border border-black/5 bg-white p-8 shadow-sm">
          <SectionHeading eyebrow="Our top 10 dishes" title="A quick taste of the house favourites" />
          <ul className="mt-6 space-y-3">
            {items.slice(0, 6).map((item) => (
              <li key={item.id} className="flex items-center justify-between rounded-2xl border border-slate-100 bg-slate-50 px-4 py-3">
                <span className="font-semibold">{item.name}</span>
                <MenuBadge label={item.is_veg ? "Veg" : "Non-veg"} tone="default" />
              </li>
            ))}
          </ul>
        </div>
        <div className="rounded-[2rem] border border-orange-100 bg-[#fff7ed] p-8 shadow-sm">
          <SectionHeading eyebrow="Starters" title="Start your meal with something irresistible" />
          <div className="mt-6 grid gap-4 sm:grid-cols-2">
            {starters.map((item) => (
              <div key={item.id} className="relative overflow-hidden rounded-[1.5rem] border border-white/70 shadow-sm">
                <img
                  src={item.image_url || "/indian_restaurant_logo.jpg"}
                  alt={item.name}
                  className="aspect-[4/3] w-full object-cover"
                  onError={(event) => {
                    event.currentTarget.src = "/indian_restaurant_logo.jpg";
                  }}
                />
                <div className="absolute inset-0 bg-gradient-to-t from-black/70 via-black/20 to-transparent" />
                <div className="absolute inset-x-0 bottom-0 p-4 text-white">
                  <p className="text-lg font-black">{item.name}</p>
                  <p className="mt-1 text-sm leading-6 text-white/90">{item.description}</p>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </section>
  );
}
