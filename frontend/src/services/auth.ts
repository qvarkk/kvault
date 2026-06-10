import http from "./http"
import type { ApiKey, Listed } from "@/types"

export const authService = {
  listKeys() {
    return http.get<Listed<ApiKey>>("/auth/keys")
  },

  renameKey(id: string, label: string) {
    return http.patch(`/auth/keys/${id}`, { label })
  },

  deleteKey(id: string) {
    return http.delete(`/auth/keys/${id}`)
  },

  logout() {
    return http.post("/auth/logout")
  },

  logoutOthers() {
    return http.post("/auth/logout-others")
  },

  updatePassword(oldPassword: string, newPassword: string) {
    return http.patch("/auth/me/password", {
      old_password: oldPassword,
      new_password: newPassword,
    })
  },

  deleteAccount(password: string) {
    return http.post("/auth/me/delete", { password })
  },
}
