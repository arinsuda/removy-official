<template>
  <Teleport to="body">
    <Transition name="admin-modal">
      <div
        v-if="modelValue"
        class="modal-backdrop"
        @click.self="handleCancel"
      >
        <div
          class="modal-card"
          role="dialog"
          aria-modal="true"
          :aria-label="title"
        >
          <!-- Header Close Button -->
          <button class="btn-close" type="button" aria-label="Close modal" @click="handleCancel">
            <X :size="18" />
          </button>

          <!-- Icon Badge -->
          <div class="modal-icon-badge" :class="`variant-${variant}`">
            <component :is="resolvedIcon" :size="24" :stroke-width="1.8" />
          </div>

          <!-- Content Header -->
          <div class="modal-header-text">
            <h3 class="modal-title">{{ title }}</h3>
            <p v-if="description" class="modal-desc">
              {{ description }}
            </p>
            <div v-if="targetName" class="target-badge">
              <span class="target-name">{{ targetName }}</span>
            </div>
          </div>

          <!-- Reason Input Box -->
          <div v-if="showReasonInput" class="reason-section">
            <label class="reason-label">
              {{ reasonLabel || 'เหตุผลประกอบ (ไม่จำเป็นต้องระบุ)' }}
            </label>
            <textarea
              v-model="reasonText"
              class="reason-textarea"
              rows="2"
              :placeholder="reasonPlaceholder || 'ระบุเหตุผลในการดำเนินรายการ...'"
            />
          </div>

          <!-- Actions -->
          <div class="modal-actions">
            <button
              type="button"
              class="btn-cancel"
              :disabled="loading"
              @click="handleCancel"
            >
              {{ cancelText || 'ยกเลิก' }}
            </button>
            <button
              type="button"
              class="btn-confirm"
              :class="`variant-${variant}`"
              :disabled="loading || confirmDisabled"
              @click="handleConfirm"
            >
              <AppSpinner v-if="loading" />
              <span>{{ loading ? 'กำลังดำเนินการ...' : (confirmText || 'ยืนยัน') }}</span>
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import AppSpinner from "@/components/ui/AppSpinner.vue"
import { ref, computed, watch, onBeforeUnmount, type Component } from 'vue'
import { X, ShieldAlert, Shield, UserX, UserCheck, Trash2, AlertTriangle } from 'lucide-vue-next'

export type AdminModalVariant = 'danger' | 'warning' | 'info' | 'success'

const props = withDefaults(
  defineProps<{
    modelValue: boolean
    title: string
    description?: string
    targetName?: string
    variant?: AdminModalVariant
    confirmText?: string
    cancelText?: string
    showReasonInput?: boolean
    reasonPlaceholder?: string
    reasonLabel?: string
    confirmDisabled?: boolean
    loading?: boolean
    icon?: Component
  }>(),
  {
    variant: 'danger',
    showReasonInput: true,
    confirmDisabled: false,
    loading: false,
  }
)

const emit = defineEmits<{
  (e: 'confirm', payload: { reason: string }): void
  (e: 'cancel'): void
  (e: 'update:modelValue', value: boolean): void
}>()

const reasonText = ref('')

const resolvedIcon = computed(() => {
  if (props.icon) return props.icon
  switch (props.variant) {
    case 'danger':
      return Trash2
    case 'warning':
      return AlertTriangle
    case 'info':
      return Shield
    case 'success':
      return UserCheck
    default:
      return ShieldAlert
  }
})

function handleCancel() {
  if (props.loading) return
  emit('cancel')
  emit('update:modelValue', false)
}

function handleConfirm() {
  if (props.loading || props.confirmDisabled) return
  emit('confirm', { reason: reasonText.value.trim() })
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape' && props.modelValue && !props.loading) {
    handleCancel()
  }
}

watch(
  () => props.modelValue,
  (val) => {
    if (val) {
      reasonText.value = ''
      document.body.style.overflow = 'hidden'
      window.addEventListener('keydown', onKeydown, { passive: true })
    } else {
      const activeBackdrops = document.querySelectorAll('.modal-backdrop')
      if (activeBackdrops.length <= 1) {
        document.body.style.overflow = ''
      }
      window.removeEventListener('keydown', onKeydown)
    }
  },
  { immediate: true }
)

onBeforeUnmount(() => {
  const activeBackdrops = document.querySelectorAll('.modal-backdrop')
  if (activeBackdrops.length <= 1) {
    document.body.style.overflow = ''
  }
  window.removeEventListener('keydown', onKeydown)
})
</script>

<style scoped>
.modal-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.78);
  backdrop-filter: blur(10px);
  -webkit-backdrop-filter: blur(10px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 99999;
  padding: 1rem;
  will-change: opacity;
  transform: translateZ(0);
  backface-visibility: hidden;
}

.modal-card {
  position: relative;
  background: #18181b;
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 20px;
  padding: 2rem 1.75rem 1.5rem;
  width: 100%;
  max-width: 420px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1.25rem;
  box-shadow:
    0 24px 64px rgba(0, 0, 0, 0.8),
    0 0 0 1px rgba(255, 255, 255, 0.05) inset;
  color: #ffffff;
  will-change: transform, opacity;
  transform: translateZ(0);
  backface-visibility: hidden;
  contain: layout style;
}

.btn-close {
  position: absolute;
  top: 1rem;
  right: 1rem;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.08);
  color: rgba(255, 255, 255, 0.6);
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: background-color 0.2s ease, color 0.2s ease, transform 0.2s ease;
  transform: translateZ(0);
}

.btn-close:hover {
  background: rgba(255, 255, 255, 0.12);
  color: #ffffff;
  transform: scale(1.05) translateZ(0);
}

/* Icon Badge Variants */
.modal-icon-badge {
  width: 56px;
  height: 56px;
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  transform: translateZ(0);
}

.modal-icon-badge.variant-danger {
  background: rgba(245, 67, 44, 0.15);
  color: #ff4d4d;
  border: 1px solid rgba(245, 67, 44, 0.3);
  box-shadow: 0 0 20px rgba(245, 67, 44, 0.15);
}

.modal-icon-badge.variant-warning {
  background: rgba(245, 158, 11, 0.15);
  color: #f59e0b;
  border: 1px solid rgba(245, 158, 11, 0.3);
  box-shadow: 0 0 20px rgba(245, 158, 11, 0.15);
}

.modal-icon-badge.variant-info {
  background: rgba(59, 130, 246, 0.15);
  color: #60a5fa;
  border: 1px solid rgba(59, 130, 246, 0.3);
  box-shadow: 0 0 20px rgba(59, 130, 246, 0.15);
}

.modal-icon-badge.variant-success {
  background: rgba(16, 185, 129, 0.15);
  color: #10b981;
  border: 1px solid rgba(16, 185, 129, 0.3);
  box-shadow: 0 0 20px rgba(16, 185, 129, 0.15);
}

/* Header Text */
.modal-header-text {
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.4rem;
  width: 100%;
}

.modal-title {
  font-size: 1.15rem;
  font-weight: 700;
  color: #ffffff;
  margin: 0;
  letter-spacing: -0.01em;
}

.modal-desc {
  font-size: 0.875rem;
  color: rgba(255, 255, 255, 0.65);
  margin: 0;
  line-height: 1.5;
}

.target-badge {
  display: inline-flex;
  align-items: center;
  padding: 0.25rem 0.6rem;
  background: rgba(255, 255, 255, 0.08);
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 6px;
  margin-top: 0.2rem;
}

.target-name {
  font-size: 0.825rem;
  font-weight: 700;
  color: #ffffff;
}

/* Reason Section */
.reason-section {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
  text-align: left;
}

.reason-label {
  font-size: 0.775rem;
  font-weight: 600;
  color: rgba(255, 255, 255, 0.6);
}

.reason-textarea {
  width: 100%;
  padding: 0.65rem 0.85rem;
  background: rgba(0, 0, 0, 0.35);
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 10px;
  color: #ffffff;
  font-size: 0.85rem;
  outline: none;
  resize: vertical;
  min-height: 54px;
  box-sizing: border-box;
  transition: border-color 0.2s ease, box-shadow 0.2s ease;
  font-family: inherit;
}

.reason-textarea:focus {
  border-color: rgba(255, 255, 255, 0.3);
  box-shadow: 0 0 0 2px rgba(255, 255, 255, 0.1);
}

.reason-textarea::placeholder {
  color: rgba(255, 255, 255, 0.3);
}

/* Actions */
.modal-actions {
  display: flex;
  gap: 0.75rem;
  width: 100%;
  margin-top: 0.25rem;
}

.btn-cancel,
.btn-confirm {
  flex: 1;
  height: 42px;
  border-radius: 10px;
  border: none;
  font-size: 0.875rem;
  font-weight: 600;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.4rem;
  transition: background-color 0.18s ease, transform 0.18s cubic-bezier(0.16, 1, 0.3, 1), box-shadow 0.18s ease;
  transform: translateZ(0);
}

.btn-cancel {
  background: rgba(255, 255, 255, 0.08);
  color: rgba(255, 255, 255, 0.8);
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.btn-cancel:hover:not(:disabled) {
  background: rgba(255, 255, 255, 0.14);
  color: #ffffff;
}

.btn-confirm.variant-danger {
  background: var(--color-brand);
  color: #ffffff;
}

.btn-confirm.variant-danger:hover:not(:disabled) {
  background: var(--color-brand-hover);
  box-shadow: 0 6px 20px rgba(245, 67, 44, 0.4);
  transform: translateY(-1px) translateZ(0);
}

.btn-confirm.variant-info {
  background: #3b82f6;
  color: #ffffff;
}

.btn-confirm.variant-info:hover:not(:disabled) {
  background: #2563eb;
  box-shadow: 0 6px 20px rgba(59, 130, 246, 0.4);
  transform: translateY(-1px) translateZ(0);
}

.btn-confirm.variant-warning {
  background: #f59e0b;
  color: #ffffff;
}

.btn-confirm.variant-warning:hover:not(:disabled) {
  background: #d97706;
  box-shadow: 0 6px 20px rgba(245, 158, 11, 0.4);
  transform: translateY(-1px) translateZ(0);
}

.btn-confirm.variant-success {
  background: #10b981;
  color: #ffffff;
}

.btn-confirm.variant-success:hover:not(:disabled) {
  background: #059669;
  box-shadow: 0 6px 20px rgba(16, 185, 129, 0.4);
  transform: translateY(-1px) translateZ(0);
}

.btn-confirm:active:not(:disabled),
.btn-cancel:active:not(:disabled) {
  transform: translateY(0) scale(0.97) translateZ(0);
}

.btn-confirm:disabled,
.btn-cancel:disabled {
  opacity: 0.5;
  cursor: not-allowed;
  transform: none !important;
  box-shadow: none !important;
}

/* Modal Transition */
.admin-modal-enter-active,
.admin-modal-leave-active {
  transition: opacity 0.2s cubic-bezier(0.16, 1, 0.3, 1);
}

.admin-modal-enter-active .modal-card,
.admin-modal-leave-active .modal-card {
  transition:
    transform 0.22s cubic-bezier(0.16, 1, 0.3, 1),
    opacity 0.2s cubic-bezier(0.16, 1, 0.3, 1);
}

.admin-modal-enter-from,
.admin-modal-leave-to {
  opacity: 0;
}

.admin-modal-enter-from .modal-card {
  transform: scale(0.93) translateY(10px) translateZ(0);
  opacity: 0;
}

.admin-modal-leave-to .modal-card {
  transform: scale(0.96) translateY(4px) translateZ(0);
  opacity: 0;
}
</style>
