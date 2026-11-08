<template>
  <div class="app-state" role="alert">
    <p class="app-state__title">{{ title }}</p>
    <p v-if="description" class="app-state__desc">{{ description }}</p>
    <AppButton variant="outline" size="sm" class="app-state__action" @click="emit('retry')">
      {{ retryLabel ?? t("common.retry") }}
    </AppButton>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n"
import AppButton from "@/components/ui/AppButton.vue"

defineProps<{ title: string; description?: string; retryLabel?: string }>()
const emit = defineEmits<{ retry: [] }>()
const { t } = useI18n()
</script>

<style scoped>
.app-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
  padding: 4rem 1rem;
  text-align: center;
}
.app-state__title {
  font-size: 0.9375rem;
  font-weight: 600;
  color: var(--color-danger);
}
.app-state__desc {
  max-width: 28rem;
  font-size: 0.8125rem;
  color: var(--color-text-muted);
}
.app-state__action {
  margin-top: 0.5rem;
}
</style>
