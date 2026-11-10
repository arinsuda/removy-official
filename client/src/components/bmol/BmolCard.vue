<template>
  <article
    class="bcard"
    :class="[`bcard--${variant}`, `bcard--${tier}`, { 'is-grabbed': grabbed, 'is-dragging': dragging, 'is-pending': pending }]"
    :style="dragStyle"
    :data-item-id="item.id"
  >
    <div class="bcard__poster">
      <RouterLink :to="detailPath" class="bcard__poster-link" :aria-label="title" :title="variant === 'shelf' && item.note ? item.note : undefined" tabindex="-1">
        <img v-if="posterSrc" :src="posterSrc" :alt="title" class="bcard__img" loading="lazy" />
        <div v-else class="bcard__img bcard__img--empty" aria-hidden="true"><Film :size="28" /></div>
      </RouterLink>
      <span class="bcard__rank" :aria-label="t('library.bmol.drag.rank', { rank: item.rank })">{{ item.rank }}</span>

      <template v-if="editing && !pending">
        <button
          type="button"
          class="bcard__handle" data-touch-ok
          :class="{ 'is-active': grabbed }"
          :aria-label="t('library.bmol.drag.handle', { title })"
          :aria-pressed="grabbed"
          @pointerdown="reorder?.onHandlePointerDown($event, item.id, item.rank)"
          @pointermove="reorder?.onHandlePointerMove($event)"
          @pointerup="reorder?.onHandlePointerUp($event)"
          @pointercancel="reorder?.cancel()"
          @keydown="reorder?.onHandleKeydown($event, item.id, item.rank)"
          @blur="grabbed && !dragging && reorder?.cancel()"
        >
          <GripVertical :size="16" aria-hidden="true" />
        </button>
        <div class="bcard__menu" @focusout="onMenuFocusOut">
          <button
            type="button"
            class="bcard__menu-btn" data-touch-ok
            aria-haspopup="menu"
            :aria-expanded="menuOpen"
            :aria-label="t('library.bmol.card.actions', { title })"
            @click="menuOpen = !menuOpen"
            @keydown.esc="menuOpen = false"
          >
            <MoreHorizontal :size="16" aria-hidden="true" />
          </button>
          <ul v-if="menuOpen" class="bcard__menu-list" role="menu" @keydown.esc="menuOpen = false">
            <li role="none"><button type="button" role="menuitem" @click="act('edit-note')"><PenLine :size="14" aria-hidden="true" />{{ t("library.bmol.card.editNote") }}</button></li>
            <li role="none"><button type="button" role="menuitem" :disabled="!canMoveUp" @click="act('move-up')"><ArrowUp :size="14" aria-hidden="true" />{{ t("library.bmol.card.moveUp") }}</button></li>
            <li role="none"><button type="button" role="menuitem" :disabled="!canMoveDown" @click="act('move-down')"><ArrowDown :size="14" aria-hidden="true" />{{ t("library.bmol.card.moveDown") }}</button></li>
            <li role="none"><button type="button" role="menuitem" class="is-danger" @click="act('remove')"><Trash2 :size="14" aria-hidden="true" />{{ t("library.bmol.card.remove") }}</button></li>
          </ul>
        </div>
      </template>
    </div>

    <div class="bcard__body">
      <RouterLink :to="detailPath" class="bcard__title">{{ title }}</RouterLink>
      <MediaRating v-if="variant !== 'shelf'" class="bcard__rating" :ratings="item.media.ratings" :vote-average="item.media.vote_average" />
      <p v-if="item.note && variant !== 'shelf'" class="bcard__note">{{ item.note }}</p>
      <button v-else-if="editing && variant !== 'shelf'" type="button" class="bcard__add-note" @click="emit('edit-note')">
        {{ t("library.bmol.card.addNote") }}
      </button>
    </div>
  </article>
</template>

<script setup lang="ts">
import { computed, ref } from "vue"
import { useI18n } from "vue-i18n"
import { ArrowDown, ArrowUp, Film, GripVertical, MoreHorizontal, PenLine, Trash2 } from "lucide-vue-next"
import MediaRating from "@/components/movie/MediaRating.vue"
import { getTmdbImageUrl } from "@/utils/image"
import type { BMOLItemResponse } from "@/types"
import { tierOf } from "./bmolRanks"
import { useBmolReorderState } from "./useBmolReorder"

const props = defineProps<{
  item: BMOLItemResponse
  variant: "hero" | "podium" | "shelf"
  /** Owner in edit mode: shows the drag handle and the action menu. */
  editing?: boolean
  canMoveUp?: boolean
  canMoveDown?: boolean
}>()
const emit = defineEmits<{ "edit-note": []; "move-up": []; "move-down": []; remove: [] }>()

const { t } = useI18n()
const reorder = useBmolReorderState()
const menuOpen = ref(false)

const title = computed(() => props.item.media.title || props.item.media.name || "")
const tier = computed(() => tierOf(props.item.rank))
const posterSrc = computed(() => getTmdbImageUrl(props.item.media.poster_url, props.variant === "hero" ? "w500" : "w342"))
const detailPath = computed(() => `/${props.item.media_type === "tv" ? "tv" : "movies"}/${props.item.media.id}`)
// Optimistic rows have a negative id until the server answers.
const pending = computed(() => props.item.id < 0)
const grabbed = computed(() => reorder?.grabbedId === props.item.id)
const dragging = computed(() => grabbed.value && reorder?.dragging === true)
const dragStyle = computed(() => (dragging.value ? { transform: `translate(${reorder!.dx}px, ${reorder!.dy}px) scale(1.03)` } : undefined))

function act(action: "edit-note" | "move-up" | "move-down" | "remove") {
  menuOpen.value = false
  if (action === "edit-note") emit("edit-note")
  else if (action === "move-up") emit("move-up")
  else if (action === "move-down") emit("move-down")
  else emit("remove")
}
function onMenuFocusOut(e: FocusEvent) {
  if (!(e.currentTarget as HTMLElement).contains(e.relatedTarget as Node | null)) menuOpen.value = false
}
</script>

<style scoped>
.bcard {
  --tier: var(--color-border-strong);
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  min-width: 0;
}
.bcard--gold { --tier: var(--color-accent); }
.bcard--silver { --tier: var(--color-text-secondary); }
.bcard--bronze { --tier: color-mix(in srgb, var(--color-warning) 55%, var(--color-brand)); }
.bcard.is-pending { opacity: 0.6; }
.bcard.is-grabbed { z-index: 5; }
.bcard.is-dragging { pointer-events: none; box-shadow: 0 16px 40px color-mix(in srgb, black 55%, transparent); opacity: 0.92; }

.bcard__poster { position: relative; }
.bcard__poster-link { display: block; }
.bcard__img {
  display: block;
  width: 100%;
  aspect-ratio: 2 / 3;
  object-fit: cover;
  background: var(--color-surface-3);
  border: 2px solid var(--tier);
  border-radius: var(--radius-card);
}
.bcard__img--empty { display: grid; place-items: center; color: var(--color-text-muted); }
.bcard--gold .bcard__img { box-shadow: 0 0 0 4px color-mix(in srgb, var(--tier) 18%, transparent), 0 12px 36px color-mix(in srgb, var(--tier) 22%, transparent); }

.bcard__rank {
  position: absolute;
  top: 0.375rem;
  left: 0.375rem;
  display: grid;
  place-items: center;
  min-width: 1.625rem;
  height: 1.625rem;
  padding: 0 0.375rem;
  font-size: var(--text-caption);
  font-weight: 800;
  color: var(--color-surface);
  background: var(--tier);
  border-radius: 999px;
  font-variant-numeric: tabular-nums;
}
.bcard--shelf .bcard__rank { color: var(--color-text-primary); background: var(--color-surface-3); border: 1px solid var(--color-border-strong); }

.bcard__handle,
.bcard__menu-btn {
  display: grid;
  place-items: center;
  width: 2rem;
  height: 2rem;
  color: var(--color-text-primary);
  background: color-mix(in srgb, var(--color-surface) 78%, transparent);
  border: 1px solid var(--color-border-strong);
  border-radius: 999px;
  cursor: pointer;
}
.bcard__handle { position: absolute; bottom: 0.375rem; left: 0.375rem; touch-action: none; cursor: grab; }
.bcard__handle.is-active { color: var(--color-surface); background: var(--color-accent); cursor: grabbing; }
/* Touch: 44px hit area. touch-action: none stays on the handle only so the page still scrolls. */
@media (pointer: coarse) {
  .bcard__handle,
  .bcard__menu-btn,
  .bcard--shelf .bcard__handle,
  .bcard--shelf .bcard__menu-btn { width: 2.75rem; height: 2.75rem; }
  .bcard__menu-list { top: 3rem; }
  .bcard__menu-list button { min-height: 2.75rem; }
}
.bcard__menu { position: absolute; top: 0.375rem; right: 0.375rem; }
.bcard__menu-list {
  position: absolute;
  top: 2.25rem;
  right: 0;
  z-index: var(--z-popover, 200);
  min-width: 11rem;
  margin: 0;
  padding: 0.25rem;
  list-style: none;
  background: var(--color-surface-2);
  border: 1px solid var(--color-border-strong);
  border-radius: var(--radius-card);
  box-shadow: 0 12px 32px color-mix(in srgb, black 50%, transparent);
}
.bcard__menu-list button {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  width: 100%;
  padding: 0.5rem 0.625rem;
  font-size: var(--text-small);
  color: var(--color-text-primary);
  text-align: left;
  background: none;
  border: 0;
  border-radius: 6px;
  cursor: pointer;
}
.bcard__menu-list button:hover:not(:disabled) { background: var(--color-surface-3); }
.bcard__menu-list button:disabled { color: var(--color-text-muted); cursor: not-allowed; }
.bcard__menu-list .is-danger { color: var(--color-danger); }

.bcard__body { display: flex; flex-direction: column; gap: 0.25rem; min-width: 0; }
.bcard__title {
  overflow: hidden;
  font-size: var(--text-small);
  font-weight: 700;
  color: var(--color-text-primary);
  text-overflow: ellipsis;
  white-space: nowrap;
}
.bcard__title:hover { text-decoration: underline; }
.bcard__note {
  margin: 0.25rem 0 0;
  padding-left: 0.625rem;
  font-size: var(--text-caption);
  line-height: 1.5;
  color: var(--color-text-secondary);
  border-left: 2px solid var(--tier);
  overflow-wrap: anywhere;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  overflow: hidden;
}
.bcard__add-note {
  align-self: flex-start;
  padding: 0;
  font-size: var(--text-caption);
  color: var(--color-text-muted);
  text-decoration: underline;
  background: none;
  border: 0;
  cursor: pointer;
}
.bcard__add-note:hover { color: var(--color-text-primary); }

/* Hero and podium show the full note and a larger title. */
.bcard--hero .bcard__title,
.bcard--podium .bcard__title {
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  white-space: normal;
  overflow-wrap: anywhere;
  font-size: var(--text-body);
}
.bcard--shelf { gap: 0.25rem; }
.bcard--shelf .bcard__title { font-size: var(--text-caption); font-weight: 600; }
.bcard--shelf .bcard__handle,
.bcard--shelf .bcard__menu-btn { width: 1.75rem; height: 1.75rem; }
@media (max-width: 767.98px) {
  .bcard--podium .bcard__note { -webkit-line-clamp: 3; line-clamp: 3; display: -webkit-box; }
}
.bcard--hero .bcard__note,
.bcard--podium .bcard__note { display: block; line-clamp: unset; -webkit-line-clamp: unset; font-size: var(--text-small); }

@media (min-width: 768px) {
  .bcard--hero { flex-direction: row; align-items: flex-start; gap: 1.25rem; }
  .bcard--hero .bcard__poster { flex: none; width: 13rem; }
  .bcard--hero .bcard__body { flex: 1; min-width: 12rem; padding-top: 0.5rem; }
  .bcard--hero .bcard__title { font-size: var(--text-title); }
}

@media (prefers-reduced-motion: no-preference) {
  .bcard { transition: transform var(--motion-fast) var(--ease-out-soft), opacity var(--motion-fast) var(--ease-out-soft); }
  .bcard.is-dragging { transition: none; }
  @media (hover: hover) {
    .bcard:not(.is-dragging):hover .bcard__img { transform: translateY(-2px); }
  }
  .bcard__img { transition: transform var(--motion-fast) var(--ease-out-soft); }
}
</style>
