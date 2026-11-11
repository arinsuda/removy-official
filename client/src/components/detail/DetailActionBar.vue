<template>
  <div class="bar" role="group" :aria-label="t('media.detail.actions.label')">
    <button type="button" class="bar__btn bar__btn--primary" data-test="action-rate" @click="onRate">
      <Star :size="20" :fill="hasOwnReview ? 'currentColor' : 'none'" aria-hidden="true" />
      <span>{{ t(hasOwnReview ? "media.detail.actions.editRating" : "media.detail.actions.rate") }}</span>
    </button>
    <button type="button" class="bar__btn" data-test="action-log" @click="onLog">
      <CalendarCheck :size="20" aria-hidden="true" />
      <span>{{ t("media.detail.actions.log") }}</span>
    </button>
    <button
      type="button"
      class="bar__btn"
      :class="{ 'bar__btn--on': watchlisted }"
      data-test="action-watchlist"
      :aria-pressed="watchlisted"
      :disabled="watchlistPending"
      @click="emit('watchlist')"
    >
      <Bookmark :size="20" :fill="watchlisted ? 'currentColor' : 'none'" aria-hidden="true" />
      <span>{{ t(watchlisted ? "media.detail.actions.inWatchlist" : "media.detail.actions.watchlist") }}</span>
    </button>
    <button type="button" class="bar__btn" data-test="action-list" @click="onAddToList">
      <ListPlus :size="20" aria-hidden="true" />
      <span>{{ t("library.customLists.sheet.action") }}</span>
    </button>
    <button type="button" class="bar__btn" data-test="action-share" @click="emit('share')">
      <Share2 :size="20" aria-hidden="true" />
      <span>{{ t("share.share") }}</span>
    </button>

    <AddToListSheet v-if="showLists" v-model="showLists" :media-id="mediaId" :media-type="mediaType" />
    <WatchLogModal v-if="showLog" :media-id="mediaId" :media-type="mediaType" @close="showLog = false" @logged="emit('logged')" />
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from "vue"
import { useI18n } from "vue-i18n"
import { Bookmark, CalendarCheck, ListPlus, Share2, Star } from "lucide-vue-next"
import AddToListSheet from "@/components/lists/AddToListSheet.vue"
import WatchLogModal from "@/components/movie/WatchLogModal.vue"
import type { AuthGateReason } from "@/composables/useAuthGate"

const props = defineProps<{
  mediaId: number
  mediaType: "movie" | "tv"
  watchlisted: boolean
  watchlistPending?: boolean
  hasOwnReview: boolean
  /** Returns true when signed in; otherwise opens the auth prompt for `reason`. */
  requireLogin: (reason?: AuthGateReason) => boolean
}>()

const emit = defineEmits<{
  rate: []
  watchlist: []
  share: []
  logged: []
}>()

const { t } = useI18n()
const showLog = ref(false)
const showLists = ref(false)

// A log must never be saved against a different title than the one it was opened for.
watch(() => [props.mediaId, props.mediaType], () => {
  showLog.value = false
  showLists.value = false
})

function onRate() {
  if (props.requireLogin("rate")) emit("rate")
}

function onAddToList() {
  if (props.requireLogin("list")) showLists.value = true
}

function onLog() {
  if (props.requireLogin("watchLog")) showLog.value = true
}
</script>

<style scoped>
.bar {
  position: fixed;
  inset: auto 0 0;
  z-index: var(--z-below-header);
  display: grid;
  grid-template-columns: repeat(5, minmax(0, 1fr));
  gap: 0.25rem;
  padding: 0.5rem max(0.75rem, env(safe-area-inset-right)) calc(0.5rem + env(safe-area-inset-bottom)) max(0.75rem, env(safe-area-inset-left));
  background: color-mix(in srgb, var(--color-surface) 92%, transparent);
  backdrop-filter: blur(12px);
  border-top: 1px solid var(--color-border);
}
.bar__btn {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.2rem;
  min-height: 48px;
  padding: 0.3rem 0.25rem;
  font-size: var(--text-caption);
  font-weight: 600;
  line-height: 1.2;
  color: var(--color-text-secondary);
  background: transparent;
  border: 0;
  border-radius: var(--radius-card);
  cursor: pointer;
  transition: color var(--motion-fast) var(--ease-water), background-color var(--motion-fast) var(--ease-water), transform var(--motion-fast) var(--ease-out-soft);
}
.bar__btn--primary { color: var(--color-brand); }
.bar__btn--on { color: var(--color-accent); }
.bar__btn:disabled { opacity: 0.6; cursor: progress; }
.bar__btn:active:not(:disabled) { transform: scale(0.95); }

@media (hover: hover) {
  .bar__btn:hover:not(:disabled) { color: var(--color-text-primary); background: var(--color-surface-3); }
}

@media (min-width: 768px) {
  .bar {
    position: static;
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
    padding: 0;
    background: none;
    backdrop-filter: none;
    border: 0;
  }
  .bar__btn {
    flex-direction: row;
    gap: 0.5rem;
    min-height: 44px;
    padding: 0 1.1rem;
    font-size: var(--text-small);
    color: var(--color-text-primary);
    background: var(--color-surface-2);
    border: 1px solid var(--color-border-strong);
    border-radius: var(--radius-btn);
  }
  .bar__btn--primary { color: var(--color-on-brand); background: var(--color-brand); border-color: var(--color-brand); }
  .bar__btn--on { color: var(--color-accent); border-color: var(--color-accent); }
}
@media (min-width: 768px) and (hover: hover) {
  .bar__btn--primary:hover:not(:disabled) { color: var(--color-on-brand); background: var(--color-brand-hover); }
}
@media (prefers-reduced-motion: reduce) {
  .bar__btn { transition: none; }
}
</style>
