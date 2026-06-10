import "@fontsource/geist/400.css"
import "@fontsource/geist/500.css"
import "@fontsource/geist/600.css"
import "@fontsource/geist/700.css"
import "@fontsource/jetbrains-mono/400.css"
import "@fontsource/jetbrains-mono/500.css"
import "@fontsource/jetbrains-mono/600.css"
import "@fontsource/jetbrains-mono/700.css"

import { createApp } from "vue"
import { createPinia } from "pinia"
import { createHead } from "@vueuse/head"
import { config } from "md-editor-v3"
import ru from "@vavt/cm-extension/dist/locale/ru"
import JP_JP from "@vavt/cm-extension/dist/locale/jp-JP"

import App from "./App.vue"
import router from "./router"
import i18n from "./i18n"

config({
  editorConfig: {
    languageUserDefined: {
      ru: ru,
      ja: JP_JP,
    },
  },
})

const appName = import.meta.env.VITE_APP_NAME ?? "kvault"
const head = createHead({
  titleTemplate: (title) => (title ? `${title} - ${appName}` : appName),
})
const app = createApp(App)

app.use(createPinia())
app.use(router)
app.use(i18n)
app.use(head)

app.mount("#app")
