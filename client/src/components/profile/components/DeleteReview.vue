<template>
  <Teleport to="body">
    <Transition name="modal">
      <div class="modal-backdrop" @click.self="$emit('close')">
        <div
          class="modal"
          role="dialog"
          aria-modal="true"
          aria-labelledby="delete-modal-title"
        >
          <!-- Icon -->
          <div class="icon-wrap">
            <Trash2 :size="20" stroke-width="1.6" />
          </div>

          <!-- Text -->
          <h2 id="delete-modal-title" class="modal-title">Delete Review</h2>
          <p class="modal-desc">
            Are you sure you want to delete your review for
            <strong class="movie-name">{{ review.media.title }}</strong
            >? This action cannot be undone.
          </p>

          <!-- Footer -->
          <div class="modal-footer">
            <button class="btn btn--ghost" @click="$emit('close')">
              Cancel
            </button>
            <button
              class="btn btn--danger"
              :disabled="deleting"
              @click="handleDelete"
            >
              <AppSpinner v-if="deleting" />
              <span v-else>Delete</span>
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import AppSpinner from "@/components/ui/AppSpinner.vue"
  import { ref, onMounted, onBeforeUnmount } from "vue"
  import { Trash2 } from "lucide-vue-next"
  import { reviewApi } from "@/api/api"
  import type { ReviewResponse } from "@/types"

  const props = defineProps<{ userId: number; review: ReviewResponse }>()
  const emit = defineEmits<{
    close: []
    deleted: [id: number]
  }>()

  const deleting = ref(false)

  async function handleDelete() {
    if (deleting.value) return
    try {
      deleting.value = true
      await reviewApi.deleteReview(props.review.id)
      emit("deleted", props.review.id)
    } catch (err) {
      console.error("Delete review failed:", err)
    } finally {
      deleting.value = false
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
    --c-border: var(--color-border);
    --c-border-h: var(--color-border-strong);
    --c-sub: var(--color-text-secondary);
    --c-muted: var(--color-text-muted);
    --c-red: var(--color-danger);
    --font: var(--font-sans);
    --ease: cubic-bezier(0.16, 1, 0.3, 1);

    font-family: var(--font-sans);
    background: #161616;
    border: 1px solid var(--c-border);
    border-radius: 14px;
    width: 100%;
    max-width: 360px;
    padding: 28px 24px 24px;
    box-shadow: 0 24px 64px rgba(0, 0, 0, 0.6);
    display: flex;
    flex-direction: column;
    align-items: center;
    text-align: center;
    gap: 0;
  }

  .icon-wrap {
    width: 48px;
    height: 48px;
    border-radius: 12px;
    background: rgba(225, 37, 27, 0.1);
    border: 1px solid rgba(225, 37, 27, 0.2);
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--c-red);
    margin-bottom: 16px;
  }

  .modal-title {
    font-size: 1rem;
    font-weight: 600;
    color: #fff;
    margin: 0 0 8px;
    line-height: 1.3;
  }

  .modal-desc {
    font-size: 0.8rem;
    color: var(--c-sub);
    line-height: 1.55;
    margin: 0 0 24px;
  }

  .movie-name {
    color: #e0e0e0;
    font-weight: 600;
  }

  .modal-footer {
    display: flex;
    gap: 8px;
    width: 100%;
  }

  .btn {
    flex: 1;
    font-family: var(--font-sans);
    font-size: 0.78rem;
    font-weight: 600;
    padding: 9px 0;
    border-radius: 8px;
    border: 1px solid transparent;
    cursor: pointer;
    transition: all 0.15s var(--ease);
    display: flex;
    align-items: center;
    justify-content: center;
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
  .btn--danger {
    color: #fff;
    background: var(--c-red);
    border-color: var(--c-red);
    min-width: 80px;
  }
  .btn--danger:hover:not(:disabled) {
    background: #c41f16;
    border-color: #c41f16;
  }
  .btn--danger:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
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
