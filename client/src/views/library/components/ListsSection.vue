<template>
  <section>
    <ListDetail
      v-if="activeList"
      :list-key="activeList"
      :user-id="userId"
      :is-owner="isOwner"
      :reload-key="reloadKey"
      @back="emit('open', null)"
      @add="emit('add')"
    />
    <ul v-else class="ls" :aria-label="t('library.tabs.lists')">
      <li v-for="c in cards" :key="c.key" class="ls__item">
        <ListCard :list-key="c.key" :to="listTo(c.key)" :count="c.count" :posters="c.posters" :updated-at="c.updatedAt" />
      </li>
    </ul>
  </section>
</template>

<script setup lang="ts">
import { computed } from "vue"
import { useI18n } from "vue-i18n"
import type { RouteLocationRaw } from "vue-router"
import type { LibraryListKey, ReviewResponse } from "@/types"
import { COLLAGE_SIZE, LIBRARY_LIST_KEYS } from "../config"
import { collagePosters, latestAt, newestFirst, type LibraryEntry } from "../libraryStats"
import ListCard from "./ListCard.vue"
import ListDetail from "./ListDetail.vue"

const props = defineProps<{
  activeList: LibraryListKey | null
  userId: number
  isOwner: boolean
  watchlist: LibraryEntry[]
  watched: LibraryEntry[]
  likes: LibraryEntry[]
  reviews: ReviewResponse[]
  reloadKey: number
  listTo: (key: LibraryListKey) => RouteLocationRaw
}>()
const emit = defineEmits<{ open: [key: LibraryListKey | null]; add: [] }>()

const { t } = useI18n()

const cards = computed(() => {
  const byKey: Record<LibraryListKey, { at: string; media: { poster_path: string } }[]> = {
    watchlist: props.watchlist,
    watched: props.watched,
    likes: props.likes,
    reviews: props.reviews.map((r) => ({ at: r.created_at, media: { poster_path: r.media.poster_url } })),
  }
  return LIBRARY_LIST_KEYS.map((key) => {
    const entries = byKey[key]
    return { key, count: entries.length, posters: collagePosters(newestFirst(entries), COLLAGE_SIZE), updatedAt: latestAt(entries) }
  })
})
</script>

<style scoped>
.ls { display: grid; grid-template-columns: minmax(0, 1fr); gap: 1rem; margin: 0; padding: 0; list-style: none; }
@media (min-width: 640px) {
  .ls { grid-template-columns: repeat(2, minmax(0, 1fr)); }
}
@media (min-width: 1280px) {
  .ls { grid-template-columns: repeat(4, minmax(0, 1fr)); }
}
</style>
