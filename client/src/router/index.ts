import { createRouter, createWebHistory } from "vue-router"
import { useAuthStore } from "@/stores/auth"
import { useAdminStore } from "@/stores/admin"
import { updateDocumentTitle } from "@/i18n"

function hasGenres(genres: string | null | undefined): boolean {
  if (!genres) return false
  if (genres === "null") return false
  // Set locally by the onboarding "skip" action; without it the guard bounces the user back.
  if (genres === "skip") return true
  try {
    const parsed = JSON.parse(genres)
    return Array.isArray(parsed) && parsed.length > 0
  } catch {
    return false
  }
}

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  scrollBehavior: () => ({ top: 0 }),
  routes: [
    // ── Auth ──────────────────────────────────────────────────
    {
      path: "/login",
      name: "login",
      component: () => import("@/views/auth/LoginView.vue"),
      meta: { guestOnly: true, titleKey: "auth.login.title" }
    },
    {
      path: "/register",
      name: "register",
      component: () => import("@/views/auth/RegisterView.vue"),
      meta: { guestOnly: true, titleKey: "auth.register.title" }
    },
    {
      path: "/check-email",
      name: "check-email",
      component: () => import("@/views/auth/CheckEmailView.vue"),
      meta: { titleKey: "auth.checkEmail.title" }
    },
    {
      path: "/auth/verify-email",
      name: "verify-email",
      component: () => import("@/views/auth/VerifyEmailView.vue"),
      meta: { titleKey: "auth.verifyEmail.title" }
    },
    {
      path: "/forgot-password",
      name: "forgot-password",
      component: () => import("@/views/auth/ForgotPasswordView.vue"),
      meta: { guestOnly: true, titleKey: "auth.forgotPassword.title" }
    },
    {
      path: "/reset-password",
      name: "reset-password",
      component: () => import("@/views/auth/ResetPasswordView.vue"),
      meta: { guestOnly: true, titleKey: "auth.resetPassword.title" }
    },

    {
      path: "/auth/handoff",
      name: "auth-handoff",
      component: () => import("@/views/auth/HandoffView.vue"),
      meta: { titleKey: "navigation.link.handoff.title" }
    },

    // ── Onboarding (protected, no MainLayout/Navbar) ───────────
    {
      path: "/onboarding",
      name: "onboarding",
      component: () => import("@/views/onboarding/FavoriteGenreView.vue"),
      meta: { requiresAuth: true, titleKey: "onboarding.title" }
    },

    // ── Main (protected, with MainLayout + Navbar) ─────────────
    {
      path: "/",
      component: () => import("@/layouts/MainLayout.vue"),
      meta: { requiresAuth: true },
      children: [
        {
          path: "",
          name: "home",
          component: () => import("@/views/HomeView.vue"),
          meta: { titleKey: "navigation.home" }
        },
        {
          path: "/notifications",
          name: "notifications",
          component: () => import("@/views/NotificationsView.vue"),
          meta: { titleKey: "notifications.title" }
        },
        {
          path: "search",
          name: "search-results",
          component: () => import("@/views/SearchResultView.vue"),
          meta: { titleKey: "media.search.title" }
        },
        {
          path: "movies",
          name: "movies",
          component: () => import("@/views/movie/MoviesView.vue"),
          meta: { titleKey: "navigation.movies" }
        },
        {
          path: "movies/:id",
          name: "movie-detail",
          component: () => import("@/views/movie/MovieDetailView.vue"),
          meta: { titleKey: "navigation.movies" }
        },
        {
          path: "tv",
          name: "tv",
          component: () => import("@/views/tv/TVSeriesView.vue"),
          meta: { titleKey: "navigation.tvSeries" }
        },
        {
          path: "tv/:id",
          name: "tv-detail",
          component: () => import("@/views/tv/TVDetailView.vue"),
          meta: { titleKey: "navigation.tvSeries" }
        },
        {
          path: "upcoming",
          name: "upcoming",
          component: () => import("@/views/movie/Upcoming.vue"),
          meta: { titleKey: "navigation.upcoming" }
        },
        {
          path: "about-us",
          name: "about-us",
          component: () => import("@/views/aboutus/AboutUsView.vue"),
          meta: { titleKey: "navigation.aboutUs" }
        },
        {
          path: "donate",
          name: "donate",
          component: () => import("@/views/aboutus/DonateView.vue"),
          meta: { titleKey: "navigation.donate" }
        },
        {
          path: "users/:userId",
          name: "user-profile",
          component: () => import("@/views/user/UserProfileView.vue"),
          meta: { titleKey: "navigation.profile" }
        },
        {
          path: "users/:userId/library",
          name: "user-library",
          component: () => import("@/views/library/LibraryView.vue"),
          meta: { titleKey: "navigation.myLibrary" }
        },
        {
          path: "users/:userId/achievements",
          name: "user-achievements",
          component: () => import("@/views/achievement/AchievementsView.vue"),
          meta: { titleKey: "navigation.achievement" }
        },
        {
          path: "feed",
          name: "feed",
          component: () => import("@/views/FeedView.vue"),
          meta: { titleKey: "navigation.feed" }
        },
        {
          path: "actors/:id",
          name: "actor-detail",
          component: () => import("@/views/actor/ActorDetailView.vue"),
          meta: { titleKey: "actor.title" }
        },
      ],
    },

    // ── Admin Workspace ────────────────────────────────────────
    {
      path: "/admin",
      component: () => import("@/layouts/AdminLayout.vue"),
      meta: { requiresAuth: true, requiresAdmin: true },
      children: [
        {
          path: "",
          name: "admin-dashboard",
          component: () => import("@/views/admin/AdminDashboardView.vue"),
          meta: { titleKey: "admin.dashboard.title" }
        },
        {
          path: "users",
          name: "admin-users",
          component: () => import("@/views/admin/AdminUsersView.vue"),
          meta: { titleKey: "admin.users.title" }
        },
        {
          path: "reviews",
          name: "admin-reviews",
          component: () => import("@/views/admin/AdminReviewsView.vue"),
          meta: { titleKey: "admin.reviews.title" }
        },
        {
          path: "audit-logs",
          name: "admin-audit-logs",
          component: () => import("@/views/admin/AdminAuditLogsView.vue"),
          meta: { titleKey: "admin.auditLogs.title" }
        },
      ],
    },

    // ── 404 ───────────────────────────────────────────────────
    {
      path: "/:pathMatch(.*)*",
      name: "not-found",
      component: () => import("@/views/NotFoundView.vue"),
      meta: { titleKey: "common.notFound.title" }
    },
  ],
})

router.beforeEach(async to => {
  const authStore = useAuthStore()
  const adminStore = useAdminStore()

  // Always initialize session state before making route decisions
  await authStore.initialize()

  // If backend error occurred during initialization, don't process redirects
  // (App.vue renders SessionBootstrapView with retry UI)
  if (authStore.initError === "network" || authStore.initError === "server") {
    return
  }

  const needsAuth = to.matched.some(r => r.meta.requiresAuth)

  // ยังไม่ login → ไป login
  if (needsAuth && !authStore.isLoggedIn)
    return { name: "login", query: { redirect: to.fullPath } }

  // login แล้วเข้า guest-only page → ไป home
  if (to.meta.guestOnly && authStore.isLoggedIn)
    return { name: "home" }

  // Admin route protection guard
  if (to.meta.requiresAdmin) {
    if (!authStore.isAdmin) {
      adminStore.reset()
      return { name: "home" }
    }
  }

  // ✅ ต้อง requiresAuth ด้วย ไม่งั้น guest route ก็โดน redirect
  if (
    needsAuth &&
    authStore.isLoggedIn &&
    !hasGenres(authStore.user?.favorite_genres) &&
    to.name !== "onboarding"
  ) {
    return { name: "onboarding" }
  }

  // ✅ ใช้ hasGenres ให้สอดคล้องกัน
  if (
    authStore.isLoggedIn &&
    hasGenres(authStore.user?.favorite_genres) &&
    to.name === "onboarding"
  ) {
    return { name: "home" }
  }
})

router.afterEach(() => {
  updateDocumentTitle()
})

export default router
