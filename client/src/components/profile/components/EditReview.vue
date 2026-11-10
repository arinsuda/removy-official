<template>
  <AppModal :model-value="true" size="sm" :title="t('reviews.editTitle')" @close="emit('close')">
    <div class="er">
      <div class="er__media">
        <span class="er__poster">
          <img v-if="review.media.poster_url" :src="review.media.poster_url" :alt="review.media.title" />
          <Film v-else :size="14" :stroke-width="1.4" aria-hidden="true" />
        </span>
        <span class="er__title">{{ review.media.title }}</span>
      </div>

      <div class="er__field">
        <span class="er__label">{{ t("reviews.composer.ratingLabel") }}</span>
        <StarRatingInput v-model="form.rating" :size="26" :aria-label="t('reviews.composer.ratingLabel')" />
      </div>

      <label class="er__field">
        <span class="er__label">{{ t("reviews.composer.bodyLabel") }}</span>
        <textarea v-model="form.body" class="er__textarea" rows="5" :placeholder="t(`reviews.composer.placeholder.${placeholderType}`)" />
      </label>

      <div class="er__field">
        <span class="er__label">{{ t("reviews.composer.visibilityLabel") }}</span>
        <VisibilitySelector v-model="form.visibility" size="sm" />
      </div>
    </div>

    <template #footer>
      <AppButton variant="outline" @click="emit('close')">{{ t("common.cancel") }}</AppButton>
      <AppButton :loading="saving" :disabled="saving || !form.rating" @click="handleSave">{{ t("reviews.update") }}</AppButton>
    </template>
  </AppModal>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from "vue"
import { useI18n } from "vue-i18n"
import { Film } from "lucide-vue-next"
import { reviewApi } from "@/api/api"
import VisibilitySelector from "@/components/common/VisibilitySelector.vue"
import StarRatingInput from "@/components/movie/Starratinginput.vue"
import AppButton from "@/components/ui/AppButton.vue"
import AppModal from "@/components/ui/AppModal.vue"
import { useUIStore } from "@/stores/ui"
import type { ReviewResponse, Visibility } from "@/types"
import { logApiError } from "@/utils/logApiError"

const props = defineProps<{ userId: number; review: ReviewResponse }>()
const emit = defineEmits<{ close: []; saved: [review: ReviewResponse] }>()

const { t } = useI18n()
const ui = useUIStore()
const saving = ref(false)
const form = reactive({
  rating: props.review.rating,
  body: props.review.body ?? "",
  visibility: (props.review.visibility || "public") as Visibility,
})
const placeholderType = computed(() => (props.review.media.media_type === "tv" ? "tv" : "movie"))

async function handleSave() {
  if (saving.value) return
  saving.value = true
  try {
    const res = await reviewApi.updateReview(props.review.id, { ...form })
    emit("saved", res.data.review)
  } catch (err) {
    logApiError("updateReview", err)
    ui.addToast(t("reviews.errors.updateFailed"), "error")
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.er { display: flex; flex-direction: column; gap: 1.25rem; }
.er__media { display: flex; align-items: center; gap: 0.75rem; padding: 0.625rem; background: var(--color-surface-3); border: 1px solid var(--color-border); border-radius: var(--radius-btn); }
.er__poster { display: grid; flex: none; place-items: center; width: 2.25rem; height: 3.25rem; overflow: hidden; color: var(--color-text-muted); background: var(--color-surface-2); border-radius: var(--radius-btn); }
.er__poster img { width: 100%; height: 100%; object-fit: cover; }
.er__title { min-width: 0; overflow: hidden; font-size: var(--text-small); font-weight: 600; color: var(--color-text-primary); text-overflow: ellipsis; white-space: nowrap; }
.er__field { display: flex; flex-direction: column; gap: 0.5rem; }
.er__label { font-size: var(--text-caption); font-weight: 600; color: var(--color-text-secondary); }
.er__textarea { width: 100%; min-height: 7.5rem; padding: 0.75rem; font: inherit; font-size: var(--text-small); line-height: 1.6; color: var(--color-text-primary); background: var(--color-surface-3); border: 1px solid var(--color-border); border-radius: var(--radius-btn); resize: vertical; }
.er__textarea::placeholder { color: var(--color-text-muted); }
</style>
