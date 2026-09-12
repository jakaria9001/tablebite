import { useEffect, useState } from "react";
import { SectionHeading } from "../common/SectionHeading";
import { getPublicReviews, type Review } from "../../api/reviews";

export function ReviewsSection() {
  const [reviews, setReviews] = useState<Review[]>([]);

  useEffect(() => {
    void getPublicReviews(8)
      .then((data) => setReviews(data?.reviews ?? []))
      .catch(() => setReviews([]));
  }, []);

  const marqueeReviews = reviews.length > 0 ? [...reviews, ...reviews] : [];

  return (
    <section id="testimonials" className="mx-auto max-w-7xl px-5 py-8 lg:px-8">
      <div className="rounded-[2rem] bg-[#fff1e6] p-8 shadow-sm md:p-10">
        <SectionHeading eyebrow="What our guests say" title="Loved by locals, perfect for every occasion" />
        <div className="mt-8 overflow-hidden rounded-[2rem] border border-orange-100 bg-white/80 p-3 shadow-inner">
          <div className="marquee flex w-max gap-4">
            {marqueeReviews.map((review, index) => (
              <article key={`${review.id}-${index}`} className="min-h-[220px] w-[300px] rounded-[1.5rem] border border-slate-200 bg-white p-6 shadow-sm">
                <div className="text-xl text-orange-500">{"★".repeat(review.rating)}</div>
                <p
                  className="mt-4 text-sm leading-7 text-slate-700"
                  style={{ display: "-webkit-box", WebkitLineClamp: 7, WebkitBoxOrient: "vertical", overflow: "hidden" }}
                >
                  “{review.review_text}”
                </p>
                <p className="mt-4 font-black text-slate-950">{review.author_name}</p>
                <p className="mt-1 text-sm text-slate-500">{review.source}</p>
              </article>
            ))}
          </div>
        </div>
        <div className="mt-5 flex justify-center">
          <a href="https://www.justdial.com/Karimganj/Indian-Restaurant-And-Sweets-Badarpur/9999P3843-3843-250219073628-Y9X1_BZDET" target="_blank" rel="noreferrer" className="rounded-full bg-orange-600 px-5 py-3 text-sm font-semibold text-white transition hover:bg-orange-700">
            Read more reviews
          </a>
        </div>
      </div>
    </section>
  );
}
