<template>
  <div class="shelf">
    <div class="shelf__grid">
      <div
        v-for="group in visibleGroups"
        :key="group.rank"
        class="shelf__group"
        :class="{ 'is-target': targetRank === group.rank }"
        :data-bmol-rank="group.rank"
      >
        <slot :items="group.items" />
      </div>
    </div>
    <AppButton v-if="hiddenCount > 0 || expanded" variant="outline" size="sm" class="shelf__toggle" :aria-expanded="expanded" @click="expanded = !expanded">
      {{ expanded ? t("library.bmol.shelf.showLess") : t("library.bmol.shelf.showAll", { count: hiddenCount }) }}
    </AppButton>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from "vue"
import { useI18n } from "vue-i18n"
import AppButton from "@/components/ui/AppButton.vue"
import { SHELF_VISIBLE_RANK, type RankGroup } from "./bmolRanks"

const props = defineProps<{ groups: RankGroup[]; targetRank?: number | null }>()
const { t } = useI18n()
const expanded = ref(false)

const hidden = computed(() => props.groups.filter((g) => g.rank > SHELF_VISIBLE_RANK || g.rank < 1))
const hiddenCount = computed(() => hidden.value.reduce((n, g) => n + g.items.length, 0))
const visibleGroups = computed(() => (expanded.value ? props.groups : props.groups.filter((g) => g.rank >= 1 && g.rank <= SHELF_VISIBLE_RANK)))
</script>

<style scoped>
.shelf { display: flex; flex-direction: column; gap: 1rem; }
/* One compact grid; each rank group is transparent to layout so cards flow inline, tied cards sharing a number. */
.shelf__grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(5.5rem, 1fr)); gap: 0.875rem 0.625rem; }
.shelf__group { display: contents; }
.shelf__group.is-target > :deep(*) { outline: 2px dashed var(--color-accent); outline-offset: 2px; border-radius: var(--radius-card); }
.shelf__toggle { align-self: center; }
@media (min-width: 768px) {
  .shelf__grid { grid-template-columns: repeat(auto-fill, minmax(7rem, 1fr)); gap: 1.125rem 0.875rem; }
}
</style>
