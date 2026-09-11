interface SectionHeadingProps {
  eyebrow: string;
  title: string;
  description?: string;
  align?: "left" | "center";
}

export function SectionHeading({ eyebrow, title, description, align = "left" }: SectionHeadingProps) {
  const alignment = align === "center" ? "items-center text-center" : "items-start text-left";

  return (
    <div className={`flex flex-col gap-2 ${alignment}`}>
      <p className="text-sm font-bold uppercase tracking-[0.25em] text-orange-600">{eyebrow}</p>
      <h2 className="text-3xl font-black text-slate-950">{title}</h2>
      {description ? <p className="max-w-2xl text-lg leading-8 text-slate-600">{description}</p> : null}
    </div>
  );
}
