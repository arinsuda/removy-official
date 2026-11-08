<template>
  <AppModal
    :model-value="modelValue"
    role="alertdialog"
    size="sm"
    :show-close="false"
    :dismiss-disabled="loading"
    :aria-labelledby="titleId"
    :aria-describedby="description ? descId : undefined"
    @close="emit('cancel')"
    @update:model-value="emit('update:modelValue', $event)"
  >
    <div class="admin-confirm">
      <div class="admin-confirm__badge" :class="`admin-confirm__badge--${variant}`">
        <component :is="resolvedIcon" :size="24" :stroke-width="1.8" aria-hidden="true" />
      </div>

      <div class="admin-confirm__text">
        <h3 :id="titleId" class="admin-confirm__title">{{ title }}</h3>
        <p v-if="description" :id="descId" class="admin-confirm__desc">{{ description }}</p>
        <span v-if="targetName" class="admin-confirm__target">{{ targetName }}</span>
      </div>

      <div v-if="showReasonInput" class="admin-confirm__reason-wrap">
        <label :for="reasonId" class="admin-confirm__label">
          {{ reasonLabel || t("admin.confirmModal.reasonLabel") }}
        </label>
        <textarea
          :id="reasonId"
          v-model="reasonText"
          class="admin-confirm__reason"
          rows="2"
          :placeholder="reasonPlaceholder || t('admin.confirmModal.reasonPlaceholder')"
        />
      </div>
    </div>

    <template #footer>
      <div class="admin-confirm__actions">
        <AppButton variant="outline" :disabled="loading" @click="handleCancel">
          {{ cancelText || t("common.cancel") }}
        </AppButton>
        <AppButton
          :variant="variant === 'danger' ? 'danger' : 'primary'"
          :loading="loading"
          :disabled="confirmDisabled"
          @click="handleConfirm"
        >
          {{ loading ? t("admin.confirmModal.processing") : confirmText || t("common.confirm") }}
        </AppButton>
      </div>
    </template>
  </AppModal>
</template>

<script setup lang="ts">
import AppButton from "@/components/ui/AppButton.vue"
import AppModal from "@/components/ui/AppModal.vue"
import { ref, computed, watch, useId, type Component } from "vue"
import { useI18n } from "vue-i18n"
import { ShieldAlert, Shield, UserCheck, Trash2, AlertTriangle } from "lucide-vue-next"

export type AdminModalVariant = "danger" | "warning" | "info" | "success"

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
    variant: "danger",
    showReasonInput: true,
    confirmDisabled: false,
    loading: false,
  },
)

const emit = defineEmits<{
  (e: "confirm", payload: { reason: string }): void
  (e: "cancel"): void
  (e: "update:modelValue", value: boolean): void
}>()

const { t } = useI18n()
const titleId = useId()
const descId = useId()
const reasonId = useId()
const reasonText = ref("")

const icons: Record<AdminModalVariant, Component> = {
  danger: Trash2,
  warning: AlertTriangle,
  info: Shield,
  success: UserCheck,
}
const resolvedIcon = computed(() => props.icon ?? icons[props.variant] ?? ShieldAlert)

// AppModal handles Escape/backdrop and emits close then update:modelValue itself.
function handleCancel() {
  if (props.loading) return
  emit("cancel")
  emit("update:modelValue", false)
}

function handleConfirm() {
  if (props.loading || props.confirmDisabled) return
  emit("confirm", { reason: reasonText.value.trim() })
}

watch(
  () => props.modelValue,
  (open) => {
    if (open) reasonText.value = ""
  },
  { immediate: true },
)
</script>

<style scoped>
.admin-confirm {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1.25rem;
  padding-top: 0.5rem;
}

.admin-confirm__badge {
  --tone: var(--color-danger);
  display: flex;
  align-items: center;
  justify-content: center;
  width: 56px;
  height: 56px;
  border-radius: 16px;
  color: var(--tone);
  background: color-mix(in srgb, var(--tone) 15%, transparent);
  border: 1px solid color-mix(in srgb, var(--tone) 30%, transparent);
}

.admin-confirm__badge--warning {
  --tone: var(--color-warning);
}

.admin-confirm__badge--info {
  --tone: var(--color-info);
}

.admin-confirm__badge--success {
  --tone: var(--color-success);
}

.admin-confirm__text {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.4rem;
  width: 100%;
  text-align: center;
}

.admin-confirm__title {
  margin: 0;
  font-size: 1.15rem;
  font-weight: 700;
  color: var(--color-text-primary);
}

.admin-confirm__desc {
  margin: 0;
  font-size: 0.875rem;
  line-height: 1.5;
  color: var(--color-text-secondary);
}

.admin-confirm__target {
  padding: 0.25rem 0.6rem;
  background: var(--color-surface-3);
  border: 1px solid var(--color-border-strong);
  border-radius: 6px;
  font-size: 0.825rem;
  font-weight: 700;
  color: var(--color-text-primary);
}

.admin-confirm__reason-wrap {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
  width: 100%;
}

.admin-confirm__label {
  font-size: 0.775rem;
  font-weight: 600;
  color: var(--color-text-secondary);
}

.admin-confirm__reason {
  width: 100%;
  min-height: 54px;
  padding: 0.65rem 0.85rem;
  box-sizing: border-box;
  background: var(--color-surface);
  border: 1px solid var(--color-border-strong);
  border-radius: 10px;
  color: var(--color-text-primary);
  font-family: var(--font-sans);
  font-size: 0.85rem;
  resize: vertical;
}

.admin-confirm__reason::placeholder {
  color: var(--color-text-muted);
}

.admin-confirm__actions {
  display: flex;
  gap: 0.75rem;
  width: 100%;
}

.admin-confirm__actions > * {
  flex: 1;
}
</style>
