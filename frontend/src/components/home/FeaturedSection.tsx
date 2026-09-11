import { useNavigate } from "react-router-dom";
import type { Category, MenuItem } from "../../types/menu";
import { SectionHeading } from "../common/SectionHeading";
import { CategoryScroller } from "../menu/CategoryScroller";
import { MenuGrid } from "../menu/MenuGrid";

interface FeaturedSectionProps {
  categories: Category[];
  items: MenuItem[];
}

export function FeaturedSection({ categories, items }: FeaturedSectionProps) {
  const navigate = useNavigate();

  return (
    <>
      <section id="menu" className="mx-auto max-w-7xl px-5 py-10 lg:px-8">
        <div className="mb-7 flex items-end justify-between gap-5">
          <SectionHeading eyebrow="Explore our menu" title="Pick a category and dive in" />
          <button onClick={() => navigate("/menu")} className="text-sm font-bold text-orange-600">See full menu →</button>
        </div>
        <CategoryScroller categories={categories} onSelect={(slug) => navigate(`/menu#${slug}`)} />
      </section>

      <section className="mx-auto max-w-7xl px-5 py-8 lg:px-8">
        <div className="rounded-[2rem] border border-orange-100 bg-[#fff3e4] p-8 shadow-sm md:p-10">
          <div className="mb-7 flex items-center justify-between gap-4">
            <SectionHeading eyebrow="Fan favourites" title="The dishes guests keep coming back for" />
            <span className="rounded-full bg-white px-4 py-2 text-sm font-semibold text-slate-700">New this week</span>
          </div>
          <MenuGrid items={items} />
        </div>
      </section>
    </>
  );
}
