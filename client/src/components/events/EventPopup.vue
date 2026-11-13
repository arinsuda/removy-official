<template>
  <AppModal
    v-model="isOpen"
    size="md"
    :title="t('events.popup.title')"
    class="event-popup"
    @close="handleDismiss"
  >
    <div v-if="currentEvent" class="event-popup__content">
      <div v-if="currentEvent.image_url" class="event-popup__banner">
        <img
          :src="currentEvent.image_url"
          alt=""
          class="event-popup__image"
        />
      </div>
      <div v-else class="event-popup__placeholder">
        <Megaphone :size="36" class="event-popup__placeholder-icon" aria-hidden="true" />
      </div>

      <div class="event-popup__body">
        <h3 class="event-popup__title">
          {{ localizedText(currentEvent.title, locale) }}
        </h3>

        <p class="event-popup__description">
          {{ localizedText(currentEvent.body, locale) }}
        </p>

        <div class="event-popup__action">
          <AppButton
            v-if="currentEvent.link_url && !isInternalLink(currentEvent.link_url)"
            variant="primary"
            size="md"
            :href="currentEvent.link_url"
            target="_blank"
            rel="noopener noreferrer"
            block
            @click="handleNavigate"
          >
            <span>{{ t("events.popup.viewEvent") }}</span>
            <ExternalLink :size="16" aria-hidden="true" />
          </AppButton>
          <AppButton
            v-else
            variant="primary"
            size="md"
            :to="currentEvent.link_url || `/events/${currentEvent.id}`"
            block
            @click="handleNavigate"
          >
            {{ t("events.popup.viewEvent") }}
          </AppButton>
        </div>
      </div>

      <div v-if="events.length > 1" class="event-popup__nav">
        <AppButton
          variant="ghost"
          size="sm"
          icon-only
          :disabled="currentIndex <= 0"
          :aria-label="t('events.popup.prev')"
          @click="prevEvent"
        >
          <ChevronLeft :size="18" aria-hidden="true" />
        </AppButton>
        <span class="event-popup__counter">
          {{ t("events.popup.counter", { current: currentIndex + 1, total: events.length }) }}
        </span>
        <AppButton
          variant="ghost"
          size="sm"
          icon-only
          :disabled="currentIndex >= events.length - 1"
          :aria-label="t('events.popup.next')"
          @click="nextEvent"
        >
          <ChevronRight :size="18" aria-hidden="true" />
        </AppButton>
      </div>
    </div>

    <template #footer>
      <div class="event-popup__footer">
        <label class="event-popup__checkbox-label">
          <input
            v-model="dontShowToday"
            type="checkbox"
            class="event-popup__checkbox"
          />
          <span>{{ t("events.popup.dontShowToday") }}</span>
        </label>
        <AppButton
          variant="ghost"
          size="sm"
          @click="closeModal"
        >
          {{ t("events.popup.close") }}
        </AppButton>
      </div>
    </template>
  </AppModal>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue"
import { useRoute } from "vue-router"
import { useI18n } from "vue-i18n"
import { ChevronLeft, ChevronRight, ExternalLink, Megaphone } from "lucide-vue-next"
import AppButton from "@/components/ui/AppButton.vue"
import AppModal from "@/components/ui/AppModal.vue"
import { eventApi } from "@/api/endpoints/event"
import type { AppEvent } from "@/types/event"
import { useAuthStore } from "@/stores/auth"
import { dismissPopup, isInternalLink, isPopupHidden, localizedText } from "@/utils/events"

const { t, locale } = useI18n()
const route = useRoute()
const authStore = useAuthStore()

const isOpen = ref(false)
const events = ref<AppEvent[]>([])
const currentIndex = ref(0)
const dontShowToday = ref(false)

const currentEvent = computed(() => events.value[currentIndex.value] ?? null)

async function checkAndShowPopup() {
  if (!authStore.isLoggedIn || !authStore.user?.id) return
  if (route.path.startsWith("/admin")) return

  const userId = authStore.user.id
  if (isPopupHidden(userId)) return

  try {
    const res = await eventApi.popup()
    const activeEvents = res.data.events || []
    if (activeEvents.length > 0) {
      events.value = activeEvents
      currentIndex.value = 0
      isOpen.value = true
    }
  } catch {
    // Non-critical background feature; ignore fetch errors
  }
}

function handleDismiss() {
  if (authStore.user?.id) {
    dismissPopup(authStore.user.id, dontShowToday.value)
  }
}

function closeModal() {
  isOpen.value = false
  handleDismiss()
}

function handleNavigate() {
  closeModal()
}

function nextEvent() {
  if (currentIndex.value < events.value.length - 1) {
    currentIndex.value++
  }
}

function prevEvent() {
  if (currentIndex.value > 0) {
    currentIndex.value--
  }
}

// Watch the primitive id: a new array each tick would re-run the check on every store change.
watch(
  () => authStore.user?.id,
  (userId) => {
    if (userId && !isOpen.value) {
      void checkAndShowPopup()
    }
  },
)

onMounted(() => {
  if (authStore.isLoggedIn) {
    void checkAndShowPopup()
  }
})
</script>

<style scoped>
.event-popup__content {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.event-popup__banner {
  width: 100%;
  aspect-ratio: 16 / 9;
  border-radius: var(--radius-card);
  background-color: var(--color-surface-3);
  overflow: hidden;
}

.event-popup__image {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.event-popup__placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  aspect-ratio: 16 / 9;
  border-radius: var(--radius-card);
  background: linear-gradient(135deg, var(--color-surface-3), var(--color-surface-2));
}

.event-popup__placeholder-icon {
  color: var(--color-text-muted);
  opacity: 0.5;
}

.event-popup__body {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.event-popup__title {
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--color-text-primary);
  line-height: 1.3;
}

.event-popup__description {
  font-size: 0.875rem;
  color: var(--color-text-secondary);
  line-height: 1.6;
  white-space: pre-wrap;
  max-height: 10rem;
  overflow-y: auto;
}

.event-popup__action {
  margin-top: 0.5rem;
}

.event-popup__nav {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.75rem;
  padding-top: 0.25rem;
}

.event-popup__counter {
  font-size: 0.8125rem;
  color: var(--color-text-muted);
  font-variant-numeric: tabular-nums;
}

.event-popup__footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}

.event-popup__checkbox-label {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.8125rem;
  color: var(--color-text-muted);
  cursor: pointer;
  user-select: none;
}

.event-popup__checkbox {
  accent-color: var(--color-brand);
  cursor: pointer;
}
</style>
