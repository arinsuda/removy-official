<template>
  <AppModal :model-value="modelValue" size="sm" :title="t('manga.progress.editTitle')" @update:model-value="emit('update:modelValue', $event)">
    <form class="manga-progress-edit" novalidate data-test="manga-progress-edit-form" @submit.prevent="submit">
      <AppInput
        v-model="draftChapter"
        type="number"
        inputmode="numeric"
        min="0"
        step="1"
        :max="chapterMax"
        :label="t('manga.progress.chapter')"
        :error="chapterError"
      />
      <AppInput
        v-model="draftVolume"
        type="number"
        inputmode="numeric"
        min="0"
        step="1"
        :max="volumeMax"
        :label="t('manga.progress.volume')"
        :error="volumeError"
      />
      <p v-if="emptyError" class="manga-progress-edit__error" role="alert">{{ emptyError }}</p>
      <div class="manga-progress-edit__actions">
        <AppButton variant="ghost" size="sm" @click="emit('update:modelValue', false)">{{ t("common.cancel") }}</AppButton>
        <AppButton type="submit" variant="primary" size="sm" :disabled="!canSave">{{ t("common.save") }}</AppButton>
      </div>
    </form>
  </AppModal>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue"
import { useI18n } from "vue-i18n"
import AppButton from "@/components/ui/AppButton.vue"
import AppInput from "@/components/ui/AppInput.vue"
import AppModal from "@/components/ui/AppModal.vue"
import { parseProgressField, progressCap } from "@/components/manga/mangaProgress"
import type { ReadingProgressInput } from "@/types/readingProgress"

const props = defineProps<{
  modelValue: boolean
  values: ReadingProgressInput
}>()

const emit = defineEmits<{
  "update:modelValue": [open: boolean]
  save: [values: ReadingProgressInput]
}>()

const { t, n } = useI18n()
// AniList's total can lag behind a real release, so the dialog only enforces the server cap.
const chapterMax = progressCap("chapter")
const volumeMax = progressCap("volume")

const draftChapter = ref("")
const draftVolume = ref("")

// Each time the dialog opens it starts from the saved position.
watch(
  () => props.modelValue,
  (open) => {
    if (!open) return
    draftChapter.value = String(props.values.chapter)
    draftVolume.value = String(props.values.volume)
  },
)

const chapter = computed(() => parseProgressField(draftChapter.value, chapterMax))
const volume = computed(() => parseProgressField(draftVolume.value, volumeMax))

const chapterError = computed(() =>
  chapter.value === null ? t("manga.progress.editInvalid", { max: n(chapterMax) }) : "",
)
const volumeError = computed(() =>
  volume.value === null ? t("manga.progress.editInvalid", { max: n(volumeMax) }) : "",
)
// The server refuses (0, 0), so the dialog says so instead of sending it.
const emptyError = computed(() => (chapter.value === 0 && volume.value === 0 ? t("manga.progress.editEmpty") : ""))
const canSave = computed(() => chapter.value !== null && volume.value !== null && !emptyError.value)

function submit() {
  const nextChapter = chapter.value
  const nextVolume = volume.value
  if (nextChapter === null || nextVolume === null || emptyError.value) return
  emit("save", { chapter: nextChapter, volume: nextVolume })
  emit("update:modelValue", false)
}
</script>

<style scoped>
.manga-progress-edit {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}
.manga-progress-edit__error {
  margin: 0;
  font-size: var(--text-small);
  color: var(--color-danger);
}
.manga-progress-edit__actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
}
</style>
