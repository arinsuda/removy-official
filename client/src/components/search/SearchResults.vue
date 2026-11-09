<template>
  <section class="sr" :aria-busy="busy">
    <SkeletonRegion v-if="busy" class="sr__grid">
      <div v-for="n in 12" :key="n" class="sr__skeleton">
        <SkeletonBlock radius="12px" aspect="2 / 3" />
        <SkeletonBlock width="80%" />
      </div>
    </SkeletonRegion>

    <AppErrorState v-else-if="failed" :title="t('media.search.error')" :description="t('media.search.errorHint')" @retry="emit('retry')" />

    <template v-else-if="isEmpty">
      <AppEmptyState
        :icon="SearchX"
        :title="filtered ? t('media.search.noMatch') : t('media.search.noResults', { query: state.q })"
        :description="filtered ? t('media.search.noMatchHint') : t('media.search.noResultsHint')"
      >
        <template #action>
          <AppButton v-if="filtered" variant="outline" size="sm" @click="emit('clear-filters')">{{ t("media.search.filters.clear") }}</AppButton>
        </template>
      </AppEmptyState>
      <SearchSuggestions v-if="!filtered" />
    </template>

    <template v-else>
      <ul v-if="state.tab === 'people'" class="sr__people">
        <li v-for="p in people" :key="p.id"><SearchPersonCard :person="p" :query="state.q" /></li>
      </ul>
      <div v-else class="sr__grid">
        <HomePosterCard
          v-for="item in items"
          :key="item.key"
          :item="item"
          rail="search"
          :sizes="SIZES"
          :highlight="state.q"
          :caption="caption(item)"
        />
      </div>

      <div v-if="hasNext" ref="sentinel" class="sr__more">
        <AppButton variant="outline" :loading="fetchingNext" @click="emit('load-more')">{{ t("browse.grid.loadMore") }}</AppButton>
        <span v-if="fetchingNext" class="sr-only" role="status">{{ t("browse.grid.loadingMore") }}</span>
      </div>
      <p v-else class="sr__end">{{ t("browse.grid.end") }}</p>
    </template>
  </section>
</template>

<script setup lang="ts">
import { computed, useTemplateRef } from "vue"
import { useI18n } from "vue-i18n"
import { SearchX } from "lucide-vue-next"
import HomePosterCard from "@/components/home/HomePosterCard.vue"
import type { HomeMedia } from "@/components/home/homeMedia"
import AppButton from "@/components/ui/AppButton.vue"
import AppEmptyState from "@/components/ui/AppEmptyState.vue"
import AppErrorState from "@/components/ui/AppErrorState.vue"
import SkeletonBlock from "@/components/ui/skeleton/SkeletonBlock.vue"
import SkeletonRegion from "@/components/ui/skeleton/SkeletonRegion.vue"
import { useInfiniteScroll } from "@/composables/useInfiniteScroll"
import type { PersonSearchResult } from "@/types"
import SearchPersonCard from "./SearchPersonCard.vue"
import SearchSuggestions from "./SearchSuggestions.vue"
import type { SearchState } from "./searchQuery"

const props = defineProps<{
  state: SearchState
  items: HomeMedia[]
  people: PersonSearchResult[]
  busy: boolean
  failed: boolean
  hasNext: boolean
  fetchingNext: boolean
}>()
const emit = defineEmits<{ retry: []; "load-more": []; "clear-filters": [] }>()

const SIZES = "(min-width: 1280px) 190px, (min-width: 1024px) 18vw, (min-width: 768px) 23vw, (min-width: 480px) 31vw, 47vw"

const { t } = useI18n()
const sentinel = useTemplateRef<HTMLElement>("sentinel")
useInfiniteScroll(sentinel, () => emit("load-more"), {
  disabled: computed(() => props.state.sort !== "default" || props.fetchingNext || !props.hasNext),
  rootMargin: "400px",
})

const isEmpty = computed(() => (props.state.tab === "people" ? !props.people.length : !props.items.length) && !props.hasNext)
const filtered = computed(() => props.state.tab !== "people" && !!(props.state.year || props.state.genre))

/** Type and year under the poster; the All tab mixes both types, so it names the type. */
function caption(item: HomeMedia): string | undefined {
  if (props.state.tab !== "all") return undefined
  const type = t(item.type === "movie" ? "navigation.movie" : "navigation.series")
  return item.date ? `${type} · ${item.date.slice(0, 4)}` : type
}
</script>

<style scoped>
.sr__grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 1.25rem 0.75rem;
}
.sr__skeleton {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}
.sr__people {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.75rem;
  margin: 0;
  padding: 0;
  list-style: none;
}
.sr__more {
  display: flex;
  justify-content: center;
  padding: 2rem 0;
}
.sr__end {
  margin: 0;
  padding: 2rem 0;
  text-align: center;
  font-size: var(--text-small);
  color: var(--color-text-muted);
}
.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  overflow: hidden;
  clip-path: inset(50%);
  white-space: nowrap;
}
@media (min-width: 480px) {
  .sr__grid,
  .sr__people {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}
@media (min-width: 768px) {
  .sr__grid {
    grid-template-columns: repeat(4, minmax(0, 1fr));
    gap: 1.5rem 1rem;
  }
  .sr__people {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }
}
@media (min-width: 1024px) {
  .sr__grid {
    grid-template-columns: repeat(5, minmax(0, 1fr));
  }
  .sr__people {
    grid-template-columns: repeat(6, minmax(0, 1fr));
  }
}
@media (min-width: 1280px) {
  .sr__grid {
    grid-template-columns: repeat(6, minmax(0, 1fr));
  }
}
</style>
