<template>
  <AppModal
    :model-value="modelValue"
    role="alertdialog"
    size="sm"
    :show-close="false"
    :aria-labelledby="titleId"
    :aria-describedby="descId"
    @close="emit('cancel')"
  >
    <div class="confirm-body">
      <div class="confirm-icon" :class="`confirm-icon--${config.tone}`">
        <component :is="config.icon" :size="22" :stroke-width="1.6" aria-hidden="true" />
      </div>
      <h3 :id="titleId" class="confirm-title">{{ t(`${prefix}.title`) }}</h3>
      <p :id="descId" class="confirm-desc">
        {{ t(`${prefix}.description`) }}
        <strong class="confirm-item">{{ itemName }}</strong>
        {{ t(`${prefix}.suffix`) }}
      </p>
    </div>

    <template #footer>
      <div class="confirm-actions">
        <AppButton variant="outline" @click="emit('cancel')">{{ t("common.cancel") }}</AppButton>
        <AppButton :variant="config.variant" :disabled="confirmDisabled" @click="emit('confirm')">
          {{ confirmDisabled ? t("profile.confirmModal.processing") : t(`${prefix}.confirm`) }}
        </AppButton>
      </div>
    </template>
  </AppModal>
</template>

<script setup lang="ts">
  import { computed, useId } from "vue"
  import { useI18n } from "vue-i18n"
  import { Heart, Bookmark, Eye, Mail, UserMinus, Trash2, Unlink } from "lucide-vue-next"
  import AppButton from "@/components/ui/AppButton.vue"
  import AppModal from "@/components/ui/AppModal.vue"
  import type { ListType } from "@/types"

  type ConfirmModalType = ListType | "email_change" | "unfollow" | "bmol_delete" | "custom_list_delete" | "unlink_account"

  const props = defineProps<{
    modelValue: boolean
    listType: ConfirmModalType
    itemName?: string
    confirmDisabled?: boolean
  }>()

  const emit = defineEmits<{
    (e: "confirm"): void
    (e: "cancel"): void
    (e: "update:modelValue", value: boolean): void
  }>()

  const { t } = useI18n()
  const titleId = useId()
  const descId = useId()

  const configs: Record<
    ConfirmModalType,
    { icon: unknown; tone: "danger" | "warning" | "info"; variant: "danger" | "outline" | "primary" }
  > = {
    likes: { icon: Heart, tone: "danger", variant: "danger" },
    watchlist: { icon: Bookmark, tone: "warning", variant: "danger" },
    watched: { icon: Eye, tone: "info", variant: "outline" },
    email_change: { icon: Mail, tone: "info", variant: "primary" },
    unfollow: { icon: UserMinus, tone: "danger", variant: "danger" },
    bmol_delete: { icon: Trash2, tone: "danger", variant: "danger" },
    custom_list_delete: { icon: Trash2, tone: "danger", variant: "danger" },
    unlink_account: { icon: Unlink, tone: "danger", variant: "danger" },
  }

  const config = computed(() => configs[props.listType])
  const prefix = computed(() => `profile.confirmModal.${props.listType}`)
</script>

<style scoped>
  .confirm-body {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 12px;
    padding-top: 12px;
    text-align: center;
  }

  .confirm-icon {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 52px;
    height: 52px;
    border-radius: 14px;
    border: 1px solid color-mix(in srgb, var(--tone) 20%, transparent);
    background: color-mix(in srgb, var(--tone) 12%, transparent);
    color: var(--tone);
  }
  .confirm-icon--danger {
    --tone: var(--color-danger);
  }
  .confirm-icon--warning {
    --tone: var(--color-warning);
  }
  .confirm-icon--info {
    --tone: var(--color-info);
  }

  .confirm-title {
    margin: 0;
    font-size: 1rem;
    font-weight: 650;
    color: var(--color-text-primary);
  }

  .confirm-desc {
    margin: 0;
    font-size: 0.82rem;
    line-height: 1.5;
    color: var(--color-text-muted);
  }

  .confirm-item {
    font-weight: 600;
    color: var(--color-text-secondary);
  }

  .confirm-actions {
    display: flex;
    gap: 8px;
    width: 100%;
  }
  .confirm-actions > * {
    flex: 1;
  }
</style>
