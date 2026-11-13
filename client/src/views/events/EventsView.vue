<template>
  <section class="events-page">
    <header class="events-page__header">
      <div class="events-page__heading">
        <h1 class="events-page__title">{{ t("events.title") }}</h1>
        <p class="events-page__subtitle">{{ t("events.subtitle") }}</p>
      </div>
      <div class="events-page__tabs">
        <AppTabs
          v-model="activeTab"
          :tabs="tabs"
          :label="t('events.title')"
          :controls="false"
        />
      </div>
    </header>

    <div v-if="loading && events.length === 0" class="events-page__loading">
      <AppSpinner :size="32" />
    </div>

    <AppErrorState
      v-else-if="error"
      :title="t('common.error')"
      :description="t('events.loadErrorDesc')"
      @retry="fetchEvents"
    />

    <AppEmptyState
      v-else-if="events.length === 0"
      :title="t(`events.empty.${activeTab}`)"
      :description="t(`events.empty.${activeTab}Desc`)"
      :icon="Megaphone"
    />

    <div v-else class="events-grid">
      <AppCard
        v-for="item in events"
        :key="item.id"
        class="event-card"
        hoverable
      >
        <div class="event-card__media">
          <img
            v-if="item.image_url"
            :src="item.image_url"
            alt=""
            class="event-card__image"
            loading="lazy"
          />
          <div v-else class="event-card__placeholder">
            <Megaphone :size="40" class="event-card__placeholder-icon" aria-hidden="true" />
          </div>
          <span
            class="event-card__badge"
            :class="`event-card__badge--${item.state}`"
          >
            {{ t(`events.tabs.${item.state}`) }}
          </span>
        </div>

        <div class="event-card__body">
          <div class="event-card__meta">
            <Calendar :size="14" aria-hidden="true" />
            <time :datetime="item.starts_at">
              {{ formatDates(item.starts_at, item.ends_at) }}
            </time>
          </div>

          <h2 class="event-card__title">
            <RouterLink :to="`/events/${item.id}`" class="event-card__title-link">
              {{ localizedText(item.title, locale) }}
            </RouterLink>
          </h2>

          <p class="event-card__description">
            {{ localizedText(item.body, locale) }}
          </p>

          <div class="event-card__actions">
            <AppButton
              variant="outline"
              size="sm"
              :to="`/events/${item.id}`"
            >
              {{ t("events.card.viewDetails") }}
            </AppButton>

            <template v-if="item.link_url">
              <AppButton
                v-if="isInternalLink(item.link_url)"
                variant="primary"
                size="sm"
                :to="item.link_url"
              >
                {{ t("events.card.openLink") }}
              </AppButton>
              <AppButton
                v-else
                variant="primary"
                size="sm"
                :href="item.link_url"
                target="_blank"
                rel="noopener noreferrer"
              >
                <span>{{ t("events.card.openLink") }}</span>
                <ExternalLink :size="14" aria-hidden="true" />
              </AppButton>
            </template>
          </div>
        </div>
      </AppCard>
    </div>

    <div v-if="pagination && pagination.total_pages > 1" class="events-page__pagination">
      <AppButton
        variant="ghost"
        size="sm"
        :disabled="page <= 1 || loading"
        @click="changePage(page - 1)"
      >
        <ChevronLeft :size="16" aria-hidden="true" />
        {{ t("admin.pagination.previous") }}
      </AppButton>
      <span class="events-page__page-info">
        {{ t("admin.pagination.page", { page: pagination.page, pages: pagination.total_pages }) }}
      </span>
      <AppButton
        variant="ghost"
        size="sm"
        :disabled="page >= pagination.total_pages || loading"
        @click="changePage(page + 1)"
      >
        {{ t("admin.pagination.next") }}
        <ChevronRight :size="16" aria-hidden="true" />
      </AppButton>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue"
import { useRoute, useRouter } from "vue-router"
import { useI18n } from "vue-i18n"
import {
  Calendar,
  ChevronLeft,
  ChevronRight,
  ExternalLink,
  Megaphone,
} from "lucide-vue-next"
import AppButton from "@/components/ui/AppButton.vue"
import AppCard from "@/components/ui/AppCard.vue"
import AppEmptyState from "@/components/ui/AppEmptyState.vue"
import AppErrorState from "@/components/ui/AppErrorState.vue"
import AppSpinner from "@/components/ui/AppSpinner.vue"
import AppTabs from "@/components/ui/AppTabs.vue"
import { eventApi } from "@/api/endpoints/event"
import type { AppEvent, EventPagination, EventState } from "@/types/event"
import {
  EVENTS_PAGE_SIZE,
  isInternalLink,
  localizedText,
} from "@/utils/events"

const { t, d, locale } = useI18n()
const route = useRoute()
const router = useRouter()

const validTabs: EventState[] = ["active", "upcoming", "past"]

// The route query is the single source for tab and page. Only the watcher below fetches.
const activeTab = computed<EventState>({
  get: () => (validTabs.includes(route.query.tab as EventState) ? (route.query.tab as EventState) : "active"),
  set: (tab) => void router.replace({ query: { ...route.query, tab, page: undefined }, hash: route.hash }),
})
const page = computed(() => parsePage(route.query.page))

const tabs = computed(() => [
  { key: "active", label: t("events.tabs.active") },
  { key: "upcoming", label: t("events.tabs.upcoming") },
  { key: "past", label: t("events.tabs.past") },
])

const events = ref<AppEvent[]>([])
const pagination = ref<EventPagination | null>(null)
const loading = ref(false)
const error = ref(false)

async function fetchEvents() {
  loading.value = true
  error.value = false
  try {
    const res = await eventApi.list({
      state: activeTab.value,
      page: page.value,
      limit: EVENTS_PAGE_SIZE,
    })
    events.value = res.data.events
    pagination.value = res.data.pagination
  } catch {
    error.value = true
  } finally {
    loading.value = false
  }
}

function parsePage(raw: unknown): number {
  const n = Number(raw)
  return Number.isInteger(n) && n > 0 ? n : 1
}

function formatDay(iso: string): string {
  const date = new Date(iso)
  return Number.isNaN(date.getTime()) ? "" : d(date, "short")
}

function formatDates(startIso: string, endIso: string): string {
  return t("events.card.period", { start: formatDay(startIso), end: formatDay(endIso) })
}

function changePage(newPage: number) {
  if (newPage < 1 || (pagination.value && newPage > pagination.value.total_pages)) return
  void router.replace({ query: { ...route.query, page: String(newPage) }, hash: route.hash })
}

watch([activeTab, page], () => void fetchEvents())

onMounted(() => {
  void fetchEvents()
})
</script>

<style scoped>
.events-page {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  width: 100%;
  max-width: var(--container-page);
  margin: 0 auto;
  padding: 1.5rem 1rem 3rem;
}

.events-page__header {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

@media (min-width: 768px) {
  .events-page__header {
    flex-direction: row;
    align-items: flex-end;
    justify-content: space-between;
  }
}

.events-page__heading {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.events-page__title {
  font-size: 1.75rem;
  font-weight: 700;
  color: var(--color-text-primary);
  line-height: 1.2;
}

.events-page__subtitle {
  font-size: 0.875rem;
  color: var(--color-text-muted);
}

.events-page__loading {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 6rem 0;
}

.events-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 1.25rem;
}

@media (min-width: 640px) {
  .events-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (min-width: 1024px) {
  .events-grid {
    grid-template-columns: repeat(3, 1fr);
  }
}

.event-card {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
  border-radius: var(--radius-card);
  background-color: var(--color-surface-2);
  border: 1px solid var(--color-border);
}

.event-card__media {
  position: relative;
  width: 100%;
  aspect-ratio: 16 / 9;
  background-color: var(--color-surface-3);
  overflow: hidden;
}

.event-card__image {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.event-card__placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  background: linear-gradient(135deg, var(--color-surface-3), var(--color-surface-2));
}

.event-card__placeholder-icon {
  color: var(--color-text-muted);
  opacity: 0.5;
}

.event-card__badge {
  position: absolute;
  top: 0.75rem;
  left: 0.75rem;
  padding: 0.25rem 0.5rem;
  font-size: 0.6875rem;
  font-weight: 600;
  border-radius: var(--radius-full);
  text-transform: uppercase;
  letter-spacing: 0.025em;
}

.event-card__badge--active {
  background-color: var(--color-brand);
  color: var(--color-on-brand);
}

.event-card__badge--upcoming {
  background-color: var(--color-warning);
  color: var(--color-surface);
}

.event-card__badge--past {
  background-color: var(--color-surface-3);
  color: var(--color-text-muted);
}

.event-card__body {
  display: flex;
  flex-direction: column;
  flex: 1;
  padding: 1.25rem;
  gap: 0.75rem;
}

.event-card__meta {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  font-size: 0.75rem;
  color: var(--color-text-muted);
}

.event-card__title {
  font-size: 1.125rem;
  font-weight: 600;
  line-height: 1.3;
}

.event-card__title-link {
  color: var(--color-text-primary);
  text-decoration: none;
}

.event-card__title-link:hover {
  color: var(--color-brand);
}

.event-card__description {
  font-size: 0.8125rem;
  color: var(--color-text-secondary);
  line-height: 1.5;
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
  flex: 1;
}

.event-card__actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.5rem;
  margin-top: auto;
  padding-top: 0.5rem;
}

.events-page__pagination {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 1rem;
  padding-top: 1rem;
}

.events-page__page-info {
  font-size: 0.8125rem;
  color: var(--color-text-muted);
}
</style>
