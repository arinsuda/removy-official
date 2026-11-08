<template>
  <main class="browse-page">
    <BrowseHeader :type="config.type" :query="state.q" @switch-type="switchType(config, $event)" @search="update({ q: $event })" />
    <BrowseFilterBar :state="state" :type="config.type" :genres="data.genres.value" :searching="data.searching.value" @update="update" />
    <BrowseTierChips
      :model-value="state.tier"
      :counts="data.tierCounts.value"
      :searching="autoLoading"
      @update:model-value="update({ tier: $event })"
    />
    <BrowseGrid
      :items="data.items.value"
      :loading="data.isLoading.value"
      :error="data.isError.value"
      :has-next="data.hasNextPage.value"
      :fetching-next="data.isFetchingNextPage.value"
      :capped="data.capped.value"
      :infinite="state.sort === 'default'"
      @retry="data.refetch()"
      @reset="reset"
      @load-more="data.fetchNextPage()"
    />
  </main>
</template>

<script setup lang="ts">
import { computed } from "vue"
import { provideHomePopup } from "@/composables/useHomePopup"
import { TIER_MIN_MATCHES, useBrowseData } from "@/composables/useBrowseData"
import { useBrowseState } from "@/composables/useBrowseState"
import type { BrowseConfig } from "./browseConfig"
import { DEFAULT_BROWSE_STATE } from "./browseQuery"
import BrowseFilterBar from "./BrowseFilterBar.vue"
import BrowseGrid from "./BrowseGrid.vue"
import BrowseHeader from "./BrowseHeader.vue"
import BrowseTierChips from "./BrowseTierChips.vue"

const props = defineProps<{ config: BrowseConfig }>()

provideHomePopup()
const { state, update, switchType } = useBrowseState()
const data = useBrowseData(props.config, state)

const autoLoading = computed(
  () => state.value.tier !== "all" && data.hasNextPage.value && data.items.value.length < TIER_MIN_MATCHES && !data.isLoading.value,
)

function reset() {
  void update({ ...DEFAULT_BROWSE_STATE })
}
</script>

<style scoped>
.browse-page {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  width: 100%;
  max-width: 1240px;
  margin: 0 auto;
  padding: 1rem 1rem 3rem;
}
@media (min-width: 768px) {
  .browse-page {
    padding: 2rem 1.5rem 4rem;
  }
}
</style>
