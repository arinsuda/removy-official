<template>
  <AppModal v-model="open" size="md" :title="t('admin.events.generateRoundModalTitle')">
    <form class="event-form" @submit.prevent="handleGenerate">
      <fieldset class="generate-fieldset">
        <legend class="form-label">{{ t("admin.events.roundKind") }}</legend>
        <div class="generate-radio-group">
          <label v-for="option in KIND_OPTIONS" :key="option.value" class="generate-radio">
            <input v-model="kind" type="radio" name="generate-round-kind" :value="option.value" />
            <span>{{ t(option.labelKey) }}</span>
          </label>
        </div>
      </fieldset>

      <fieldset class="generate-fieldset">
        <legend class="form-label">{{ t("admin.events.mediaType") }}</legend>
        <div class="generate-radio-group">
          <label class="generate-radio">
            <input v-model="mediaType" type="radio" name="generate-media-type" value="movie" />
            <span>{{ t("admin.events.movie") }}</span>
          </label>
          <label class="generate-radio">
            <input v-model="mediaType" type="radio" name="generate-media-type" value="tv" />
            <span>{{ t("admin.events.tv") }}</span>
          </label>
        </div>
      </fieldset>

      <div class="form-field">
        <label class="form-label" for="gen-period">{{ t(periodLabelKey) }}</label>
        <input
          id="gen-period"
          v-model="period"
          type="text"
          class="form-input"
          :class="{ 'is-invalid': periodInvalid }"
          :placeholder="t(periodPlaceholderKey)"
          :aria-invalid="periodInvalid || undefined"
          :aria-describedby="periodInvalid ? 'gen-period-error' : undefined"
          required
          @input="periodInvalid = false"
        />
        <span v-if="periodInvalid" id="gen-period-error" class="form-error">
          {{ t(periodErrorKey) }}
        </span>
      </div>

      <template v-if="isMonthly">
        <div class="generate-preview-action">
          <AppButton
            variant="outline"
            size="sm"
            :loading="loadingSuggestions"
            @click.prevent="loadSuggestions"
          >
            {{ t("admin.events.previewNominees") }}
          </AppButton>
        </div>

        <div v-if="suggestions.length" class="suggested-list">
          <div v-for="(nom, idx) in suggestions" :key="nom.media_id" class="suggested-item">
            <span class="suggested-item__idx">#{{ idx + 1 }}</span>
            <div class="suggested-item__info">
              <span class="suggested-item__title">{{ nom.media_title }}</span>
              <span class="suggested-item__source">{{ formatSource(nom.source) }}</span>
            </div>
          </div>
        </div>
      </template>
      <p v-else class="form-hint">{{ t("admin.events.yearlyNomineesHint") }}</p>

      <div class="form-actions">
        <AppButton variant="ghost" @click="open = false">
          {{ t("admin.events.form.cancel") }}
        </AppButton>
        <AppButton variant="primary" type="submit" :loading="generating">
          {{ t("admin.events.generateRound") }}
        </AppButton>
      </div>
    </form>
  </AppModal>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue"
import { useI18n } from "vue-i18n"
import AppButton from "@/components/ui/AppButton.vue"
import AppModal from "@/components/ui/AppModal.vue"
import { useToast } from "@/composables/useToast"
import { adminEventApi } from "@/api/endpoints/event"
import type { SuggestedNominee } from "@/types/event"
import {
  eventErrorKey,
  isValidMonthlyPeriod,
  isValidYearlyPeriod,
  nextMonthlyPeriod,
  nextYearlyPeriod,
} from "@/utils/events"
import "./adminEventForm.css"

type GenerateMediaType = "movie" | "tv"
type RoundKind = "monthly" | "yearly"

/** Round kinds offered in the modal; labels reuse the event type names. */
const KIND_OPTIONS: { value: RoundKind; labelKey: string }[] = [
  { value: "monthly", labelKey: "admin.events.types.vote_monthly" },
  { value: "yearly", labelKey: "admin.events.types.vote_yearly" },
]

/** Next round's default period for each kind, computed when the modal opens or the kind changes. */
const DEFAULT_PERIOD: Record<RoundKind, () => string> = {
  monthly: () => nextMonthlyPeriod(),
  yearly: () => nextYearlyPeriod(),
}

const open = defineModel<boolean>({ required: true })
const emit = defineEmits<{ generated: [] }>()

const { t } = useI18n()
const toast = useToast()

const kind = ref<RoundKind>("monthly")
const mediaType = ref<GenerateMediaType>("movie")
const period = ref("")
const periodInvalid = ref(false)
const suggestions = ref<SuggestedNominee[]>([])
const loadingSuggestions = ref(false)
const generating = ref(false)

const isMonthly = computed(() => kind.value === "monthly")
const periodLabelKey = computed(() => (isMonthly.value ? "admin.events.period" : "admin.events.yearlyPeriod"))
const periodPlaceholderKey = computed(() =>
  isMonthly.value ? "admin.events.periodPlaceholder" : "admin.events.yearlyPeriodPlaceholder",
)
const periodErrorKey = computed(() =>
  isMonthly.value ? "admin.events.errors.periodInvalid" : "admin.events.errors.yearlyPeriodInvalid",
)

/** Resets the period to the next round of the current kind and clears the field and preview state. */
function resetForKind() {
  period.value = DEFAULT_PERIOD[kind.value]()
  periodInvalid.value = false
  suggestions.value = []
}

// Each opening starts from the next monthly round's defaults.
watch(open, (isOpen) => {
  if (!isOpen) return
  kind.value = "monthly"
  mediaType.value = "movie"
  resetForKind()
})

// Switching kind swaps the period format and the default; a preview only applies to monthly rounds.
watch(kind, resetForKind)

/** Checks the period against the format of the current kind; sets the field error when it fails. */
function checkPeriod(): boolean {
  const valid = isMonthly.value ? isValidMonthlyPeriod(period.value) : isValidYearlyPeriod(period.value)
  periodInvalid.value = !valid
  return valid
}

function formatSource(source: string): string {
  switch (source) {
    case "community":
      return t("admin.events.sourceCommunity")
    case "release":
      return t("admin.events.sourceRelease")
    case "popular":
      return t("admin.events.sourcePopular")
    default:
      return t("admin.events.sourceFallback")
  }
}

async function loadSuggestions() {
  if (!checkPeriod()) return
  loadingSuggestions.value = true
  try {
    const res = await adminEventApi.suggestNominees(mediaType.value, period.value)
    suggestions.value = res.data.nominees
  } catch (err) {
    toast.error(t(eventErrorKey(err)))
  } finally {
    loadingSuggestions.value = false
  }
}

async function requestRound() {
  if (isMonthly.value) {
    await adminEventApi.generateMonthly(mediaType.value, period.value)
  } else {
    await adminEventApi.generateYearlyRound({ media_type: mediaType.value, period: period.value })
  }
}

async function handleGenerate() {
  if (!checkPeriod()) return
  generating.value = true
  try {
    await requestRound()
    toast.success(t("admin.events.generateSuccess"))
    open.value = false
    emit("generated")
  } catch (err) {
    toast.error(t(eventErrorKey(err)))
  } finally {
    generating.value = false
  }
}
</script>
