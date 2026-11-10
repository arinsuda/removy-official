<template>
  <AdminPage>
    <AdminPageHeader :title="t('admin.users.title')" />

    <AdminFilters>
      <AdminSearch v-model="search" :label="t('admin.table.search')" :placeholder="t('admin.users.searchPlaceholder')" />
      <AdminSelect v-model="roleFilter" :label="t('admin.users.colRole')" :options="roleOptions" />
      <AdminSelect v-model="statusFilter" :label="t('admin.users.colStatus')" :options="statusOptions" />
    </AdminFilters>

    <AdminTable
      :columns="columns"
      :label="t('admin.users.title')"
      :loading="adminStore.isLoadingUsers"
      :error="loadError"
      :empty="users.length === 0"
      :filtered="isFiltered"
      :empty-title="t('admin.users.emptyTitle')"
      :empty-desc="t('admin.users.emptyDesc')"
      :sort-by="sortBy"
      :sort-order="sortOrder"
      @sort="onSort"
      @retry="fetchData(adminStore.usersPage)"
      @clear="clearFilters"
    >
      <tr v-for="u in users" :key="u.id">
        <td class="cell-id">#{{ u.id }}</td>
        <td>
          <div class="user-cell">
            <AppAvatar :src="u.avatar_url" :name="u.display_name || u.username" size="sm" decorative />
            <div class="user-meta">
              <span class="user-name">
                {{ u.display_name || u.username }}
                <AdminBadge v-if="u.id === authStore.user?.id" tone="info">{{ t("admin.users.you") }}</AdminBadge>
              </span>
              <span class="user-handle">@{{ u.username }}</span>
            </div>
          </div>
        </td>
        <td class="cell-muted">{{ u.email }}</td>
        <td>
          <AdminBadge :tone="u.role === 'admin' ? 'brand' : 'neutral'">
            {{ u.role === "admin" ? t("admin.users.roleAdmin") : t("admin.users.roleUser") }}
          </AdminBadge>
        </td>
        <td>
          <div v-if="u.deleted_at" class="status-cell">
            <AdminBadge tone="danger">{{ t("admin.users.badgeDeleted") }}</AdminBadge>
            <span class="cell-muted">{{ t("admin.users.deletedOn", { date: formatDate(u.deleted_at) }) }}</span>
          </div>
          <div v-else-if="u.erasure_due_at" class="status-cell">
            <AdminBadge tone="danger">{{ t("admin.users.badgeErasurePending") }}</AdminBadge>
            <span class="cell-muted">{{ t("admin.users.erasureDueOn", { date: formatDate(u.erasure_due_at) }) }}</span>
          </div>
          <AdminBadge v-else :tone="u.is_active ? 'success' : 'danger'">
            {{ u.is_active ? t("admin.users.statusActive") : t("admin.users.statusInactive") }}
          </AdminBadge>
        </td>
        <td class="cell-num">{{ u.review_count }}</td>
        <td class="cell-muted">{{ formatDate(u.created_at) }}</td>
        <td class="is-end">
          <AdminRowMenu
            :label="t('admin.table.rowActions', { name: `@${u.username}` })"
            :items="menuItems(u)"
            @select="(key) => (key === 'erase' ? startErase(u) : startAction(u, key))"
          />
        </td>
      </tr>
    </AdminTable>

    <AdminPagination
      :page="adminStore.usersPage"
      :total-pages="adminStore.usersTotalPages"
      :total="adminStore.usersTotal"
      :limit="adminStore.usersLimit"
      @change="fetchData"
    />

    <AdminConfirmModal
      v-model="confirmModal.show"
      :title="modal.title"
      :description="modal.description"
      :target-name="confirmModal.user ? `@${confirmModal.user.username}` : ''"
      :variant="modal.variant"
      :icon="modal.icon"
      :confirm-text="modal.confirm"
      :reason-placeholder="t('admin.userModal.reasonNote')"
      :reason-hint="t('admin.userModal.reasonHint')"
      :loading="confirmModal.loading"
      @confirm="onConfirmAction"
      @cancel="confirmModal.show = false"
    />

    <AdminEraseUserModal
      v-model="eraseModal.show"
      :user="eraseModal.user"
      @erased="onErased"
      @stale="fetchData(adminStore.usersPage)"
    />
  </AdminPage>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch, type Component } from "vue"
import { useI18n } from "vue-i18n"
import { Flame, Shield, ShieldOff, Trash2, UserCheck, UserX } from "lucide-vue-next"
import AdminBadge from "@/components/admin/AdminBadge.vue"
import AdminConfirmModal from "@/components/admin/AdminConfirmModal.vue"
import AdminEraseUserModal from "@/components/admin/AdminEraseUserModal.vue"
import AdminFilters from "@/components/admin/AdminFilters.vue"
import AdminPage from "@/components/admin/AdminPage.vue"
import AdminPageHeader from "@/components/admin/AdminPageHeader.vue"
import AdminPagination from "@/components/admin/AdminPagination.vue"
import AdminRowMenu, { type RowMenuItem } from "@/components/admin/AdminRowMenu.vue"
import AdminSearch from "@/components/admin/AdminSearch.vue"
import AdminSelect from "@/components/admin/AdminSelect.vue"
import AdminTable, { type AdminColumn } from "@/components/admin/AdminTable.vue"
import AppAvatar from "@/components/ui/AppAvatar.vue"
import { useToast } from "@/composables/useToast"
import { useAdminStore } from "@/stores/admin"
import { useAuthStore } from "@/stores/auth"
import type { AdminUserRow, UserStatusFilter } from "@/types"
import { adminErrorKey } from "@/utils/adminError"

type ActionType = "role" | "status" | "delete"

const { t, locale } = useI18n()
const adminStore = useAdminStore()
const authStore = useAuthStore()
const toast = useToast()
const users = computed(() => adminStore.users)

const search = ref("")
const roleFilter = ref("all")
const statusFilter = ref<UserStatusFilter>("all")
const sortBy = ref("created_at")
const sortOrder = ref<"asc" | "desc">("desc")
const loadError = ref(false)

const columns = computed<AdminColumn[]>(() => [
  { key: "id", label: t("admin.users.colId") },
  { key: "username", label: t("admin.users.colUser"), sortable: true },
  { key: "email", label: t("admin.users.colEmail"), sortable: true },
  { key: "role", label: t("admin.users.colRole") },
  { key: "status", label: t("admin.users.colStatus") },
  { key: "review_count", label: t("admin.users.colReviews"), sortable: true },
  { key: "created_at", label: t("admin.users.colJoined"), sortable: true },
  { key: "actions", label: t("admin.users.colActions"), align: "end" },
])
const roleOptions = computed(() => [
  { value: "all", label: t("admin.users.roleAll") },
  { value: "admin", label: t("admin.users.roleAdmin") },
  { value: "user", label: t("admin.users.roleUser") },
])
const statusOptions = computed(() => [
  { value: "all", label: t("admin.users.statusAll") },
  { value: "active", label: t("admin.users.statusActive") },
  { value: "inactive", label: t("admin.users.statusInactive") },
  { value: "deleted", label: t("admin.users.statusDeleted") },
])
const isFiltered = computed(() => !!search.value.trim() || roleFilter.value !== "all" || statusFilter.value !== "all")

async function fetchData(page = 1) {
  loadError.value = false
  try {
    await adminStore.fetchUsers({
      page,
      limit: 20,
      search: search.value.trim(),
      role: roleFilter.value,
      status: statusFilter.value,
      sort_by: sortBy.value,
      sort_order: sortOrder.value,
    })
  } catch {
    loadError.value = true
  }
}

let debounceTimer: ReturnType<typeof setTimeout> | null = null
watch(search, () => {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => fetchData(1), 300)
})
watch([roleFilter, statusFilter], () => fetchData(1))

function onSort(key: string) {
  if (sortBy.value === key) sortOrder.value = sortOrder.value === "asc" ? "desc" : "asc"
  else {
    sortBy.value = key
    sortOrder.value = key === "created_at" || key === "review_count" ? "desc" : "asc"
  }
  fetchData(1)
}

function clearFilters() {
  search.value = ""
  roleFilter.value = "all"
  statusFilter.value = "all"
}

function menuItems(u: AdminUserRow): RowMenuItem[] {
  // A deleted account only has the second step left; every other action would 404.
  if (u.deleted_at) return [{ key: "erase", label: t("admin.users.actionErase"), icon: Flame, tone: "danger" }]
  return [
    {
      key: "role",
      label: u.role === "admin" ? t("admin.users.actionMakeUser") : t("admin.users.actionMakeAdmin"),
      icon: u.role === "admin" ? ShieldOff : Shield,
      disabled: !u.is_active,
      hint: t("admin.users.roleInactiveHint"),
    },
    u.is_active
      ? { key: "status", label: t("admin.users.actionDeactivate"), icon: UserX, tone: "warning" }
      : { key: "status", label: t("admin.users.actionReactivate"), icon: UserCheck },
    { key: "delete", label: t("admin.users.actionDelete"), icon: Trash2, tone: "danger" },
  ]
}

// Every destructive action goes through the confirm modal; nothing here calls the API directly.
const confirmModal = ref<{ show: boolean; type: ActionType; user: AdminUserRow | null; loading: boolean }>({
  show: false,
  type: "role",
  user: null,
  loading: false,
})

function startAction(user: AdminUserRow, key: string) {
  confirmModal.value = { show: true, type: key as ActionType, user, loading: false }
}

const eraseModal = ref<{ show: boolean; user: AdminUserRow | null }>({ show: false, user: null })

function startErase(user: AdminUserRow) {
  eraseModal.value = { show: true, user }
}

function onErased(user: AdminUserRow) {
  toast.success(t("admin.userModal.erased", { username: user.username }))
  fetchData(adminStore.usersPage)
}

const modal = computed<{
  title: string
  description: string
  confirm: string
  variant: "danger" | "warning" | "info" | "success"
  icon: Component
}>(() => {
  const { type, user } = confirmModal.value
  const username = user ? `@${user.username}` : ""
  if (type === "delete") {
    return {
      title: t("admin.userModal.deleteTitle"),
      description: t("admin.userModal.deleteDesc", { username }),
      confirm: t("admin.userModal.deleteConfirm"),
      variant: "danger",
      icon: Trash2,
    }
  }
  if (type === "status") {
    return user?.is_active
      ? {
          title: t("admin.userModal.deactivateTitle"),
          description: t("admin.userModal.deactivateDesc", { username }),
          confirm: t("admin.userModal.deactivateConfirm"),
          variant: "warning",
          icon: UserX,
        }
      : {
          title: t("admin.userModal.reactivateTitle"),
          description: t("admin.userModal.reactivateDesc", { username }),
          confirm: t("admin.userModal.reactivateConfirm"),
          variant: "success",
          icon: UserCheck,
        }
  }
  return user?.role === "admin"
    ? {
        title: t("admin.userModal.toUserTitle"),
        description: t("admin.userModal.toUserDesc", { username }),
        confirm: t("admin.userModal.toUserConfirm"),
        variant: "warning",
        icon: ShieldOff,
      }
    : {
        title: t("admin.userModal.toAdminTitle"),
        description: t("admin.userModal.toAdminDesc", { username }),
        confirm: t("admin.userModal.toAdminConfirm"),
        variant: "info",
        icon: Shield,
      }
})

async function onConfirmAction(payload: { reason: string }) {
  const { user, type } = confirmModal.value
  if (!user) return
  confirmModal.value.loading = true
  const reason = payload.reason || undefined
  const username = user.username
  try {
    if (type === "role") {
      await adminStore.updateUserRole(user.id, user.role === "admin" ? "user" : "admin", reason)
      toast.success(t("admin.userModal.roleUpdated", { username }))
    } else if (type === "status") {
      await adminStore.updateUserStatus(user.id, !user.is_active, reason)
      toast.success(t(user.is_active ? "admin.userModal.deactivated" : "admin.userModal.reactivated", { username }))
    } else {
      await adminStore.deleteUser(user.id, reason)
      toast.success(t("admin.userModal.deleted", { username }))
    }
    confirmModal.value.show = false
    fetchData(adminStore.usersPage)
  } catch (err: unknown) {
    toast.error(t(adminErrorKey(err)))
    if ((err as { response?: { status?: number } }).response?.status === 404) {
      confirmModal.value.show = false
      fetchData(adminStore.usersPage)
    }
  } finally {
    confirmModal.value.loading = false
  }
}

function formatDate(iso: string) {
  return iso ? new Intl.DateTimeFormat(locale.value, { dateStyle: "medium" }).format(new Date(iso)) : "-"
}

onMounted(() => fetchData(1))
onBeforeUnmount(() => {
  if (debounceTimer) clearTimeout(debounceTimer)
})
</script>

<style scoped>
.user-cell {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  min-width: 11rem;
}

.user-meta {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.user-name {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  font-weight: 600;
}

.status-cell {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 0.25rem;
}

.user-handle,
.cell-muted {
  font-size: var(--text-caption);
  color: var(--color-text-muted);
}

.cell-id,
.cell-num {
  font-variant-numeric: tabular-nums;
}

.cell-id {
  color: var(--color-text-muted);
}
</style>
