import { Link } from "react-router-dom";
import { useRestaurantSettings } from "../../hooks/useRestaurantSettings";
import { FallbackImage } from "../common/FallbackImage";

const quickLinks = [
  { label: "Home", href: "/" },
  { label: "Menu", href: "/menu" },
  { label: "About", href: "/#about" },
  { label: "Gallery", href: "/#gallery" },
  { label: "Contact", href: "/#contact" },
];

export function Footer() {
  const { settings } = useRestaurantSettings();
  const restaurantName = settings.restaurant_name?.trim() || "Restaurant";
  const tagline = settings.tagline?.trim() || "Fresh flavours, warm hospitality, and a seamless dining experience from the first bite to the last.";
  const addressLines = settings.address?.split("\n") ?? [];
  const phone = settings.phone?.trim();
  const email = settings.email?.trim();
  const instagram = settings.instagram?.trim();
  const logo = settings.logo?.trim();

  return (
    <footer className="border-t border-black/5 bg-[#1f1a17] text-slate-200">
      <div className="mx-auto grid max-w-7xl gap-8 px-5 py-10 lg:grid-cols-[1.2fr_0.8fr_0.8fr] lg:px-8">
        <div>
          <div className="flex items-center gap-3">
            <FallbackImage src={logo || "/indian_restaurant_logo.jpg"} alt={`${restaurantName} logo`} className="h-14 w-auto rounded-full object-contain" />
            <p className="text-xl font-black tracking-tight text-orange-400">{restaurantName}</p>
          </div>
          <p className="mt-4 max-w-md text-sm leading-7 text-slate-400">{tagline}</p>
        </div>

        <div>
          <h3 className="text-sm font-bold uppercase tracking-[0.3em] text-white">Quick Links</h3>
          <ul className="mt-4 space-y-3 text-sm text-slate-400">
            {quickLinks.map((link) => (
              <li key={link.label}>
                {link.href.startsWith("/") && !link.href.includes("#") ? (
                  <Link to={link.href} className="transition hover:text-orange-400">
                    {link.label}
                  </Link>
                ) : (
                  <a href={link.href} className="transition hover:text-orange-400">
                    {link.label}
                  </a>
                )}
              </li>
            ))}
          </ul>
        </div>

        <div>
          <h3 className="text-sm font-bold uppercase tracking-[0.3em] text-white">Visit Us</h3>
          <ul className="mt-4 space-y-3 text-sm text-slate-400">
            {addressLines.map((line) => (
              <li key={line}>{line}</li>
            ))}
            {phone ? <li>{phone}</li> : null}
            {email ? <li>{email}</li> : null}
          </ul>
        </div>
      </div>

      <div className="border-t border-white/10">
        <div className="mx-auto flex max-w-7xl flex-col gap-3 px-5 py-4 text-sm text-slate-500 sm:flex-row sm:items-center sm:justify-between lg:px-8">
          <p>© 2026 TableBite. All rights reserved.</p>
          <div className="flex gap-4">
            <a href="/" className="transition hover:text-orange-400">Privacy</a>
            <a href="/" className="transition hover:text-orange-400">Terms</a>
            {instagram ? <a href={instagram} className="transition hover:text-orange-400">Instagram</a> : null}
          </div>
        </div>
      </div>
    </footer>
  );
}
