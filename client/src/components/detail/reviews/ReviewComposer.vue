<template>
  <form class="composer" :aria-label="t('reviews.composer.title')" @submit.prevent="onSubmit">
    <h3 class="composer__title">{{ t("reviews.composer.title") }}</h3>

    <div class="composer__rating">
      <span class="composer__label">{{ t("reviews.composer.ratingLabel") }}</span>
      <StarRatingInput v-model="form.rating.value" :size="28" :aria-label="t('reviews.composer.ratingLabel')" />
      <span v-if="form.rating.value === 0" class="composer__hint">{{ t("reviews.composer.ratingRequired") }}</span>
    </div>

    <textarea
      ref="textarea"
      v-model="form.body.value"
      class="composer__input"
      rows="3"
      :aria-label="t('reviews.composer.bodyLabel')"
      :placeholder="t(`reviews.composer.placeholder.${mediaType}`)"
    />

    <div class="composer__footer">
      <VisibilitySelector v-model="form.visibility.value" size="sm" />
      <AppButton type="submit" :loading="form.isSubmitting.value" :disabled="!form.canSubmit.value">
        {{ t("reviews.submit") }}
      </AppButton>
    </div>
  </form>
</template>

<script setup lang="ts">
import { useTemplateRef } from "vue"
import { useI18n } from "vue-i18n"
import AppButton from "@/components/ui/AppButton.vue"
import VisibilitySelector from "@/components/common/VisibilitySelector.vue"
import StarRatingInput from "@/components/movie/Starratinginput.vue"
import { useReviewForm } from "@/composables/useReviewForm"
import type { ReviewResponse } from "@/types/review"

const props = defineProps<{ mediaId: number; mediaType: "movie" | "tv" }>()
const emit = defineEmits<{ submitted: [review: ReviewResponse] }>()

const { t } = useI18n()
const textarea = useTemplateRef<HTMLTextAreaElement>("textarea")
// Getters keep the form pointed at the title shown now, not the first one mounted.
const form = useReviewForm(() => props.mediaId, () => props.mediaType)

async function onSubmit() {
  const created = await form.submit()
  if (created) emit("submitted", created)
}

defineExpose({ focus: () => textarea.value?.focus() })
</script>

<style scoped>
.composer { display: flex; flex-direction: column; gap: 0.75rem; padding: 1rem; background: var(--color-surface-2); border: 1px solid var(--color-border); border-radius: var(--radius-card); }
.composer__title { font-size: var(--text-body); font-weight: 700; color: var(--color-text-primary); }
.composer__rating { display: flex; flex-wrap: wrap; align-items: center; gap: 0.25rem 0.75rem; }
.composer__label { font-size: var(--text-small); color: var(--color-text-secondary); }
.composer__hint { font-size: var(--text-caption); color: var(--color-text-muted); }
.composer__input { width: 100%; padding: 0.75rem; font: inherit; font-size: var(--text-small); color: var(--color-text-primary); background: var(--color-surface); border: 1px solid var(--color-border-strong); border-radius: var(--radius-card); resize: vertical; box-sizing: border-box; }
.composer__footer { display: flex; flex-wrap: wrap; align-items: center; justify-content: space-between; gap: 0.75rem; }
</style>
