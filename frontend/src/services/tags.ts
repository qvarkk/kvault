import type { AxiosRequestConfig } from "axios"
import http from "./http"
import type { Tag, Paginated } from "@/types"

export const tagsService = {
  list(params?: Record<string, unknown>, config?: AxiosRequestConfig) {
    return http.get<Paginated<Tag>>("/tags", { params, ...config })
  },

  create(name: string) {
    return http.post<Tag>("/tags", { name })
  },

  update(id: string, name: string) {
    return http.patch<Tag>(`/tags/${id}`, { name })
  },

  remove(id: string) {
    return http.delete(`/tags/${id}`)
  },
}
