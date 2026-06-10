import { useAuthStore } from "@/stores/auth"
import { createRouter, createWebHistory } from "vue-router"

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: "/login",
      name: "login",
      component: () => import("@/views/LoginView.vue"),
      meta: { requiresGuest: true },
    },
    {
      path: "/signup",
      name: "signUp",
      component: () => import("@/views/SignUpView.vue"),
      meta: { requiresGuest: true },
    },
    {
      path: "/",
      name: "home",
      component: () => import("@/views/HomeView.vue"),
      meta: { requiresAuth: true },
    },
    {
      path: "/notes",
      name: "notes",
      component: () => import("@/views/NotesView.vue"),
      meta: { requiresAuth: true },
    },
    {
      path: "/notes/:id",
      name: "note",
      component: () => import("@/views/NoteView.vue"),
      meta: { requiresAuth: true },
      props: true,
    },
    {
      path: "/files",
      name: "files",
      component: () => import("@/views/FilesView.vue"),
      meta: { requiresAuth: true },
    },
    {
      path: "/files/:id",
      name: "file",
      component: () => import("@/views/FileView.vue"),
      meta: { requiresAuth: true },
      props: true,
    },
    {
      path: "/search",
      name: "search",
      component: () => import("@/views/SearchView.vue"),
      meta: { requiresAuth: true },
    },
    {
      path: "/about",
      name: "about",
      component: () => import("@/views/AboutView.vue"),
      meta: { requiresAuth: true },
    },
  ],
})

router.beforeEach(async (to) => {
  const auth = useAuthStore()
  await auth.init()

  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    return { name: "login" }
  }

  if (to.meta.requiresGuest && auth.isAuthenticated) {
    return { name: "home" }
  }
})

export default router
