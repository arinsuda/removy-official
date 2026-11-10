<template>
  <AppModal :model-value="true" size="sm" :title="t('reviews.delete.title')" @close="emit('close')">
    <div class="dr">
      <span class="dr__icon" aria-hidden="true"><Trash2 :size="20" :stroke-width="1.6" /></span>
      <p class="dr__media">{{ review.media.title }}</p>
      <p class="dr__text">{{ t("reviews.delete.message") }}</p>
    </div>
    <template #footer>
      <AppButton variant="outline" @click="emit('close')">{{ t("common.cancel") }}</AppButton>
      <AppButton variant="danger" :loading="deleting" :disabled="deleting" @click="handleDelete">{{ t("reviews.delete.confirm") }}</AppButton>
    </template>
  </AppModal>
</template>

<script setup lang="ts">
import { ref } from "vue"
import { useI18n } from "vue-i18n"
import { Trash2 } from "lucide-vue-next"
import { reviewApi } from "@/api/api"
import AppButton from "@/components/ui/AppButton.vue"
import AppModal from "@/components/ui/AppModal.vue"
import { useUIStore } from "@/stores/ui"
import type { ReviewResponse } from "@/types"
import { logApiError } from "@/utils/logApiError"

const props = defineProps<{ userId: number; review: ReviewResponse }>()
const emit = defineEmits<{ close: []; deleted: [id: number] }>()

const { t } = useI18n()
const ui = useUIStore()
const deleting = ref(false)

async function handleDelete() {
  if (deleting.value) return
  deleting.value = true
  try {
    await reviewApi.deleteReview(props.review.id)
    emit("deleted", props.review.id)
  } catch (err) {
    logApiError("deleteReview", err)
    ui.addToast(t("reviews.errors.deleteFailed"), "error")
  } finally {
    deleting.value = false
  }
}
</script>

<style scoped>
.dr { display: flex; flex-direction: column; align-items: center; gap: 0.5rem; text-align: center; }
.dr__icon { display: grid; place-items: center; width: 2.75rem; height: 2.75rem; margin-bottom: 0.25rem; color: var(--color-danger); background: color-mix(in srgb, var(--color-danger) 12%, transparent); border: 1px solid color-mix(in srgb, var(--color-danger) 24%, transparent); border-radius: 999px; }
.dr__media { margin: 0; font-weight: 700; color: var(--color-text-primary); }
.dr__text { margin: 0; font-size: var(--text-small); line-height: 1.6; color: var(--color-text-secondary); }
</style>
