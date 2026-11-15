<template>
  <AppModal v-model="open" size="md" :title="t('admin.events.generateMonthlyModalTitle')">
    <form class="event-form" @submit.prevent="handleGenerate">
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
        <label class="form-label" for="gen-period">{{ t("admin.events.period") }}</label>
        <input
          id="gen-period"
          v-model="period"
          type="text"
          class="form-input"
          :class="{ 'is-invalid': periodInvalid }"
          :placeholder="t('admin.events.periodPlaceholder')"
          :aria-invalid="periodInvalid || undefined"
          :aria-describedby="periodInvalid ? 'gen-period-error' : undefined"
          required
          @input="periodInvalid = false"
        />
        <span v-if="periodInvalid" id="gen-period-error" class="form-error">
          {{ t("admin.events.errors.periodInvalid") }}
        </span>
      </div>

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
import { ref, watch } from "vue"
import { useI18n } from "vue-i18n"
import AppButton from "@/components/ui/AppButton.vue"
import AppModal from "@/components/ui/AppModal.vue"
import { useToast } from "@/composables/useToast"
import { adminEventApi } from "@/api/endpoints/event"
import type { SuggestedNominee } from "@/types/event"
import { eventErrorKey, isValidMonthlyPeriod, nextMonthlyPeriod } from "@/utils/events"
import "./adminEventForm.css"

type GenerateMediaType = "movie" | "tv"

const open = defineModel<boolean>({ required: true })
const emit = defineEmits<{ generated: [] }>()

const { t } = useI18n()
const toast = useToast()

const mediaType = ref<GenerateMediaType>("movie")
const period = ref("")
const periodInvalid = ref(false)
const suggestions = ref<SuggestedNominee[]>([])
const loadingSuggestions = ref(false)
const generating = ref(false)

// Each opening starts from the next round's defaults.
watch(open, (isOpen) => {
  if (!isOpen) return
  mediaType.value = "movie"
  period.value = nextMonthlyPeriod()
  periodInvalid.value = false
  suggestions.value = []
})

/** Checks the period against the server format; sets the field error when it fails. */
function checkPeriod(): boolean {
  periodInvalid.value = !isValidMonthlyPeriod(period.value)
  return !periodInvalid.value
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

async function handleGenerate() {
  if (!checkPeriod()) return
  generating.value = true
  try {
    await adminEventApi.generateMonthly(mediaType.value, period.value)
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
