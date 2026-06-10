import { ref } from "vue"
import type { ApiError, Paginated } from "@/types"
import type { AxiosResponse } from "axios"

interface Service<T> {
  list(params?: Record<string, unknown>): Promise<AxiosResponse<Paginated<T>>>
}

export function useEntities<T>(service: Service<T>) {
  const entities = ref<T[]>([])
  const loading = ref(false)
  const error = ref<ApiError | null>(null)

  async function fetchEntities(params?: Record<string, unknown>) {
    error.value = null
    loading.value = true

    try {
      const response = await service.list(params)
      entities.value = response.data.data
    } catch (e) {
      error.value = e as ApiError
    } finally {
      loading.value = false
    }
  }

  return { entities, loading, error, fetchEntities }
}
