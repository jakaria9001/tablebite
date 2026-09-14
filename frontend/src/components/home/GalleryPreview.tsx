import { useEffect, useState } from "react";
import { getPublicGalleryImages } from "../../api/gallery";
import { SectionHeading } from "../common/SectionHeading";

export function GalleryPreview() {
  const [images, setImages] = useState<Array<{ image_url: string; caption: string }>>([]);

  useEffect(() => {
    const load = async () => {
      try {
        const data = await getPublicGalleryImages(3);
        setImages(data.images);
      } catch {
        setImages([]);
      }
    };

    void load();
  }, []);

  return (
    <section id="gallery" className="scroll-mt-24 mx-auto max-w-7xl px-5 py-8 lg:px-8">
      <div className="mb-6">
        <SectionHeading eyebrow="Gallery" title="A snapshot of the atmosphere and the ambience" />
      </div>
      <div className="grid gap-5 md:grid-cols-3">
        {images.map((image, index) => (
          <div key={`${image.image_url}-${index}`} className="overflow-hidden rounded-[2rem] border border-black/5 bg-white shadow-sm">
            <img src={image.image_url} alt={image.caption || "Gallery view"} className="aspect-[4/3] w-full object-cover" />
            <div className="p-5 text-center">
              <p className="text-lg font-bold">Gallery view {index + 1}</p>
              <p className="mt-2 text-sm text-slate-600">{image.caption || "Freshly plated, warmly presented, and always ready to share."}</p>
            </div>
          </div>
        ))}
      </div>
    </section>
  );
}
