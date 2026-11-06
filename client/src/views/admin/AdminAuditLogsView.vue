<template>
  <div class="admin-audit-logs-view">
    <div class="view-header">
      <h1 class="page-title">{{ $t("admin.auditLogs.title") }}</h1>
    </div>

    <!-- Filter Bar -->
    <div class="filter-bar">
      <div class="filter-controls">
        <select v-model="actionFilter" class="filter-select" @change="fetchData(1)">
          <option value="all">{{ $t("admin.auditLogs.actionAll") }}</option>
          <option value="USER_DEACTIVATED">USER_DEACTIVATED</option>
          <option value="USER_REACTIVATED">USER_REACTIVATED</option>
          <option value="USER_ROLE_CHANGED">USER_ROLE_CHANGED</option>
          <option value="REVIEW_DELETED">REVIEW_DELETED</option>
        </select>

        <select v-model="targetTypeFilter" class="filter-select" @change="fetchData(1)">
          <option value="all">{{ $t("admin.auditLogs.targetAll") }}</option>
          <option value="user">User</option>
          <option value="review">Review</option>
        </select>
      </div>
    </div>

    <!-- Table -->
    <div class="table-container">
      <table class="admin-table">
        <thead>
          <tr>
            <th>{{ $t("admin.auditLogs.colId") }}</th>
            <th>{{ $t("admin.auditLogs.colAdmin") }}</th>
            <th>{{ $t("admin.auditLogs.colAction") }}</th>
            <th>{{ $t("admin.auditLogs.colTarget") }}</th>
            <th>{{ $t("admin.auditLogs.colReason") }}</th>
            <th>{{ $t("admin.auditLogs.colMetadata") }}</th>
            <th>{{ $t("admin.auditLogs.colTimestamp") }}</th>
          </tr>
        </thead>
        <tbody>
          <SkeletonTableRows v-if="adminStore.isLoadingAuditLogs" :columns="7" />
          <tr v-else-if="auditLogs.length === 0">
            <td colspan="7" class="text-center py-8 text-muted">
              No audit logs recorded yet.
            </td>
          </tr>
          <tr v-for="log in auditLogs" :key="log.id">
            <td>#{{ log.id }}</td>
            <td class="font-semibold">@{{ log.admin_username }}</td>
            <td>
              <span class="action-badge" :class="log.action">
                {{ log.action }}
              </span>
            </td>
            <td>
              <span class="target-tag">
                {{ log.target_type }} #{{ log.target_id }}
              </span>
            </td>
            <td class="text-secondary">{{ log.reason || '-' }}</td>
            <td>
              <code class="meta-code" v-if="log.meta_data">
                {{ JSON.stringify(log.meta_data) }}
              </code>
              <span v-else class="text-muted">-</span>
            </td>
            <td class="text-secondary">{{ formatDate(log.created_at) }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Pagination Bar -->
    <div v-if="adminStore.auditLogsTotalPages > 1" class="pagination-bar">
      <span class="pagination-info">
        {{ $t("admin.pagination.showing", {
          from: (adminStore.auditLogsPage - 1) * adminStore.auditLogsLimit + 1,
          to: Math.min(adminStore.auditLogsPage * adminStore.auditLogsLimit, adminStore.auditLogsTotal),
          total: adminStore.auditLogsTotal
        }) }}
      </span>

      <div class="pagination-buttons">
        <button
          class="page-btn"
          :disabled="adminStore.auditLogsPage <= 1"
          @click="fetchData(adminStore.auditLogsPage - 1)"
        >
          {{ $t("admin.pagination.previous") }}
        </button>
        <span class="page-num">{{ adminStore.auditLogsPage }} / {{ adminStore.auditLogsTotalPages }}</span>
        <button
          class="page-btn"
          :disabled="adminStore.auditLogsPage >= adminStore.auditLogsTotalPages"
          @click="fetchData(adminStore.auditLogsPage + 1)"
        >
          {{ $t("admin.pagination.next") }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import SkeletonTableRows from "@/components/ui/skeleton/SkeletonTableRows.vue"
import { ref, computed, onMounted } from "vue"
import { useAdminStore } from "@/stores/admin"

const adminStore = useAdminStore()
const auditLogs = computed(() => adminStore.auditLogs)

const actionFilter = ref("all")
const targetTypeFilter = ref("all")

function fetchData(page = 1) {
  adminStore.fetchAuditLogs({
    page,
    limit: 20,
    action: actionFilter.value,
    target_type: targetTypeFilter.value,
  })
}

function formatDate(isoStr: string) {
  if (!isoStr) return "-"
  return new Date(isoStr).toLocaleString()
}

onMounted(() => {
  fetchData(1)
})
</script>

<style scoped>
.admin-audit-logs-view {
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
  justify-content: flex-end;
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

.table-container {
  background: var(--color-surface-2);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-card);
  overflow-x: auto;
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

.font-semibold { font-weight: 600; }

.action-badge {
  display: inline-block;
  padding: 0.2rem 0.5.rem;
  border-radius: 0.35rem;
  font-size: 0.75rem;
  font-weight: 700;
  text-transform: uppercase;
}
.action-badge.USER_DEACTIVATED { background: rgba(239, 68, 68, 0.15); color: #ef4444; }
.action-badge.USER_REACTIVATED { background: rgba(16, 185, 129, 0.15); color: #10b981; }
.action-badge.USER_ROLE_CHANGED { background: rgba(59, 130, 246, 0.15); color: #3b82f6; }
.action-badge.REVIEW_DELETED { background: rgba(245, 158, 11, 0.15); color: #f59e0b; }

.target-tag {
  font-size: 0.775rem;
  font-weight: 600;
  padding: 0.2rem 0.4rem;
  border-radius: 0.25rem;
  background: var(--color-surface-3);
  color: var(--color-text-secondary);
}

.meta-code {
  font-family: monospace;
  font-size: 0.75rem;
  background: rgba(0, 0, 0, 0.3);
  padding: 0.15rem 0.4rem;
  border-radius: 0.25rem;
  color: var(--color-text-secondary);
}

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
}
.page-btn:disabled { opacity: 0.4; cursor: not-allowed; }
</style>
