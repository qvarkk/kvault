import http from "./http"
import type {
  BindTagPayload,
  Listed,
  Note,
  NoteUpdatePayload,
  Paginated,
  TagRef,
} from "@/types"

export const notesService = {
  list(params?: Record<string, unknown>) {
    return http.get<Paginated<Note>>("/items", { params })
  },

  createText(title: string) {
    return http.post<Note>("/items", { title, type: "text" })
  },

  createUrl(title: string, sourceUrl: string) {
    return http.post<Note>("/items", { title, type: "url", source_url: sourceUrl })
  },

  remove(id: string) {
    return http.delete(`/items/${id}`)
  },

  rename(id: string, title: string) {
    return http.patch<Note>(`/items/${id}`, { title })
  },

  getById(id: string) {
    return http.get<Note>(`/items/${id}`)
  },

  update(id: string, payload: NoteUpdatePayload) {
    return http.patch<Note>(`/items/${id}`, payload)
  },

  bindTag(itemId: string, payload: BindTagPayload) {
    return http.post(`/items/${itemId}/tags`, payload)
  },

  unbindTag(itemId: string, tagId: string) {
    return http.delete(`/items/${itemId}/tags/${tagId}`)
  },

  autotag(id: string, number: number) {
    return http.post<Listed<TagRef>>(`/items/${id}/autotag`, { number })
  },

  listDeleted(params?: Record<string, unknown>) {
    return http.get<Paginated<Note>>("/items/deleted", { params })
  },

  restore(id: string) {
    return http.post(`/items/${id}/restore`)
  },

  permanentDelete(id: string) {
    return http.delete(`/items/deleted/${id}`)
  },

  emptyTrash() {
    return http.delete("/items/deleted")
  },

  updateSourceUrl(id: string, sourceUrl: string) {
    return http.patch<Note>(`/items/${id}`, { source_url: sourceUrl })
  },

  refetch(id: string) {
    return http.post(`/items/${id}/refetch`)
  },
}
