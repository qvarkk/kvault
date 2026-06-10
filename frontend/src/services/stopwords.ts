import http from "./http"
import type { Listed, Stopword, StopwordCreatePayload } from "@/types"

export const stopwordsService = {
  list(params?: Record<string, unknown>) {
    return http.get<Listed<Stopword>>("/stopwords", { params })
  },

  create(payload: StopwordCreatePayload) {
    return http.post<Stopword>("/stopwords", payload)
  },

  remove(word: string) {
    return http.delete(`/stopwords/${word}`)
  },

  enable(word: string) {
    return http.post(`/stopwords/${word}/enable`)
  },

  disable(word: string) {
    return http.post(`/stopwords/${word}/disable`)
  },
}
