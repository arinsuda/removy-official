<template>
  <div class="admin-shell" :class="{ 'is-collapsed': collapsed }">
    <header class="admin-topbar">
      <div class="admin-topbar__brand">
        <AppButton
          class="admin-topbar__menu"
          variant="ghost"
          size="sm"
          icon-only
          :aria-label="t('admin.nav.openMenu')"
          :aria-expanded="drawerOpen"
          aria-controls="admin-sidebar"
          @click="openDrawer"
        >
          <Menu :size="20" aria-hidden="true" />
        </AppButton>
        <RouterLink to="/" class="admin-logo" :aria-label="t('admin.nav.backToApp')">
          <img src="/popcorn-icon-192.png" alt="" class="admin-logo__icon" />
          <!-- hardcode-ok: brand name -->
          <span class="admin-logo__name">POP<span>CORN</span></span>
        </RouterLink>
        <AdminBadge tone="brand"><Shield :size="12" aria-hidden="true" />{{ t("admin.nav.adminMode") }}</AdminBadge>
      </div>

      <AppButton variant="outline" size="sm" @click="handleExitAdmin">
        <LogOut :size="16" aria-hidden="true" />
        <span class="admin-topbar__exit">{{ t("admin.nav.exitAdmin") }}</span>
      </AppButton>
    </header>

    <div v-if="drawerOpen" class="admin-scrim" @click="closeDrawer" />

    <div class="admin-body">
      <aside
        id="admin-sidebar"
        ref="sidebarRef"
        class="admin-sidebar"
        :class="{ 'is-open': drawerOpen }"
        :aria-label="t('admin.nav.primary')"
        @keydown.esc="closeDrawer"
      >
        <div class="admin-sidebar__head">
          <span class="admin-sidebar__title">{{ t("admin.nav.adminMode") }}</span>
          <AppButton
            class="admin-sidebar__close"
            variant="ghost"
            size="sm"
            icon-only
            :aria-label="t('admin.nav.closeMenu')"
            @click="closeDrawer"
          >
            <X :size="18" aria-hidden="true" />
          </AppButton>
        </div>

        <nav class="admin-nav">
          <RouterLink
            v-for="item in NAV"
            :key="item.name"
            :to="item.to"
            class="admin-nav__item"
            :class="{ 'is-active': route.name === item.name }"
            :aria-current="route.name === item.name ? 'page' : undefined"
            :title="collapsed ? t(item.label) : undefined"
          >
            <component :is="item.icon" :size="18" aria-hidden="true" />
            <span class="admin-nav__label">{{ t(item.label) }}</span>
          </RouterLink>
        </nav>

        <button
          type="button"
          class="admin-sidebar__toggle"
          :aria-label="collapsed ? t('admin.nav.expand') : t('admin.nav.collapse')"
          :aria-pressed="collapsed"
          @click="toggleCollapsed"
        >
          <PanelLeftOpen v-if="collapsed" :size="18" aria-hidden="true" />
          <PanelLeftClose v-else :size="18" aria-hidden="true" />
          <span class="admin-nav__label">{{ collapsed ? t("admin.nav.expand") : t("admin.nav.collapse") }}</span>
        </button>
      </aside>

      <main class="admin-content" :inert="drawerOpen || undefined">
        <RouterView />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { nextTick, onBeforeUnmount, ref, watch } from "vue"
import { useRoute, useRouter } from "vue-router"
import { useI18n } from "vue-i18n"
import {
  FileText,
  LayoutDashboard,
  LogOut,
  Menu,
  MessageSquare,
  PanelLeftClose,
  PanelLeftOpen,
  Shield,
  Users,
  X,
} from "lucide-vue-next"
import AdminBadge from "@/components/admin/AdminBadge.vue"
import AppButton from "@/components/ui/AppButton.vue"
import { lockScroll, unlockScroll } from "@/composables/scrollLock"
import { useAdminStore } from "@/stores/admin"

const NAV = [
  { name: "admin-dashboard", to: "/admin", label: "admin.nav.dashboard", icon: LayoutDashboard },
  { name: "admin-users", to: "/admin/users", label: "admin.nav.users", icon: Users },
  { name: "admin-reviews", to: "/admin/reviews", label: "admin.nav.reviews", icon: MessageSquare },
  { name: "admin-audit-logs", to: "/admin/audit-logs", label: "admin.nav.auditLogs", icon: FileText },
]
const COLLAPSE_KEY = "admin_sidebar_collapsed"

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const adminStore = useAdminStore()

const collapsed = ref(localStorage.getItem(COLLAPSE_KEY) === "true")
const drawerOpen = ref(false)
const sidebarRef = ref<HTMLElement | null>(null)

function toggleCollapsed() {
  collapsed.value = !collapsed.value
  localStorage.setItem(COLLAPSE_KEY, String(collapsed.value))
}

async function openDrawer() {
  drawerOpen.value = true
  lockScroll()
  await nextTick()
  sidebarRef.value?.querySelector<HTMLElement>("a, button")?.focus()
}

function closeDrawer() {
  if (!drawerOpen.value) return
  drawerOpen.value = false
  unlockScroll()
  document.querySelector<HTMLElement>(".admin-topbar__menu")?.focus()
}

// Following a link closes the drawer without stealing focus back to the menu button.
watch(
  () => route.fullPath,
  () => {
    if (!drawerOpen.value) return
    drawerOpen.value = false
    unlockScroll()
  },
)

onBeforeUnmount(() => {
  if (drawerOpen.value) unlockScroll()
})

function handleExitAdmin() {
  adminStore.exitAdminMode()
  router.push({ name: "home" })
}
</script>

<style scoped>
.admin-shell {
  --sidebar-width: 15rem;
  display: flex;
  flex-direction: column;
  min-height: 100dvh;
  background: var(--color-surface);
  color: var(--color-text-primary);
}

.admin-shell.is-collapsed {
  --sidebar-width: 4.25rem;
}

.admin-topbar {
  position: sticky;
  top: 0;
  z-index: var(--z-header);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  height: 3.5rem;
  padding: 0 1rem;
  background: var(--color-surface-2);
  border-bottom: 1px solid var(--color-border);
}

.admin-topbar__brand {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  min-width: 0;
}

.admin-topbar__exit {
  display: none;
}

.admin-logo {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  color: var(--color-text-primary);
  text-decoration: none;
}

.admin-logo__icon {
  width: 1.75rem;
  height: 1.75rem;
}

.admin-logo__name {
  font-size: 1.125rem;
  font-weight: 900;
  letter-spacing: -0.02em;
}

.admin-logo__name span {
  color: var(--color-brand);
}

.admin-scrim {
  position: fixed;
  inset: 0;
  z-index: var(--z-above-header);
  background: color-mix(in srgb, var(--color-surface) 70%, transparent);
}

.admin-body {
  display: flex;
  flex: 1;
  min-width: 0;
}

/* Phones: the sidebar is an off-canvas drawer. */
.admin-sidebar {
  position: fixed;
  top: 0;
  bottom: 0;
  left: 0;
  z-index: calc(var(--z-above-header) + 1);
  display: flex;
  flex-direction: column;
  gap: 1rem;
  width: min(18rem, 85vw);
  padding: 0.75rem;
  background: var(--color-surface-2);
  border-right: 1px solid var(--color-border);
  box-shadow: var(--shadow-sheet);
  transform: translateX(-100%);
  visibility: hidden;
  transition:
    transform var(--motion-base) var(--ease-out-soft),
    visibility 0s linear var(--motion-base);
}

.admin-sidebar.is-open {
  transform: none;
  visibility: visible;
  transition-delay: 0s;
}

.admin-sidebar__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.25rem 0.5rem;
}

.admin-sidebar__title {
  font-size: var(--text-caption);
  font-weight: 700;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--color-text-muted);
}

.admin-nav {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.admin-nav__item,
.admin-sidebar__toggle {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  min-height: 2.75rem;
  padding: 0 0.875rem;
  background: none;
  border: 0;
  border-radius: 0.625rem;
  font: inherit;
  font-size: var(--text-small);
  font-weight: 600;
  text-decoration: none;
  color: var(--color-text-secondary);
  cursor: pointer;
  transition:
    background var(--motion-fast) var(--ease-out-soft),
    color var(--motion-fast) var(--ease-out-soft);
}

.admin-nav__item.is-active {
  background: color-mix(in srgb, var(--color-brand) 16%, transparent);
  color: var(--color-text-primary);
  box-shadow: inset 3px 0 0 var(--color-brand);
}

.admin-sidebar__toggle {
  display: none;
  margin-top: auto;
}

@media (hover: hover) {
  .admin-nav__item:hover,
  .admin-sidebar__toggle:hover {
    background: var(--color-surface-3);
    color: var(--color-text-primary);
  }
}

.admin-content {
  flex: 1;
  min-width: 0;
  padding: 1rem;
  box-sizing: border-box;
}

@media (min-width: 640px) {
  .admin-topbar {
    padding: 0 1.5rem;
  }

  .admin-topbar__exit {
    display: inline;
  }
}

@media (min-width: 768px) {
  .admin-topbar__menu,
  .admin-sidebar__close,
  .admin-scrim {
    display: none;
  }

  .admin-sidebar {
    position: sticky;
    top: 3.5rem;
    z-index: auto;
    align-self: flex-start;
    flex: 0 0 var(--sidebar-width);
    width: var(--sidebar-width);
    height: calc(100dvh - 3.5rem);
    box-shadow: none;
    transform: none;
    visibility: visible;
    transition: width var(--motion-base) var(--ease-out-soft);
  }

  .admin-sidebar__head {
    display: none;
  }

  .admin-sidebar__toggle {
    display: flex;
  }

  .is-collapsed .admin-nav__label {
    position: absolute;
    width: 1px;
    height: 1px;
    overflow: hidden;
    clip-path: inset(50%);
    white-space: nowrap;
  }

  .is-collapsed .admin-nav__item,
  .is-collapsed .admin-sidebar__toggle {
    justify-content: center;
    padding: 0;
  }

  .admin-content {
    padding: 1.5rem 2rem;
  }
}

@media (prefers-reduced-motion: reduce) {
  .admin-sidebar {
    transition: none;
  }
}
</style>
