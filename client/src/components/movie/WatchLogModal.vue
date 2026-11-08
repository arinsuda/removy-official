<template>
  <AppModal
    v-model="open"
    size="sm"
    :title="t('watchLog.recordWatch')"
    :dismiss-disabled="saving"
    :close-on-escape="!menuOpen"
    @update:model-value="onOpenChange"
  >
    <div class="field">
      <label class="field-label">{{ t('watchLog.watchedDate') }}</label>
      <VueDatePicker
        v-model="watchedOn"
        :max-date="new Date()"
        :enable-time-picker="false"
        format="dd/MM/yyyy"
        dark
        auto-apply
        :placeholder="t('watchLog.watchedDate')"
        teleport
        @open="menuOpen = true"
        @closed="menuOpen = false"
      />
    </div>

    <div class="field">
      <label class="field-label">{{ t('visibility.label') }}</label>
      <VisibilitySelector v-model="visibility" size="md" />
    </div>

    <template #footer>
      <AppButton variant="ghost" :disabled="saving" @click="emit('close')">{{ t('common.cancel') }}</AppButton>
      <AppButton :loading="saving" @click="handleSave">{{ t('watchLog.save') }}</AppButton>
    </template>
  </AppModal>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { VueDatePicker } from '@vuepic/vue-datepicker'
import '@vuepic/vue-datepicker/dist/main.css'
import AppButton from '@/components/ui/AppButton.vue'
import AppModal from '@/components/ui/AppModal.vue'
import VisibilitySelector from '@/components/common/VisibilitySelector.vue'
import { watchLogApi } from '@/api/endpoints/watchLog'
import type { Visibility } from '@/types'
import { useToast } from '@/composables/useToast'

const props = defineProps<{
  mediaId: number
  mediaType: "movie" | "tv"
}>()

const emit = defineEmits<{
  close: []
  logged: []
}>()

const { t } = useI18n()
const { success: toastSuccess, error: toastError } = useToast()

// Parents mount this with v-if, so it is open for its whole life.
const open = ref(true)
const menuOpen = ref(false)
const saving = ref(false)
const watchedOn = ref<Date>(new Date())
const visibility = ref<Visibility>('public')

function onOpenChange(value: boolean) {
  if (!value) emit('close')
}

function toDateString(d: Date): string {
  const yyyy = d.getFullYear()
  const mm = String(d.getMonth() + 1).padStart(2, '0')
  const dd = String(d.getDate()).padStart(2, '0')
  return `${yyyy}-${mm}-${dd}`
}

async function handleSave() {
  if (saving.value) return
  saving.value = true
  try {
    await watchLogApi.createWatchLog(props.mediaType, props.mediaId, {
      watched_on: toDateString(watchedOn.value),
      visibility: visibility.value,
    })
  } catch (err) {
    console.error('Failed to create watch log:', err)
    toastError(t('watchLog.createFailed'))
    saving.value = false
    return
  }
  toastSuccess(t('watchLog.createSuccess'))
  saving.value = false
  emit('logged')
  emit('close')
}
</script>

<style scoped>
.field {
  margin-bottom: 20px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.field-label {
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--color-text-muted);
  text-transform: uppercase;
}
</style>
