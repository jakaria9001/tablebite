import { useEffect, useState } from "react";
import { getPublicBanners, type Banner } from "../../api/banners";

export function PromoBanner() {
  const [banners, setBanners] = useState<Banner[]>([]);
  const [activeIndex, setActiveIndex] = useState(0);
  const [dragOffset, setDragOffset] = useState(0);

  useEffect(() => {
    void getPublicBanners().then((data) => {
      const active = data.banners.filter((item) => item.is_active).sort((a, b) => b.id - a.id);
      setBanners(active);
      setActiveIndex(0);
    });
  }, []);

  const currentBanner = banners[activeIndex];

  const goToNext = () => {
    setActiveIndex((current) => (current + 1) % banners.length);
  };

  const goToPrev = () => {
    setActiveIndex((current) => (current - 1 + banners.length) % banners.length);
  };

  const handleTouchStart = (event: React.TouchEvent<HTMLDivElement>) => {
    setDragOffset(0);
    const touch = event.touches[0];
    setDragOffset(touch.clientX);
  };

  const handleTouchEnd = (event: React.TouchEvent<HTMLDivElement>) => {
    if (dragOffset === 0) return;
    const touch = event.changedTouches[0];
    const delta = touch.clientX - dragOffset;
    if (delta < -60) {
      goToNext();
    } else if (delta > 60) {
      goToPrev();
    }
    setDragOffset(0);
  };

  if (!currentBanner) return null;

  return (
    <section className="mx-auto max-w-7xl px-5 pt-8 lg:px-8">
      <div className="overflow-hidden rounded-[2rem] border border-orange-100 bg-white text-slate-900 shadow-sm">
        <div className="overflow-hidden" onTouchStart={handleTouchStart} onTouchEnd={handleTouchEnd}>
          <div className="transition-transform duration-300 ease-out" style={{ transform: `translateX(-${activeIndex * 100}%)` }}>
            <div className="flex">
              {banners.map((banner) => (
                <div key={banner.id} className="min-w-full">
                  <div className="relative min-h-[320px] w-full overflow-hidden">
                    {banner.image_url ? <img src={banner.image_url} alt={banner.title} className="absolute inset-0 h-full w-full object-cover" /> : null}
                    <div className="absolute inset-0 bg-gradient-to-t from-black/35 via-black/10 to-transparent" />
                    <div className="relative flex h-full min-h-[320px] flex-col justify-end gap-5 p-8 md:flex-row md:items-end md:justify-between md:p-10">
                      <div className="max-w-2xl text-white">
                        <h2 className="mt-2 text-3xl font-black">{banner.title}</h2>
                        <p className="mt-3 text-base leading-7 text-slate-100">{banner.subtitle}</p>
                      </div>
                      {banner.cta_url ? (
                        <a href={banner.cta_url} target={banner.cta_url.startsWith("http") ? "_blank" : undefined} rel={banner.cta_url.startsWith("http") ? "noreferrer" : undefined} className="rounded-full bg-orange-500 px-5 py-3 font-bold text-white transition hover:bg-orange-400">
                          {banner.cta_label || "Learn more"}
                        </a>
                      ) : null}
                    </div>
                    {banners.length > 1 ? (
                      <div className="absolute inset-x-0 bottom-4 z-10 flex items-center justify-center gap-2">
                        {banners.map((banner, index) => (
                          <button key={banner.id} type="button" onClick={() => setActiveIndex(index)} className={`rounded-full border transition-all ${index === activeIndex ? "h-3.5 w-3.5 border-orange-500 bg-orange-500" : "h-3 w-3 border-white/70 bg-white/80"}`} aria-label={`Go to banner ${index + 1}`} />
                        ))}
                      </div>
                    ) : null}
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
