interface VegIndicatorProps {
  isVeg: boolean;
}

export function VegIndicator({ isVeg }: VegIndicatorProps) {
  return (
    <span title={isVeg ? "Vegetarian" : "Non-vegetarian"} className={`mt-1 inline-flex h-4 w-4 shrink-0 rounded-sm border-2 ${isVeg ? "border-emerald-600" : "border-red-600"}`}>
      <span className={`m-auto h-1.5 w-1.5 rounded-full ${isVeg ? "bg-emerald-600" : "bg-red-600"}`} />
    </span>
  );
}
