import { zhCN, type LocaleKeys } from "./zh-CN"
import { enUS } from "./en-US"
import { useLocaleStore, type Locale } from "@/stores/localeStore"

const locales: Record<Locale, LocaleKeys> = {
  "zh-CN": zhCN,
  "en-US": enUS,
}

export function useTranslation() {
  const locale = useLocaleStore((state) => state.locale)
  return locales[locale]
}

export function t(key: string): string {
  const locale = useLocaleStore.getState().locale
  const keys = key.split(".")
  let value: unknown = locales[locale]
  
  for (const k of keys) {
    if (value && typeof value === "object" && k in value) {
      value = (value as Record<string, unknown>)[k]
    } else {
      return key
    }
  }
  
  return typeof value === "string" ? value : key
}

export { zhCN, enUS, type LocaleKeys }
export type { Locale }