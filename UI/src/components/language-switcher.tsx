import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { useLocaleStore, type Locale } from "@/stores/localeStore"

const localeNames: Record<Locale, string> = {
  "zh-CN": "简体中文",
  "en-US": "English",
}

export function LanguageSwitcher() {
  const locale = useLocaleStore((state) => state.locale)
  const setLocale = useLocaleStore((state) => state.setLocale)

  return (
    <Select value={locale} onValueChange={(value) => setLocale(value as Locale)}>
      <SelectTrigger className="w-[120px]">
        <SelectValue>{localeNames[locale]}</SelectValue>
      </SelectTrigger>
      <SelectContent>
        <SelectItem value="zh-CN">{localeNames["zh-CN"]}</SelectItem>
        <SelectItem value="en-US">{localeNames["en-US"]}</SelectItem>
      </SelectContent>
    </Select>
  )
}