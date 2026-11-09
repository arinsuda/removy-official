<template>
  <AppEmptyState
    v-if="forbidden"
    :icon="Lock"
    :title="t('library.bmol.locked.title')"
    :description="t('library.bmol.locked.body')"
  />
  <section v-else class="board" :aria-labelledby="headingId">
    <header class="board__head">
      <div class="board__titles">
        <h2 :id="headingId" class="board__title"><Trophy :size="22" aria-hidden="true" />{{ t("library.bmol.title") }}</h2>
        <p class="board__desc">{{ isOwner ? t("library.bmol.descriptionOwner") : t("library.bmol.descriptionVisitor") }}</p>
      </div>
      <div v-if="isOwner && !isLoading && !isError" class="board__actions">
        <AppButton size="sm" @click="openAdd()"><Plus :size="16" aria-hidden="true" />{{ t("library.bmol.actions.add") }}</AppButton>
        <AppButton v-if="typeItems.length && view === 'board'" size="sm" variant="outline" :aria-pressed="editing" @click="editing = !editing">
          <component :is="editing ? Check : PenLine" :size="16" aria-hidden="true" />{{ editing ? t("library.bmol.actions.done") : t("library.bmol.actions.edit") }}
        </AppButton>
        <AppButton v-if="typeItems.length" size="sm" variant="outline" @click="showShare = true">
          <Share2 :size="16" aria-hidden="true" />{{ t("library.bmol.actions.share") }}
        </AppButton>
      </div>
    </header>

    <div class="board__bar">
      <AppTabs v-model="mediaType" :tabs="tabs" :controls="false" :label="t('library.bmol.tabs.label')" class="board__tabs" />
      <div v-if="typeItems.length" class="board__view" role="group" :aria-label="t('library.bmol.view.label')">
        <button v-for="v in VIEWS" :key="v.key" type="button" class="board__view-btn" :class="{ 'is-active': view === v.key }" :aria-pressed="view === v.key" data-bmol-view @click="view = v.key">
          <component :is="v.icon" :size="15" aria-hidden="true" />{{ t(`library.bmol.view.${v.key}`) }}
        </button>
      </div>
    </div>

    <SkeletonGrid v-if="isLoading" :count="8" min-width="140px" />
    <AppErrorState v-else-if="isError" :title="t('library.bmol.error.title')" :description="t('library.bmol.error.body')" @retry="refetch()" />

    <template v-else-if="!typeItems.length">
      <AppEmptyState v-if="isOwner" :icon="Trophy" :title="t('library.bmol.onboarding.title')" :description="t('library.bmol.onboarding.body')">
        <template #action>
          <AppButton @click="openAdd(1)"><Plus :size="16" aria-hidden="true" />{{ t("library.bmol.onboarding.cta") }}</AppButton>
        </template>
      </AppEmptyState>
      <AppEmptyState v-else :icon="Trophy" :title="t('library.bmol.empty.title')" :description="t('library.bmol.empty.body')" />
    </template>

    <BmolTimeline v-else-if="view === 'timeline'" :items="typeItems">
      <template #default="{ items: yearItems }">
        <BmolCard v-for="item in yearItems" :key="item.id" :item="item" variant="shelf" />
      </template>
    </BmolTimeline>

    <div v-else class="board__tiers" :class="{ 'is-reordering': reorderActive }">
      <BmolHero v-if="heroGroup" :group="heroGroup" :is-target="reorder.targetRank === 1 && reorder.grabbedId !== null">
        <template #default="{ items }">
          <BmolCard v-for="item in items" :key="item.id" v-bind="cardBind(item, 'hero')" v-on="cardOn(item)" />
        </template>
      </BmolHero>

      <BmolPodium v-if="podiumGroups.length" :groups="podiumGroups" :target-rank="activeTarget">
        <template #default="{ items }">
          <BmolCard v-for="item in items" :key="item.id" v-bind="cardBind(item, 'podium')" v-on="cardOn(item)" />
        </template>
      </BmolPodium>

      <BmolShelf v-if="shelfGroups.length" :groups="shelfGroups" :target-rank="activeTarget">
        <template #default="{ items }">
          <BmolCard v-for="item in items" :key="item.id" v-bind="cardBind(item, 'shelf')" v-on="cardOn(item)" />
        </template>
      </BmolShelf>
    </div>

    <p class="sr-only" role="status" aria-live="assertive">{{ announcement }}</p>

    <BmolAddSheet
      v-if="isOwner"
      v-model="showAdd"
      :user-id="userId"
      :media-type="mediaType"
      :items="items"
      :default-rank="addDefaultRank"
      @add="onAdd"
    />
    <BmolNoteEditor v-if="isOwner" v-model="showNote" :media-title="noteTarget?.media.title ?? ''" :note="noteTarget?.note ?? ''" :saving="noteSaving" @save="onSaveNote" />
    <ConfirmModal
      v-if="isOwner"
      v-model="showRemove"
      list-type="bmol_delete"
      :item-name="removeTarget?.media.title ?? ''"
      @confirm="onConfirmRemove"
      @cancel="showRemove = false"
    />
    <BolShareSheet v-if="isOwner" v-model="showShare" :media-type="mediaType" :groups="groups" :username="username" />
  </section>
</template>

<script setup lang="ts">
import { computed, ref, useId, watch } from "vue"
import { useI18n } from "vue-i18n"
import { CalendarRange, Check, LayoutGrid, Lock, PenLine, Plus, Share2, Trophy } from "lucide-vue-next"
import AppButton from "@/components/ui/AppButton.vue"
import AppEmptyState from "@/components/ui/AppEmptyState.vue"
import AppErrorState from "@/components/ui/AppErrorState.vue"
import AppTabs from "@/components/ui/AppTabs.vue"
import SkeletonGrid from "@/components/ui/skeleton/SkeletonGrid.vue"
import ConfirmModal from "@/components/profile/components/ConfirmModal.vue"
import BolShareSheet from "@/components/share/BolShareSheet.vue"
import { useAuthStore } from "@/stores/auth"
import type { BMOLItemResponse, MediaType } from "@/types"
import BmolAddSheet from "./BmolAddSheet.vue"
import BmolCard from "./BmolCard.vue"
import BmolHero from "./BmolHero.vue"
import BmolNoteEditor from "./BmolNoteEditor.vue"
import BmolPodium from "./BmolPodium.vue"
import BmolShelf from "./BmolShelf.vue"
import BmolTimeline from "./BmolTimeline.vue"
import { groupByRank, moveTarget } from "./bmolRanks"
import { useBmol, type BmolCandidate } from "./useBmol"
import { provideBmolReorder } from "./useBmolReorder"

const props = defineProps<{ userId: number }>()

const { t } = useI18n()
const auth = useAuthStore()
const headingId = useId()
const { items, forbidden, isLoading, isError, refetch, add, move, saveNote, remove } = useBmol(() => props.userId)

const isOwner = computed(() => auth.user?.id === props.userId)
const username = computed(() => auth.user?.display_name || auth.user?.username || "")
const mediaType = ref<MediaType>("movie")
const editing = ref(false)
const VIEWS = [
  { key: "board" as const, icon: LayoutGrid },
  { key: "timeline" as const, icon: CalendarRange },
]
const view = ref<(typeof VIEWS)[number]["key"]>("board")
watch(view, () => (editing.value = false))

const typeItems = computed(() => items.value.filter((i) => i.media_type === mediaType.value))
const groups = computed(() => groupByRank(typeItems.value))
const heroGroup = computed(() => groups.value.find((g) => g.rank === 1))
const podiumGroups = computed(() => groups.value.filter((g) => g.rank === 2 || g.rank === 3))
const shelfGroups = computed(() => groups.value.filter((g) => g.rank > 3 || g.rank < 1))
const tabs = computed(() => [
  { key: "movie" as const, label: `${t("library.bmol.tabs.movie")} (${items.value.filter((i) => i.media_type === "movie").length})` },
  { key: "tv" as const, label: `${t("library.bmol.tabs.tv")} (${items.value.filter((i) => i.media_type === "tv").length})` },
])

// Drag and keyboard reordering; a drop on a rank group is one PUT.
const { reorder, announcement, active: reorderActive } = provideBmolReorder(
  groups,
  (id, rank) => void move(id, rank),
  (id) => items.value.find((i) => i.id === id)?.media.title ?? "",
)
const activeTarget = computed(() => (reorder.grabbedId !== null ? reorder.targetRank : null))

function cardBind(item: BMOLItemResponse, variant: "hero" | "podium" | "shelf") {
  return {
    item,
    variant,
    editing: isOwner.value && editing.value,
    canMoveUp: moveTarget(groups.value, item.rank, "up") !== null,
    canMoveDown: moveTarget(groups.value, item.rank, "down") !== null,
  }
}
function cardOn(item: BMOLItemResponse) {
  return {
    "edit-note": () => openNote(item),
    "move-up": () => stepItem(item, "up"),
    "move-down": () => stepItem(item, "down"),
    remove: () => askRemove(item),
  }
}
function stepItem(item: BMOLItemResponse, direction: "up" | "down") {
  const target = moveTarget(groups.value, item.rank, direction)
  if (target !== null) void move(item.id, target)
}

// Add
const showAdd = ref(false)
const addDefaultRank = ref<number | undefined>()
function openAdd(rank?: number) {
  addDefaultRank.value = rank
  showAdd.value = true
}
async function onAdd(candidates: BmolCandidate[], rank: number) {
  // One at a time: they share a rank and each answer refreshes the list.
  for (const candidate of candidates) await add(candidate, rank)
}

// Note
const showNote = ref(false)
const noteTarget = ref<BMOLItemResponse | null>(null)
const noteSaving = ref(false)
function openNote(item: BMOLItemResponse) {
  noteTarget.value = item
  showNote.value = true
}
async function onSaveNote(note: string) {
  if (!noteTarget.value) return
  noteSaving.value = true
  const ok = await saveNote(noteTarget.value.id, note)
  noteSaving.value = false
  if (ok) showNote.value = false
}

// Remove
const showRemove = ref(false)
const removeTarget = ref<BMOLItemResponse | null>(null)
function askRemove(item: BMOLItemResponse) {
  removeTarget.value = item
  showRemove.value = true
}
async function onConfirmRemove() {
  const target = removeTarget.value
  showRemove.value = false
  if (target) await remove(target.id)
}

const showShare = ref(false)

watch(
  () => props.userId,
  () => {
    editing.value = false
    mediaType.value = "movie"
  },
)
watch(isOwner, (owner) => {
  if (!owner) editing.value = false
})
</script>

<style scoped>
.board { display: flex; flex-direction: column; gap: 1.25rem; min-width: 0; }
.board__head { display: flex; flex-direction: column; gap: 0.875rem; }
.board__title { display: flex; align-items: center; gap: 0.5rem; font-size: var(--text-title); font-weight: 800; color: var(--color-text-primary); }
.board__desc { margin-top: 0.25rem; font-size: var(--text-small); color: var(--color-text-secondary); }
.board__actions { display: flex; flex-wrap: wrap; gap: 0.5rem; }
.board__bar { display: flex; flex-wrap: wrap; align-items: center; justify-content: space-between; gap: 0.75rem; }
.board__tabs { overflow-x: auto; }
.board__view { display: inline-flex; padding: 2px; background: var(--color-surface-2); border: 1px solid var(--color-border); border-radius: var(--radius-btn); }
.board__view-btn { display: inline-flex; align-items: center; gap: 0.375rem; min-height: 2rem; padding: 0 0.75rem; font: inherit; font-size: var(--text-caption); color: var(--color-text-muted); background: none; border: 0; border-radius: var(--radius-btn); cursor: pointer; transition: color var(--motion-fast) var(--ease-water), background-color var(--motion-fast) var(--ease-water); }
.board__view-btn.is-active { color: var(--color-text-primary); background: var(--color-surface-3); }
@media (prefers-reduced-motion: reduce) { .board__view-btn { transition: none; } }
.board__tiers { display: flex; flex-direction: column; gap: 1.5rem; }
@media (min-width: 768px) {
  .board__head { flex-direction: row; align-items: flex-start; justify-content: space-between; }
}
</style>
