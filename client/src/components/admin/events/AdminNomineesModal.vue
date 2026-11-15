<template>
  <AppModal
    v-model="open"
    size="md"
    :title="t('admin.events.nomineesModalTitle', { title: event ? eventTitle(event, t, locale) : '' })"
  >
    <div class="event-form">
      <div class="generate-preview-action">
        <AppButton
          variant="outline"
          size="sm"
          :loading="loadingSuggestions"
          @click.prevent="loadSuggestions"
        >
          {{ t("admin.events.loadSuggestedNominees") }}
        </AppButton>
      </div>

      <div v-if="nominees.length" class="suggested-list">
        <div v-for="(nom, idx) in nominees" :key="nom.media_id" class="suggested-item">
          <span class="suggested-item__idx">#{{ idx + 1 }}</span>
          <div class="suggested-item__info">
            <span class="suggested-item__title">{{ nom.media_title }}</span>
          </div>
          <div class="suggested-item__actions">
            <AppButton
              variant="ghost"
              size="sm"
              icon-only
              :aria-label="t('admin.events.moveUp')"
              :disabled="idx === 0"
              @click="moveNominee(idx, -1)"
            >
              <ArrowUp :size="14" />
            </AppButton>
            <AppButton
              variant="ghost"
              size="sm"
              icon-only
              :aria-label="t('admin.events.moveDown')"
              :disabled="idx === nominees.length - 1"
              @click="moveNominee(idx, 1)"
            >
              <ArrowDown :size="14" />
            </AppButton>
            <AppButton
              variant="ghost"
              size="sm"
              icon-only
              :aria-label="t('admin.events.removeNominee')"
              @click="removeNomineeAt(idx)"
            >
              <Trash2 :size="14" />
            </AppButton>
          </div>
        </div>
      </div>
      <p v-if="nominees.length < minNominees" class="form-hint">
        {{ t("admin.events.saveNomineesEmptyHint", { min: minNominees }, minNominees) }}
      </p>
      <div class="form-actions">
        <AppButton variant="ghost" @click="open = false">
          {{ t("admin.events.form.cancel") }}
        </AppButton>
        <AppButton
          variant="primary"
          :loading="saving"
          :disabled="nominees.length < minNominees"
          @click="handleSave"
        >
          {{ t("admin.events.saveNominees") }}
        </AppButton>
      </div>
    </div>
  </AppModal>
</template>

<script setup lang="ts">
import { ref, watch } from "vue"
import { useI18n } from "vue-i18n"
import { ArrowDown, ArrowUp, Trash2 } from "lucide-vue-next"
import AppButton from "@/components/ui/AppButton.vue"
import AppModal from "@/components/ui/AppModal.vue"
import { useToast } from "@/composables/useToast"
import { adminEventApi } from "@/api/endpoints/event"
import type { AdminEvent, NomineeInput } from "@/types/event"
import {
  EVENT_MIN_NOMINEES_TO_PUBLISH,
  eventErrorKey,
  eventTitle,
  nextMonthlyPeriod,
  publishErrorKey,
} from "@/utils/events"
import "./adminEventForm.css"

const props = defineProps<{
  /** Round whose nominees are being edited; null while no round is selected. */
  event: AdminEvent | null
}>()

const open = defineModel<boolean>({ required: true })
const emit = defineEmits<{ saved: [] }>()

const { t, locale } = useI18n()
const toast = useToast()

const nominees = ref<NomineeInput[]>([])
const loadingSuggestions = ref(false)
const saving = ref(false)

// One rule for every status, so a draft is always ready to publish.
const minNominees = EVENT_MIN_NOMINEES_TO_PUBLISH

// Each opening starts from the round's saved nominees.
watch(open, (isOpen) => {
  if (isOpen) nominees.value = nomineesFromEvent(props.event)
})

function nomineesFromEvent(ev: AdminEvent | null): NomineeInput[] {
  return (ev?.nominees ?? []).map((n, idx) => ({
    id: n.id,
    media_id: n.media_id,
    media_type: n.media_type,
    position: idx + 1,
    media_title: n.media_title,
    poster_path: n.poster_path ?? "",
    backdrop_path: n.backdrop_path ?? "",
    overview: n.overview ?? "",
  }))
}

function moveNominee(idx: number, delta: -1 | 1) {
  const target = idx + delta
  if (target < 0 || target >= nominees.value.length) return
  const list = [...nominees.value]
  const [item] = list.splice(idx, 1)
  if (!item) return
  list.splice(target, 0, item)
  nominees.value = list.map((n, i) => ({ ...n, position: i + 1 }))
}

function removeNomineeAt(idx: number) {
  nominees.value = nominees.value
    .filter((_, i) => i !== idx)
    .map((n, i) => ({ ...n, position: i + 1 }))
}

async function loadSuggestions() {
  if (!props.event) return
  loadingSuggestions.value = true
  try {
    const mediaType = props.event.media_type ?? "movie"
    const period = props.event.period ?? nextMonthlyPeriod()
    const res = await adminEventApi.suggestNominees(mediaType, period)
    nominees.value = res.data.nominees.map((s, idx) => ({
      media_id: s.media_id,
      media_type: s.media_type,
      position: idx + 1,
      media_title: s.media_title,
      poster_path: s.poster_path,
      backdrop_path: s.backdrop_path,
      overview: s.overview,
    }))
  } catch (err) {
    toast.error(t(eventErrorKey(err)))
  } finally {
    loadingSuggestions.value = false
  }
}

async function handleSave() {
  if (!props.event || nominees.value.length < minNominees) return
  saving.value = true
  try {
    await adminEventApi.setNominees(props.event.id, nominees.value)
    toast.success(t("admin.events.nomineesSaved"))
    open.value = false
    emit("saved")
  } catch (err) {
    toast.error(t(publishErrorKey(err), { min: EVENT_MIN_NOMINEES_TO_PUBLISH }))
  } finally {
    saving.value = false
  }
}
</script>
