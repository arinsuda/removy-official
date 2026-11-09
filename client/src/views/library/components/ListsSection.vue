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
    <CustomListDetail v-else-if="customListId" :list-id="customListId" :current-user-id="currentUserId" @back="emit('openCustom', null)" />
    <template v-else>
      <ul class="ls" :aria-label="t('library.tabs.lists')">
        <li v-for="c in cards" :key="c.key" class="ls__item">
          <ListCard :list-key="c.key" :to="listTo(c.key)" :count="c.count" :posters="c.posters" :updated-at="c.updatedAt" />
        </li>
      </ul>

      <section class="custom" :aria-labelledby="headingId">
        <header class="custom__head">
          <h2 :id="headingId" class="custom__title">{{ t("library.customLists.title") }}</h2>
          <AppButton v-if="isOwner" size="sm" data-action="new-list" :disabled="atLimit" @click="showEditor = true">
            <Plus :size="16" aria-hidden="true" />{{ t("library.customLists.new") }}
          </AppButton>
        </header>
        <p v-if="isOwner && atLimit" class="custom__note">{{ t("library.customLists.limit", { max: MAX_CUSTOM_LISTS }) }}</p>

        <SkeletonGrid v-if="custom.isLoading.value" :count="4" min-width="240px" />
        <AppErrorState v-else-if="custom.isError.value" :title="t('library.customLists.loadFailed')" @retry="custom.refetch()" />
        <p v-else-if="custom.forbidden.value" class="custom__note" role="status"><Lock :size="14" aria-hidden="true" />{{ t("library.customLists.locked") }}</p>
        <AppEmptyState
          v-else-if="!custom.lists.value.length"
          :icon="ListIcon"
          :title="t(isOwner ? 'library.customLists.emptyOwnerTitle' : 'library.customLists.emptyVisitorTitle')"
          :description="t(isOwner ? 'library.customLists.emptyOwnerBody' : 'library.customLists.emptyVisitorBody')"
        />
        <ul v-else class="ls" :aria-label="t('library.customLists.title')">
          <li v-for="l in custom.lists.value" :key="l.id" class="ls__item">
            <CustomListCard :list="l" :to="customListTo(l.id)" />
          </li>
        </ul>
      </section>
    </template>

    <ListEditorModal v-if="isOwner" v-model="showEditor" @saved="onCreated" />
  </section>
</template>

<script setup lang="ts">
import { computed, ref, useId } from "vue"
import { useI18n } from "vue-i18n"
import type { RouteLocationRaw } from "vue-router"
import { List as ListIcon, Lock, Plus } from "lucide-vue-next"
import AppButton from "@/components/ui/AppButton.vue"
import AppEmptyState from "@/components/ui/AppEmptyState.vue"
import AppErrorState from "@/components/ui/AppErrorState.vue"
import SkeletonGrid from "@/components/ui/skeleton/SkeletonGrid.vue"
import { useAuthStore } from "@/stores/auth"
import type { LibraryListKey, ListSummary, ReviewResponse } from "@/types"
import { COLLAGE_SIZE, LIBRARY_LIST_KEYS, MAX_CUSTOM_LISTS } from "../config"
import { collagePosters, latestAt, newestFirst, type LibraryEntry } from "../libraryStats"
import { useCustomLists } from "../useCustomLists"
import CustomListCard from "./CustomListCard.vue"
import CustomListDetail from "./CustomListDetail.vue"
import ListCard from "./ListCard.vue"
import ListDetail from "./ListDetail.vue"
import ListEditorModal from "./ListEditorModal.vue"

const props = defineProps<{
  activeList: LibraryListKey | null
  customListId: number | null
  userId: number
  isOwner: boolean
  watchlist: LibraryEntry[]
  watched: LibraryEntry[]
  likes: LibraryEntry[]
  reviews: ReviewResponse[]
  reloadKey: number
  listTo: (key: LibraryListKey) => RouteLocationRaw
  customListTo: (id: number) => RouteLocationRaw
}>()
const emit = defineEmits<{ open: [key: LibraryListKey | null]; openCustom: [id: number | null]; add: [] }>()

const { t } = useI18n()
const auth = useAuthStore()
const headingId = useId()
const currentUserId = computed(() => auth.user?.id ?? null)
const custom = useCustomLists(() => props.userId, () => props.isOwner)
const atLimit = computed(() => custom.lists.value.length >= MAX_CUSTOM_LISTS)
const showEditor = ref(false)

function onCreated(list: ListSummary) {
  emit("openCustom", list.id)
}

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
.custom { margin-top: 2rem; }
.custom__head { display: flex; align-items: center; justify-content: space-between; gap: 0.75rem; margin-bottom: 1rem; }
.custom__title { margin: 0; font-size: var(--text-title); font-weight: 700; color: var(--color-text-primary); }
.custom__note { display: flex; align-items: center; gap: 0.375rem; margin: 0 0 1rem; font-size: var(--text-small); color: var(--color-text-secondary); }
@media (min-width: 640px) {
  .ls { grid-template-columns: repeat(2, minmax(0, 1fr)); }
}
@media (min-width: 1280px) {
  .ls { grid-template-columns: repeat(4, minmax(0, 1fr)); }
}
</style>
