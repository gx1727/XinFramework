import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export function resolveAssetUrl(url: string | undefined | null): string {
  if (!url) return ""
  if (url.startsWith("http://") || url.startsWith("https://") || url.startsWith("blob:") || url.startsWith("data:")) {
    return url
  }
  const base = import.meta.env.VITE_ASSET_BASE_URL || ""
  return `${base}${url.startsWith("/") ? "" : "/"}${url}`
}
