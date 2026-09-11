import type { MenuItem } from "../../types/menu";
import { MenuCard } from "./MenuCard";

interface MenuGridProps {
  items: MenuItem[];
  columns?: "default" | "compact";
}

export function MenuGrid({ items, columns = "default" }: MenuGridProps) {
  const layoutClass = columns === "compact" ? "grid gap-4 md:grid-cols-2" : "grid gap-5 md:grid-cols-2 xl:grid-cols-3";

  return <div className={layoutClass}>{items.map((item) => <MenuCard key={item.id} item={item} />)}</div>;
}
