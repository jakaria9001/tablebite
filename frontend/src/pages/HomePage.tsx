import { useMemo } from "react";
import { useMenu } from "../hooks/useMenu";
import { useRestaurantSettings } from "../hooks/useRestaurantSettings";
import { Footer } from "../components/layout/Footer";
import { Header } from "../components/layout/Header";
import { Hero } from "../components/home/Hero";
import { FeaturedSection } from "../components/home/FeaturedSection";
import { PromoBanner } from "../components/home/PromoBanner";
import { Top10Section } from "../components/home/Top10Section";
import { SectionHeading } from "../components/common/SectionHeading";
import { GalleryPreview } from "../components/home/GalleryPreview";
import { ReviewsSection } from "../components/home/ReviewsSection";
import { SeoHead } from "../components/seo/SeoHead";
import { StructuredData } from "../components/seo/StructuredData";

export default function HomePage() {
  const { data, loading, error } = useMenu();
  const { settings } = useRestaurantSettings();
  const heroTitle = settings.restaurant_name?.trim() || "Restaurant";
  const address = settings.address?.trim() || "Deorail, Badarpur, Karimganj - 788806";
  const hours = settings.hours_monday?.trim() || "Open daily from 09:00 AM to 11:00 PM";
  const googleMapsUrl = settings.google_maps_url?.trim() || "https://www.google.com/maps/embed?pb=!1m18!1m12!1m3!1d3619.86365547279!2d92.54719528595412!3d24.868505888029183!2m3!1f0!2f0!3f0!3m2!1i1024!2i768!4f13.1!3m3!1m2!1s0x374e2d0012d9376b%3A0x35e5aa792d50570a!2sIndian%20Restaurant%20and%20Sweets!5e0!3m2!1sen!2sus!4v1788726138104!5m2!1sen!2sus";

  const categories = useMemo(() => data?.categories.slice(0, 8) ?? [], [data]);
  const featuredItems = useMemo(() => data?.items.filter((item) => item.is_available).slice(0, 6) ?? [], [data]);
  const starterItems = useMemo(() => {
    if (!data?.items) return [];

    const starterCategoryIds = new Set([9, 10]);
    const starters = data.items.filter((item) => item.is_available && starterCategoryIds.has(item.category_id));

    if (starters.length <= 4) return starters;

    const shuffled = [...starters];
    for (let index = shuffled.length - 1; index > 0; index -= 1) {
      const swapIndex = Math.floor(Math.random() * (index + 1));
      [shuffled[index], shuffled[swapIndex]] = [shuffled[swapIndex], shuffled[index]];
    }

    return shuffled.slice(0, 4);
  }, [data]);
  const topDishes = useMemo(() => data?.items.filter((item) => item.is_available).slice(0, 8) ?? [], [data]);

  return (
    <div className="min-h-screen bg-[#fffaf5] text-slate-900">
      <SeoHead title="Home" description="Indian Restaurant & Sweets in Badarpur, Karimganj offers fresh Indian food, Chinese starters, shakes, coffee, and sweets with a modern dine-in QR ordering experience." />
      <StructuredData />
      <Header />
      <main className="pb-16">
        <Hero />

        {loading ? <p className="mx-auto max-w-7xl px-5 py-10 lg:px-8">Loading categories…</p> : error ? <p className="mx-auto max-w-7xl px-5 py-10 text-red-600 lg:px-8">{error}</p> : (
          <FeaturedSection categories={categories} items={featuredItems} />
        )}

        <PromoBanner />

        {!loading && !error ? <Top10Section items={topDishes} starters={starterItems} /> : null}

        <section id="about" className="mx-auto max-w-7xl px-5 py-8 lg:px-8">
          <div className="rounded-[2rem] border border-black/5 bg-white p-8 shadow-sm md:p-10">
            <SectionHeading eyebrow="About us" title="A warm, memorable dining experience from first bite to last." description="We’re dedicated to serving comfort food with a modern touch, blending traditional flavours with a seamless digital experience for guests and staff alike." />
          </div>
        </section>

        <GalleryPreview />
        <ReviewsSection />

        <section id="contact" className="mx-auto max-w-7xl px-5 py-8 lg:px-8">
          <div className="grid gap-6 rounded-[2rem] border border-black/5 bg-white p-8 shadow-sm md:grid-cols-[1fr_.9fr] md:p-10">
            <div>
              <p className="text-sm font-bold uppercase tracking-[0.25em] text-orange-600">Visit us</p>
              <h2 className="mt-2 text-3xl font-black">Come by for a meal, a drink, or a long chat.</h2> <br/>
              <p className="mt-4 text-lg leading-8 text-slate-600">{address}<br />{hours}</p>
            </div>
            <div className="overflow-hidden rounded-[1.5rem] border border-orange-200 bg-orange-50">
              <iframe
                src={googleMapsUrl}
                title={`${heroTitle} location`}
                className="h-64 w-full border-0"
                loading="lazy"
                referrerPolicy="strict-origin-when-cross-origin"
              />
            </div>
          </div>
        </section>
      </main>

      <Footer />
    </div>
  );
}
