<template>
  <div class="admin-shell">
    <!-- Topbar -->
    <header class="admin-topbar">
      <div class="admin-brand">
        <RouterLink to="/" class="nav-logo">
          <span class="logo-movie">POP</span><span class="logo-hub">CORN</span>
        </RouterLink>
        <span class="admin-badge">
          <Shield :size="14" /> {{ $t("admin.nav.adminMode") }}
        </span>
      </div>

      <div class="admin-topbar-right">
        <button class="exit-admin-btn" @click="handleExitAdmin">
          <LogOut :size="15" />
          <span>{{ $t("admin.nav.exitAdmin") }}</span>
        </button>
      </div>
    </header>

    <!-- Main Workspace Container -->
    <div class="admin-body">
      <!-- Sidebar Navigation -->
      <aside class="admin-sidebar">
        <nav class="admin-nav">
          <RouterLink
            to="/admin"
            exact
            class="admin-nav-item"
            :class="{ active: route.name === 'admin-dashboard' }"
          >
            <LayoutDashboard :size="18" />
            <span>{{ $t("admin.nav.dashboard") }}</span>
          </RouterLink>

          <RouterLink
            to="/admin/users"
            class="admin-nav-item"
            :class="{ active: route.name === 'admin-users' }"
          >
            <Users :size="18" />
            <span>{{ $t("admin.nav.users") }}</span>
          </RouterLink>

          <RouterLink
            to="/admin/reviews"
            class="admin-nav-item"
            :class="{ active: route.name === 'admin-reviews' }"
          >
            <MessageSquare :size="18" />
            <span>{{ $t("admin.nav.reviews") }}</span>
          </RouterLink>

          <RouterLink
            to="/admin/audit-logs"
            class="admin-nav-item"
            :class="{ active: route.name === 'admin-audit-logs' }"
          >
            <FileText :size="18" />
            <span>{{ $t("admin.nav.auditLogs") }}</span>
          </RouterLink>
        </nav>
      </aside>

      <!-- Content Area -->
      <main class="admin-content">
        <RouterView />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRoute, useRouter } from "vue-router"
import { useAdminStore } from "@/stores/admin"
import {
  Shield,
  LayoutDashboard,
  Users,
  MessageSquare,
  FileText,
  LogOut,
} from "lucide-vue-next"

const route = useRoute()
const router = useRouter()
const adminStore = useAdminStore()

function handleExitAdmin() {
  adminStore.exitAdminMode()
  router.push({ name: "home" })
}
</script>

<style scoped>
.admin-shell {
  min-height: 100vh;
  background-color: var(--color-surface);
  color: var(--color-text-primary);
  display: flex;
  flex-direction: column;
}

/* ── Topbar ─────────────────────────────────────────── */
.admin-topbar {
  height: 64px;
  background-color: var(--color-surface-2);
  border-bottom: 1px solid var(--color-border);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 1.5rem;
  position: sticky;
  top: 0;
  z-index: 100;
}

.admin-brand {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.nav-logo {
  font-family: var(--font-sans);
  font-size: 1.4rem;
  font-weight: 900;
  text-decoration: none;
  letter-spacing: -0.5px;
}
.logo-movie { color: var(--color-text-primary); }
.logo-hub { color: var(--color-brand); }

.admin-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  background: rgba(245, 67, 44, 0.15);
  border: 1px solid rgba(245, 67, 44, 0.4);
  color: var(--color-brand);
  font-size: 0.75rem;
  font-weight: 700;
  padding: 0.25rem 0.6rem;
  border-radius: var(--radius-btn);
  letter-spacing: 0.5px;
  text-transform: uppercase;
}

.exit-admin-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  background: var(--color-surface-3);
  border: 1px solid var(--color-border);
  color: var(--color-text-secondary);
  font-size: 0.825rem;
  font-weight: 600;
  padding: 0.45rem 0.9rem;
  border-radius: var(--radius-btn);
  cursor: pointer;
  transition: all 0.2s;
}
.exit-admin-btn:hover {
  color: var(--color-text-primary);
  border-color: var(--color-brand);
  background: rgba(245, 67, 44, 0.1);
}

/* ── Body Layout ────────────────────────────────────── */
.admin-body {
  display: flex;
  flex: 1;
}

.admin-sidebar {
  width: 240px;
  background-color: var(--color-surface-2);
  border-right: 1px solid var(--color-border);
  padding: 1.5rem 0.75rem;
  flex-shrink: 0;
}

.admin-nav {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.admin-nav-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem 1rem;
  color: var(--color-text-secondary);
  font-size: 0.875rem;
  font-weight: 600;
  text-decoration: none;
  border-radius: 0.5rem;
  transition: all 0.2s;
}
.admin-nav-item:hover {
  color: var(--color-text-primary);
  background-color: var(--color-surface-3);
}
.admin-nav-item.active {
  color: var(--color-text-primary);
  background-color: var(--color-brand);
}

.admin-content {
  flex: 1;
  padding: 2rem;
  max-width: 1400px;
  width: 100%;
  box-sizing: border-box;
}

@media (max-width: 768px) {
  .admin-body {
    flex-direction: column;
  }
  .admin-sidebar {
    width: 100%;
    border-right: none;
    border-bottom: 1px solid var(--color-border);
    padding: 0.5rem;
  }
  .admin-nav {
    flex-direction: row;
    justify-content: space-around;
  }
  .admin-nav-item {
    font-size: 0.75rem;
    padding: 0.5rem;
    flex-direction: column;
    gap: 0.2rem;
  }
  .admin-content {
    padding: 1rem;
  }
}
</style>
