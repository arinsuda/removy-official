<template>
  <Teleport to="body">
    <Transition name="modal">
      <div class="modal-backdrop" @click.self="$emit('close')">
        <div class="modal" role="dialog" aria-modal="true" :aria-label="t('watchLog.recordWatch')">
          <div class="modal-header">
            <h2 class="modal-title">{{ t('watchLog.recordWatch') }}</h2>
            <button class="modal-close" :aria-label="t('common.close')" @click="$emit('close')">
              <X class="icon" />
            </button>
          </div>

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
            />
          </div>

          <div class="field">
            <label class="field-label">{{ t('visibility.label') }}</label>
            <VisibilitySelector v-model="visibility" size="md" />
          </div>

          <div class="modal-footer">
            <button class="btn btn--ghost" @click="$emit('close')">{{ t('common.cancel') }}</button>
            <button class="btn btn--primary" :disabled="saving" @click="handleSave">
              {{ saving ? t('watchLog.saving') : t('watchLog.save') }}
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import { X } from 'lucide-vue-next'
import { VueDatePicker } from '@vuepic/vue-datepicker'
import '@vuepic/vue-datepicker/dist/main.css'
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

const saving = ref(false)
const watchedOn = ref<Date>(new Date())
const visibility = ref<Visibility>('public')

function toDateString(d: Date): string {
  const yyyy = d.getFullYear()
  const mm = String(d.getMonth() + 1).padStart(2, '0')
  const dd = String(d.getDate()).padStart(2, '0')
  return `${yyyy}-${mm}-${dd}`
}

async function handleSave() {
  if (saving.value) return
  saving.value = true
  const data = {
    watched_on: toDateString(watchedOn.value),
    visibility: visibility.value
  }

  // Close modal immediately and notify parent for smooth, instant UX
  emit('logged')
  emit('close')

  try {
    await watchLogApi.createWatchLog(props.mediaType, props.mediaId, data)
    toastSuccess(t('watchLog.createSuccess'))
  } catch (err) {
    console.error('Failed to create watch log:', err)
    toastError(t('watchLog.createFailed'))
  } finally {
    saving.value = false
  }
}

onMounted(() => { document.body.style.overflow = 'hidden' })
onBeforeUnmount(() => { document.body.style.overflow = '' })
</script>

<style scoped>
.modal-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.72);
  backdrop-filter: blur(6px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 200;
  padding: 16px;
}
.modal {
  background: #161616;
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 14px;
  width: 100%;
  max-width: 400px;
  box-shadow: 0 24px 64px rgba(0, 0, 0, 0.6);
  font-family: var(--font-sans);
}
.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px;
}
.modal-title {
  font-size: 1.1rem;
  font-weight: 600;
  color: #fff;
  margin: 0;
}
.modal-close {
  background: transparent;
  border: none;
  color: #8a8a8e;
  cursor: pointer;
}
.modal-close:hover {
  color: #fff;
}
.icon {
  width: 1.25rem;
  height: 1.25rem;
}
.field {
  padding: 0 20px;
  margin-bottom: 20px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.field-label {
  font-size: 0.75rem;
  font-weight: 600;
  color: #8a8a8e;
  text-transform: uppercase;
}
.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  padding: 16px 20px;
  border-top: 1px solid rgba(255, 255, 255, 0.08);
}
.btn {
  font-size: 0.875rem;
  font-weight: 600;
  padding: 8px 16px;
  border-radius: 8px;
  border: 1px solid transparent;
  cursor: pointer;
}
.btn--ghost {
  color: #8a8a8e;
  background: transparent;
}
.btn--ghost:hover {
  color: #fff;
  background: rgba(255, 255, 255, 0.05);
}
.btn--primary {
  color: #000;
  background: #fff;
}
.btn--primary:hover:not(:disabled) {
  background: #e0e0e0;
}
.btn--primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
.modal-enter-active, .modal-leave-active {
  transition: all 0.2s;
}
.modal-enter-from, .modal-leave-to {
  opacity: 0;
  transform: scale(0.95);
}
</style>
