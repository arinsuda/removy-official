<template>
  <AppModal v-model="open" size="sm" :title="t('library.customLists.sheet.title')">
    <SkeletonList v-if="loading" :count="3" />
    <AppErrorState v-else-if="failed" :title="t('library.customLists.loadFailed')" @retry="load" />
    <template v-else>
      <p v-if="!lists.length" class="sheet__note">{{ t("library.customLists.sheet.empty") }}</p>
      <ul v-else class="sheet__list" :aria-label="t('library.customLists.sheet.title')">
        <li v-for="l in lists" :key="l.id" class="sheet__row">
          <label class="sheet__label">
            <input
              type="checkbox"
              class="sheet__check"
              :checked="contains[l.id]"
              :disabled="busy.has(l.id)"
              :data-list-id="l.id"
              @change="toggle(l, $event)"
            />
            <span class="sheet__name">{{ l.name }}</span>
            <span class="sheet__count">{{ t("library.lists.count", { n: counts[l.id] ?? l.item_count }, counts[l.id] ?? l.item_count) }}</span>
          </label>
          <AppSpinner v-if="busy.has(l.id)" :size="14" />
        </li>
      </ul>

      <form class="sheet__new" @submit.prevent="create">
        <AppInput v-model="newName" :label="t('library.customLists.sheet.newLabel')" :disabled="creating || atLimit" :maxlength="LIST_NAME_MAX" :placeholder="t('library.customLists.sheet.newPlaceholder')" autocomplete="off" data-field="new-list" />
        <p v-if="atLimit" class="sheet__note">{{ t("library.customLists.limit", { max: MAX_CUSTOM_LISTS }) }}</p>
        <AppButton type="submit" variant="outline" :loading="creating" :disabled="!newName.trim() || atLimit">
          <Plus :size="16" aria-hidden="true" />{{ t("library.customLists.sheet.create") }}
        </AppButton>
      </form>
    </template>
  </AppModal>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from "vue"
import { useI18n } from "vue-i18n"
import { Plus } from "lucide-vue-next"
import { listApi } from "@/api/endpoints/list"
import AppButton from "@/components/ui/AppButton.vue"
import AppErrorState from "@/components/ui/AppErrorState.vue"
import AppInput from "@/components/ui/AppInput.vue"
import AppModal from "@/components/ui/AppModal.vue"
import AppSpinner from "@/components/ui/AppSpinner.vue"
import SkeletonList from "@/components/ui/skeleton/SkeletonList.vue"
import { useToast } from "@/composables/useToast"
import type { ListSummary, MediaType } from "@/types"
import { LIST_NAME_MAX, MAX_CUSTOM_LISTS } from "@/views/library/config"
import { useListActions } from "./useListActions"

const props = defineProps<{ modelValue: boolean; mediaId: number; mediaType: MediaType }>()
const emit = defineEmits<{ "update:modelValue": [value: boolean] }>()

const { t } = useI18n()
const actions = useListActions()
const toast = useToast()
const open = computed({ get: () => props.modelValue, set: (v) => emit("update:modelValue", v) })

const lists = ref<ListSummary[]>([])
const loading = ref(false)
const failed = ref(false)
const creating = ref(false)
const newName = ref("")
const busy = reactive(new Set<number>())
const contains = reactive<Record<number, boolean>>({})
const counts = reactive<Record<number, number>>({})
/** Item ids learned this session (added here, or found when unchecking). */
const itemIds = new Map<number, number>()

const atLimit = computed(() => lists.value.length >= MAX_CUSTOM_LISTS)
const media = computed(() => ({ media_id: props.mediaId, media_type: props.mediaType }))

async function load() {
  loading.value = true
  failed.value = false
  try {
    const { data } = await listApi.getOwn({ media_type: props.mediaType, media_id: props.mediaId })
    lists.value = data.lists ?? []
    for (const l of lists.value) {
      contains[l.id] = !!l.contains_media
      counts[l.id] = l.item_count
      if (l.item_id) itemIds.set(l.id, l.item_id)
    }
  } catch {
    failed.value = true
  } finally {
    loading.value = false
  }
}

async function toggle(l: ListSummary, e: Event) {
  const box = e.target as HTMLInputElement
  if (busy.has(l.id)) {
    box.checked = !!contains[l.id]
    return
  }
  busy.add(l.id)
  const adding = !contains[l.id]
  try {
    if (adding) {
      const item = await actions.addItem(l.id, media.value)
      if (item) {
        itemIds.set(l.id, item.id)
        contains[l.id] = true
        counts[l.id] = (counts[l.id] ?? 0) + 1
      }
    } else {
      const itemId = itemIds.get(l.id) ?? null
      if (itemId === null) {
        toast.error(t("library.customLists.sheet.removeFailed"))
      } else if (await actions.removeItem(l.id, itemId)) {
        itemIds.delete(l.id)
        contains[l.id] = false
        counts[l.id] = Math.max((counts[l.id] ?? 1) - 1, 0)
      }
    }
  } finally {
    busy.delete(l.id)
    // The checkbox flipped on click; show what the server state actually is.
    box.checked = !!contains[l.id]
  }
}

async function create() {
  const name = newName.value.trim()
  if (!name || creating.value) return
  creating.value = true
  const list = await actions.create({ name })
  if (list) {
    newName.value = ""
    const item = await actions.addItem(list.id, media.value)
    lists.value = [list, ...lists.value]
    contains[list.id] = !!item
    counts[list.id] = item ? 1 : 0
    if (item) itemIds.set(list.id, item.id)
  }
  creating.value = false
}

watch(
  () => [props.modelValue, props.mediaId, props.mediaType] as const,
  ([isOpen]) => {
    if (!isOpen) return
    itemIds.clear()
    void load()
  },
  { immediate: true },
)
</script>

<style scoped>
.sheet__list { display: flex; flex-direction: column; gap: 0.25rem; max-height: 16rem; margin: 0 0 1rem; padding: 0; overflow-y: auto; list-style: none; }
.sheet__row { display: flex; align-items: center; gap: 0.5rem; }
.sheet__label { display: flex; flex: 1; align-items: center; gap: 0.75rem; min-width: 0; padding: 0.625rem 0.5rem; cursor: pointer; border-radius: var(--radius-btn); }
.sheet__label:hover { background: var(--color-surface-3); }
.sheet__check { flex: none; width: 1.125rem; height: 1.125rem; accent-color: var(--color-brand); }
.sheet__name { flex: 1; min-width: 0; overflow: hidden; font-size: var(--text-small); font-weight: 600; color: var(--color-text-primary); text-overflow: ellipsis; white-space: nowrap; }
.sheet__count { flex: none; font-size: var(--text-caption); color: var(--color-text-muted); }
.sheet__note { margin: 0 0 1rem; font-size: var(--text-small); color: var(--color-text-secondary); }
.sheet__new { display: flex; flex-direction: column; gap: 0.625rem; padding-top: 1rem; border-top: 1px solid var(--color-border); }
</style>
