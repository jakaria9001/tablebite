interface MenuBadgeProps {
  label: string;
  tone?: "default" | "accent" | "success";
}

export function MenuBadge({ label, tone = "default" }: MenuBadgeProps) {
  const toneStyles = {
    default: "bg-slate-100 text-slate-700",
    accent: "bg-orange-100 text-orange-700",
    success: "bg-emerald-100 text-emerald-800",
  };

  return <span className={`rounded-full px-3 py-1 text-sm font-semibold ${toneStyles[tone]}`}>{label}</span>;
}
