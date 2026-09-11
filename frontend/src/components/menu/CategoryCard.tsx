import { FallbackImage } from "../common/FallbackImage";
import type { Category } from "../../types/menu";

interface Props {
  category: Category;
  onClick: () => void;
}

export function CategoryCard({ category, onClick }: Props) {
  return (
    <button onClick={onClick} className="group w-full overflow-hidden rounded-3xl border border-black/5 bg-white text-left shadow-sm transition hover:-translate-y-1 hover:shadow-lg">
      <div className="aspect-[1.1] overflow-hidden bg-orange-50 sm:aspect-[1.2]">
        {category.image_url ? (
          <FallbackImage src={category.image_url} alt={category.name} className="h-full w-full object-cover transition duration-500 group-hover:scale-105" />
        ) : (
          <FallbackImage src="/indian_restaurant_logo.jpg" alt={category.name} className="h-full w-full object-cover transition duration-500 group-hover:scale-105" />
        )}
      </div>
      <div className="px-3 py-3 text-sm font-bold sm:px-4">{category.name}</div>
    </button>
  );
}
