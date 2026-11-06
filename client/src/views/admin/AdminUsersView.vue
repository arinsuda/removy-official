<template>
  <div class="admin-users-view">
    <div class="view-header">
      <h1 class="page-title">{{ $t("admin.users.title") }}</h1>
    </div>

    <!-- Filters & Search Bar -->
    <div class="filter-bar">
      <div class="search-input-wrapper">
        <Search :size="16" class="search-icon" />
        <input
          v-model="search"
          type="text"
          class="search-input"
          :placeholder="$t('admin.users.searchPlaceholder')"
          @input="onSearchInput"
        />
      </div>

      <div class="filter-controls">
        <select v-model="roleFilter" class="filter-select" @change="fetchData(1)">
          <option value="all">{{ $t("admin.users.roleAll") }}</option>
          <option value="admin">{{ $t("admin.users.roleAdmin") }}</option>
          <option value="user">{{ $t("admin.users.roleUser") }}</option>
        </select>

        <select v-model="statusFilter" class="filter-select" @change="fetchData(1)">
          <option value="all">{{ $t("admin.users.statusAll") }}</option>
          <option value="active">{{ $t("admin.users.statusActive") }}</option>
          <option value="inactive">{{ $t("admin.users.statusInactive") }}</option>
        </select>

        <select v-model="sortBy" class="filter-select" @change="fetchData(1)">
          <option value="created_at">Sort by Joined</option>
          <option value="username">Sort by Username</option>
          <option value="email">Sort by Email</option>
          <option value="review_count">Sort by Reviews</option>
        </select>

        <button class="sort-order-btn" @click="toggleSortOrder">
          <ArrowUpDown :size="16" />
          <span>{{ sortOrder.toUpperCase() }}</span>
        </button>
      </div>
    </div>

    <!-- Table -->
    <div class="table-container">
      <table class="admin-table">
        <thead>
          <tr>
            <th>{{ $t("admin.users.colId") }}</th>
            <th>{{ $t("admin.users.colUser") }}</th>
            <th>{{ $t("admin.users.colEmail") }}</th>
            <th>{{ $t("admin.users.colRole") }}</th>
            <th>{{ $t("admin.users.colStatus") }}</th>
            <th>{{ $t("admin.users.colReviews") }}</th>
            <th>{{ $t("admin.users.colJoined") }}</th>
            <th class="text-right">{{ $t("admin.users.colActions") }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="adminStore.isLoadingUsers">
            <td colspan="8" class="text-center py-8">
              <div class="spinner inline" />
            </td>
          </tr>
          <tr v-else-if="users.length === 0">
            <td colspan="8" class="text-center py-8 text-muted">
              No users found.
            </td>
          </tr>
          <tr v-for="u in users" :key="u.id" class="user-row">
            <td>#{{ u.id }}</td>
            <td>
              <div class="user-cell">
                <img
                  v-if="u.avatar_url"
                  :src="u.avatar_url"
                  class="user-avatar"
                  :alt="u.username"
                  loading="lazy"
                  decoding="async"
                />
                <div v-else class="user-avatar placeholder">
                  {{ u.username.charAt(0).toUpperCase() }}
                </div>
                <div class="user-meta">
                  <span class="user-name">{{ u.display_name || u.username }}</span>
                  <span class="user-handle">@{{ u.username }}</span>
                </div>
              </div>
            </td>
            <td class="text-secondary">{{ u.email }}</td>
            <td>
              <span class="role-badge" :class="u.role">
                {{ u.role === "admin" ? $t("admin.users.roleAdmin") : $t("admin.users.roleUser") }}
              </span>
            </td>
            <td>
              <span class="status-badge" :class="u.is_active ? 'active' : 'inactive'">
                {{ u.is_active ? $t("admin.users.statusActive") : $t("admin.users.statusInactive") }}
              </span>
            </td>
            <td>{{ u.review_count }}</td>
            <td class="text-secondary">{{ formatDate(u.created_at) }}</td>
            <td class="text-right">
              <div class="action-buttons">
                <!-- Change Role Button -->
                <button
                  class="action-btn"
                  :disabled="!u.is_active"
                  :title="!u.is_active ? 'Cannot change role of inactive user' : 'Change Role'"
                  @click="handleRoleChange(u)"
                >
                  <Shield :size="14" />
                </button>

                <!-- Toggle Status Button -->
                <button
                  class="action-btn"
                  :class="u.is_active ? 'warning' : 'success'"
                  :title="u.is_active ? 'Deactivate' : 'Reactivate'"
                  @click="handleStatusChange(u)"
                >
                  <UserX v-if="u.is_active" :size="14" />
                  <UserCheck v-else :size="14" />
                </button>

                <!-- Delete User Account Button -->
                <button
                  class="action-btn danger"
                  title="Delete Account"
                  @click="handleUserDelete(u)"
                >
                  <Trash2 :size="14" />
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Pagination Controls -->
    <div v-if="adminStore.usersTotalPages > 1" class="pagination-bar">
      <span class="pagination-info">
        {{ $t("admin.pagination.showing", {
          from: (adminStore.usersPage - 1) * adminStore.usersLimit + 1,
          to: Math.min(adminStore.usersPage * adminStore.usersLimit, adminStore.usersTotal),
          total: adminStore.usersTotal
        }) }}
      </span>

      <div class="pagination-buttons">
        <button
          class="page-btn"
          :disabled="adminStore.usersPage <= 1"
          @click="fetchData(adminStore.usersPage - 1)"
        >
          {{ $t("admin.pagination.previous") }}
        </button>
        <span class="page-num">{{ adminStore.usersPage }} / {{ adminStore.usersTotalPages }}</span>
        <button
          class="page-btn"
          :disabled="adminStore.usersPage >= adminStore.usersTotalPages"
          @click="fetchData(adminStore.usersPage + 1)"
        >
          {{ $t("admin.pagination.next") }}
        </button>
      </div>
    </div>

    <!-- Official Admin Confirm Modal -->
    <AdminConfirmModal
      v-model="confirmModal.show"
      :title="modalTitle"
      :description="modalDescription"
      :target-name="confirmModal.user ? `@${confirmModal.user.username}` : ''"
      :variant="modalVariant"
      :icon="modalIcon"
      :confirm-text="modalConfirmText"
      :loading="confirmModal.loading"
      @confirm="onConfirmAction"
      @cancel="confirmModal.show = false"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue"
import { useAdminStore } from "@/stores/admin"
import { useToast } from "@/composables/useToast"
import type { AdminUserRow } from "@/types"
import { Search, ArrowUpDown, Shield, UserX, UserCheck, Trash2 } from "lucide-vue-next"
import AdminConfirmModal from "@/components/admin/AdminConfirmModal.vue"

const adminStore = useAdminStore()
const toast = useToast()
const users = computed(() => adminStore.users)

const search = ref("")
const roleFilter = ref("all")
const statusFilter = ref("all")
const sortBy = ref("created_at")
const sortOrder = ref<"asc" | "desc">("desc")

let debounceTimer: ReturnType<typeof setTimeout> | null = null

interface ConfirmModalState {
  show: boolean
  type: "role" | "status" | "delete"
  user: AdminUserRow | null
  targetRole?: string
  newStatus?: boolean
  loading: boolean
}

const confirmModal = ref<ConfirmModalState>({
  show: false,
  type: "role",
  user: null,
  loading: false,
})

const modalTitle = computed(() => {
  if (!confirmModal.value.user) return ""
  if (confirmModal.value.type === "role") {
    return "เปลี่ยนสิทธิ์ผู้ใช้งาน"
  }
  if (confirmModal.value.type === "delete") {
    return "ลบบัญชีผู้ใช้"
  }
  return confirmModal.value.newStatus ? "เปิดคืนสิทธิ์ใช้งานบัญชี" : "ระงับการใช้งานบัญชี"
})

const modalDescription = computed(() => {
  if (!confirmModal.value.user) return ""
  if (confirmModal.value.type === "role") {
    return `ต้องการเปลี่ยนสิทธิ์ของ @${confirmModal.value.user.username} จาก '${confirmModal.value.user.role}' เป็น '${confirmModal.value.targetRole}' หรือไม่?`
  }
  if (confirmModal.value.type === "delete") {
    return `คุณแน่ใจหรือไม่ที่จะลบบัญชี @${confirmModal.value.user.username}? ข้อมูลผู้ใช้และรีวิวที่เกี่ยวข้องจะถูก Soft Delete และบันทึกประวัติลง Audit Log`
  }
  return confirmModal.value.newStatus
    ? `คุณแน่ใจหรือไม่ที่จะเปิดคืนสิทธิ์ใช้งานบัญชี @${confirmModal.value.user.username}?`
    : `คุณแน่ใจหรือไม่ที่จะระงับบัญชี @${confirmModal.value.user.username}? ผู้ใช้จะไม่สามารถเข้าสู่ระบบได้`
})

const modalVariant = computed<"danger" | "warning" | "info" | "success">(() => {
  if (confirmModal.value.type === "role") return "info"
  if (confirmModal.value.type === "delete") return "danger"
  return confirmModal.value.newStatus ? "success" : "warning"
})

const modalIcon = computed(() => {
  if (confirmModal.value.type === "role") return Shield
  if (confirmModal.value.type === "delete") return Trash2
  return confirmModal.value.newStatus ? UserCheck : UserX
})

const modalConfirmText = computed(() => {
  if (confirmModal.value.type === "role") return "เปลี่ยนสิทธิ์"
  if (confirmModal.value.type === "delete") return "ยืนยันลบบัญชี"
  return confirmModal.value.newStatus ? "เปิดใช้งานบัญชี" : "ระงับบัญชี"
})

function onSearchInput() {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => fetchData(1), 300)
}

function toggleSortOrder() {
  sortOrder.value = sortOrder.value === "asc" ? "desc" : "asc"
  fetchData(1)
}

function fetchData(page = 1) {
  adminStore.fetchUsers({
    page,
    limit: 20,
    search: search.value.trim(),
    role: roleFilter.value,
    status: statusFilter.value,
    sort_by: sortBy.value,
    sort_order: sortOrder.value,
  })
}

function handleRoleChange(user: AdminUserRow) {
  const targetRole = user.role === "admin" ? "user" : "admin"
  confirmModal.value = {
    show: true,
    type: "role",
    user,
    targetRole,
    loading: false,
  }
}

function handleStatusChange(user: AdminUserRow) {
  const newStatus = !user.is_active
  confirmModal.value = {
    show: true,
    type: "status",
    user,
    newStatus,
    loading: false,
  }
}

function handleUserDelete(user: AdminUserRow) {
  confirmModal.value = {
    show: true,
    type: "delete",
    user,
    loading: false,
  }
}

async function onConfirmAction(payload: { reason: string }) {
  if (!confirmModal.value.user) return
  confirmModal.value.loading = true
  const user = confirmModal.value.user
  const reason = payload.reason || undefined

  try {
    if (confirmModal.value.type === "role" && confirmModal.value.targetRole) {
      await adminStore.updateUserRole(user.id, confirmModal.value.targetRole, reason)
      toast.success(`อัปเดตสิทธิ์ของ @${user.username} เรียบร้อยแล้ว`)
    } else if (confirmModal.value.type === "status" && confirmModal.value.newStatus !== undefined) {
      await adminStore.updateUserStatus(user.id, confirmModal.value.newStatus, reason)
      const statusText = confirmModal.value.newStatus ? "เปิดคืนสิทธิ์ใช้งาน" : "ระงับการใช้งาน"
      toast.success(`${statusText} @${user.username} เรียบร้อยแล้ว`)
    } else if (confirmModal.value.type === "delete") {
      await adminStore.deleteUser(user.id, reason)
      toast.success(`ลบบัญชี @${user.username} เรียบร้อยแล้ว`)
    }
    confirmModal.value.show = false
    fetchData(adminStore.usersPage)
  } catch (err: unknown) {
    const e = err as { response?: { data?: { error?: string } } }
    toast.error(e.response?.data?.error || "เกิดข้อผิดพลาดในการดำเนินการ")
  } finally {
    confirmModal.value.loading = false
  }
}

function formatDate(isoStr: string) {
  if (!isoStr) return "-"
  return new Date(isoStr).toLocaleDateString()
}

onMounted(() => {
  fetchData(1)
})
</script>

<style scoped>
.admin-users-view {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.page-title {
  font-size: 1.6rem;
  font-weight: 800;
  margin: 0;
  color: var(--color-text-primary);
}

.filter-bar {
  display: flex;
  flex-wrap: wrap;
  gap: 1rem;
  justify-content: space-between;
  align-items: center;
}

.search-input-wrapper {
  position: relative;
  flex: 1;
  min-width: 260px;
}

.search-icon {
  position: absolute;
  left: 0.85rem;
  top: 50%;
  transform: translateY(-50%);
  color: var(--color-text-secondary);
}

.search-input {
  width: 100%;
  padding: 0.6rem 0.85rem 0.6rem 2.4rem;
  background: var(--color-surface-2);
  border: 1px solid var(--color-border);
  border-radius: 0.5rem;
  color: var(--color-text-primary);
  outline: none;
  font-size: 0.875rem;
  box-sizing: border-box;
}

.filter-controls {
  display: flex;
  align-items: center;
  gap: 0.6rem;
}

.filter-select {
  padding: 0.6rem 0.85rem;
  background: var(--color-surface-2);
  border: 1px solid var(--color-border);
  border-radius: 0.5rem;
  color: var(--color-text-primary);
  font-size: 0.85rem;
  outline: none;
  cursor: pointer;
}

.sort-order-btn {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.6rem 0.85rem;
  background: var(--color-surface-2);
  border: 1px solid var(--color-border);
  border-radius: 0.5rem;
  color: var(--color-text-primary);
  font-size: 0.85rem;
  cursor: pointer;
}

.table-container {
  background: var(--color-surface-2);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-card);
  overflow-x: auto;
  contain: layout style;
}

.admin-table {
  width: 100%;
  border-collapse: collapse;
  text-align: left;
  font-size: 0.875rem;
}

.admin-table th {
  padding: 0.85rem 1rem;
  background: rgba(255, 255, 255, 0.02);
  border-bottom: 1px solid var(--color-border);
  color: var(--color-text-secondary);
  font-weight: 600;
}

.admin-table td {
  padding: 0.85rem 1rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  color: var(--color-text-primary);
}

.user-row {
  transition: background-color 0.15s ease;
  will-change: background-color;
}

.user-row:hover {
  background-color: rgba(255, 255, 255, 0.02);
}

.user-cell {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.user-avatar {
  width: 34px;
  height: 34px;
  border-radius: 50%;
  object-fit: cover;
}

.user-avatar.placeholder {
  background: var(--color-surface-3);
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  color: var(--color-brand);
}

.user-meta {
  display: flex;
  flex-direction: column;
}

.user-name { font-weight: 600; }
.user-handle { font-size: 0.775rem; color: var(--color-text-muted); }

.role-badge {
  display: inline-block;
  padding: 0.2rem 0.5rem;
  border-radius: 0.35rem;
  font-size: 0.75rem;
  font-weight: 700;
  text-transform: uppercase;
}
.role-badge.admin { background: rgba(245, 67, 44, 0.2); color: var(--color-brand); }
.role-badge.user { background: rgba(255, 255, 255, 0.08); color: var(--color-text-secondary); }

.status-badge {
  display: inline-block;
  padding: 0.2rem 0.5rem;
  border-radius: 0.35rem;
  font-size: 0.75rem;
  font-weight: 700;
}
.status-badge.active { background: rgba(16, 185, 129, 0.15); color: #10b981; }
.status-badge.inactive { background: rgba(239, 68, 68, 0.15); color: #ef4444; }

.action-buttons {
  display: flex;
  justify-content: flex-end;
  gap: 0.4rem;
}

.action-btn {
  padding: 0.4rem;
  background: var(--color-surface-3);
  border: 1px solid var(--color-border);
  color: var(--color-text-secondary);
  border-radius: 0.35rem;
  cursor: pointer;
  transition: color 0.15s ease, border-color 0.15s ease, transform 0.15s ease;
  transform: translateZ(0);
}
.action-btn:hover:not(:disabled) { color: var(--color-text-primary); border-color: var(--color-brand); transform: translateY(-1px) translateZ(0); }
.action-btn.danger:hover:not(:disabled) { color: #ef4444; border-color: #ef4444; }
.action-btn.warning:hover:not(:disabled) { color: #f59e0b; border-color: #f59e0b; }
.action-btn.success:hover:not(:disabled) { color: #10b981; border-color: #10b981; }
.action-btn:disabled { opacity: 0.4; cursor: not-allowed; transform: none !important; }

.text-right { text-align: right; }
.text-center { text-align: center; }
.text-secondary { color: var(--color-text-secondary); }
.text-muted { color: var(--color-text-muted); }
.py-8 { padding-top: 2rem; padding-bottom: 2rem; }

.pagination-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.85rem;
  color: var(--color-text-secondary);
}

.pagination-buttons {
  display: flex;
  align-items: center;
  gap: 0.6rem;
}

.page-btn {
  padding: 0.4rem 0.85rem;
  background: var(--color-surface-2);
  border: 1px solid var(--color-border);
  color: var(--color-text-primary);
  border-radius: 0.35rem;
  cursor: pointer;
  transition: background-color 0.15s ease;
}
.page-btn:disabled { opacity: 0.4; cursor: not-allowed; }
</style>
