<template>
  <AdminPage>
    <AdminPageHeader :title="t('admin.auditLogs.title')" :subtitle="t('admin.auditLogs.subtitle')" />

    <AdminFilters>
      <AdminSelect v-model="actionFilter" :label="t('admin.auditLogs.colAction')" :options="actionOptions" />
      <AdminSelect v-model="targetTypeFilter" :label="t('admin.auditLogs.colTarget')" :options="targetOptions" />
    </AdminFilters>

    <SkeletonList v-if="adminStore.isLoadingAuditLogs" :count="6" />
    <AppErrorState
      v-else-if="loadError"
      :title="t('admin.table.errorTitle')"
      :description="t('admin.table.errorDesc')"
      @retry="fetchData(adminStore.auditLogsPage)"
    />
    <AppEmptyState
      v-else-if="auditLogs.length === 0"
      :icon="isFiltered ? SearchX : ScrollText"
      :title="isFiltered ? t('admin.table.noResultsTitle') : t('admin.auditLogs.emptyTitle')"
      :description="isFiltered ? t('admin.table.noResultsDesc') : t('admin.auditLogs.emptyDesc')"
    >
      <template v-if="isFiltered" #action>
        <AppButton variant="outline" size="sm" @click="clearFilters">{{ t("admin.table.clearFilters") }}</AppButton>
      </template>
    </AppEmptyState>

    <div v-else class="timeline">
      <section v-for="group in groups" :key="group.day" class="timeline__day">
        <h2 class="timeline__heading">{{ group.label }}</h2>
        <ol class="timeline__list">
          <li v-for="log in group.items" :key="log.id" class="entry" :class="`entry--${tone(log.action)}`">
            <span class="entry__dot" aria-hidden="true"><component :is="icon(log.action)" :size="14" /></span>
            <div class="entry__body">
              <div class="entry__head">
                <strong class="entry__action">{{ actionLabel(log.action) }}</strong>
                <AdminBadge>{{ t("admin.auditLogs.target", { type: targetLabel(log.target_type), id: log.target_id }) }}</AdminBadge>
                <time class="entry__time" :datetime="log.created_at">{{ formatTime(log.created_at) }}</time>
              </div>
              <p class="entry__by">{{ actorLabel(log) }}</p>
              <p class="entry__reason" :class="{ 'is-empty': !log.reason }">{{ log.reason || t("admin.auditLogs.noReason") }}</p>
              <details v-if="hasMeta(log)" class="entry__meta">
                <summary>{{ t("admin.auditLogs.details") }}</summary>
                <dl>
                  <template v-for="(value, key) in log.meta_data" :key="key">
                    <dt>{{ key }}</dt>
                    <dd>{{ formatMeta(value) }}</dd>
                  </template>
                </dl>
              </details>
            </div>
          </li>
        </ol>
      </section>
    </div>

    <AdminPagination
      :page="adminStore.auditLogsPage"
      :total-pages="adminStore.auditLogsTotalPages"
      :total="adminStore.auditLogsTotal"
      :limit="adminStore.auditLogsLimit"
      @change="fetchData"
    />
  </AdminPage>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch, type Component } from "vue"
import { useI18n } from "vue-i18n"
import { Flame, MessageSquareOff, ScrollText, SearchX, Shield, Trash2, UserCheck, UserX } from "lucide-vue-next"
import AdminBadge from "@/components/admin/AdminBadge.vue"
import AdminFilters from "@/components/admin/AdminFilters.vue"
import AdminPage from "@/components/admin/AdminPage.vue"
import AdminPageHeader from "@/components/admin/AdminPageHeader.vue"
import AdminPagination from "@/components/admin/AdminPagination.vue"
import AdminSelect from "@/components/admin/AdminSelect.vue"
import AppButton from "@/components/ui/AppButton.vue"
import AppEmptyState from "@/components/ui/AppEmptyState.vue"
import AppErrorState from "@/components/ui/AppErrorState.vue"
import SkeletonList from "@/components/ui/skeleton/SkeletonList.vue"
import { useAdminStore } from "@/stores/admin"
import type { AdminAuditLogRow } from "@/types"

const { t, te, locale } = useI18n()
const adminStore = useAdminStore()
const auditLogs = computed(() => adminStore.auditLogs)

const actionFilter = ref("all")
const targetTypeFilter = ref("all")
const loadError = ref(false)
const isFiltered = computed(() => actionFilter.value !== "all" || targetTypeFilter.value !== "all")

const ACTIONS = ["USER_DEACTIVATED", "USER_REACTIVATED", "USER_ROLE_CHANGED", "USER_DELETED", "USER_ERASED", "USER_ERASURE_REQUESTED", "USER_ERASED_SCHEDULED", "REVIEW_DELETED"]
const actionOptions = computed(() => [
  { value: "all", label: t("admin.auditLogs.actionAll") },
  ...ACTIONS.map((a) => ({ value: a, label: t(`admin.auditLogs.action.${a}`) })),
])
const targetOptions = computed(() => [
  { value: "all", label: t("admin.auditLogs.targetAll") },
  { value: "user", label: t("admin.auditLogs.targetUser") },
  { value: "review", label: t("admin.auditLogs.targetReview") },
])

const VISUALS: Record<string, { tone: "danger" | "warning" | "success" | "info"; icon: Component }> = {
  USER_DEACTIVATED: { tone: "warning", icon: UserX },
  USER_REACTIVATED: { tone: "success", icon: UserCheck },
  USER_ROLE_CHANGED: { tone: "info", icon: Shield },
  USER_DELETED: { tone: "danger", icon: Trash2 },
  USER_ERASED: { tone: "danger", icon: Flame },
  USER_ERASURE_REQUESTED: { tone: "warning", icon: UserX },
  USER_ERASED_SCHEDULED: { tone: "danger", icon: Flame },
  REVIEW_DELETED: { tone: "danger", icon: MessageSquareOff },
}
// Actions written without an admin: the account owner asked, or the purge worker ran.
const ACTOR_LABELS: Record<string, string> = {
  USER_ERASURE_REQUESTED: "admin.auditLogs.byAccountOwner",
  USER_ERASED_SCHEDULED: "admin.auditLogs.bySystem",
}
function actorLabel(log: AdminAuditLogRow): string {
  if (log.admin_username) return t("admin.auditLogs.byAdmin", { admin: log.admin_username })
  const key = ACTOR_LABELS[log.action]
  return key ? t(key) : t("admin.auditLogs.byDeletedAdmin")
}
const tone = (action: string) => VISUALS[action]?.tone ?? "info"
const icon = (action: string) => VISUALS[action]?.icon ?? ScrollText
// Unknown future actions fall back to the raw code instead of a missing-key warning.
const actionLabel = (action: string) => (te(`admin.auditLogs.action.${action}`) ? t(`admin.auditLogs.action.${action}`) : action)
const targetLabel = (type: string) =>
  type === "user" ? t("admin.auditLogs.targetUser") : type === "review" ? t("admin.auditLogs.targetReview") : type

const hasMeta = (log: AdminAuditLogRow) => !!log.meta_data && Object.keys(log.meta_data).length > 0
const formatMeta = (v: unknown) => (typeof v === "object" && v !== null ? JSON.stringify(v) : String(v))
const formatTime = (iso: string) => new Intl.DateTimeFormat(locale.value, { timeStyle: "short" }).format(new Date(iso))

const groups = computed(() => {
  const dayFmt = new Intl.DateTimeFormat(locale.value, { dateStyle: "full" })
  const out: { day: string; label: string; items: AdminAuditLogRow[] }[] = []
  for (const log of auditLogs.value) {
    const d = new Date(log.created_at)
    const day = Number.isNaN(d.getTime()) ? log.created_at : d.toDateString()
    let g = out[out.length - 1]
    if (!g || g.day !== day) {
      g = { day, label: Number.isNaN(d.getTime()) ? day : dayFmt.format(d), items: [] }
      out.push(g)
    }
    g.items.push(log)
  }
  return out
})

async function fetchData(page = 1) {
  loadError.value = false
  try {
    await adminStore.fetchAuditLogs({
      page,
      limit: 20,
      action: actionFilter.value,
      target_type: targetTypeFilter.value,
    })
  } catch {
    loadError.value = true
  }
}

function clearFilters() {
  actionFilter.value = "all"
  targetTypeFilter.value = "all"
}

watch([actionFilter, targetTypeFilter], () => fetchData(1))
onMounted(() => fetchData(1))
</script>

<style scoped>
.timeline {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.timeline__heading {
  margin: 0 0 0.75rem;
  font-size: var(--text-caption);
  font-weight: 700;
  letter-spacing: 0.04em;
  color: var(--color-text-muted);
}

.timeline__list {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  margin: 0;
  padding: 0;
  list-style: none;
}

/* The rail runs behind the dots. */
.timeline__list::before {
  content: "";
  position: absolute;
  top: 0.75rem;
  bottom: 0.75rem;
  left: 1rem;
  width: 1px;
  background: var(--color-border-strong);
}

.entry {
  --tone: var(--color-info);
  position: relative;
  display: flex;
  gap: 0.875rem;
}

.entry--warning {
  --tone: var(--color-warning);
}

.entry--success {
  --tone: var(--color-success);
}

.entry--danger {
  --tone: var(--color-danger);
}

.entry__dot {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  width: 2rem;
  height: 2rem;
  border: 1px solid color-mix(in srgb, var(--tone) 45%, transparent);
  border-radius: 50%;
  background: var(--color-surface);
  color: var(--tone);
}

.entry__body {
  flex: 1;
  min-width: 0;
  padding: 0.75rem 1rem;
  background: var(--color-surface-2);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-card);
}

.entry__head {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.5rem;
}

.entry__action {
  font-size: var(--text-small);
  color: var(--color-text-primary);
}

.entry__time {
  margin-left: auto;
  font-size: var(--text-caption);
  font-variant-numeric: tabular-nums;
  color: var(--color-text-muted);
}

.entry__by,
.entry__reason {
  margin: 0.25rem 0 0;
  font-size: var(--text-small);
  color: var(--color-text-secondary);
  overflow-wrap: anywhere;
}

.entry__reason.is-empty {
  font-style: italic;
  color: var(--color-text-muted);
}

.entry__meta {
  margin-top: 0.5rem;
  font-size: var(--text-caption);
  color: var(--color-text-secondary);
}

.entry__meta summary {
  cursor: pointer;
  color: var(--color-text-muted);
}

.entry__meta dl {
  display: grid;
  grid-template-columns: max-content 1fr;
  gap: 0.25rem 0.75rem;
  margin: 0.5rem 0 0;
}

.entry__meta dt {
  font-weight: 600;
}

.entry__meta dd {
  margin: 0;
  overflow-wrap: anywhere;
}
</style>
