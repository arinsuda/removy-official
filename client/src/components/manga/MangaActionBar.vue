<template>
  <div class="manga-actions" role="group" :aria-label="t('manga.detail.actions.label')">
    <AppButton variant="primary" size="sm" data-test="manga-action-rate" @click="onRate">
      <Star :size="16" :fill="hasOwnReview ? 'currentColor' : 'none'" aria-hidden="true" />
      <span>{{ t(hasOwnReview ? "manga.detail.actions.editRating" : "manga.detail.actions.rate") }}</span>
    </AppButton>
    <AppButton
      variant="outline"
      size="sm"
      data-test="manga-action-want"
      :aria-pressed="activity.isWatchlisted.value"
      :loading="activity.pending.value.has('watchlist')"
      @click="onToggle(activity.toggleWatchlist)"
    >
      <Bookmark :size="16" :fill="activity.isWatchlisted.value ? 'currentColor' : 'none'" aria-hidden="true" />
      <span>{{ t("manga.detail.actions.wantToRead") }}</span>
    </AppButton>
    <AppButton
      variant="outline"
      size="sm"
      data-test="manga-action-read"
      :aria-pressed="activity.isWatched.value"
      :loading="activity.pending.value.has('watched')"
      @click="onToggle(activity.toggleWatched)"
    >
      <CheckCircle2 :size="16" :fill="activity.isWatched.value ? 'currentColor' : 'none'" aria-hidden="true" />
      <span>{{ t("manga.detail.actions.read") }}</span>
    </AppButton>
    <AppButton
      variant="outline"
      size="sm"
      data-test="manga-action-like"
      :aria-pressed="activity.isLiked.value"
      :loading="activity.pending.value.has('like')"
      @click="onToggle(activity.toggleLike)"
    >
      <Heart :size="16" :fill="activity.isLiked.value ? 'currentColor' : 'none'" aria-hidden="true" />
      <span>{{ t(activity.isLiked.value ? "manga.detail.actions.unlike" : "manga.detail.actions.like") }}</span>
    </AppButton>
    <AppButton variant="outline" size="sm" data-test="manga-action-list" @click="onAddToList">
      <ListPlus :size="16" aria-hidden="true" />
      <span>{{ t("manga.detail.actions.addToList") }}</span>
    </AppButton>
    <AppButton variant="outline" size="sm" data-test="manga-action-bmol" @click="onAddToBmol">
      <Trophy :size="16" aria-hidden="true" />
      <span>{{ t("manga.detail.actions.addToBmol") }}</span>
    </AppButton>

    <AddToListSheet v-if="showLists" v-model="showLists" :media-id="manga.id" media-type="manga" />
    <MangaBmolSheet v-if="showBmol && userId" v-model="showBmol" :user-id="userId" :manga="manga" />
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from "vue"
import { useI18n } from "vue-i18n"
import { Bookmark, CheckCircle2, Heart, ListPlus, Star, Trophy } from "lucide-vue-next"
import AddToListSheet from "@/components/lists/AddToListSheet.vue"
import AppButton from "@/components/ui/AppButton.vue"
import MangaBmolSheet from "@/components/manga/MangaBmolSheet.vue"
import type { useMediaActivity } from "@/composables/useMediaActivity"
import type { MangaItem } from "@/types/manga"

const props = defineProps<{
  manga: MangaItem
  /** Viewer's id, or null for guests (the BMOL sheet needs an owner). */
  userId: number | null
  hasOwnReview: boolean
  activity: ReturnType<typeof useMediaActivity>
}>()

const emit = defineEmits<{ rate: [] }>()
const { t } = useI18n()
const showLists = ref(false)
const showBmol = ref(false)

// Sheets belong to one title; never let them stay open across a switch.
watch(() => props.manga.id, () => {
  showLists.value = false
  showBmol.value = false
})

/** The activity toggles check sign-in themselves and open the auth prompt for guests. */
function onToggle(toggle: () => Promise<void>) {
  void toggle()
}

function onRate() {
  if (props.activity.requireLogin("rate")) emit("rate")
}

function onAddToList() {
  if (props.activity.requireLogin("list")) showLists.value = true
}

function onAddToBmol() {
  if (props.activity.requireLogin("bmol")) showBmol.value = true
}
</script>

<style scoped>
.manga-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}
</style>
