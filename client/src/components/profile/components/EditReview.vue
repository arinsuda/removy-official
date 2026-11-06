<template>
  <Teleport to="body">
    <Transition name="modal">
      <div class="modal-backdrop" @click.self="$emit('close')">
        <div
          class="modal"
          role="dialog"
          aria-modal="true"
          aria-labelledby="edit-modal-title"
        >
          <!-- Header -->
          <div class="modal-header">
            <div class="modal-title-group">
              <span class="modal-eyebrow">Review</span>
              <h2 id="edit-modal-title" class="modal-title">Edit Review</h2>
            </div>
            <button
              class="modal-close"
              @click="$emit('close')"
              aria-label="Close"
            >
              <X :size="14" />
            </button>
          </div>

          <!-- Movie info -->
          <div class="movie-row">
            <div class="movie-poster">
              <img
                v-if="review.media.poster_url"
                :src="review.media.poster_url"
                :alt="review.media.title"
              />
              <Film v-else :size="14" :stroke-width="1.4" />
            </div>
            <span class="movie-title">{{ review.media.title }}</span>
          </div>

          <!-- Rating -->
          <div class="field">
            <label class="field-label">Rating</label>
            <div class="star-picker">
              <div
                v-for="n in 5"
                :key="n"
                class="star-slot"
                @mouseleave="hoverRating = 0"
              >
                <div
                  class="half-zone left"
                  @click="form.rating = n - 0.5"
                  @mouseenter="hoverRating = n - 0.5"
                />
                <div
                  class="half-zone right"
                  @click="form.rating = n"
                  @mouseenter="hoverRating = n"
                />
                <!-- Empty star base -->
                <svg
                  class="star-base"
                  viewBox="0 0 24 24"
                  xmlns="http://www.w3.org/2000/svg"
                >
                  <path
                    d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"
                    stroke="currentColor"
                    stroke-width="1.5"
                    stroke-linejoin="round"
                    fill="none"
                  />
                </svg>
                <!-- Filled star with clip -->
                <svg
                  class="star-fill"
                  viewBox="0 0 24 24"
                  xmlns="http://www.w3.org/2000/svg"
                >
                  <defs>
                    <clipPath :id="`clip-edit-${n}-half`">
                      <rect x="0" y="0" width="12" height="24" />
                    </clipPath>
                    <clipPath :id="`clip-edit-${n}-full`">
                      <rect x="0" y="0" width="24" height="24" />
                    </clipPath>
                  </defs>
                  <path
                    d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"
                    fill="currentColor"
                    :clip-path="
                      starClipPath(n, hoverRating || form.rating, 'edit')
                    "
                    :class="starFillClass(n, hoverRating || form.rating)"
                  />
                </svg>
              </div>
              <span class="rating-label">{{
                (hoverRating || form.rating).toFixed(1)
              }}</span>
            </div>
          </div>

          <!-- Body -->
          <div class="field">
            <label class="field-label" for="review-body">Review</label>
            <textarea
              id="review-body"
              v-model="form.body"
              class="textarea"
              rows="5"
              placeholder="What did you think?"
            />
            <span class="char-count">{{ form.body.length }} / 1000</span>
          </div>

          <div class="field">
            <label class="field-label">Visibility</label>
            <VisibilitySelector v-model="form.visibility" size="sm" />
          </div>

          <!-- Footer -->
          <div class="modal-footer">
            <button class="btn btn--ghost" @click="$emit('close')">
              Cancel
            </button>
            <button
              class="btn btn--primary"
              :disabled="saving"
              @click="handleSave"
            >
              <AppSpinner v-if="saving" />
              <span v-else>Save Changes</span>
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import AppSpinner from "@/components/ui/AppSpinner.vue"
  import { reactive, ref, onMounted, onBeforeUnmount } from "vue"
  import { Star, Film, X } from "lucide-vue-next"
  import { reviewApi } from "@/api/api"
  import type { ReviewResponse } from "@/types"
  import type { Visibility } from "@/types"
  import VisibilitySelector from "@/components/common/VisibilitySelector.vue"

  const props = defineProps<{ userId: number; review: ReviewResponse }>()
  const emit = defineEmits<{
    close: []
    saved: [review: ReviewResponse]
  }>()

  const saving = ref(false)
  const hoverRating = ref(0)

  const form = reactive({
    rating: props.review.rating,
    body: props.review.body ?? "",
    visibility: props.review.visibility || "public",
  })

  function starFillClass(n: number, rating: number) {
    if (rating >= n - 0.5) return "fill--on"
    return "fill--off"
  }

  function starClipPath(n: number, rating: number, prefix: string) {
    if (rating >= n) return `url(#clip-${prefix}-${n}-full)`
    if (rating >= n - 0.5) return `url(#clip-${prefix}-${n}-half)`
    return "none"
  }

  async function handleSave() {
    if (saving.value) return
    try {
      saving.value = true
      const res = await reviewApi.updateReview(props.review.id, {
        rating: form.rating,
        body: form.body,
        visibility: form.visibility,
      })
      emit("saved", res.data.review)
    } catch (err) {
      console.error("Update review failed:", err)
    } finally {
      saving.value = false
    }
  }

  onMounted(() => {
    document.body.style.overflow = "hidden"
  })

  onBeforeUnmount(() => {
    const activeBackdrops = document.querySelectorAll(".modal-backdrop")
    if (activeBackdrops.length <= 1) {
      document.body.style.overflow = ""
    }
  })
</script>

<style scoped>
  .modal-backdrop {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.72);
    backdrop-filter: blur(6px);
    -webkit-backdrop-filter: blur(6px);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 200;
    padding: 16px;
  }

  .modal {
    --c-border: rgba(255, 255, 255, 0.08);
    --c-border-h: rgba(255, 255, 255, 0.14);
    --c-sub: #8a8a8e;
    --c-muted: #3a3a3c;
    --c-star: #fbbf24;
    --c-red: #e1251b;
    --font: var(--font-sans);
    --ease: cubic-bezier(0.16, 1, 0.3, 1);

    font-family: var(--font-sans);
    background: #161616;
    border: 1px solid var(--c-border);
    border-radius: 14px;
    width: 100%;
    max-width: 440px;
    box-shadow: 0 24px 64px rgba(0, 0, 0, 0.6);
    overflow: hidden;
  }

  /* Header */
  .modal-header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    padding: 22px 22px 0;
  }
  .modal-title-group {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }
  .modal-eyebrow {
    font-size: 0.58rem;
    font-weight: 700;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--c-muted);
  }
  .modal-title {
    font-size: 1.05rem;
    font-weight: 600;
    color: #fff;
    margin: 0;
    line-height: 1.2;
  }
  .modal-close {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 26px;
    height: 26px;
    border-radius: 6px;
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid var(--c-border);
    color: var(--c-sub);
    cursor: pointer;
    transition: all 0.15s;
    flex-shrink: 0;
  }
  .modal-close:hover {
    color: #fff;
    background: rgba(255, 255, 255, 0.1);
    border-color: var(--c-border-h);
  }

  /* Movie info */
  .movie-row {
    display: flex;
    align-items: center;
    gap: 10px;
    margin: 16px 22px;
    padding: 10px 12px;
    background: rgba(255, 255, 255, 0.03);
    border: 1px solid var(--c-border);
    border-radius: 8px;
  }
  .movie-poster {
    width: 28px;
    height: 40px;
    border-radius: 4px;
    overflow: hidden;
    background: #222;
    border: 1px solid var(--c-border);
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
    color: var(--c-muted);
  }
  .movie-poster img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
  .movie-title {
    font-size: 0.82rem;
    font-weight: 500;
    color: #e0e0e0;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  /* Fields */
  .field {
    padding: 0 22px;
    margin-bottom: 16px;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  .field-label {
    font-size: 0.7rem;
    font-weight: 600;
    letter-spacing: 0.05em;
    text-transform: uppercase;
    color: var(--c-sub);
  }

  /* Star picker */
  .star-picker {
    display: flex;
    align-items: center;
    gap: 4px;
  }
  .star-slot {
    position: relative;
    width: 28px;
    height: 28px;
    cursor: pointer;
    transition: transform 0.12s var(--ease);
  }
  .star-slot:hover {
    transform: scale(1.18);
  }
  .star-base,
  .star-fill {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    pointer-events: none;
  }
  .star-base {
    color: var(--c-muted);
  }
  .fill--on {
    color: var(--c-star);
  }
  .fill--off {
    color: transparent;
  }
  .half-zone {
    position: absolute;
    top: 0;
    height: 100%;
    width: 50%;
    cursor: pointer;
    z-index: 1;
  }
  .half-zone.left {
    left: 0;
  }
  .half-zone.right {
    right: 0;
  }
  .rating-label {
    font-size: 0.85rem;
    font-weight: 700;
    color: var(--c-star);
    margin-left: 6px;
    font-variant-numeric: tabular-nums;
    min-width: 24px;
  }

  /* Textarea */
  .textarea {
    font-family: var(--font-sans);
    font-size: 0.82rem;
    color: #e0e0e0;
    background: rgba(255, 255, 255, 0.03);
    border: 1px solid var(--c-border);
    border-radius: 8px;
    padding: 11px 13px;
    resize: vertical;
    min-height: 110px;
    transition: border-color 0.15s;
    outline: none;
  }
  .textarea:focus {
    border-color: rgba(255, 255, 255, 0.22);
  }
  .textarea::placeholder {
    color: var(--c-muted);
  }
  .char-count {
    font-size: 0.65rem;
    color: var(--c-muted);
    text-align: right;
    margin-top: -4px;
  }

  /* Footer */
  .modal-footer {
    display: flex;
    justify-content: flex-end;
    gap: 8px;
    padding: 16px 22px 22px;
    border-top: 1px solid var(--c-border);
    margin-top: 4px;
  }
  .btn {
    font-family: var(--font-sans);
    font-size: 0.78rem;
    font-weight: 600;
    padding: 8px 18px;
    border-radius: 8px;
    border: 1px solid transparent;
    cursor: pointer;
    transition: all 0.15s var(--ease);
    display: flex;
    align-items: center;
    gap: 6px;
  }
  .btn--ghost {
    color: var(--c-sub);
    background: transparent;
    border-color: var(--c-border);
  }
  .btn--ghost:hover {
    color: #fff;
    border-color: var(--c-border-h);
    background: rgba(255, 255, 255, 0.04);
  }
  .btn--primary {
    color: #000;
    background: #fff;
    border-color: #fff;
    min-width: 110px;
    justify-content: center;
  }
  .btn--primary:hover:not(:disabled) {
    background: #e0e0e0;
  }
  .btn--primary:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
  /* Transitions */
  .modal-enter-active {
    transition:
      opacity 0.2s ease,
      transform 0.25s var(--ease);
  }
  .modal-leave-active {
    transition:
      opacity 0.15s ease,
      transform 0.15s ease;
  }
  .modal-enter-from {
    opacity: 0;
    transform: scale(0.96) translateY(8px);
  }
  .modal-leave-to {
    opacity: 0;
    transform: scale(0.96) translateY(4px);
  }
</style>
