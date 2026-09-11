import { useState, type ComponentPropsWithoutRef } from "react";

function isUsableImageSource(src: string | undefined) {
  if (!src) {
    return false;
  }

  const trimmed = src.trim();
  if (!trimmed || trimmed === "undefined" || trimmed === "null") {
    return false;
  }

  if (trimmed.includes("://undefined") || trimmed.includes("http://undefined") || trimmed.includes("https://undefined")) {
    return false;
  }

  if (trimmed.startsWith("/") || trimmed.startsWith("./") || trimmed.startsWith("../") || trimmed.startsWith("blob:") || trimmed.startsWith("data:")) {
    return true;
  }

  try {
    const url = new URL(trimmed, window.location.origin);
    return url.protocol === "http:" || url.protocol === "https:" || url.protocol === "data:";
  } catch {
    return false;
  }
}

function buildResponsiveImageUrl(src: string | undefined) {
  if (!src) {
    return src;
  }

  try {
    const url = new URL(src);
    if (!url.hostname.includes("cloudinary.com")) {
      return src;
    }

    const params = new URLSearchParams(url.search);
    const hasTransformation = params.has("q") || params.has("f") || params.has("c") || params.has("w") || params.has("dpr");
    if (hasTransformation) {
      return src;
    }

    params.set("q", "auto");
    params.set("f", "auto");
    params.set("c", "fill");
    params.set("w", "800");
    return `${url.origin}${url.pathname}?${params.toString()}`;
  } catch {
    return src;
  }
}

function buildResponsiveImageSrcSet(src: string | undefined) {
  if (!src) {
    return undefined;
  }

  try {
    const url = new URL(src);
    if (!url.hostname.includes("cloudinary.com")) {
      return undefined;
    }

    const params = new URLSearchParams(url.search);
    const hasTransformation = params.has("q") || params.has("f") || params.has("c") || params.has("w") || params.has("dpr");
    if (hasTransformation) {
      return undefined;
    }

    const build = (width: string) => {
      const nextParams = new URLSearchParams(url.search);
      nextParams.set("q", "auto");
      nextParams.set("f", "auto");
      nextParams.set("c", "fill");
      nextParams.set("w", width);
      return `${url.origin}${url.pathname}?${nextParams.toString()} ${width}w`;
    };

    return [build("400"), build("800")].join(", ");
  } catch {
    return undefined;
  }
}

interface FallbackImageProps extends ComponentPropsWithoutRef<"img"> {
  fallbackSrc?: string;
  fallbackClassName?: string;
}

export function FallbackImage({
  src,
  alt,
  fallbackSrc = "/indian_restaurant_logo.jpg",
  className,
  fallbackClassName,
  ...props
}: FallbackImageProps) {
  const [currentSrc, setCurrentSrc] = useState<string | undefined>(src ?? fallbackSrc);
  const [showFallback, setShowFallback] = useState(false);

  const resolvedSrc = showFallback ? fallbackSrc : currentSrc ?? fallbackSrc;
  const responsiveSrc = buildResponsiveImageUrl(resolvedSrc);
  const responsiveSrcSet = buildResponsiveImageSrcSet(resolvedSrc);
  const isSafeSrc = isUsableImageSource(resolvedSrc);

  if (showFallback || !isSafeSrc) {
    return (
      <img
        {...props}
        src={fallbackSrc}
        alt={alt}
        loading={props.loading ?? "lazy"}
        decoding="async"
        className={fallbackClassName ?? className}
      />
    );
  }

  return (
    <img
      {...props}
      src={responsiveSrc}
      srcSet={responsiveSrcSet}
      sizes={props.sizes ?? "(max-width: 640px) 100vw, (max-width: 1024px) 50vw, 33vw"}
      alt={alt}
      loading={props.loading ?? "lazy"}
      decoding="async"
      className={className}
      onError={() => {
        if (currentSrc !== fallbackSrc) {
          setCurrentSrc(fallbackSrc);
          return;
        }
        setShowFallback(true);
      }}
    />
  );
}
