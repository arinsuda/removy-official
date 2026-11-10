<template>
  <main class="search-page">
    <header class="search-head">
      <h1 class="search-head__title">
        <Search :size="22" aria-hidden="true" />
        <span v-if="state.q">{{ t("media.search.title", { query: state.q }) }}</span>
        <span v-else>{{ t("media.search.promptTitle") }}</span>
      </h1>
      <p class="search-head__count" role="status" aria-live="polite">
        <template v-if="state.q && !data.busy.value && !data.failed.value && !(data.countsApproximate.value && state.tab !== 'people')">
          {{ t("media.search.count", { n: data.counts.value[state.tab].toLocaleString(locale) }, data.counts.value[state.tab]) }}
        </template>
      </p>
    </header>

    <AppEmptyState v-if="!state.q" :icon="Search" :title="t('media.search.promptTitle')" :description="t('media.search.promptHint')" />

    <template v-else>
      <SearchFilterBar :state="state" :counts="data.counts.value" :counts-approximate="data.countsApproximate.value" :years="data.years.value" :genres="data.genres.value" @update="update" />
      <SearchResults
        :state="state"
        :items="data.items.value"
        :people="data.peopleItems.value"
        :busy="data.busy.value"
        :failed="data.failed.value"
        :has-next="data.hasNext.value"
        :fetching-next="data.fetchingNext.value"
        @retry="data.retry"
        @load-more="data.fetchNext"
        @clear-filters="update({ year: null, genre: null })"
      />
    </template>
  </main>
</template>

<script setup lang="ts">
import { computed } from "vue"
import { useRoute, useRouter } from "vue-router"
import { useI18n } from "vue-i18n"
import { Search } from "lucide-vue-next"
import SearchFilterBar from "@/components/search/SearchFilterBar.vue"
import SearchResults from "@/components/search/SearchResults.vue"
import { mergeSearchQuery, parseSearchQuery, type SearchState } from "@/components/search/searchQuery"
import AppEmptyState from "@/components/ui/AppEmptyState.vue"
import { useSearchResults } from "@/composables/useSearchResults"
import { provideHomePopup } from "@/composables/useHomePopup"

const { t, locale } = useI18n()
const route = useRoute()
const router = useRouter()

provideHomePopup()
// The URL is the single source of truth for q, tab, sort, year and genre.
const state = computed<SearchState>(() => parseSearchQuery(route.query))
const data = useSearchResults(state)

// Carry route.hash: OAuth and handoff tokens can arrive in the fragment.
function update(patch: Partial<SearchState>) {
  return router.replace({ query: mergeSearchQuery(route.query, { ...state.value, ...patch }), hash: route.hash })
}
</script>

<style scoped>
.search-page {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  width: 100%;
  max-width: 1240px;
  margin: 0 auto;
  padding: 1rem 1rem 3rem;
}
.search-head__title {
  display: flex;
  align-items: center;
  gap: 0.625rem;
  margin: 0;
  font-size: var(--text-heading);
  font-weight: 700;
  line-height: 1.3;
  color: var(--color-text-primary);
  overflow-wrap: anywhere;
}
.search-head__count {
  min-height: 1.25rem;
  margin: 0.25rem 0 0;
  font-size: var(--text-small);
  color: var(--color-text-muted);
}
@media (min-width: 768px) {
  .search-page {
    padding: 2rem 1.5rem 4rem;
  }
  .search-head__title {
    font-size: var(--text-page);
  }
}
</style>
