<template>
  <button
    class="follow-btn"
    :class="[stateClass, { 'is-loading': isMutating }]"
    :disabled="isMutating || status.isLoading.value"
    @mouseenter="isHovering = true"
    @mouseleave="isHovering = false"
    @click="handleClick"
  >
    <i :class="iconClass"></i>
    <span>{{ label }}</span>
  </button>
</template>

<script setup lang="ts">
  import { computed, ref } from "vue"
  import { useI18n } from "vue-i18n"
  import { useFollowActions, useFollowStatus } from "@/composables/useFollow"

  const props = defineProps<{
    userId: number
  }>()

  const { t } = useI18n()
  const status = useFollowStatus(props.userId)
  const { follow, unfollow } = useFollowActions(props.userId)

  const isHovering = ref(false)

  const isMutating = computed(
    () => follow.isPending.value || unfollow.isPending.value,
  )

  const relationship = computed(() => {
    if (status.data.value?.is_following) return "following"
    if (status.data.value?.follow_status === "pending") return "pending"
    return "none"
  })

  const stateClass = computed(() => `is-${relationship.value}`)

  const label = computed(() => {
    if (relationship.value === "following") {
      return isHovering.value ? t("profile.follow.unfollow") : t("profile.follow.following")
    }
    if (relationship.value === "pending") {
      return isHovering.value ? t("profile.follow.cancelRequest") : t("profile.follow.requested")
    }
    return t("profile.follow.follow")
  })

  const iconClass = computed(() => {
    if (relationship.value === "following") {
      return isHovering.value ? "pi pi-times" : "pi pi-check"
    }
    if (relationship.value === "pending") return "pi pi-clock"
    return "pi pi-plus"
  })

  function handleClick() {
    if (relationship.value === "none") {
      follow.mutate()
    } else {
      unfollow.mutate()
    }
  }
</script>

<style scoped>
  .follow-btn {
    display: inline-flex;
    align-items: center;
    gap: 0.4rem;
    padding: 0.5rem 1.2rem;
    border-radius: 999px;
    font-size: 0.85rem;
    font-weight: 600;
    cursor: pointer;
    border: 1px solid transparent;
    transition:
      background 0.15s,
      border-color 0.15s,
      color 0.15s;
    white-space: nowrap;
  }

  .follow-btn.is-loading {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .follow-btn.is-none {
    background: var(--color-brand);
    color: var(--color-on-brand);
  }
  .follow-btn.is-none:hover:not(.is-loading) {
    background: var(--color-brand-hover);
  }

  .follow-btn.is-following {
    background: var(--color-border);
    color: var(--color-text-primary);
    border-color: var(--color-border-strong);
  }
  .follow-btn.is-following:hover:not(.is-loading),
  .follow-btn.is-pending:hover:not(.is-loading) {
    background: color-mix(in srgb, var(--color-brand) 12%, transparent);
    border-color: color-mix(in srgb, var(--color-brand) 35%, transparent);
    color: var(--color-brand);
  }

  .follow-btn.is-pending {
    background: color-mix(in srgb, var(--color-rating) 10%, transparent);
    color: var(--color-rating);
    border-color: color-mix(in srgb, var(--color-rating) 25%, transparent);
  }
</style>
