<template>
  <section class="cd" :aria-labelledby="titleId">
    <header class="cd__head">
      <AppButton variant="ghost" size="sm" @click="emit('back')">
        <ChevronLeft :size="16" aria-hidden="true" />{{ t("library.lists.back") }}
      </AppButton>
      <div v-if="list" class="cd__titles">
        <h2 :id="titleId" class="cd__title">{{ list.name }}</h2>
        <p v-if="list.description" class="cd__desc">{{ list.description }}</p>
        <p class="cd__vis"><component :is="LIST_VISIBILITY_ICONS[list.visibility]" :size="13" aria-hidden="true" />{{ t(`visibility.${list.visibility}`) }}</p>
      </div>
      <div v-if="list && isOwner" class="cd__actions">
        <AppButton size="sm" variant="outline" data-action="edit-list" @click="showEditor = true"><PenLine :size="15" aria-hidden="true" />{{ t("library.customLists.detail.edit") }}</AppButton>
        <AppButton size="sm" variant="outline" data-action="delete-list" @click="showDelete = true"><Trash2 :size="15" aria-hidden="true" />{{ t("library.customLists.detail.delete") }}</AppButton>
      </div>
    </header>

    <SkeletonGrid v-if="loading" :count="10" min-width="140px" />
    <AppErrorState v-else-if="error === 'notFound' || error === 'forbidden'" :title="t(`library.customLists.detail.${error}Title`)" :description="t(`library.customLists.detail.${error}Body`)" />
    <AppErrorState v-else-if="error" :title="t('library.customLists.detail.errorTitle')" :description="t('library.customLists.detail.errorBody')" @retry="load()" />

    <template v-else-if="list">
      <AppEmptyState v-if="!items.length" :icon="ListIcon" :title="t(isOwner ? 'library.customLists.detail.emptyOwnerTitle' : 'library.customLists.detail.emptyVisitorTitle')" :description="isOwner ? t('library.customLists.detail.emptyOwnerBody') : undefined" />
      <template v-else>
        <TabToolbar v-model:query="tabList.query.value" v-model:sort="tabList.sort.value" :count-label="t('library.lists.count', { n: total }, total)" :sorts="sorts" />
        <p v-if="isOwner && tabList.sort.value !== 'position'" class="cd__note">{{ t("library.customLists.detail.reorderHint") }}</p>
        <TabEmptyState v-if="!tabList.filtered.value.length" filtered :icon="ListIcon" :is-owner="isOwner" @clear="tabList.clear" />
        <MediaGrid v-else :items="tabList.visible.value.map((e) => e.media)" :rail="`custom-${list.id}`" :label="list.name" :cell-state="cellStateOf">
          <template #footer="{ item }">
            <span class="cd__meta"><time :datetime="entryOf(item).at">{{ formatDate(entryOf(item).at) }}</time></span>
            <span v-if="isOwner" class="cd__tools">
              <template v-if="canReorder">
                <TabIconButton
                  data-action="drag-handle"
                  class="cd__handle"
                  :label="t('library.customLists.detail.dragHandle', { title: item.title })"
                  @pointerdown="drag.onPointerDown($event, item.key)"
                  @pointermove="drag.onPointerMove($event)"
                  @pointerup="drag.onPointerUp($event)"
                  @pointercancel="drag.cancel()"
                >
                  <GripVertical :size="15" aria-hidden="true" />
                </TabIconButton>
                <TabIconButton data-action="move-up" :disabled="indexOf(item) === 0" :label="t('library.customLists.detail.moveUp', { title: item.title })" @click="step(item, -1)">
                  <ArrowUp :size="15" aria-hidden="true" />
                </TabIconButton>
                <TabIconButton data-action="move-down" :disabled="indexOf(item) === items.length - 1" :label="t('library.customLists.detail.moveDown', { title: item.title })" @click="step(item, 1)">
                  <ArrowDown :size="15" aria-hidden="true" />
                </TabIconButton>
              </template>
              <TabIconButton data-action="remove" danger :label="t('library.customLists.detail.remove', { title: item.title })" @click="onRemove(entryOf(item))">
                <Trash2 :size="15" aria-hidden="true" />
              </TabIconButton>
            </span>
          </template>
        </MediaGrid>
        <div v-if="hasMore" class="cd__more">
          <AppButton variant="outline" :loading="loadingMore" @click="loadMore()">{{ t("profile.tabContent.common.loadMore") }}</AppButton>
        </div>
      </template>
    </template>

    <p class="sr-only" role="status" aria-live="polite">{{ announcement }}</p>

    <ListEditorModal v-if="isOwner && list" v-model="showEditor" :list="list" :items="items" @saved="onSaved" />
    <ConfirmModal v-if="isOwner && list" v-model="showDelete" list-type="custom_list_delete" :item-name="list.name" :confirm-disabled="deleting" @confirm="onDelete" @cancel="showDelete = false" />
  </section>
</template>

<script setup lang="ts">
import { computed, ref, useId } from "vue"
import { useI18n } from "vue-i18n"
import { ArrowDown, ArrowUp, ChevronLeft, GripVertical, List as ListIcon, PenLine, Trash2 } from "lucide-vue-next"
import { listApi } from "@/api/endpoints/list"
import AppButton from "@/components/ui/AppButton.vue"
import AppEmptyState from "@/components/ui/AppEmptyState.vue"
import AppErrorState from "@/components/ui/AppErrorState.vue"
import SkeletonGrid from "@/components/ui/skeleton/SkeletonGrid.vue"
import ConfirmModal from "@/components/profile/components/ConfirmModal.vue"
import MediaGrid from "@/components/profile/tabs/MediaGrid.vue"
import TabEmptyState from "@/components/profile/tabs/TabEmptyState.vue"
import TabIconButton from "@/components/profile/tabs/TabIconButton.vue"
import TabToolbar from "@/components/profile/tabs/TabToolbar.vue"
import { summaryToHomeMedia, type ProfileEntry } from "@/components/profile/tabs/profileMedia"
import { useTabList } from "@/components/profile/tabs/useTabList"
import type { HomeMedia } from "@/components/home/homeMedia"
import { listErrorKey } from "@/components/lists/listErrors"
import { useListActions } from "@/components/lists/useListActions"
import { useToast } from "@/composables/useToast"
import type { ListSummary } from "@/types"
import { LIST_VISIBILITY_ICONS } from "../config"
import { moveIndex, useListReorder } from "../useListReorder"
import { useCustomListDetail } from "../useCustomListDetail"
import ListEditorModal from "./ListEditorModal.vue"

interface ListEntry extends ProfileEntry {
  position: number
}

const props = defineProps<{ listId: number; currentUserId: number | null }>()
const emit = defineEmits<{ back: [] }>()

const { t, locale } = useI18n()
const toast = useToast()
const actions = useListActions()
const titleId = useId()
const { list, items, loading, loadingMore, error, hasMore, total, load, loadMore, loadAll, mutate } = useCustomListDetail(() => props.listId)

const isOwner = computed(() => !!list.value && list.value.user_id === props.currentUserId)
const showEditor = ref(false)
const showDelete = ref(false)
const deleting = ref(false)
const announcement = ref("")

const entries = computed<ListEntry[]>(() => items.value.map((i) => ({ id: i.id, position: i.position, media: summaryToHomeMedia(i.media), at: i.added_at })))
const byKey = computed(() => new Map(entries.value.map((e) => [e.media.key, e])))
// Only items from `visible` reach the footer slot, and each has an entry.
const entryOf = (item: HomeMedia) => byKey.value.get(item.key)!
const indexOf = (item: HomeMedia) => entries.value.findIndex((e) => e.media.key === item.key)

const SORT_KEYS = ["position", "newest", "title", "release"] as const
const releaseTime = (e: ListEntry) => Date.parse(e.media.date) || Number.NEGATIVE_INFINITY
const tabList = useTabList<ListEntry>({
  items: entries,
  searchText: (e) => e.media.title,
  sorts: {
    position: (a, b) => a.position - b.position,
    newest: (a, b) => Date.parse(b.at) - Date.parse(a.at),
    title: (a, b) => a.media.title.localeCompare(b.media.title),
    release: (a, b) => releaseTime(b) - releaseTime(a) || 0,
  },
  defaultSort: "position",
  pageSize: 48,
})
const sorts = computed(() => SORT_KEYS.map((key) => ({ key, label: t(`library.customLists.detail.sorts.${key}`) })))
const formatDate = (iso: string) => new Intl.DateTimeFormat(locale.value, { dateStyle: "medium" }).format(new Date(iso))

// Reorder only makes sense in the saved order with nothing filtered out.
const canReorder = computed(() => tabList.sort.value === "position" && !tabList.query.value.trim())

async function move(from: number, to: number) {
  if (!(await loadAll())) {
    toast.error(t("library.customLists.detail.reorderLoadFailed"))
    return
  }
  const next = moveIndex(items.value, from, to)
  if (next === items.value || !list.value) return
  announcement.value = t("library.customLists.detail.moved", { title: items.value[from]?.media.title ?? "", position: to + 1 })
  const undo = mutate(next.map((i, idx) => ({ ...i, position: idx + 1 })))
  try {
    await listApi.reorder(list.value.id, next.map((i) => i.id))
  } catch (err) {
    undo()
    toast.error(t(listErrorKey(err)))
    void load()
  }
}
const step = (item: HomeMedia, delta: number) => move(indexOf(item), indexOf(item) + delta)

const drag = useListReorder(
  () => entries.value.map((e) => e.media.key),
  (from, to) => void move(from, to),
)
const cellStateOf = (item: HomeMedia) => drag.cellState(item.key)

async function onRemove(entry: ListEntry) {
  if (!list.value) return
  const undo = mutate(items.value.filter((i) => i.id !== entry.id))
  total.value -= 1
  const ok = await actions.removeItem(list.value.id, entry.id)
  if (!ok) {
    undo()
    total.value += 1
  }
}

function onSaved(saved: ListSummary) {
  list.value = saved
}

async function onDelete() {
  if (!list.value) return
  deleting.value = true
  const ok = await actions.remove(list.value.id)
  deleting.value = false
  if (!ok) return
  showDelete.value = false
  toast.success(t("library.customLists.detail.deleted"))
  emit("back")
}
</script>

<style scoped>
.cd__head { display: flex; flex-wrap: wrap; align-items: flex-start; gap: 0.5rem 0.75rem; margin-bottom: 1.25rem; }
.cd__titles { flex: 1 1 14rem; order: 2; min-width: 0; }
.cd__title { margin: 0; font-size: var(--text-title); font-weight: 700; color: var(--color-text-primary); overflow-wrap: anywhere; }
.cd__desc { margin: 0.25rem 0 0; font-size: var(--text-small); color: var(--color-text-secondary); overflow-wrap: anywhere; }
.cd__vis { display: inline-flex; align-items: center; gap: 0.3rem; margin: 0.375rem 0 0; font-size: var(--text-caption); color: var(--color-text-muted); }
.cd__actions { display: flex; gap: 0.5rem; order: 3; }
.cd__note { margin: 0 0 0.75rem; font-size: var(--text-caption); color: var(--color-text-muted); }
.cd__meta { display: none; min-width: 0; font-size: var(--text-caption); color: var(--color-text-muted); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.cd__tools { display: inline-flex; margin-left: auto; flex: none; align-items: center; }
.cd__handle { cursor: grab; touch-action: none; }
@media (min-width: 1024px) {
  .cd__meta { display: block; }
}
.cd__more { display: flex; justify-content: center; padding: 2rem 0 0.5rem; }
.sr-only { position: absolute; width: 1px; height: 1px; overflow: hidden; clip-path: inset(50%); white-space: nowrap; }
@media (min-width: 640px) {
  .cd__titles { order: 0; }
}
</style>
