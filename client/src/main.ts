import { createApp } from "vue"
import { createPinia } from "pinia"
import { VueQueryPlugin } from "@tanstack/vue-query"

import App from "./App.vue"
import router from "./router"
import { initSentry } from "./monitoring/sentry"
import { installStaleChunkReload } from "./monitoring/staleChunkReload"

import { VueDatePicker } from "@vuepic/vue-datepicker"
import "@vuepic/vue-datepicker/dist/main.css"
import "./assets/styles/main.css"
import "./assets/styles/motion.css"
import "primeicons/primeicons.css"
import { i18n, getSavedLocale, setLocale } from "./i18n"

async function bootstrap() {
  const app = createApp(App)

  // set saved or default locale
  const savedLocale = getSavedLocale()
  await setLocale(savedLocale)

  initSentry(app, router)
  installStaleChunkReload(router)

  const pinia = createPinia()
  app.use(pinia)
  app.use(i18n)
  app.component("VueDatePicker", VueDatePicker)

  // No blocking fetchMe() — auth initialization is deferred to
  // the router guard and only awaited for protected routes.

  app.use(router)

  app.use(VueQueryPlugin, {
    queryClientConfig: {
      defaultOptions: {
        queries: {
          staleTime: 1000 * 60 * 5,
          retry: 1,
        },
      },
    },
  })

  app.mount("#app")
}

bootstrap()
