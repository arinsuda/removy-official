<template>
  <AppModal :model-value="modelValue" role="alertdialog" size="sm" :title="t('manga.adult.confirmTitle')" @update:model-value="emit('update:modelValue', $event)">
    <div class="adult-confirm">
      <p class="adult-confirm__text">{{ t("manga.adult.confirmBody") }}</p>
      <div class="adult-confirm__actions">
        <AppButton variant="ghost" size="sm" @click="emit('update:modelValue', false)">{{ t("manga.adult.cancel") }}</AppButton>
        <AppButton variant="primary" size="sm" @click="confirm">{{ t("manga.adult.confirm") }}</AppButton>
      </div>
    </div>
  </AppModal>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n"
import AppButton from "@/components/ui/AppButton.vue"
import AppModal from "@/components/ui/AppModal.vue"

defineProps<{ modelValue: boolean }>()
const emit = defineEmits<{ "update:modelValue": [open: boolean]; confirm: [] }>()

const { t } = useI18n()

function confirm() {
  emit("confirm")
  emit("update:modelValue", false)
}
</script>

<style scoped>
.adult-confirm {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}
.adult-confirm__text {
  margin: 0;
  font-size: var(--text-small);
  color: var(--color-text-secondary);
}
.adult-confirm__actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
}
</style>
