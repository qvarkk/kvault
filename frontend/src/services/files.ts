import http from "./http"
import type { File, Paginated, PresignedUrl } from "@/types"

export const filesService = {
  list(params?: Record<string, unknown>) {
    return http.get<Paginated<File>>("/files", { params })
  },

  getDownloadUrl(id: string) {
    return http.get<{ url: string }>(`/files/${id}`)
  },

  getViewUrl(id: string) {
    return http.get<PresignedUrl>(`/files/${id}/view`)
  },

  getInfo(id: string) {
    return http.get<File>(`/files/${id}/info`)
  },

  upload(file: globalThis.File, onProgress?: (percent: number) => void, signal?: AbortSignal) {
    const formData = new FormData()
    formData.append("file", file)
    return http.post<File>("/files/upload", formData, {
      signal,
      headers: { "Content-Type": "multipart/form-data" },
      onUploadProgress(event) {
        if (onProgress && event.total) {
          onProgress(Math.round((event.loaded / event.total) * 100))
        }
      },
    })
  },

  remove(id: string) {
    return http.delete(`/files/${id}`)
  },

  restore(id: string) {
    return http.post(`/files/${id}/restore`)
  },

  listDeleted(params?: Record<string, unknown>) {
    return http.get<Paginated<File>>("/files/deleted", { params })
  },

  permanentDelete(id: string) {
    return http.delete(`/files/deleted/${id}`)
  },

  emptyTrash() {
    return http.delete("/files/deleted")
  },
}
