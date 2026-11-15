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
            {{ t(`events.detail.${event.type}`) }}
          </span>
        </div>

        <h1 class="event-detail__title">
          {{ eventTitle(event, t, locale) }}
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
          {{ eventBody(event, t, locale) }}
        </p>

        <!-- Voting Section (for vote_monthly & vote_yearly) -->
        <section v-if="event.nominees && event.nominees.length" class="event-voting">
          <div class="event-voting__header">
            <div>
              <h2 class="event-voting__title">{{ t("events.voting.title") }}</h2>
              <p class="event-voting__subtitle">{{ t("events.voting.subtitle") }}</p>
            </div>
            <div v-if="event.total_votes != null" class="event-voting__total">
              {{ formatTotalVotes(event.total_votes) }}
            </div>
          </div>

          <!-- Status / Secrecy Notice -->
          <div v-if="event.state === 'active'" class="event-voting__notice event-voting__notice--info">
            <Lock :size="16" aria-hidden="true" />
            <span>{{ t("events.voting.secrecyNotice", { date: formatDay(event.ends_at) }) }}</span>
          </div>
          <div v-else-if="event.state === 'past'" class="event-voting__notice event-voting__notice--past">
            <Trophy :size="16" aria-hidden="true" />
            <span>{{ t("events.voting.pastNotice") }}</span>
          </div>
          <div v-else class="event-voting__notice event-voting__notice--upcoming">
            <Calendar :size="16" aria-hidden="true" />
            <span>{{ t("events.voting.upcomingNotice", { date: formatDay(event.starts_at) }) }}</span>
          </div>

          <!-- Active Vote Banner: show who the user voted for with option to cancel -->
          <div v-if="event.state === 'active' && event.my_vote_nominee_id" class="event-voting__voted-banner">
            <div class="event-voting__voted-text">
              <Check :size="18" class="event-voting__check-icon" aria-hidden="true" />
              <span>{{ t("events.voting.votedNotice", { title: votedNomineeTitle }) }}</span>
            </div>
            <AppButton
              variant="ghost"
              size="sm"
              :loading="cancellingVote"
              :disabled="voteBusy"
              @click="handleCancelVote"
            >
              {{ t("events.voting.cancelVote") }}
            </AppButton>
          </div>

          <!-- Nominee Cards Grid -->
          <div class="event-voting__grid">
            <div
              v-for="nominee in event.nominees"
              :key="nominee.id"
              class="event-nominee-card"
              :class="{
                'event-nominee-card--voted': event.my_vote_nominee_id === nominee.id,
                'event-nominee-card--winner': nominee.is_winner,
              }"
            >
              <div class="event-nominee-card__media">
                <RouterLink
                  :to="`/${nominee.media_type === 'tv' ? 'tv' : 'movies'}/${nominee.media_id}`"
                  class="event-nominee-card__poster-link"
                >
                  <img
                    v-if="nominee.poster_path"
                    :src="getTmdbImageUrl(nominee.poster_path, NOMINEE_POSTER_SIZE) ?? undefined"
                    :alt="nominee.media_title"
                    class="event-nominee-card__poster"
                    loading="lazy"
                  />
                  <div v-else class="event-nominee-card__poster-fallback">
                    {{ nominee.media_title }}
                  </div>
                </RouterLink>

                <span class="event-nominee-card__position">
                  #{{ nominee.position }}
                </span>

                <span v-if="nominee.is_winner" class="event-nominee-card__winner-badge">
                  <Trophy :size="14" aria-hidden="true" />
                  {{ isTieWinner ? t("events.voting.coWinner") : t("events.voting.winner") }}
                </span>
              </div>

              <div class="event-nominee-card__info">
                <h3 class="event-nominee-card__title">
                  <RouterLink :to="`/${nominee.media_type === 'tv' ? 'tv' : 'movies'}/${nominee.media_id}`">
                    {{ nominee.media_title }}
                  </RouterLink>
                </h3>
                <p v-if="nominee.overview" class="event-nominee-card__overview">
                  {{ nominee.overview }}
                </p>

                <!-- Vote results after poll closes -->
                <div v-if="event.state === 'past'" class="event-nominee-card__results">
                  <div class="event-nominee-card__bar">
                    <div
                      class="event-nominee-card__bar-fill"
                      :style="{ width: `${nominee.percentage ?? 0}%` }"
                    />
                  </div>
                  <div class="event-nominee-card__vote-label">
                    {{ formatVoteResult(nominee.vote_count, nominee.percentage) }}
                  </div>
                </div>

                <!-- Vote button during active period -->
                <div v-else-if="event.state === 'active'" class="event-nominee-card__actions">
                  <AppButton
                    v-if="event.my_vote_nominee_id === nominee.id"
                    variant="primary"
                    size="md"
                    block
                    disabled
                    class="event-nominee-card__voted-btn"
                  >
                    <Check :size="16" aria-hidden="true" />
                    {{ t("events.voting.voted") }}
                  </AppButton>
                  <AppButton
                    v-else
                    variant="outline"
                    size="md"
                    block
                    :loading="votingId === nominee.id"
                    :disabled="voteBusy"
                    @click="handleVote(nominee.id)"
                  >
                    {{ event.my_vote_nominee_id ? t("events.voting.changeVote") : t("events.voting.vote") }}
                  </AppButton>
                </div>
              </div>
            </div>
          </div>
        </section>

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
import { computed, onMounted, ref, watch } from "vue"
import { useRoute } from "vue-router"
import { isAxiosError } from "axios"
import { useI18n } from "vue-i18n"
import { Calendar, Check, ChevronLeft, ExternalLink, Lock, Trophy } from "lucide-vue-next"
import AppButton from "@/components/ui/AppButton.vue"
import AppErrorState from "@/components/ui/AppErrorState.vue"
import AppSpinner from "@/components/ui/AppSpinner.vue"
import { useAuthGate } from "@/composables/useAuthGate"
import { useToast } from "@/composables/useToast"
import { eventApi } from "@/api/endpoints/event"
import type { AppEvent } from "@/types/event"
import {
  HTTP_NOT_FOUND,
  NOMINEE_POSTER_SIZE,
  eventBody,
  eventTitle,
  isInternalLink,
  voteErrorKey,
} from "@/utils/events"
import { getTmdbImageUrl } from "@/utils/image"

/** Server percentages are 0-100; Intl's percent style expects a fraction. */
const PERCENT_DIVISOR = 100

const { t, d, n, locale } = useI18n()
const route = useRoute()
const toast = useToast()
const { requireAuth } = useAuthGate()

const event = ref<AppEvent | null>(null)
const loading = ref(true)
const error = ref(false)
const notFound = ref(false)

const votingId = ref<number | null>(null)
const cancellingVote = ref(false)
// One vote request at a time: any in-flight vote or cancel locks every vote control.
const voteBusy = computed(() => votingId.value !== null || cancellingVote.value)

const percentFormatter = computed(() => new Intl.NumberFormat(locale.value, { style: "percent" }))

function formatTotalVotes(count: number): string {
  return t("events.voting.totalVotes", { votes: n(count) }, count)
}

function formatVoteResult(count: number | null | undefined, percentage: number | null | undefined): string {
  if (!count) return t("events.voting.noVotes", { votes: n(0) }, 0)
  const percent = percentFormatter.value.format((percentage ?? 0) / PERCENT_DIVISOR)
  return t("events.voting.votesCount", { votes: n(count), percent }, count)
}

const votedNomineeTitle = computed(() => {
  if (!event.value?.nominees || !event.value?.my_vote_nominee_id) return ""
  const match = event.value.nominees.find((n) => n.id === event.value?.my_vote_nominee_id)
  return match?.media_title ?? ""
})

const isTieWinner = computed(() => {
  const winners = event.value?.nominees?.filter((n) => n.is_winner) ?? []
  return winners.length > 1
})

async function handleVote(nomineeId: number) {
  if (!event.value || voteBusy.value) return
  if (!requireAuth("generic")) return

  votingId.value = nomineeId
  try {
    await eventApi.vote(event.value.id, nomineeId)
    event.value.my_vote_nominee_id = nomineeId
    toast.success(t("events.voting.voteSuccess"))
  } catch (err) {
    toast.error(t(voteErrorKey(err)))
    if (isAxiosError(err) && err.response?.data?.code === "VOTING_CLOSED") {
      void fetchEvent()
    }
  } finally {
    votingId.value = null
  }
}

async function handleCancelVote() {
  if (!event.value || voteBusy.value) return
  if (!requireAuth("generic")) return

  cancellingVote.value = true
  try {
    await eventApi.cancelVote(event.value.id)
    event.value.my_vote_nominee_id = null
    toast.success(t("events.voting.cancelSuccess"))
  } catch (err) {
    toast.error(t(voteErrorKey(err)))
    if (isAxiosError(err) && err.response?.data?.code === "VOTING_CLOSED") {
      void fetchEvent()
    }
  } finally {
    cancellingVote.value = false
  }
}

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

/* Voting Section */
.event-voting {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  padding: 1.5rem;
  background-color: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-card);
}

.event-voting__header {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  justify-content: space-between;
  gap: 1rem;
}

.event-voting__title {
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--color-text-primary);
  margin-bottom: 0.25rem;
}

.event-voting__subtitle {
  font-size: 0.875rem;
  color: var(--color-text-muted);
}

.event-voting__total {
  padding: 0.25rem 0.75rem;
  font-size: 0.8125rem;
  font-weight: 600;
  background-color: var(--color-surface-3);
  color: var(--color-text-secondary);
  border-radius: var(--radius-full);
}

.event-voting__notice {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem 1rem;
  font-size: 0.875rem;
  border-radius: var(--radius-card);
  line-height: 1.4;
}

.event-voting__notice--info {
  background-color: color-mix(in srgb, var(--color-info) 12%, transparent);
  color: var(--color-info);
  border: 1px solid color-mix(in srgb, var(--color-info) 30%, transparent);
}

.event-voting__notice--past {
  background-color: color-mix(in srgb, var(--color-rating) 12%, transparent);
  color: var(--color-rating);
  border: 1px solid color-mix(in srgb, var(--color-rating) 30%, transparent);
}

.event-voting__notice--upcoming {
  background-color: var(--color-surface-3);
  color: var(--color-text-muted);
}

.event-voting__voted-banner {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  padding: 0.75rem 1rem;
  background-color: color-mix(in srgb, var(--color-brand) 12%, transparent);
  border: 1px solid color-mix(in srgb, var(--color-brand) 35%, transparent);
  border-radius: var(--radius-card);
}

.event-voting__voted-text {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--color-text-primary);
}

.event-voting__check-icon {
  color: var(--color-brand);
}

.event-voting__grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 1.25rem;
}

@media (min-width: 640px) {
  .event-voting__grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (min-width: 1024px) {
  .event-voting__grid {
    grid-template-columns: repeat(5, 1fr);
  }
}

.event-nominee-card {
  display: flex;
  flex-direction: column;
  background-color: var(--color-surface-2);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-card);
  overflow: hidden;
  transition: transform 0.15s ease, border-color 0.15s ease;
}

.event-nominee-card:hover {
  border-color: var(--color-border-strong);
}

.event-nominee-card--voted {
  border-color: var(--color-brand);
  box-shadow: 0 0 0 1px var(--color-brand);
}

.event-nominee-card--winner {
  border-color: var(--color-rating);
  box-shadow: 0 0 12px color-mix(in srgb, var(--color-rating) 30%, transparent);
}

.event-nominee-card__media {
  position: relative;
  width: 100%;
  aspect-ratio: 2 / 3;
  background-color: var(--color-surface-3);
  overflow: hidden;
}

.event-nominee-card__poster-link {
  display: block;
  width: 100%;
  height: 100%;
}

.event-nominee-card__poster {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.event-nominee-card__poster-fallback {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1rem;
  text-align: center;
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--color-text-muted);
}

.event-nominee-card__position {
  position: absolute;
  top: 0.5rem;
  left: 0.5rem;
  padding: 0.125rem 0.5rem;
  font-size: 0.75rem;
  font-weight: 700;
  background-color: color-mix(in srgb, var(--color-surface) 80%, transparent);
  color: var(--color-text-primary);
  border-radius: var(--radius-sm);
  backdrop-filter: blur(4px);
}

.event-nominee-card__winner-badge {
  position: absolute;
  top: 0.5rem;
  right: 0.5rem;
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  padding: 0.25rem 0.5rem;
  font-size: 0.6875rem;
  font-weight: 700;
  background-color: var(--color-rating);
  color: var(--color-surface);
  border-radius: var(--radius-sm);
  text-transform: uppercase;
  letter-spacing: 0.025em;
}

.event-nominee-card__info {
  display: flex;
  flex-direction: column;
  flex: 1;
  padding: 1rem;
  gap: 0.5rem;
}

.event-nominee-card__title {
  font-size: 0.9375rem;
  font-weight: 600;
  color: var(--color-text-primary);
  line-height: 1.35;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.event-nominee-card__title a {
  color: inherit;
  text-decoration: none;
}

.event-nominee-card__title a:hover {
  color: var(--color-brand);
}

.event-nominee-card__overview {
  font-size: 0.75rem;
  color: var(--color-text-muted);
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  margin-bottom: auto;
}

.event-nominee-card__results {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
  margin-top: 0.5rem;
}

.event-nominee-card__bar {
  width: 100%;
  height: 6px;
  background-color: var(--color-surface-3);
  border-radius: var(--radius-full);
  overflow: hidden;
}

.event-nominee-card__bar-fill {
  height: 100%;
  background-color: var(--color-brand);
  border-radius: var(--radius-full);
  transition: width 0.3s ease;
}

.event-nominee-card--winner .event-nominee-card__bar-fill {
  background-color: var(--color-rating);
}

.event-nominee-card__vote-label {
  font-size: 0.75rem;
  color: var(--color-text-secondary);
  font-weight: 500;
}

.event-nominee-card__actions {
  margin-top: 0.5rem;
}

.event-nominee-card__voted-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.375rem;
}
</style>
