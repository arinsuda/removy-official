<template>
  <AppModal v-model="open" size="md" :title="t(isEdit ? 'library.customLists.editor.editTitle' : 'library.customLists.editor.createTitle')" :dismiss-disabled="saving">
    <form id="list-editor-form" class="ed" novalidate @submit.prevent="submit">
      <div class="ed__field">
        <AppInput v-model="name" :label="t('library.customLists.editor.name')" :error="nameError" :disabled="saving" autocomplete="off" data-field="name" @blur="touched = true" />
        <p class="ed__count" :class="{ 'is-over': nameLength > LIST_NAME_MAX }">{{ t("library.customLists.editor.counter", { n: nameLength, max: LIST_NAME_MAX }) }}</p>
      </div>

      <div class="ed__field">
        <label :for="descId" class="ed__label">{{ t("library.customLists.editor.description") }}</label>
        <textarea :id="descId" v-model="description" class="ed__textarea" rows="3" :disabled="saving" :aria-invalid="descTooLong || undefined" :placeholder="t('library.customLists.editor.descriptionPlaceholder')" data-field="description" />
        <p class="ed__count" :class="{ 'is-over': descTooLong }">{{ t("library.customLists.editor.counter", { n: descLength, max: LIST_DESC_MAX }) }}</p>
        <p v-if="descTooLong" class="ed__error" role="alert">{{ t("library.customLists.editor.descTooLong", { max: LIST_DESC_MAX }) }}</p>
      </div>

      <fieldset class="ed__fieldset">
        <legend class="ed__label">{{ t("visibility.label") }}</legend>
        <label v-for="v in LIST_VISIBILITIES" :key="v" class="ed__vis" :class="{ 'is-on': visibility === v }">
          <input v-model="visibility" type="radio" class="ed__radio" name="list-visibility" :value="v" :disabled="saving" />
          <component :is="LIST_VISIBILITY_ICONS[v]" :size="16" aria-hidden="true" />
          <span class="ed__vis-text"><strong>{{ t(`visibility.${v}`) }}</strong><small>{{ t(`visibility.${v}Desc`) }}</small></span>
        </label>
      </fieldset>

      <fieldset v-if="isEdit && items.length" class="ed__fieldset">
        <legend class="ed__label">{{ t("library.customLists.editor.cover") }}</legend>
        <p class="ed__hint">{{ t("library.customLists.editor.coverHint") }}</p>
        <div class="ed__covers">
          <label class="ed__cover ed__cover--auto" :class="{ 'is-on': coverItemId === 0 }">
            <input v-model="coverItemId" type="radio" class="ed__radio" name="list-cover" :value="0" :disabled="saving" />
            <span>{{ t("library.customLists.editor.coverAuto") }}</span>
          </label>
          <label v-for="item in items" :key="item.id" class="ed__cover" :class="{ 'is-on': coverItemId === item.id }">
            <input v-model="coverItemId" type="radio" class="ed__radio" name="list-cover" :value="item.id" :disabled="saving" :aria-label="item.media.title" />
            <img v-if="posterOf(item)" :src="posterOf(item)" alt="" class="ed__cover-img" loading="lazy" />
            <Film v-else :size="18" aria-hidden="true" />
          </label>
        </div>
      </fieldset>
    </form>
    <template #footer>
      <div class="ed__actions">
        <AppButton variant="outline" :disabled="saving" @click="open = false">{{ t("common.cancel") }}</AppButton>
        <AppButton type="submit" form="list-editor-form" :loading="saving" :disabled="!canSave">{{ t(isEdit ? "library.customLists.editor.save" : "library.customLists.editor.create") }}</AppButton>
      </div>
    </template>
  </AppModal>
</template>

<script setup lang="ts">
import { computed, ref, useId, watch } from "vue"
import { useI18n } from "vue-i18n"
import { Film } from "lucide-vue-next"
import AppButton from "@/components/ui/AppButton.vue"
import AppInput from "@/components/ui/AppInput.vue"
import AppModal from "@/components/ui/AppModal.vue"
import { useListActions } from "@/components/lists/useListActions"
import { getTmdbImageUrl } from "@/utils/image"
import type { ListItem, ListSummary, Visibility } from "@/types"
import { COVER_PICK_SIZE, DEFAULT_LIST_VISIBILITY, LIST_DESC_MAX, LIST_NAME_MAX, LIST_VISIBILITIES, LIST_VISIBILITY_ICONS } from "../config"

const props = defineProps<{
  modelValue: boolean
  /** Present to edit that list; absent to create one. */
  list?: ListSummary | null
  /** Loaded items of the list, offered as cover choices. */
  items?: ListItem[]
}>()
const emit = defineEmits<{ "update:modelValue": [value: boolean]; saved: [list: ListSummary] }>()

const { t } = useI18n()
const actions = useListActions()
const descId = useId()
const open = computed({ get: () => props.modelValue, set: (v) => emit("update:modelValue", v) })
const isEdit = computed(() => !!props.list)
const items = computed(() => props.items ?? [])

const name = ref("")
const description = ref("")
const visibility = ref<Visibility>(DEFAULT_LIST_VISIBILITY)
const coverItemId = ref(0)
const saving = ref(false)
const touched = ref(false)

// Runes, not UTF-16 units, as the server counts them.
const runes = (s: string) => [...s].length
const nameLength = computed(() => runes(name.value.trim()))
const descLength = computed(() => runes(description.value.trim()))
const descTooLong = computed(() => descLength.value > LIST_DESC_MAX)
const nameError = computed(() => {
  if (nameLength.value > LIST_NAME_MAX) return t("library.customLists.editor.nameTooLong", { max: LIST_NAME_MAX })
  if (touched.value && nameLength.value === 0) return t("library.customLists.editor.nameRequired")
  return ""
})
const valid = computed(() => nameLength.value > 0 && nameLength.value <= LIST_NAME_MAX && !descTooLong.value)
const canSave = computed(() => !saving.value && valid.value)

const posterOf = (item: ListItem) => getTmdbImageUrl(item.media.poster_url, COVER_PICK_SIZE) ?? ""

// Fresh form each time the modal opens.
watch(
  () => props.modelValue,
  (isOpen) => {
    if (!isOpen) return
    touched.value = false
    name.value = props.list?.name ?? ""
    description.value = props.list?.description ?? ""
    visibility.value = props.list?.visibility ?? DEFAULT_LIST_VISIBILITY
    coverItemId.value = props.list?.cover_item_id ?? 0
  },
  { immediate: true },
)

async function submit() {
  touched.value = true
  if (!valid.value || saving.value) return
  saving.value = true
  const body = { name: name.value.trim(), description: description.value.trim(), visibility: visibility.value }
  const saved = props.list
    ? await actions.update(props.list.id, { ...body, cover_item_id: coverItemId.value })
    : await actions.create(body)
  saving.value = false
  if (!saved) return
  emit("saved", saved)
  open.value = false
}
</script>

<style scoped>
.ed { display: flex; flex-direction: column; gap: 1.125rem; }
.ed__field { display: flex; flex-direction: column; gap: 0.4rem; }
.ed__label { font-size: var(--text-small); font-weight: 600; color: var(--color-text-secondary); }
.ed__textarea { padding: 0.625rem 0.75rem; font: inherit; font-size: var(--text-small); color: var(--color-text-primary); resize: vertical; background: var(--color-surface-2); border: 1px solid var(--color-border-strong); border-radius: var(--radius-btn); }
.ed__textarea[aria-invalid="true"] { border-color: var(--color-danger); }
.ed__count { align-self: flex-end; margin: 0; font-size: var(--text-caption); color: var(--color-text-muted); }
.ed__count.is-over, .ed__error { color: var(--color-danger); }
.ed__error { margin: 0; font-size: var(--text-caption); }
.ed__fieldset { display: flex; flex-direction: column; gap: 0.5rem; min-width: 0; margin: 0; padding: 0; border: 0; }
.ed__hint { margin: 0; font-size: var(--text-caption); color: var(--color-text-muted); }
.ed__radio { position: absolute; opacity: 0; pointer-events: none; }
.ed__vis { position: relative; display: flex; align-items: center; gap: 0.75rem; padding: 0.625rem 0.75rem; color: var(--color-text-secondary); cursor: pointer; background: var(--color-surface-2); border: 1px solid var(--color-border); border-radius: var(--radius-btn); }
.ed__vis.is-on { color: var(--color-text-primary); border-color: var(--color-brand); }
.ed__vis:focus-within, .ed__cover:focus-within { outline: 2px solid var(--color-focus); outline-offset: 2px; }
.ed__vis-text { display: flex; flex-direction: column; font-size: var(--text-small); }
.ed__vis-text small { font-size: var(--text-caption); color: var(--color-text-muted); }
.ed__covers { display: grid; grid-template-columns: repeat(auto-fill, minmax(3.25rem, 1fr)); gap: 0.5rem; max-height: 11rem; padding: 2px; overflow-y: auto; }
.ed__cover { position: relative; display: grid; place-items: center; aspect-ratio: 2 / 3; overflow: hidden; color: var(--color-text-muted); cursor: pointer; background: var(--color-surface-3); border: 2px solid transparent; border-radius: var(--radius-btn); }
.ed__cover.is-on { border-color: var(--color-brand); }
.ed__cover--auto { padding: 0.25rem; font-size: 0.6875rem; text-align: center; }
.ed__cover-img { width: 100%; height: 100%; object-fit: cover; }
.ed__actions { display: flex; justify-content: flex-end; gap: 0.5rem; }
</style>
