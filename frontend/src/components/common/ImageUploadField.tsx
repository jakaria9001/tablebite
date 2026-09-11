import { useState, type ChangeEvent } from "react";

interface ImageUploadFieldProps {
  value: string;
  onChange: (value: string) => void;
  label: string;
  placeholder?: string;
}

const MAX_FILE_SIZE_BYTES = 5 * 1024 * 1024;
const ACCEPTED_TYPES = ["image/jpeg", "image/png", "image/webp"];

export function ImageUploadField({ value, onChange, label, placeholder }: ImageUploadFieldProps) {
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleFileChange = async (event: ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (!file) {
      return;
    }

    if (!ACCEPTED_TYPES.includes(file.type)) {
      setError("Please choose a JPG, PNG, or WebP image.");
      event.target.value = "";
      return;
    }

    if (file.size > MAX_FILE_SIZE_BYTES) {
      setError("Please keep the file under 5 MB.");
      event.target.value = "";
      return;
    }

    setUploading(true);
    setError(null);

    try {
      const formData = new FormData();
      formData.append("file", file);
      const baseUrl = import.meta.env.VITE_API_BASE_URL ?? "";
      const response = await fetch(`${baseUrl}/api/v1/admin/upload`, {
        method: "POST",
        credentials: "include",
        body: formData,
      });

      const responseText = await response.text();
      let payload: unknown = {};
      try {
        payload = responseText ? JSON.parse(responseText) : {};
      } catch {
        payload = {};
      }

      if (!response.ok) {
        const message = typeof payload === "object" && payload && "error" in (payload as Record<string, unknown>)
          ? String((payload as Record<string, unknown>).error)
          : responseText || "Image upload failed.";
        throw new Error(message);
      }

      const uploadedUrl = typeof payload === "object" && payload && "url" in (payload as Record<string, unknown>) ? String((payload as Record<string, unknown>).url) : "";
      if (!uploadedUrl) {
        throw new Error("The server did not return an image URL.");
      }

      onChange(uploadedUrl);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Unable to upload image.");
    } finally {
      setUploading(false);
      event.target.value = "";
    }
  };

  return (
    <div className="space-y-2">
      <label className="text-sm font-semibold text-slate-700">{label}</label>
      <input
        value={value}
        onChange={(event) => {
          setError(null);
          onChange(event.target.value);
        }}
        placeholder={placeholder}
        className="mt-2 w-full rounded-full border border-slate-200 px-4 py-3 text-sm outline-none"
      />
      <div>
        <label className="inline-flex cursor-pointer items-center gap-2 rounded-full border border-slate-200 bg-slate-50 px-3 py-2 text-sm font-semibold text-slate-700 transition hover:border-orange-300 hover:text-orange-600">
          <input type="file" accept="image/png,image/jpeg,image/webp" onChange={handleFileChange} className="sr-only" />
          {uploading ? "Uploading..." : "Upload image"}
        </label>
        <p className="mt-2 text-xs text-slate-500">Max 5 MB • JPG, PNG, or WebP • images are auto-optimized and stored in Cloudinary.</p>
      </div>
      {error ? <p className="text-sm font-medium text-rose-600">{error}</p> : null}
      {value ? (
        <div className="overflow-hidden rounded-[1rem] border border-slate-200 bg-slate-50 p-2">
          <img src={value} alt={label} className="aspect-[4/3] w-full rounded-[0.75rem] object-cover" loading="lazy" />
        </div>
      ) : null}
    </div>
  );
}
