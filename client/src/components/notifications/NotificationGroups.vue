<template>
  <div class="noti-groups">
    <section v-for="group in groups" :key="group.bucket" class="noti-group" :aria-labelledby="`${baseId}-${group.bucket}`">
      <h2 :id="`${baseId}-${group.bucket}`" class="noti-group__title">{{ t(`notifications.groups.${group.bucket}`) }}</h2>
      <TransitionGroup name="noti-list" tag="ul" class="noti-group__list">
        <NotificationItem
          v-for="n in group.items"
          :key="n.id"
          :notification="n"
          :fresh="freshIds.has(n.id)"
          @open="(x) => emit('open', x)"
          @remove="(x) => emit('remove', x)"
          @responded="(x) => emit('responded', x)"
        />
      </TransitionGroup>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, useId } from "vue"
import { useI18n } from "vue-i18n"
import type { AppNotification } from "@/types/notification"
import { groupByDay } from "@/utils/dayGroups"
import NotificationItem from "./NotificationItem.vue"

const props = defineProps<{
  items: AppNotification[]
  freshIds: Set<number>
}>()

const emit = defineEmits<{
  open: [notification: AppNotification]
  remove: [notification: AppNotification]
  responded: [notification: AppNotification]
}>()

const { t } = useI18n()
const baseId = useId()
const groups = computed(() => groupByDay(props.items, (n) => n.created_at))
</script>

<style scoped>
.noti-group + .noti-group {
  margin-top: 1.5rem;
}
.noti-group__title {
  margin: 0 0 0.5rem;
  font-size: var(--text-caption);
  font-weight: 700;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: var(--color-text-muted);
}
.noti-group__list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  margin: 0;
  padding: 0;
  list-style: none;
}

.noti-list-enter-active,
.noti-list-leave-active {
  transition:
    opacity var(--motion-base) var(--ease-out-soft),
    transform var(--motion-base) var(--ease-out-soft);
}
.noti-list-enter-from {
  opacity: 0;
  transform: translateY(-10px);
}
.noti-list-leave-to {
  opacity: 0;
  transform: translateX(16px);
}
.noti-list-move {
  transition: transform var(--motion-base) var(--ease-out-soft);
}

@media (prefers-reduced-motion: reduce) {
  .noti-list-enter-from,
  .noti-list-leave-to {
    transform: none;
  }
  .noti-list-move {
    transition: none;
  }
}
</style>
