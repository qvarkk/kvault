import axios, { AxiosError } from "axios"
import { camelizeKeys } from "./camelize"
import type { ApiError } from "@/types"
import i18n from "@/i18n"

const http = axios.create({
  baseURL: import.meta.env.VITE_API_URL,
  timeout: 10000,
  headers: {
    "Content-Type": "application/json",
  },
})

http.defaults.paramsSerializer = {
  indexes: null,
}

http.interceptors.request.use((config) => {
  const token = localStorage.getItem("token")

  if (token) {
    config.headers.Authorization = token
  }

  config.headers["Accept-Language"] = i18n.global.locale.value

  return config
})

http.interceptors.response.use(
  (response) => {
    response.data = camelizeKeys(response.data)
    return response
  },
  (error: AxiosError<ApiError>) => {
    const data = error.response?.data
    // A well-formed API error has a string `detail`. Anything else (e.g. a dev
    // proxy returning a 500 HTML page when the API is down) is normalized to an
    // internal error so downstream toasts always have a message.
    if (data && typeof (data as ApiError).detail === "string") {
      return Promise.reject(camelizeKeys(data) as ApiError)
    }

    if (error.response) {
      return Promise.reject({
        title: i18n.global.t("errors.internal.title"),
        detail: i18n.global.t("errors.internal.detail"),
        status: error.response.status,
      } as ApiError)
    }

    return Promise.reject({
      title: i18n.global.t("errors.network.title"),
      detail: i18n.global.t("errors.network.detail"),
      status: 0,
    } as ApiError)
  },
)

export default http
