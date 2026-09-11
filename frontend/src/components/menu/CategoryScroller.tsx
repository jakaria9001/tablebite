import { useMemo, useRef, useState } from "react";
import type { Category } from "../../types/menu";
import { CategoryCard } from "./CategoryCard";

interface CategoryScrollerProps {
  categories: Category[];
  onSelect: (slug: string) => void;
}

export function CategoryScroller({ categories, onSelect }: CategoryScrollerProps) {
  const [startIndex, setStartIndex] = useState(0);
  const containerRef = useRef<HTMLDivElement>(null);

  const visibleCount = useMemo(() => {
    if (typeof window === "undefined") return 4;
    if (window.innerWidth < 640) return 2;
    if (window.innerWidth < 1024) return 3;
    return 4;
  }, []);

  const maxStartIndex = Math.max(0, categories.length - visibleCount);

  const handlePrev = () => setStartIndex((value) => Math.max(0, value - 1));
  const handleNext = () => setStartIndex((value) => Math.min(maxStartIndex, value + 1));

  const visibleCategories = categories.slice(startIndex, startIndex + visibleCount);

  return (
    <div className="flex items-center gap-2 sm:gap-3">
      <button
        type="button"
        onClick={handlePrev}
        disabled={startIndex === 0}
        className="grid h-9 w-9 shrink-0 place-items-center rounded-full border border-slate-200 bg-white text-lg font-bold text-slate-700 shadow-sm transition hover:border-orange-300 hover:text-orange-600 disabled:cursor-not-allowed disabled:opacity-40 sm:h-11 sm:w-11 sm:text-xl"
        aria-label="Scroll categories left"
      >
        ←
      </button>

      <div ref={containerRef} className="flex-1 overflow-hidden">
        <div className="flex gap-3 transition-all duration-200 sm:gap-4">
          {visibleCategories.map((category) => (
            <div key={category.id} className="w-[calc(50%-0.375rem)] shrink-0 sm:w-[calc(33.333%-0.666rem)] lg:w-[calc(25%-0.75rem)]">
              <CategoryCard category={category} onClick={() => onSelect(category.slug)} />
            </div>
          ))}
        </div>
      </div>

      <button
        type="button"
        onClick={handleNext}
        disabled={startIndex >= maxStartIndex}
        className="grid h-9 w-9 shrink-0 place-items-center rounded-full border border-slate-200 bg-white text-lg font-bold text-slate-700 shadow-sm transition hover:border-orange-300 hover:text-orange-600 disabled:cursor-not-allowed disabled:opacity-40 sm:h-11 sm:w-11 sm:text-xl"
        aria-label="Scroll categories right"
      >
        →
      </button>
    </div>
  );
}
