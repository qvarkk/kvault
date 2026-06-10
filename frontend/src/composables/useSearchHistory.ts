import { useStorage } from "@vueuse/core"

const MAX_ENTRIES = 10
const STORAGE_KEY = "kvault-search-history"

// Shared across all callers so the palette and the view stay in sync.
const history = useStorage<string[]>(STORAGE_KEY, [])

export function useSearchHistory() {
  function add(query: string) {
    const q = query.trim()
    if (!q) return
    history.value = [q, ...history.value.filter((entry) => entry !== q)].slice(
      0,
      MAX_ENTRIES,
    )
  }

  function remove(query: string) {
    history.value = history.value.filter((entry) => entry !== query)
  }

  function clear() {
    history.value = []
  }

  return { history, add, remove, clear }
}
