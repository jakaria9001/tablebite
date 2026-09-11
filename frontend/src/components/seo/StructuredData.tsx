import { useMemo } from "react";
import { useRestaurantSettings } from "../../hooks/useRestaurantSettings";

export function StructuredData() {
  const { settings } = useRestaurantSettings();

  const data = useMemo(() => {
    const restaurantName = settings.restaurant_name?.trim() || "Indian Restaurant & Sweets";
    const address = settings.address?.trim() || "Deorail, Badarpur, Karimganj - 788806";
    const phone = settings.phone?.trim() || "+91 00000 00000";
    const cuisine = settings.cuisine?.trim() || "Indian, Chinese";

    return {
      "@context": "https://schema.org",
      "@type": "Restaurant",
      name: restaurantName,
      image: ["/indian_restaurant_logo.jpg"],
      telephone: phone,
      address: {
        "@type": "PostalAddress",
        streetAddress: address,
        addressLocality: "Badarpur",
        addressRegion: "Karimganj",
        postalCode: "788806",
        addressCountry: "IN",
      },
      servesCuisine: cuisine.split(/[,/]+/).map((entry) => entry.trim()).filter(Boolean),
      openingHoursSpecification: [
        {
          "@type": "OpeningHoursSpecification",
          dayOfWeek: ["Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"],
          opens: "09:00",
          closes: "23:00",
        },
      ],
      priceRange: "₹₹",
      aggregateRating: {
        "@type": "AggregateRating",
        ratingValue: "4.8",
        reviewCount: "100",
      },
      url: "http://localhost:5173/",
    };
  }, [settings]);

  return <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(data) }} />;
}
