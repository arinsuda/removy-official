<template>
  <nav v-if="total > 0" class="admin-pagination" :aria-label="t('admin.pagination.label')">
    <p class="admin-pagination__info">
      {{ t("admin.pagination.showing", { from: (page - 1) * limit + 1, to: Math.min(page * limit, total), total }) }}
    </p>
    <div class="admin-pagination__controls">
      <AppButton variant="outline" size="sm" :disabled="page <= 1" @click="emit('change', page - 1)">
        <ChevronLeft :size="16" aria-hidden="true" />
        {{ t("admin.pagination.previous") }}
      </AppButton>
      <span class="admin-pagination__page" aria-current="page">
        {{ t("admin.pagination.page", { page, pages: totalPages }) }}
      </span>
      <AppButton variant="outline" size="sm" :disabled="page >= totalPages" @click="emit('change', page + 1)">
        {{ t("admin.pagination.next") }}
        <ChevronRight :size="16" aria-hidden="true" />
      </AppButton>
    </div>
  </nav>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n"
import { ChevronLeft, ChevronRight } from "lucide-vue-next"
import AppButton from "@/components/ui/AppButton.vue"

defineProps<{ page: number; totalPages: number; total: number; limit: number }>()
const emit = defineEmits<{ change: [page: number] }>()
const { t } = useI18n()
</script>

<style scoped>
.admin-pagination {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
}

.admin-pagination__info {
  margin: 0;
  font-size: var(--text-small);
  color: var(--color-text-secondary);
}

.admin-pagination__controls {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.admin-pagination__page {
  font-size: var(--text-small);
  font-variant-numeric: tabular-nums;
  color: var(--color-text-muted);
}
</style>
