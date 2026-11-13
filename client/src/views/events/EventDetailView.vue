<template>
  <section class="event-detail-page">
    <div class="event-detail-page__nav">
      <AppButton
        variant="ghost"
        size="sm"
        to="/events"
        class="event-detail-page__back"
      >
        <ChevronLeft :size="16" aria-hidden="true" />
        {{ t("events.detail.backToEvents") }}
      </AppButton>
    </div>

    <div v-if="loading" class="event-detail-page__loading">
      <AppSpinner :size="32" />
    </div>

    <AppErrorState
      v-else-if="notFound"
      :title="t('events.detail.notFoundTitle')"
      :description="t('events.detail.notFoundDesc')"
    />

    <AppErrorState
      v-else-if="error"
      :title="t('common.error')"
      :description="t('events.loadErrorDesc')"
      @retry="fetchEvent"
    />

    <article v-else-if="event" class="event-detail">
      <div v-if="event.image_url" class="event-detail__banner">
        <img
          :src="event.image_url"
          alt=""
          class="event-detail__banner-img"
        />
      </div>

      <header class="event-detail__header">
        <div class="event-detail__tags">
          <span
            class="event-detail__badge"
            :class="`event-detail__badge--${event.state}`"
          >
            {{ t(`events.tabs.${event.state}`) }}
          </span>
          <span class="event-detail__type-tag">
            {{ t("events.detail.announcement") }}
          </span>
        </div>

        <h1 class="event-detail__title">
          {{ localizedText(event.title, locale) }}
        </h1>

        <div class="event-detail__dates">
          <Calendar :size="16" aria-hidden="true" />
          <span>{{ t("events.detail.dates") }}</span>
          <time :datetime="event.starts_at">
            {{ formatDates(event.starts_at, event.ends_at) }}
          </time>
        </div>
      </header>

      <div class="event-detail__content">
        <p class="event-detail__body">
          {{ localizedText(event.body, locale) }}
        </p>

        <div v-if="event.link_url" class="event-detail__action">
          <AppButton
            v-if="isInternalLink(event.link_url)"
            variant="primary"
            size="lg"
            :to="event.link_url"
          >
            {{ t("events.detail.link") }}
          </AppButton>
          <AppButton
            v-else
            variant="primary"
            size="lg"
            :href="event.link_url"
            target="_blank"
            rel="noopener noreferrer"
          >
            <span>{{ t("events.detail.link") }}</span>
            <ExternalLink :size="16" aria-hidden="true" />
          </AppButton>
        </div>
      </div>
    </article>
  </section>
</template>

<script setup lang="ts">
import { onMounted, ref, watch } from "vue"
import { useRoute } from "vue-router"
import { isAxiosError } from "axios"
import { useI18n } from "vue-i18n"
import { Calendar, ChevronLeft, ExternalLink } from "lucide-vue-next"
import AppButton from "@/components/ui/AppButton.vue"
import AppErrorState from "@/components/ui/AppErrorState.vue"
import AppSpinner from "@/components/ui/AppSpinner.vue"
import { eventApi } from "@/api/endpoints/event"
import type { AppEvent } from "@/types/event"
import { HTTP_NOT_FOUND, isInternalLink, localizedText } from "@/utils/events"

const { t, d, locale } = useI18n()
const route = useRoute()

const event = ref<AppEvent | null>(null)
const loading = ref(true)
const error = ref(false)
const notFound = ref(false)

async function fetchEvent() {
  const id = Number(route.params.id)
  loading.value = true
  error.value = false
  notFound.value = false
  if (!id) {
    notFound.value = true
    loading.value = false
    return
  }

  try {
    const res = await eventApi.get(id)
    event.value = res.data.event
  } catch (err) {
    event.value = null
    if (isAxiosError(err) && err.response?.status === HTTP_NOT_FOUND) {
      notFound.value = true
    } else {
      error.value = true
    }
  } finally {
    loading.value = false
  }
}

function formatDay(iso: string): string {
  const date = new Date(iso)
  return Number.isNaN(date.getTime()) ? "" : d(date, "medium")
}

function formatDates(startIso: string, endIso: string): string {
  return t("events.card.period", { start: formatDay(startIso), end: formatDay(endIso) })
}

// Same page component serves every /events/:id, so refetch when the id changes.
watch(() => route.params.id, () => void fetchEvent())

onMounted(() => {
  void fetchEvent()
})
</script>

<style scoped>
.event-detail-page {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  width: 100%;
  max-width: var(--container-article);
  margin: 0 auto;
  padding: 1.5rem 1rem 4rem;
}

.event-detail-page__nav {
  display: flex;
  align-items: center;
}

.event-detail-page__loading {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 6rem 0;
}

.event-detail {
  display: flex;
  flex-direction: column;
  gap: 2rem;
  background-color: var(--color-surface-2);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-card);
  overflow: hidden;
}

.event-detail__banner {
  width: 100%;
  aspect-ratio: 21 / 9;
  background-color: var(--color-surface-3);
  overflow: hidden;
}

.event-detail__banner-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.event-detail__header {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  padding: 0 1.5rem;
}

@media (min-width: 768px) {
  .event-detail__header {
    padding: 0 2rem;
  }
}

.event-detail__tags {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.5rem;
}

.event-detail__badge {
  padding: 0.25rem 0.625rem;
  font-size: 0.6875rem;
  font-weight: 600;
  border-radius: var(--radius-full);
  text-transform: uppercase;
  letter-spacing: 0.025em;
}

.event-detail__badge--active {
  background-color: var(--color-brand);
  color: var(--color-on-brand);
}

.event-detail__badge--upcoming {
  background-color: var(--color-warning);
  color: var(--color-surface);
}

.event-detail__badge--past {
  background-color: var(--color-surface-3);
  color: var(--color-text-muted);
}

.event-detail__type-tag {
  padding: 0.25rem 0.625rem;
  font-size: 0.6875rem;
  font-weight: 500;
  border-radius: var(--radius-full);
  background-color: var(--color-surface-3);
  color: var(--color-text-secondary);
}

.event-detail__title {
  font-size: 1.75rem;
  font-weight: 700;
  color: var(--color-text-primary);
  line-height: 1.25;
}

@media (min-width: 768px) {
  .event-detail__title {
    font-size: 2.25rem;
  }
}

.event-detail__dates {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.875rem;
  color: var(--color-text-muted);
}

.event-detail__content {
  display: flex;
  flex-direction: column;
  gap: 2rem;
  padding: 0 1.5rem 2rem;
}

@media (min-width: 768px) {
  .event-detail__content {
    padding: 0 2rem 2.5rem;
  }
}

.event-detail__body {
  font-size: 1rem;
  line-height: 1.75;
  color: var(--color-text-secondary);
  white-space: pre-wrap;
}

.event-detail__action {
  display: flex;
  align-items: center;
  padding-top: 1rem;
  border-top: 1px solid var(--color-border);
}
</style>
