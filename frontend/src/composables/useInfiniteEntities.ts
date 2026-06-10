import type { ApiError, Paginated } from "@/types"
import type { AxiosRequestConfig, AxiosResponse } from "axios"
import { ref, shallowRef } from "vue"

interface InfiniteService<T> {
  list(
    params?: Record<string, unknown>,
    config?: AxiosRequestConfig,
  ): Promise<AxiosResponse<Paginated<T>>>
}

export function useInfiniteEntities<T>(service: InfiniteService<T>) {
  const entities = shallowRef<T[]>([])
  const loading = ref(false)
  const loadingMore = ref(false)
  const error = ref<ApiError | null>(null)
  const page = ref(1)
  const hasMore = ref(true)

  let currentRequest: AbortController | null = null

  async function fetchEntities(params?: Record<string, unknown>) {
    if (currentRequest) currentRequest.abort()
    currentRequest = new AbortController()

    error.value = null
    page.value = 1
    hasMore.value = true

    let loadingTimer = setTimeout(() => {
      loading.value = true
      entities.value = []
    }, 300)

    try {
      const response = await service.list(
        { ...params, page: 1 },
        { signal: currentRequest.signal },
      )
      entities.value = response.data.data
      hasMore.value =
        response.data.page * response.data.pageSize < response.data.total
    } catch (e) {
      if ((e as any)?.code === "ERR_CANCELED") return
      error.value = e as ApiError
    } finally {
      clearTimeout(loadingTimer)
      loading.value = false
    }
  }

  async function loadMore(params?: Record<string, unknown>) {
    if (!hasMore.value || loadingMore.value || loading.value) return
    loadingMore.value = true

    try {
      page.value++
      const response = await service.list({ ...params, page: page.value })
      entities.value = [...entities.value, ...response.data.data]
      hasMore.value =
        response.data.page * response.data.pageSize < response.data.total
    } catch (e) {
      error.value = e as ApiError
    } finally {
      loadingMore.value = false
    }
  }

  return {
    entities,
    loading,
    loadingMore,
    error,
    hasMore,
    fetchEntities,
    loadMore,
  }
}
