<template>
  <main class="manga-page">
    <h1 class="manga-title">{{ t("manga.title") }}</h1>
    <MangaFilterBar
      :state="state"
      :genres="data.genres.value"
      :adult-allowed="eligible"
      :acknowledged="acknowledged"
      @update="update"
      @request-adult="adultConfirmOpen = true"
    />
    <MangaGrid
      :items="data.loaded.value"
      :loading="data.isLoading.value"
      :error="data.isError.value"
      :has-next="data.hasNextPage.value"
      :fetching-next="data.isFetchingNextPage.value"
      :capped="data.capped.value"
      @retry="data.refetch()"
      @reset="reset"
      @load-more="data.fetchNextPage()"
    />
    <MangaAttribution />
    <MangaAdultConfirmModal v-model="adultConfirmOpen" @confirm="confirmAdult" />
  </main>
</template>

<script setup lang="ts">
import { computed, ref } from "vue"
import { useI18n } from "vue-i18n"
import MangaAdultConfirmModal from "@/components/manga/MangaAdultConfirmModal.vue"
import MangaAttribution from "@/components/manga/MangaAttribution.vue"
import MangaFilterBar from "@/components/manga/MangaFilterBar.vue"
import MangaGrid from "@/components/manga/MangaGrid.vue"
import { DEFAULT_MANGA_STATE } from "@/components/manga/mangaQuery"
import { useMangaAdult } from "@/composables/useMangaAdult"
import { useMangaBrowse } from "@/composables/useMangaBrowse"
import { useMangaRouteState } from "@/composables/useMangaRouteState"

const { t } = useI18n()
const { state, update } = useMangaRouteState()
const { eligible, acknowledged, canShowAdult, acknowledge } = useMangaAdult()
/** adult=1 in the URL is not enough on its own: the viewer must be eligible and have confirmed. */
const data = useMangaBrowse(state, computed(() => state.value.adult && canShowAdult.value))
const adultConfirmOpen = ref(false)

/** Acknowledging first keeps the URL change from ever reaching the API unconfirmed. */
function confirmAdult() {
  acknowledge()
  void update({ adult: true })
}

/** Clears every filter. */
function reset() {
  void update({ ...DEFAULT_MANGA_STATE })
}
</script>

<style scoped>
.manga-page {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  width: 100%;
  max-width: 1240px;
  margin: 0 auto;
  padding: 1rem 1rem 3rem;
}
.manga-title {
  margin: 0;
  font-size: var(--text-title);
  font-weight: 700;
  color: var(--color-text-primary);
}
@media (min-width: 768px) {
  .manga-page {
    padding: 2rem 1.5rem 4rem;
  }
}
</style>
