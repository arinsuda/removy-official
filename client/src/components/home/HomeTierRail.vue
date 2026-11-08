<template>
  <HomeRail
    :title="t('home.popcorn.title')"
    :subtitle="t('home.popcorn.subtitle')"
    :loading="loading"
    :error="error"
    :empty="!shown.length"
    :empty-title="t('home.popcorn.empty')"
    @retry="emit('retry')"
  >
    <template #toolbar>
      <div class="tier-toolbar">
        <AppTabs v-model="tier" :tabs="tabs" :controls="false" :label="t('home.popcorn.tabs')" />
        <HomeGenreChips />
      </div>
    </template>
    <template v-for="(item, i) in shown" :key="item.key">
      <div v-if="i === backfillAt" class="tier-divider" role="separator">
        <span>{{ t("home.popcorn.backfill") }}</span>
      </div>
      <HomePosterCard :item="item" rail="popcorn" />
    </template>
  </HomeRail>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue"
import { useI18n } from "vue-i18n"
import AppTabs from "@/components/ui/AppTabs.vue"
import HomeGenreChips from "./HomeGenreChips.vue"
import HomePosterCard from "./HomePosterCard.vue"
import HomeRail from "./HomeRail.vue"
import type { HomeMedia } from "./homeMedia"
import { countByTier, tierWithBackfill } from "./popcornHome"

type TabTier = "hot" | "fresh" | "stale"
const TIERS: TabTier[] = ["hot", "fresh", "stale"]

const props = defineProps<{ items: HomeMedia[]; loading?: boolean; error?: boolean }>()
const emit = defineEmits<{ retry: [] }>()
const { t } = useI18n()

const tier = ref<TabTier>("hot")
const counts = computed(() => countByTier(props.items))
const tabs = computed(() =>
  TIERS.map((key) => ({ key, label: t(`home.popcorn.tiers.${key}`), disabled: counts.value[key] === 0 })),
)
const filled = computed(() => tierWithBackfill(props.items, tier.value))
const shown = computed(() => filled.value.items)
const backfillAt = computed(() => filled.value.backfillAt)

// Land on the first tier that has titles once data arrives.
watch(
  counts,
  (c) => {
    if (c[tier.value] === 0) tier.value = TIERS.find((k) => c[k] > 0) ?? "hot"
  },
  { immediate: true },
)
</script>

<style scoped>
.tier-divider.tier-divider {
  flex: 0 0 112px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  align-self: stretch;
  gap: 8px;
  padding: 0 12px;
  border-left: 1px dashed var(--color-border-strong);
  color: var(--color-text-muted);
  font-size: var(--text-caption);
  line-height: 1.4;
  text-align: center;
}
.tier-toolbar {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin-bottom: 1rem;
}
.tier-toolbar :deep(.app-tabs) {
  overflow-x: auto;
  scrollbar-width: none;
}
@media (min-width: 768px) {
  .tier-toolbar {
    flex-direction: row;
    align-items: center;
    justify-content: space-between;
  }
  .tier-toolbar :deep(.genre-chips) {
    justify-content: flex-end;
  }
}
</style>
