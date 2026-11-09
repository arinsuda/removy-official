<template>
  <AppModal v-model="open" size="sm" :title="t('library.bmol.note.title')" :dismiss-disabled="saving" initial-focus="textarea">
    <p class="note__for">{{ mediaTitle }}</p>
    <label class="note__label" :for="fieldId">{{ t("library.bmol.note.label") }}</label>
    <textarea
      :id="fieldId"
      v-model="draft"
      class="note__field"
      rows="4"
      :placeholder="t('library.bmol.note.placeholder')"
      :aria-describedby="counterId"
      @keydown.ctrl.enter.prevent="submit"
      @keydown.meta.enter.prevent="submit"
    />
    <p :id="counterId" class="note__count" :class="{ 'is-over': over }" aria-live="polite">
      {{ t("library.bmol.note.counter", { count, max: MAX_NOTE }) }}
    </p>
    <template #footer>
      <div class="note__actions">
        <AppButton variant="outline" :disabled="saving" @click="open = false">{{ t("common.cancel") }}</AppButton>
        <AppButton :loading="saving" :disabled="over || unchanged" @click="submit">{{ t("common.save") }}</AppButton>
      </div>
    </template>
  </AppModal>
</template>

<script setup lang="ts">
import { computed, ref, useId, watch } from "vue"
import { useI18n } from "vue-i18n"
import AppButton from "@/components/ui/AppButton.vue"
import AppModal from "@/components/ui/AppModal.vue"

const MAX_NOTE = 280

const props = defineProps<{ modelValue: boolean; mediaTitle: string; note: string; saving?: boolean }>()
const emit = defineEmits<{ "update:modelValue": [value: boolean]; save: [note: string] }>()

const { t } = useI18n()
const fieldId = useId()
const counterId = useId()
const draft = ref(props.note)
const open = computed({ get: () => props.modelValue, set: (v) => emit("update:modelValue", v) })

watch(() => props.modelValue, (visible) => { if (visible) draft.value = props.note })

// Count characters the way the server does (runes), so Thai and emoji are not over-counted.
const count = computed(() => Array.from(draft.value.trim()).length)
const over = computed(() => count.value > MAX_NOTE)
const unchanged = computed(() => draft.value.trim() === props.note)

function submit() {
  if (over.value || unchanged.value) return
  emit("save", draft.value.trim())
}
</script>

<style scoped>
.note__for { margin-bottom: 0.75rem; font-size: var(--text-small); font-weight: 700; color: var(--color-text-primary); }
.note__label { display: block; margin-bottom: 0.375rem; font-size: var(--text-caption); color: var(--color-text-secondary); }
.note__field {
  width: 100%;
  padding: 0.625rem 0.75rem;
  font: inherit;
  font-size: var(--text-small);
  color: var(--color-text-primary);
  resize: vertical;
  background: var(--color-surface);
  border: 1px solid var(--color-border-strong);
  border-radius: var(--radius-card);
}
.note__count { margin-top: 0.375rem; font-size: var(--text-caption); text-align: right; color: var(--color-text-muted); font-variant-numeric: tabular-nums; }
.note__count.is-over { color: var(--color-danger); }
.note__actions { display: flex; justify-content: flex-end; gap: 0.5rem; width: 100%; }
</style>
