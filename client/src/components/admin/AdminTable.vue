<template>
  <div class="admin-table">
    <div
      class="admin-table__scroll"
      role="region"
      tabindex="0"
      :aria-label="label"
      :aria-busy="loading || undefined"
    >
      <table class="admin-table__table">
        <caption class="sr-only">{{ label }}</caption>
        <thead>
          <tr>
            <th
              v-for="c in columns"
              :key="c.key"
              scope="col"
              :class="{ 'is-end': c.align === 'end' }"
              :aria-sort="c.sortable && sortBy === c.key ? (sortOrder === 'asc' ? 'ascending' : 'descending') : undefined"
            >
              <button
                v-if="c.sortable"
                type="button"
                class="admin-table__sort"
                :class="{ 'is-active': sortBy === c.key }"
                @click="emit('sort', c.key)"
              >
                {{ c.label }}
                <ArrowUp v-if="sortBy === c.key && sortOrder === 'asc'" :size="14" aria-hidden="true" />
                <ArrowDown v-else-if="sortBy === c.key" :size="14" aria-hidden="true" />
                <ArrowUpDown v-else :size="14" aria-hidden="true" />
              </button>
              <template v-else>{{ c.label }}</template>
            </th>
          </tr>
        </thead>
        <tbody v-if="loading">
          <SkeletonTableRows :columns="columns.length" />
        </tbody>
        <tbody v-else-if="!error && !empty">
          <slot />
        </tbody>
      </table>

      <AppErrorState v-if="error && !loading" :title="t('admin.table.errorTitle')" :description="t('admin.table.errorDesc')" @retry="emit('retry')" />
      <AppEmptyState
        v-else-if="empty && !loading"
        :icon="filtered ? SearchX : Inbox"
        :title="filtered ? t('admin.table.noResultsTitle') : (emptyTitle ?? t('admin.table.emptyTitle'))"
        :description="filtered ? t('admin.table.noResultsDesc') : emptyDesc"
      >
        <template v-if="filtered" #action>
          <AppButton variant="outline" size="sm" @click="emit('clear')">{{ t("admin.table.clearFilters") }}</AppButton>
        </template>
      </AppEmptyState>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n"
import { ArrowDown, ArrowUp, ArrowUpDown, Inbox, SearchX } from "lucide-vue-next"
import AppButton from "@/components/ui/AppButton.vue"
import AppEmptyState from "@/components/ui/AppEmptyState.vue"
import AppErrorState from "@/components/ui/AppErrorState.vue"
import SkeletonTableRows from "@/components/ui/skeleton/SkeletonTableRows.vue"

export interface AdminColumn {
  key: string
  label: string
  sortable?: boolean
  align?: "end"
}

withDefaults(
  defineProps<{
    columns: AdminColumn[]
    label: string
    loading?: boolean
    error?: boolean
    empty?: boolean
    /** True when search or filters are active, so an empty list means "no matches". */
    filtered?: boolean
    emptyTitle?: string
    emptyDesc?: string
    sortBy?: string
    sortOrder?: "asc" | "desc"
  }>(),
  { loading: false, error: false, empty: false, filtered: false },
)
const emit = defineEmits<{ sort: [key: string]; retry: []; clear: [] }>()
const { t } = useI18n()
</script>

<style scoped>
.admin-table {
  background: var(--color-surface-2);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-card);
  overflow: hidden;
}

/* The scroll area owns both axes so the header stays pinned while rows scroll. */
.admin-table__scroll {
  max-height: min(72dvh, 46rem);
  overflow: auto;
}

.admin-table__table {
  width: 100%;
  border-collapse: separate;
  border-spacing: 0;
  text-align: left;
  font-size: var(--text-small);
}

.admin-table__table :deep(th) {
  position: sticky;
  top: 0;
  z-index: 2;
  padding: 0.75rem 1rem;
  background: var(--color-surface-3);
  border-bottom: 1px solid var(--color-border-strong);
  font-size: var(--text-caption);
  font-weight: 700;
  letter-spacing: 0.02em;
  white-space: nowrap;
  color: var(--color-text-secondary);
}

.admin-table__table :deep(th.is-end),
.admin-table__table :deep(td.is-end) {
  text-align: right;
}

.admin-table__table :deep(td) {
  padding: 0.75rem 1rem;
  border-bottom: 1px solid var(--color-border);
  vertical-align: middle;
  color: var(--color-text-primary);
}

.admin-table__table :deep(tbody tr:last-child td) {
  border-bottom: 0;
}

.admin-table__sort {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  padding: 0;
  background: none;
  border: 0;
  font: inherit;
  letter-spacing: inherit;
  color: inherit;
  cursor: pointer;
}

.admin-table__sort.is-active {
  color: var(--color-text-primary);
}

/* On phones the row actions stay reachable while the other columns scroll sideways. */
@media (max-width: 767px) {
  .admin-table__table :deep(th.is-end),
  .admin-table__table :deep(td.is-end) {
    position: sticky;
    right: 0;
    background: var(--color-surface-2);
    box-shadow: -1px 0 0 var(--color-border);
  }

  .admin-table__table :deep(th.is-end) {
    z-index: 3;
    background: var(--color-surface-3);
  }
}

@media (hover: hover) {
  .admin-table__table :deep(tbody tr:hover td) {
    background: var(--color-surface-3);
  }

  .admin-table__sort:hover {
    color: var(--color-text-primary);
  }
}
</style>
