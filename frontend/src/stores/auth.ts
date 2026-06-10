import http from "@/services/http"
import { authService } from "@/services/auth"
import type { User, LoginPayload, SignupPayload } from "@/types"
import { defineStore } from "pinia"
import { computed, ref } from "vue"

export const useAuthStore = defineStore("auth", () => {
  const user = ref<User | null>(null)
  const isAuthenticated = computed(() => user.value !== null)

  async function login(payload: LoginPayload) {
    const response = await http.post<User>("/auth/login", payload)
    // login rotates the key and returns the fresh one — persist it as the token.
    if (response.data.apiKey) localStorage.setItem("token", response.data.apiKey)
    user.value = response.data
  }

  async function signup(payload: SignupPayload) {
    const response = await http.post<User>("/auth/register", payload)
    if (response.data.apiKey) localStorage.setItem("token", response.data.apiKey)
    user.value = response.data
  }

  async function logout() {
    // Revoke the current key server-side; clear locally regardless of outcome.
    try {
      await authService.logout()
    } catch {
      // Token may already be invalid — still clear the client session.
    }
    localStorage.removeItem("token")
    user.value = null
  }

  async function init() {
    const token = localStorage.getItem("token")
    if (!token) return

    try {
      const response = await http.get<User>("/auth/me")
      user.value = response.data
    } catch {
      localStorage.removeItem("token")
      user.value = null
    }
  }

  return { user, isAuthenticated, login, signup, logout, init }
})
