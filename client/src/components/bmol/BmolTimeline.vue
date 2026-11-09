<template>
  <div class="tl">
    <section v-for="group in groups" :key="group.year" class="tl__year" :aria-labelledby="`${uid}-${group.year}`">
      <h3 :id="`${uid}-${group.year}`" class="tl__heading">
        {{ group.year === UNKNOWN_YEAR ? t("library.bmol.timeline.unknown") : group.year }}
        <span class="tl__count">{{ t("library.bmol.timeline.count", { count: group.items.length }) }}</span>
      </h3>
      <div class="tl__grid">
        <slot :items="group.items" />
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, useId } from "vue"
import { useI18n } from "vue-i18n"
import type { BMOLItemResponse } from "@/types"
import { groupByReleaseYear, UNKNOWN_YEAR } from "./bmolTimeline"

const props = defineProps<{ items: BMOLItemResponse[] }>()
const { t } = useI18n()
const uid = useId()
const groups = computed(() => groupByReleaseYear(props.items))
</script>

<style scoped>
.tl { display: flex; flex-direction: column; gap: 1.75rem; }
.tl__heading { display: flex; align-items: baseline; gap: 0.625rem; padding-bottom: 0.5rem; margin-bottom: 0.875rem; font-size: var(--text-title); font-weight: 800; color: var(--color-text-primary); border-bottom: 1px solid var(--color-border); }
.tl__count { font-size: var(--text-caption); font-weight: 500; color: var(--color-text-muted); }
.tl__grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(5.5rem, 1fr)); gap: 0.875rem 0.625rem; }
@media (min-width: 768px) {
  .tl__grid { grid-template-columns: repeat(auto-fill, minmax(7rem, 1fr)); gap: 1.125rem 0.875rem; }
}
</style>
