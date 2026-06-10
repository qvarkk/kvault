import { toast } from "vue-sonner"
import i18n from "@/i18n"
import type { ApiError } from "@/types"

/**
 * Resolve any thrown value into a user-facing message, with robust fallbacks so
 * a toast is never empty. Covers: our RFC7807-ish ApiError, the network-error
 * shape (status 0), and anything malformed (e.g. a proxy returning a 500 HTML
 * page when the API is down, which has no `detail`).
 */
export function resolveApiError(e: unknown): {
  title: string
  description?: string
} {
  const t = i18n.global.t
  const err = e as Partial<ApiError> | null | undefined

  // Network / no-response errors are tagged with status 0 by the http interceptor.
  if (err && err.status === 0) {
    return {
      title: t("errors.network.title"),
      description: err.detail || t("errors.network.detail"),
    }
  }

  // Well-formed API error with a usable message.
  if (err && typeof err.detail === "string" && err.detail.trim() !== "") {
    return {
      title: err.title || t("errors.internal.title"),
      description: err.detail,
    }
  }

  // Anything else (malformed body, thrown non-ApiError, missing detail).
  return {
    title: t("errors.internal.title"),
    description: t("errors.internal.detail"),
  }
}

/** Show an error toast for any thrown value, never empty. */
export function toastApiError(e: unknown): void {
  const { title, description } = resolveApiError(e)
  toast.error(title, { description })
}
