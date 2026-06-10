import { formatDistanceToNow, format } from "date-fns"
import { enUS, ru, ja } from "date-fns/locale"
import i18n from "@/i18n"

const locales = { en: enUS, ru, ja }

function currentLocale() {
  return locales[i18n.global.locale.value as keyof typeof locales] ?? enUS
}

export function timeAgo(date: string): string {
  return formatDistanceToNow(new Date(date), {
    addSuffix: true,
    locale: currentLocale(),
  })
}

/** Absolute, human-readable timestamp, e.g. "May 30, 2026, 14:03". */
export function formatDateTime(date: string): string {
  return format(new Date(date), "PPp", { locale: currentLocale() })
}

export function formatBytes(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}
