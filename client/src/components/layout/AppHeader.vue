<template>
  <header class="hdr" :class="{ 'hdr--scrolled': scrolled }">
    <RouterLink to="/" class="hdr-logo" @click="mobileMenuOpen = false">
      <img src="/popcorn-icon-192.png" alt="" class="hdr-logo-icon" /><span>POP</span><span class="hdr-logo-brand">CORN</span> <!-- hardcode-ok: brand name -->
    </RouterLink>

    <HeaderNav class="hdr-nav" />

    <div class="hdr-right">
      <HeaderSearch />

      <LanguageToggle class="hdr-lang" />

      <template v-if="authStore.user">
        <div class="hdr-bell">
          <button
            type="button"
            class="hdr-icon"
            :class="{ 'hdr-icon--on': notificationStore.isPanelOpen }"
            :aria-label="badge ? t('navigation.header.notificationsUnread', { count: badge }) : t('navigation.header.notifications')"
            :title="t('navigation.header.notifications')"
            aria-haspopup="dialog"
            :aria-expanded="notificationStore.isPanelOpen"
            @click.stop="toggleNotificationPanel"
          >
            <Bell :size="20" aria-hidden="true" />
            <span v-if="badge" class="hdr-badge" aria-hidden="true">{{ badge }}</span>
          </button>

          <Transition name="dropdown">
            <NotificationDropdown v-if="notificationStore.isPanelOpen" @close="notificationStore.closePanel()" />
          </Transition>
        </div>

        <AccountMenu
          v-model:open="userMenuOpen"
          class="hdr-account"
          :user="authStore.user"
          :is-admin="authStore.isAdmin"
          @switch-planzi="handleSwitchToPlanzi"
          @logout="handleLogout"
          @enter-admin="handleEnterAdmin"
        />
      </template>
      <div v-else class="hdr-guest">
        <AppButton :to="{ path: '/login', query: buildAuthQuery(route.fullPath) }" variant="ghost" size="sm">{{ t("navigation.logIn") }}</AppButton>
        <AppButton :to="{ path: '/register', query: buildAuthQuery(route.fullPath) }" size="sm">{{ t("navigation.signUp") }}</AppButton>
      </div>

      <button
        type="button"
        class="hdr-icon hdr-burger" data-touch-ok
        :aria-label="t('navigation.header.openMenu')"
        :title="t('navigation.header.openMenu')"
        aria-haspopup="dialog"
        :aria-expanded="mobileMenuOpen"
        @click="toggleMobileMenu"
      >
        <Menu :size="22" aria-hidden="true" />
      </button>
    </div>
  </header>

  <MobileNavSheet
    v-model="mobileMenuOpen"
    :user="authStore.user"
    :is-admin="authStore.isAdmin"
    @switch-planzi="handleSwitchToPlanzi"
    @logout="handleLogout"
    @enter-admin="handleEnterAdmin"
  />
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from "vue"
import { useRoute, useRouter } from "vue-router"
import { useI18n } from "vue-i18n"
import { buildAuthQuery } from "@/utils/safeRedirect"
import { Bell, Menu } from "lucide-vue-next"
import { useAuthStore } from "@/stores/auth"
import { useAdminStore } from "@/stores/admin"
import { useNotificationStore } from "@/stores/notification"
import { useSessionReset } from "@/composables/useSessionReset"
import { i18n } from "@/i18n"
import NotificationDropdown from "@/components/notifications/NotificationDropdown.vue"
import HeaderSearch from "@/components/search/HeaderSearch.vue"
import AppButton from "@/components/ui/AppButton.vue"
import { linkApi } from "@/api/endpoints/link"
import { playPlatformExit } from "@/utils/platformTransition"
import HeaderNav from "./HeaderNav.vue"
import LanguageToggle from "./LanguageToggle.vue"
import AccountMenu from "./AccountMenu.vue"
import MobileNavSheet from "./MobileNavSheet.vue"
import { formatBadge } from "./navItems"
import { isAtLeast } from "@/constants/breakpoints"

const { t } = useI18n()
const resetSession = useSessionReset()
const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const adminStore = useAdminStore()
const notificationStore = useNotificationStore()

const scrolled = ref(false)
const userMenuOpen = ref(false)
const mobileMenuOpen = ref(false)
const badge = computed(() => formatBadge(notificationStore.unreadCount))

async function handleSwitchToPlanzi() {
  userMenuOpen.value = false
  mobileMenuOpen.value = false
  try {
    const { data } = await linkApi.createHandoff("planzi")
    await playPlatformExit()
    window.location.href = data.redirect_url
  } catch {
    window.alert(i18n.global.t("navigation.link.switchFailed"))
  }
}

function toggleNotificationPanel() {
  userMenuOpen.value = false
  mobileMenuOpen.value = false
  notificationStore.togglePanel()
}

function onScroll() {
  scrolled.value = window.scrollY > 20
}

// The mobile sheet and the notification panel must not be open together.
function toggleMobileMenu() {
  mobileMenuOpen.value = !mobileMenuOpen.value
  if (mobileMenuOpen.value) {
    userMenuOpen.value = false
    notificationStore.closePanel()
  }
}

async function handleLogout() {
  userMenuOpen.value = false
  mobileMenuOpen.value = false
  resetSession()
  await authStore.logout()
  router.push({ name: "login" })
}

function handleEnterAdmin() {
  userMenuOpen.value = false
  mobileMenuOpen.value = false
  adminStore.enterAdminMode()
  router.push("/admin")
}

// Close the nav sheet once the inline nav takes over (lg and up).
function onResize() {
  if (isAtLeast("lg") && mobileMenuOpen.value) mobileMenuOpen.value = false
}

onMounted(() => {
  onScroll()
  window.addEventListener("scroll", onScroll, { passive: true })
  window.addEventListener("resize", onResize)
})
onUnmounted(() => {
  window.removeEventListener("scroll", onScroll)
  window.removeEventListener("resize", onResize)
})
</script>

<style scoped>
.hdr {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: var(--z-header);
  display: flex;
  align-items: center;
  gap: 2rem;
  height: 3.5rem;
  padding: 0 max(1rem, env(safe-area-inset-right)) 0 max(1rem, env(safe-area-inset-left));
  border-bottom: 1px solid transparent;
  transition: background var(--motion-base), backdrop-filter var(--motion-base), border-color var(--motion-base);
}
.hdr--scrolled {
  background: color-mix(in srgb, var(--color-surface) 86%, transparent);
  backdrop-filter: blur(16px) saturate(1.4);
  -webkit-backdrop-filter: blur(16px) saturate(1.4);
  border-color: var(--color-border);
}
.hdr-logo {
  display: flex;
  align-items: center;
  flex-shrink: 0;
  font-size: 1.25rem;
  font-weight: 900;
  letter-spacing: -0.5px;
  color: var(--color-text-primary);
  text-decoration: none;
}
.hdr-logo-icon {
  width: 1.75rem;
  height: 1.75rem;
  margin-right: 0.375rem;
  object-fit: contain;
}
.hdr-logo-brand {
  color: var(--color-brand);
}
.hdr-nav {
  display: none;
}
.hdr-right {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  margin-left: auto;
}
.hdr-lang,
.hdr-account,
.hdr-guest {
  display: none;
}
.hdr-guest {
  gap: 0.5rem;
  margin-left: 0.25rem;
}

.hdr-icon {
  position: relative;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 2.75rem;
  height: 2.75rem;
  flex-shrink: 0;
  border: none;
  border-radius: 9999px;
  background: none;
  color: var(--color-text-secondary);
  cursor: pointer;
  transition: color var(--motion-fast), background var(--motion-fast);
}
.hdr-icon--on {
  color: var(--color-text-primary);
  background: color-mix(in srgb, var(--color-text-primary) 10%, transparent);
}
@media (hover: hover) {
  .hdr-icon:hover {
    color: var(--color-text-primary);
    background: color-mix(in srgb, var(--color-text-primary) 8%, transparent);
  }
}
.hdr-bell {
  position: relative;
  display: flex;
  align-items: center;
}
.hdr-badge {
  position: absolute;
  top: 0;
  right: -0.2rem;
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 1.125rem;
  height: 1.125rem;
  padding: 0 0.3rem;
  border-radius: 9999px;
  border: 2px solid var(--color-surface);
  box-sizing: content-box;
  background: var(--color-brand);
  color: var(--color-on-brand);
  font-size: 0.625rem;
  font-weight: 700;
  line-height: 1;
  animation: hdr-badge-pop 320ms var(--ease-spring);
}
@keyframes hdr-badge-pop {
  from {
    transform: scale(0.5);
  }
  to {
    transform: scale(1);
  }
}

.dropdown-enter-active,
.dropdown-leave-active {
  transition: opacity var(--motion-fast), transform var(--motion-fast);
}
.dropdown-enter-from,
.dropdown-leave-to {
  opacity: 0;
  transform: translateY(-6px);
}

/* md-lg: tablet bar keeps search, bell and account; the burger opens the nav sheet. */
@media (min-width: 768px) {
  .hdr {
    height: 4rem;
    padding: 0 1.5rem;
    gap: 1.25rem;
  }
  .hdr-account {
    display: block;
    margin-left: 0.25rem;
  }
  .hdr-guest {
    display: flex;
  }
}
/* lg and up: inline nav + language toggle replace the burger. */
@media (min-width: 1024px) {
  .hdr {
    gap: 2rem;
    padding: 0 max(2rem, env(safe-area-inset-right)) 0 max(2rem, env(safe-area-inset-left));
  }
  .hdr-nav {
    display: flex;
  }
  .hdr-lang {
    display: inline-flex;
  }
  .hdr-burger {
    display: none;
  }
}
@media (min-width: 1280px) {
  .hdr {
    padding: 0 max(3rem, env(safe-area-inset-right)) 0 max(3rem, env(safe-area-inset-left));
  }
}
@media (prefers-reduced-motion: reduce) {
  .hdr,
  .dropdown-enter-active,
  .dropdown-leave-active {
    transition: none;
  }
  .hdr-badge {
    animation: none;
  }
}
</style>
