<template>
  <section class="event-voting">
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
        :disabled="voteBusy"
        @click="openCancelConfirm"
      >
        {{ t("events.voting.cancelVote") }}
      </AppButton>
    </div>

    <div class="event-voting__grid">
      <EventNomineeCard
        v-for="nominee in nominees"
        :key="nominee.id"
        :nominee="nominee"
        :state="event.state"
        :is-voted="event.my_vote_nominee_id === nominee.id"
        :has-vote="Boolean(event.my_vote_nominee_id)"
        :voting="votingId === nominee.id"
        :disabled="voteBusy"
        :co-winner="isTieWinner"
        @vote="handleVote"
      />
    </div>

    <AppModal
      v-model="cancelConfirmOpen"
      role="alertdialog"
      size="sm"
      :show-close="false"
      :dismiss-disabled="cancellingVote"
      :title="t('events.voting.cancelVote')"
      :aria-label="t('events.voting.cancelVote')"
      :aria-describedby="cancelDescId"
    >
      <p :id="cancelDescId" class="event-voting__confirm-text">
        {{ t("events.voting.cancelConfirm") }}
      </p>

      <template #footer>
        <div class="event-voting__confirm-actions">
          <AppButton variant="outline" :disabled="cancellingVote" @click="cancelConfirmOpen = false">
            {{ t("common.cancel") }}
          </AppButton>
          <AppButton variant="primary" :loading="cancellingVote" @click="handleCancelVote">
            {{ t("events.voting.cancelVote") }}
          </AppButton>
        </div>
      </template>
    </AppModal>
  </section>
</template>

<script setup lang="ts">
import { computed, ref, useId } from "vue"
import { isAxiosError } from "axios"
import { useI18n } from "vue-i18n"
import { Calendar, Check, Lock, Trophy } from "lucide-vue-next"
import AppButton from "@/components/ui/AppButton.vue"
import AppModal from "@/components/ui/AppModal.vue"
import EventNomineeCard from "@/components/events/EventNomineeCard.vue"
import { useAuthGate } from "@/composables/useAuthGate"
import { useToast } from "@/composables/useToast"
import { eventApi } from "@/api/endpoints/event"
import type { AppEvent } from "@/types/event"
import { formatEventDay, voteErrorKey } from "@/utils/events"

const props = defineProps<{ event: AppEvent }>()

const emit = defineEmits<{
  /** The vote was rejected because voting closed; the parent should refetch the event. */
  refresh: []
  /** The viewer's vote changed; null after a cancel. */
  "vote-change": [nomineeId: number | null]
}>()

const { t, d, n } = useI18n()
const toast = useToast()
const { requireAuth } = useAuthGate()
const cancelDescId = useId()

const nominees = computed(() => props.event.nominees ?? [])

const votingId = ref<number | null>(null)
const cancellingVote = ref(false)
const cancelConfirmOpen = ref(false)
// One vote request at a time: any in-flight vote or cancel locks every vote control.
const voteBusy = computed(() => votingId.value !== null || cancellingVote.value)

const votedNomineeTitle = computed(() => {
  const votedId = props.event.my_vote_nominee_id
  if (!votedId) return ""
  return nominees.value.find((nominee) => nominee.id === votedId)?.media_title ?? ""
})

const isTieWinner = computed(() => nominees.value.filter((nominee) => nominee.is_winner).length > 1)

function formatTotalVotes(count: number): string {
  return t("events.voting.totalVotes", { votes: n(count) }, count)
}

function formatDay(iso: string): string {
  return formatEventDay(iso, (date) => d(date, "medium"))
}

async function handleVote(nomineeId: number) {
  if (voteBusy.value) return
  if (!requireAuth("generic")) return

  votingId.value = nomineeId
  try {
    await eventApi.vote(props.event.id, nomineeId)
    emit("vote-change", nomineeId)
    toast.success(t("events.voting.voteSuccess"))
  } catch (err) {
    toast.error(t(voteErrorKey(err)))
    if (isAxiosError(err) && err.response?.data?.code === "VOTING_CLOSED") emit("refresh")
  } finally {
    votingId.value = null
  }
}

function openCancelConfirm() {
  if (voteBusy.value || !requireAuth("generic")) return
  cancelConfirmOpen.value = true
}

async function handleCancelVote() {
  if (cancellingVote.value) return

  cancellingVote.value = true
  try {
    await eventApi.cancelVote(props.event.id)
    emit("vote-change", null)
    toast.success(t("events.voting.cancelSuccess"))
  } catch (err) {
    toast.error(t(voteErrorKey(err)))
    if (isAxiosError(err) && err.response?.data?.code === "VOTING_CLOSED") emit("refresh")
  } finally {
    cancellingVote.value = false
    cancelConfirmOpen.value = false
  }
}
</script>

<style scoped>
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

.event-voting__confirm-text {
  font-size: 0.875rem;
  line-height: 1.5;
  color: var(--color-text-secondary);
}

.event-voting__confirm-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
}
</style>
