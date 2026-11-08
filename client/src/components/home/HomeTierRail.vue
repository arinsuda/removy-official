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
    <HomePosterCard v-for="item in shown" :key="item.key" :item="item" rail="popcorn" />
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
import { countByTier, filterByTier } from "./popcornHome"

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
const shown = computed(() => filterByTier(props.items, tier.value))

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
