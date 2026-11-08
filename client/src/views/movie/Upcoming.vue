<template>
  <main class="upcoming-page">
    <UpcomingHeader :genre="genre" :genres="data.genres.value" @update:genre="setGenre" />
    <UpcomingHero v-if="!data.isError.value" :items="data.items.value" :today="today" />
    <UpcomingTimeline
      :items="data.items.value"
      :today="today"
      :loading="data.isLoading.value"
      :error="data.isError.value"
      :has-next="data.hasNextPage.value"
      :fetching-next="data.isFetchingNextPage.value"
      :capped="data.capped.value"
      @retry="data.refetch()"
      @load-more="data.fetchNextPage()"
    />
  </main>
</template>

<script setup lang="ts">
import { computed } from "vue"
import { useRoute, useRouter } from "vue-router"
import UpcomingHeader from "@/components/upcoming/UpcomingHeader.vue"
import UpcomingHero from "@/components/upcoming/UpcomingHero.vue"
import UpcomingTimeline from "@/components/upcoming/UpcomingTimeline.vue"
import { startOfDay } from "@/components/upcoming/upcomingGroups"
import { provideHomePopup } from "@/composables/useHomePopup"
import { useUpcomingData } from "@/composables/useUpcomingData"

provideHomePopup()
const route = useRoute()
const router = useRouter()
const today = startOfDay(new Date())

// The URL owns the genre filter; ?genre= must be a positive integer.
const genre = computed(() => {
  const raw = Array.isArray(route.query.genre) ? route.query.genre[0] : route.query.genre
  const n = Number(raw)
  return Number.isInteger(n) && n > 0 ? n : null
})
const data = useUpcomingData(genre)

// Carry route.hash: OAuth and handoff tokens arrive in the fragment.
function setGenre(next: number | null) {
  const { genre: _drop, ...rest } = route.query
  void router.replace({ query: next ? { ...rest, genre: String(next) } : rest, hash: route.hash })
}
</script>

<style scoped>
.upcoming-page {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  width: 100%;
  max-width: 1240px;
  margin: 0 auto;
  padding: 1rem 1rem 3rem;
}
@media (min-width: 768px) {
  .upcoming-page {
    padding: 2rem 1.5rem 4rem;
  }
}
</style>
