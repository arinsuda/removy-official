<template>
  <Teleport to="body">
    <Transition name="modal">
      <div
        v-if="modelValue"
        class="modal-backdrop"
        @click.self="emit('cancel')"
      >
        <div
          class="modal-card"
          role="dialog"
          :aria-label="config.ariaLabel"
          aria-modal="true"
        >
          <!-- Icon -->
          <div class="modal-icon" :class="config.iconClass">
            <component :is="config.icon" :size="22" :stroke-width="1.6" />
          </div>

          <!-- Content -->
          <div class="modal-body">
            <h3 class="modal-title">{{ config.title }}</h3>
            <p class="modal-desc">
              {{ config.description }}
              <strong class="item-name">{{ itemName }}</strong>
              {{ config.descriptionSuffix }}
            </p>
          </div>

          <!-- Actions -->
          <div class="modal-actions">
            <button class="btn-cancel" @click="emit('cancel')">Cancel</button>
            <button
              class="btn-confirm"
              :class="config.confirmClass"
              :disabled="confirmDisabled"
              @click="emit('confirm')"
            >
              {{ confirmDisabled ? "Processing..." : config.confirmLabel }}
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
  import { computed, watch, onBeforeUnmount } from "vue"
  import { Heart, Bookmark, Eye, Mail, UserMinus, Trash2 } from "lucide-vue-next"
  import type { ListType } from "@/types"

  type ConfirmModalType = ListType | "email_change" | "unfollow" | "bmol_delete"

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

  const configs: Record<
    ConfirmModalType,
    {
      icon: unknown
      iconClass: string
      ariaLabel: string
      title: string
      description: string
      descriptionSuffix: string
      confirmLabel: string
      confirmClass: string
    }
  > = {
    likes: {
      icon: Heart,
      iconClass: "icon-heart",
      ariaLabel: "Unlike confirmation",
      title: "Remove from Favorites?",
      description: "You're about to unlike",
      descriptionSuffix: ". You can always like it again later.",
      confirmLabel: "Unlike",
      confirmClass: "confirm-red",
    },
    watchlist: {
      icon: Bookmark,
      iconClass: "icon-bookmark",
      ariaLabel: "Remove from watchlist confirmation",
      title: "Remove from Watchlist?",
      description: "Remove",
      descriptionSuffix: " from your watchlist. You can add it back anytime.",
      confirmLabel: "Remove",
      confirmClass: "confirm-red",
    },
    watched: {
      icon: Eye,
      iconClass: "icon-eye",
      ariaLabel: "Remove from watched confirmation",
      title: "Remove from Watched?",
      description: "Mark",
      descriptionSuffix: " as unwatched. Your watch history will be updated.",
      confirmLabel: "Mark Unwatched",
      confirmClass: "confirm-muted",
    },
    email_change: {
      icon: Mail,
      iconClass: "icon-mail",
      ariaLabel: "Change email confirmation",
      title: "Change Email?",
      description: "We'll send a verification code to",
      descriptionSuffix: ". Enter it to confirm your identity.",
      confirmLabel: "Send OTP",
      confirmClass: "confirm-blue",
    },
    unfollow: {
      icon: UserMinus,
      iconClass: "icon-user-minus",
      ariaLabel: "Unfollow confirmation",
      title: "Unfollow User?",
      description: "You're about to unfollow",
      descriptionSuffix: ". You can follow them again at any time.",
      confirmLabel: "Unfollow",
      confirmClass: "confirm-red",
    },
    bmol_delete: {
      icon: Trash2,
      iconClass: "icon-trash",
      ariaLabel: "Delete BMOL item confirmation",
      title: "Remove from Best Movies?",
      description: "Remove",
      descriptionSuffix: " from your Best Movies of Life list.",
      confirmLabel: "Remove",
      confirmClass: "confirm-red",
    },
  }

  const config = computed(() => configs[props.listType])

  watch(() => props.modelValue, (newVal) => {
    if (newVal) {
      document.body.style.overflow = "hidden"
    } else {
      const activeBackdrops = document.querySelectorAll(".modal-backdrop")
      if (activeBackdrops.length <= 1) {
        document.body.style.overflow = ""
      }
    }
  }, { immediate: true })

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
    background: rgba(0, 0, 0, 0.75);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 9999;
    padding: 16px;
  }

  .modal-card {
    background: #1c1c1e;
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 16px;
    padding: 28px 24px 24px;
    width: 100%;
    max-width: 360px;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 16px;
    box-shadow:
      0 24px 64px rgba(0, 0, 0, 0.6),
      0 0 0 1px rgba(255, 255, 255, 0.04) inset;
    will-change: transform, opacity;
  }

  /* Icon */
  .modal-icon {
    width: 52px;
    height: 52px;
    border-radius: 14px;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }
  .icon-heart {
    background: rgba(225, 37, 27, 0.12);
    color: #e1251b;
    border: 1px solid rgba(225, 37, 27, 0.2);
  }
  .icon-bookmark {
    background: rgba(255, 184, 0, 0.1);
    color: #ffb800;
    border: 1px solid rgba(255, 184, 0, 0.2);
  }
  .icon-eye {
    background: rgba(99, 99, 255, 0.1);
    color: #8585ff;
    border: 1px solid rgba(99, 99, 255, 0.2);
  }

  /* Body */
  .modal-body {
    text-align: center;
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .modal-title {
    font-size: 1rem;
    font-weight: 650;
    color: #f0f0f0;
    margin: 0;
    letter-spacing: -0.01em;
    font-family: var(--font-sans);
  }
  .modal-desc {
    font-size: 0.82rem;
    color: #8a8a8e;
    margin: 0;
    line-height: 1.5;
    font-family: var(--font-sans);
  }
  .item-name {
    color: #d0d0d0;
    font-weight: 600;
  }

  /* Actions */
  .modal-actions {
    display: flex;
    gap: 8px;
    width: 100%;
    margin-top: 4px;
  }

  .btn-cancel,
  .btn-confirm {
    flex: 1;
    height: 40px;
    border-radius: 10px;
    border: none;
    font-size: 0.82rem;
    font-weight: 600;
    cursor: pointer;
    transition: background-color 0.18s ease, transform 0.18s ease, box-shadow 0.18s ease, opacity 0.18s ease;
    font-family: var(--font-sans);
    letter-spacing: 0.01em;
  }

  .btn-cancel {
    background: rgba(255, 255, 255, 0.06);
    color: #a0a0a4;
    border: 1px solid rgba(255, 255, 255, 0.07);
  }
  .btn-cancel:hover {
    background: rgba(255, 255, 255, 0.1);
    color: #f0f0f0;
  }

  .confirm-red {
    background: #e1251b;
    color: #fff;
  }
  .confirm-red:hover {
    background: #c92017;
    transform: translateY(-1px);
    box-shadow: 0 6px 20px rgba(225, 37, 27, 0.35);
  }

  .confirm-muted {
    background: rgba(133, 133, 255, 0.18);
    color: #8585ff;
    border: 1px solid rgba(133, 133, 255, 0.25);
  }
  .confirm-muted:hover {
    background: rgba(133, 133, 255, 0.28);
    transform: translateY(-1px);
  }

  .btn-cancel:active,
  .btn-confirm:active {
    transform: translateY(0) scale(0.97);
  }

  /* Transition */
  .modal-enter-active,
  .modal-leave-active {
    transition: opacity 0.2s ease;
  }
  .modal-enter-active .modal-card,
  .modal-leave-active .modal-card {
    transition:
      transform 0.22s cubic-bezier(0.16, 1, 0.3, 1),
      opacity 0.2s ease;
  }
  .modal-enter-from,
  .modal-leave-to {
    opacity: 0;
  }
  .modal-enter-from .modal-card {
    transform: scale(0.93) translateY(8px);
    opacity: 0;
  }
  .modal-leave-to .modal-card {
    transform: scale(0.95) translateY(4px);
    opacity: 0;
  }
  .icon-mail {
    background: rgba(10, 132, 255, 0.1);
    color: #0a84ff;
    border: 1px solid rgba(10, 132, 255, 0.2);
  }
  .confirm-blue {
    background: #0a84ff;
    color: #fff;
  }
  .confirm-blue:hover {
    background: #0071e3;
    transform: translateY(-1px);
    box-shadow: 0 6px 20px rgba(10, 132, 255, 0.35);
  }
  .icon-user-minus {
    background: rgba(225, 37, 27, 0.12);
    color: #e1251b;
    border: 1px solid rgba(225, 37, 27, 0.2);
  }
  .icon-trash {
    background: rgba(239, 68, 68, 0.1);
    color: #ef4444;
    border: 1px solid rgba(239, 68, 68, 0.2);
  }

  .btn-confirm:disabled,
  .btn-cancel:disabled {
    opacity: 0.55;
    cursor: not-allowed;
    transform: none;
  }
</style>
