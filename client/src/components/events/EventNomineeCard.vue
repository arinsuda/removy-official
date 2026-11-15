<template>
  <div
    class="event-nominee-card"
    :class="{
      'event-nominee-card--voted': isVoted,
      'event-nominee-card--winner': nominee.is_winner,
    }"
  >
    <div class="event-nominee-card__media">
      <RouterLink :to="mediaPath" class="event-nominee-card__poster-link">
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
        {{ coWinner ? t("events.voting.coWinner") : t("events.voting.winner") }}
      </span>
    </div>

    <div class="event-nominee-card__info">
      <h3 class="event-nominee-card__title">
        <RouterLink :to="mediaPath">
          {{ nominee.media_title }}
        </RouterLink>
      </h3>
      <p v-if="nominee.overview" class="event-nominee-card__overview">
        {{ nominee.overview }}
      </p>

      <!-- Vote results after poll closes -->
      <div v-if="state === 'past'" class="event-nominee-card__results">
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
      <div v-else-if="state === 'active'" class="event-nominee-card__actions">
        <AppButton
          v-if="isVoted"
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
          :loading="voting"
          :disabled="disabled"
          @click="emit('vote', nominee.id)"
        >
          {{ hasVote ? t("events.voting.changeVote") : t("events.voting.vote") }}
        </AppButton>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue"
import { useI18n } from "vue-i18n"
import { Check, Trophy } from "lucide-vue-next"
import AppButton from "@/components/ui/AppButton.vue"
import type { EventState, Nominee } from "@/types/event"
import { NOMINEE_POSTER_SIZE } from "@/utils/events"
import { getTmdbImageUrl } from "@/utils/image"

/** Server percentages are 0-100; Intl's percent style expects a fraction. */
const PERCENT_DIVISOR = 100

const props = defineProps<{
  nominee: Nominee
  state: EventState
  /** This nominee is the viewer's current vote. */
  isVoted: boolean
  /** The viewer has voted for some nominee, so the button reads "change vote". */
  hasVote: boolean
  /** This nominee's vote request is in flight. */
  voting: boolean
  /** Another vote request is in flight, so vote buttons are locked. */
  disabled: boolean
  /** More than one nominee shares the win. */
  coWinner: boolean
}>()

const emit = defineEmits<{ vote: [nomineeId: number] }>()

const { t, n, locale } = useI18n()

const percentFormatter = computed(() => new Intl.NumberFormat(locale.value, { style: "percent" }))

const mediaPath = computed(
  () => `/${props.nominee.media_type === "tv" ? "tv" : "movies"}/${props.nominee.media_id}`,
)

function formatVoteResult(count: number | null | undefined, percentage: number | null | undefined): string {
  if (!count) return t("events.voting.noVotes", { votes: n(0) }, 0)
  const percent = percentFormatter.value.format((percentage ?? 0) / PERCENT_DIVISOR)
  return t("events.voting.votesCount", { votes: n(count), percent }, count)
}
</script>

<style scoped>
.event-nominee-card {
  display: flex;
  flex-direction: column;
  background-color: var(--color-surface-2);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-card);
  overflow: hidden;
  transition: transform var(--motion-fast) var(--ease-out-soft), border-color var(--motion-fast) var(--ease-out-soft);
}

.event-nominee-card:hover {
  border-color: var(--color-border-strong);
}

.event-nominee-card--voted {
  border-color: var(--color-brand);
  box-shadow: var(--shadow-ring-brand);
}

.event-nominee-card--winner {
  border-color: var(--color-rating);
  box-shadow: var(--shadow-glow-rating);
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
  transition: width var(--motion-base) var(--ease-out-soft);
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
