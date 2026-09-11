import { useEffect } from "react";
import { useLocation } from "react-router-dom";

interface SeoHeadProps {
  title?: string;
  description?: string;
  canonicalPath?: string;
  imageUrl?: string;
  type?: string;
  noIndex?: boolean;
  structuredData?: Record<string, unknown> | Record<string, unknown>[];
}

const defaultTitle = "Indian Restaurant & Sweets | Badarpur, Karimganj";
const defaultDescription = "Indian Restaurant & Sweets in Badarpur, Karimganj serves fresh Indian food, Chinese starters, shakes, coffee, and sweets with a modern dine-in QR ordering experience.";
const defaultImage = "/indian_restaurant_logo.jpg";
const defaultKeywords = "Indian Restaurant Badarpur, Indian Restaurant & Sweets, Karimganj restaurant, Chinese starters, shakes, coffee, sweets, dine-in QR ordering";

function setMeta(name: string, content: string | null, attr = "name") {
  if (typeof document === "undefined") {
    return;
  }

  const tag = document.querySelector(`meta[${attr}="${name}"]`);
  if (!tag) {
    const nextTag = document.createElement("meta");
    nextTag.setAttribute(attr, name);
    document.head.appendChild(nextTag);
  }

  const activeTag = document.querySelector(`meta[${attr}="${name}"]`);
  if (activeTag) {
    activeTag.setAttribute("content", content ?? "");
  }
}

export function SeoHead({
  title,
  description,
  canonicalPath,
  imageUrl = defaultImage,
  type = "website",
  noIndex = false,
  structuredData,
}: SeoHeadProps) {
  const location = useLocation();

  useEffect(() => {
    if (typeof window === "undefined") {
      return;
    }

    const origin = window.location.origin;
    const currentPath = `${location.pathname}${location.search}`;
    const canonicalUrl = new URL(canonicalPath ?? currentPath, origin).toString();
    const pageTitle = title ? `${title} | Indian Restaurant & Sweets` : defaultTitle;
    const pageDescription = description ?? defaultDescription;
    const image = new URL(imageUrl, origin).toString();

    document.title = pageTitle;

    setMeta("description", pageDescription);
    setMeta("keywords", defaultKeywords);
    setMeta("robots", noIndex ? "noindex, nofollow" : "index, follow");
    setMeta("theme-color", "#fffaf5");
    setMeta("og:title", pageTitle, "property");
    setMeta("og:description", pageDescription, "property");
    setMeta("og:type", type, "property");
    setMeta("og:url", canonicalUrl, "property");
    setMeta("og:image", image, "property");
    setMeta("og:site_name", "Indian Restaurant & Sweets", "property");
    setMeta("twitter:card", "summary_large_image");
    setMeta("twitter:title", pageTitle);
    setMeta("twitter:description", pageDescription);
    setMeta("twitter:image", image);

    let canonicalLink = document.querySelector("link[rel='canonical']") as HTMLLinkElement | null;
    if (!canonicalLink) {
      canonicalLink = document.createElement("link");
      canonicalLink.rel = "canonical";
      document.head.appendChild(canonicalLink);
    }
    canonicalLink.href = canonicalUrl;

    const existingSchema = document.getElementById("seo-structured-data");
    if (existingSchema) {
      existingSchema.remove();
    }

    if (structuredData) {
      const script = document.createElement("script");
      script.id = "seo-structured-data";
      script.type = "application/ld+json";
      script.textContent = JSON.stringify(structuredData);
      document.head.appendChild(script);
    }
  }, [canonicalPath, description, imageUrl, location.pathname, location.search, noIndex, structuredData, title, type]);

  return null;
}
