<template>
  <section class="rail" :aria-labelledby="titleId">
    <header class="rail__header">
      <div class="rail__heading">
        <h2 :id="titleId" class="rail__title">{{ title }}</h2>
        <p v-if="subtitle" class="rail__subtitle">{{ subtitle }}</p>
      </div>
      <RouterLink v-if="seeAllTo" :to="seeAllTo" class="rail__see-all link-nudge">
        <span>{{ t("home.rail.seeAll") }}</span>
        <ChevronRight :size="16" aria-hidden="true" />
      </RouterLink>
    </header>

    <slot name="toolbar" />

    <div v-if="loading" class="rail__grid" aria-hidden="true">
      <SkeletonBlock v-for="i in skeletonCount" :key="i" aspect="2 / 3" radius="12px" />
    </div>
    <AppErrorState v-else-if="error" :title="t('home.rail.error')" @retry="emit('retry')" />
    <AppEmptyState v-else-if="empty" :title="emptyTitle ?? t('home.rail.empty')" :description="emptyDescription">
      <template v-if="$slots.empty" #action><slot name="empty" /></template>
    </AppEmptyState>
    <div v-else class="rail__grid"><slot /></div>
  </section>
</template>

<script setup lang="ts">
import { useId } from "vue"
import { useI18n } from "vue-i18n"
import { ChevronRight } from "lucide-vue-next"
import type { RouteLocationRaw } from "vue-router"
import AppEmptyState from "@/components/ui/AppEmptyState.vue"
import AppErrorState from "@/components/ui/AppErrorState.vue"
import SkeletonBlock from "@/components/ui/skeleton/SkeletonBlock.vue"

withDefaults(
  defineProps<{
    title: string
    subtitle?: string
    seeAllTo?: RouteLocationRaw
    loading?: boolean
    error?: boolean
    empty?: boolean
    emptyTitle?: string
    emptyDescription?: string
    skeletonCount?: number
  }>(),
  { skeletonCount: 10 },
)
const emit = defineEmits<{ retry: [] }>()

const { t } = useI18n()
const titleId = useId()
</script>

<style scoped>
.rail__header {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 1rem;
}
.rail__title {
  margin: 0;
  font-size: 1.05rem;
  font-weight: 700;
  line-height: 1.2;
  color: var(--color-text-primary);
}
.rail__subtitle {
  display: none;
  margin: 2px 0 0;
  font-size: var(--text-caption);
  color: var(--color-text-muted);
}
.rail__see-all {
  display: flex;
  flex-shrink: 0;
  align-items: center;
  gap: 4px;
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--color-text-secondary);
  text-decoration: none;
  transition: color 0.2s;
}
.rail__see-all:hover {
  color: var(--color-brand);
}
.rail__grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
  align-items: start;
}
@media (min-width: 640px) {
  .rail__title {
    font-size: 1.25rem;
  }
  .rail__subtitle {
    display: block;
  }
}
@media (min-width: 768px) {
  .rail__grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}
@media (min-width: 1024px) {
  .rail__grid {
    grid-template-columns: repeat(4, minmax(0, 1fr));
    gap: 14px;
  }
}
@media (min-width: 1280px) {
  .rail__grid {
    grid-template-columns: repeat(5, minmax(0, 1fr));
    gap: 16px;
  }
}
</style>
